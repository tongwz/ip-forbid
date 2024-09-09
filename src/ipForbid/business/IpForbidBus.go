package business

import (
	"bufio"
	"fmt"
	"io"
	"ipForbid/consts"
	"ipForbid/http/rsp"
	"ipForbid/pkg/logrus"
	"ipForbid/pkg/types"
	"ipForbid/pkg/viper"
	"ipForbid/utils"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

var IpForbidBusObj *IpForbidBus
var IpForbidBusOnce sync.Once

type IpForbidBus struct {
	mtx           *sync.Mutex
	CloudflareIps []string                 // 它的ip是带子网掩码的
	QuicIps       []string                 // 它是纯ip
	forbidMaps    map[string]map[int64]int // 第一个key IP 第二个key 分钟的整型 value=一分钟触发次数
}

// 单独实例化一次
func GetIpForbidBusObj() *IpForbidBus {
	IpForbidBusOnce.Do(func() {
		IpForbidBusObj = &IpForbidBus{}
		IpForbidBusObj.mtx = &sync.Mutex{}
		IpForbidBusObj.CloudflareIps = CloudFlareListGet()
		IpForbidBusObj.QuicIps = QuicListGet()
		IpForbidBusObj.forbidMaps = make(map[string]map[int64]int)

		// 加一个定时任务，定时刷新最新cdn ips
		go IpForbidBusObj.refreshCDNIps()
	})
	return IpForbidBusObj
}

/**
 * @note: 获取所有服务配置
 * @auth: tongWz
 * @date: 2024/8/29 10:13
**/
func GetAllService() map[string]map[string]string {
	allServiceSlice := viper.Viper.GetStringSlice("app.deny_service")
	if len(allServiceSlice) == 0 {
		logrus.LogrusObj.Infof("没有查询到需要进行ip检测的服务")
		return nil
	}
	serviceMap := make(map[string]map[string]string)
	for _, service := range allServiceSlice {
		serviceMap[service] = viper.Viper.GetStringMapString(service)
	}
	return serviceMap

}

// 获取单个service的map
func GetService(key string) map[string]string {
	return viper.Viper.GetStringMapString(key)
}

// 2024年9月9日17:47:23 加入ips更新心跳
func (i *IpForbidBus) refreshCDNIps() {
	timer := time.NewTicker(20 * time.Minute)
	for {
		select {
		case <-timer.C:
			cloudFlareIps := CloudFlareListGet()
			if len(cloudFlareIps) > 0 {
				IpForbidBusObj.CloudflareIps = cloudFlareIps
			}
			quicIps := QuicListGet()
			if len(quicIps) > 0 {
				IpForbidBusObj.QuicIps = quicIps
			}
		}
	}
}

// nginx_logs_dir = "xxxx"
// nginx_deny_conf_dir = "yyyy"
// start_line_index = 0

/**
 * @note: 运行具体的服务查询封禁IP
 * @auth: tongWz
 * @date: 2024/8/30 14:08
**/
func (i *IpForbidBus) ServiceErrorWatch(serviceName string, serviceMap map[string]string, wg *sync.WaitGroup) {
	defer wg.Done()
	if serviceMap == nil {
		return
	}
	denyPath, ok := serviceMap["nginx_deny_conf_dir"]
	if !ok || denyPath == "" {
		fmt.Printf("denyPath异常:%s \n", denyPath)
		return
	}
	nginxLogsPath, ok := serviceMap["nginx_logs_dir"]
	if !ok || nginxLogsPath == "" {
		fmt.Printf("nginxLogsPath异常:%s \n", nginxLogsPath)
		return
	}
	nginxStartLineIndex, ok := serviceMap["start_line_index"]
	if !ok || nginxStartLineIndex == "" {
		fmt.Printf("nginxStartLineIndex异常:%s \n", nginxStartLineIndex)
		return
	}

	lineIndex, err := strconv.Atoi(nginxStartLineIndex)
	if err != nil {
		logrus.LogrusObj.Errorf("转化行数异常：%s", err)
		return
	}

	// 读取已封禁ip列表
	// 一般长这样 deny 127.0.0.1;#2024-08-29 17:06:51
	denyIps := BlockNginxFile(denyPath)
	// fmt.Printf("封禁的ip:%#v \n", denyIps)

	// 读取当前文件index后100行 或者一行一行的读取
	triggerMap, readedIndex := ReadNginxLog(lineIndex, nginxLogsPath, denyIps)
	fmt.Printf("readedIndex = %d\n", readedIndex)

	needAddIps := []types.BlockIpType{}
	// 一分钟最大请求次数
	minMaxErrNum := viper.Viper.GetInt("app.one_min_error_num")
	// 先检测是否是我们的 cdn ip范围
	for tmpIp, _ := range triggerMap {
		isIn := i.IsIpInCDN(tmpIp)
		// 在里面我们需要去掉这个ip进入黑名单
		if isIn {
			delete(triggerMap, tmpIp)
		}
	}
	for tmpIp, minMap := range triggerMap {
		isNeedDelMapKey := 0
		for minuteInt64, times := range minMap {
			// 如果一分钟直接超过次数 直接跳过后续判断直接拉黑
			if times >= minMaxErrNum {
				tmpTime, _ := time.ParseInLocation(consts.TimeYmdHi, strconv.FormatInt(minuteInt64, 10), time.Local)
				blockIpTmp := types.BlockIpType{
					Ip:      tmpIp,
					AddTime: &tmpTime,
				}
				needAddIps = append(needAddIps, blockIpTmp)
				isNeedDelMapKey = 1
				break
			}
		}
		// 如果已经进入拉黑名单了 我们后续就不要做二次筛选确认是否拉黑了
		if isNeedDelMapKey == 1 {
			delete(triggerMap, tmpIp)
		}
	}
	// 如果第一次筛选数据后 还有单分钟没有触发的 我们搜索连续分钟是否有超过阈值的访问请求
	for tmpIp, minMap := range triggerMap {
		for minuteInt64, times := range minMap {
			nextMinTimes, ok := minMap[minuteInt64+1]
			if !ok {
				continue
			}
			if (nextMinTimes + times) >= minMaxErrNum {
				tmpTime, _ := time.ParseInLocation(consts.TimeYmdHi, strconv.FormatInt(minuteInt64, 10), time.Local)
				blockIpTmp := types.BlockIpType{
					Ip:      tmpIp,
					AddTime: &tmpTime,
				}
				needAddIps = append(needAddIps, blockIpTmp)
				break
			}
		}
	}
	// 写入nginx的block的 配置conf文件
	if len(needAddIps) > 0 {
		file, err := os.OpenFile(denyPath, os.O_APPEND|os.O_WRONLY, 0644)
		defer file.Close()
		if err != nil {
			logrus.LogrusObj.Errorf("读取文件失败：path:%s,%s", denyPath, err.Error())
			utils.RobotMsg("读取文件失败：path:%s,%s", denyPath, err.Error())
			return
		}
		for _, blockIpInfo := range needAddIps {
			tmpString := "deny " + blockIpInfo.Ip + ";#" + blockIpInfo.AddTime.Format(consts.TimeYmdHis) + "\n"
			_, err := file.WriteString(tmpString)
			if err != nil {
				logrus.LogrusObj.Errorf("写入文件失败：path:%s,%s", denyPath, err.Error())
				return
			}
			utils.RobotMsg("服务器ip:%s, 新增封禁IP：%s, 封禁开始时间为：%s", viper.Viper.GetString("app.ip"), blockIpInfo.Ip, blockIpInfo.AddTime.Format(consts.TimeYmdHis))
		}
		err = utils.ReloadNginx()
		if err != nil {
			utils.RobotMsg("服务器ip:%s, nginx配置文件修改异常：%s", viper.Viper.GetString("app.ip"), err.Error())
			return
		}
	}

	if readedIndex != lineIndex {
		// 将行数写入config.toml中去
		viper.Viper.Set(serviceName+".nginx_deny_conf_dir", denyPath)
		viper.Viper.Set(serviceName+".nginx_logs_dir", nginxLogsPath)
		viper.Viper.Set(serviceName+".start_line_index", readedIndex)
		viper.SetConfig()
	}
}

// 2024年9月9日16:31:58 是否是供应商的ip
func (i *IpForbidBus) IsIpInCDN(ip string) bool {
	i.mtx.Lock()
	defer i.mtx.Unlock()
	// 供应商1
	isIn, err := utils.IsIpInSlice(ip, i.CloudflareIps)
	if isIn {
		return true
	}
	if err != nil {
		logrus.LogrusObj.Errorf("ip 错误%s", err.Error())
	}
	// 供应商人2
	for _, v := range i.QuicIps {
		if ip == v {
			return true
		}
	}

	return false
}

// 获取供应商1的列表
func CloudFlareListGet() []string {
	// 供应商1
	cdnIpConfigCloud := viper.Viper.GetString("cdn.cloudflare")
	ipsRsp := rsp.CloudFlareRsp{}
	err := utils.ApiGet(cdnIpConfigCloud, &ipsRsp, 5*time.Second)
	if err != nil {
		logrus.LogrusObj.Errorf("请求cloudflare ip列表异常：%s", err.Error())
	} else {
		return ipsRsp.Result.Ipv4Cidrs
	}
	return []string{}
}

// 获取ip列表
func QuicListGet() []string {
	cdnIpConfigQuic := viper.Viper.GetString("cdn.quic")
	ipBytes, _ := utils.GetNoJsonReturnReq(cdnIpConfigQuic)
	if len(ipBytes) == 0 {
		return []string{}
	}
	return strings.Split(string(ipBytes), "\n")
}

/**
 * @note: 已经被配置阻止ip的服务器
 * @auth: tongWz
 * @date: 2024/8/30 9:40
**/
func BlockNginxFile(path string) map[string]types.BlockIpType {
	// 打开文件
	file, err := os.Open(path)
	defer file.Close()
	if err != nil {
		logrus.LogrusObj.Errorf("读取文件失败：path:%s,%s", path, err.Error())
		return nil
	}
	// 创建一个带缓冲的读取器
	reader := bufio.NewReader(file)

	denyIps := make(map[string]types.BlockIpType)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				// fmt.Println("文件结束")
				break // 文件结束
			}
			logrus.LogrusObj.Errorf("读取文件异常%s", err.Error())
			break
		}
		if line == "" {
			continue
		}
		denyIpAndTime := strings.Split(line, ";")
		// 必须是 一个deny属性  一个日期属性
		if len(denyIpAndTime) != 2 {
			continue
		}
		denyIp, ok := strings.CutPrefix(denyIpAndTime[0], "deny ")
		denyIp = strings.TrimSpace(denyIp)
		if !ok {
			// logrus.LogrusObj.Errorf("黑名单ip格式异常 原始字符串是：%s", line)
			continue
		}
		blockIpStruct := types.BlockIpType{
			Ip: denyIp,
		}
		// 获取日期
		entryTime, ok := strings.CutPrefix(denyIpAndTime[1], "#")
		entryTime = strings.TrimSpace(entryTime)
		if !ok {
			logrus.LogrusObj.Errorf("黑名单ip格式异常 原始字符串是：%s", line)
			continue
		}
		// 将日期格式化
		lineTime, err := time.ParseInLocation(consts.TimeYmdHis, entryTime, time.Local)
		if err != nil {
			logrus.LogrusObj.Errorf("时间转化异常 原始时间字符串是：%s, error :%s", entryTime, err.Error())
			continue
		}
		blockIpStruct.AddTime = &lineTime
		denyIps[denyIp] = blockIpStruct
	}
	return denyIps
}

/**
 * @note: 读取nginx日志 自后进行错误匹配
 * @auth: tongWz
 * @date: 2024/8/30 11:49
**/
func ReadNginxLog(startLineIndex int, nginxFilePath string, blockMap map[string]types.BlockIpType) (map[string]map[int64]int, int) {
	// 打开文件
	file, err := os.Open(nginxFilePath)
	defer file.Close()
	if err != nil {
		logrus.LogrusObj.Errorf("读取文件失败：%s", err.Error())
		return nil, 0
	}

	// 创建一个带缓冲的读取器
	reader := bufio.NewReader(file)
	// 我们的读取index 如果>500行 我们都要 -200 以免漏了之前时间访问的数据
	realLineIndex := 0
	if startLineIndex > 200 {
		realLineIndex = startLineIndex - 200
	}

	// 第一个key IP 第二个map key 分钟的整型 值 - 最新的那一分钟
	triggerIpsMap := make(map[string]map[int64]int)
	index := 0
	// 一次最多读取1w行就结束这次任务
	maxRead := 0
	for {
		if maxRead > 10000 {
			break
		}
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				// fmt.Println("文件结束")
				break // 文件结束
			}
			logrus.LogrusObj.Errorf("读取文件异常%s", err.Error())
			break
		}
		// 行数自增
		index++

		if line == "" {
			continue
		}
		// 如果当前行数
		if index < realLineIndex {
			continue
		}
		// 前面跳过的行不做自增 从开始真正读取文件开始算
		// 大于1w行下次再读取
		maxRead++
		isTrigger, triggerTime, ip, _ := MatchTriggerIp(line)
		// 不是我们的异常情况
		if !isTrigger {
			continue
		}
		// 已经在禁止名单里的 不进入黑名单
		if _, ok := blockMap[ip]; ok {
			continue
		}
		// 异常情况 我们统计异常数据
		timeMinute := triggerTime.Format(consts.TimeYmdHi)
		timeMinuteInt, _ := strconv.ParseInt(timeMinute, 10, 64)
		minMaps, ok := triggerIpsMap[ip]
		if !ok {
			minMaps = make(map[int64]int)
		}
		minTimes, ok := minMaps[timeMinuteInt]
		if !ok {
			minMaps[timeMinuteInt] = 1
		} else {
			minMaps[timeMinuteInt] = minTimes + 1
		}
		// 别忘记赋值
		triggerIpsMap[ip] = minMaps
	}

	return triggerIpsMap, index
}

/**
 * @note: 匹配出是否有 错误信息之后 收集 请求时间 ip 请求全域名
 * @auth: tongWz
 * @date: 2024/8/30 11:19
**/
func MatchTriggerIp(lineStr string) (isTrigger bool, triggerTime time.Time, ip string, request string) {
	isHad := strings.Contains(lineStr, "Primary script unknown")
	if isHad {
		var err error
		// re := regexp.MustCompile(`\[(\d{4}/\d{2}/\d{2} \d{2}:\d{2}:\d{2})\] .*? client: ([\d\w\.]+), .*? server: (.+), .*? request: "(.+?)"`)
		re := regexp.MustCompile(`(\d{4}/\d{2}/\d{2} \d{2}:\d{2}:\d{2}) .+ client: (\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}),.+server: (.*?),.+request: "(GET|POST) (.+) HTTP/\d\.\d"`)
		// 使用FindStringSubmatch查找匹配项
		matches := re.FindStringSubmatch(lineStr)
		// for _, match := range matches {
		// 	fmt.Println(match)
		// }

		if len(matches) < 6 {
			logrus.LogrusObj.Errorf("没有匹配到日期，客户端ip，服务器域名+url信息, 全部str:%s", lineStr)
			return false, time.Time{}, "", ""
		}

		// 提取匹配的字段
		accessTime := matches[1] // 日志时间
		triggerTime, err = time.ParseInLocation(consts.TimeYmdHisBias, accessTime, time.Local)
		if err != nil {
			logrus.LogrusObj.Errorf("没有匹配到日期 全部str:%s, error:%s", lineStr, err.Error())
			return false, time.Time{}, "", ""
		}
		clientIP := matches[2] // 客户端IP
		host := matches[3]     // 服务器域名
		// 4 被GET或者POST占用了
		requestURL := matches[5] // 请求的URL
		return true, triggerTime, clientIP, host + requestURL
	}
	return false, time.Time{}, "", ""
}
