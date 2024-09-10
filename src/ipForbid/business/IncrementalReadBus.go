package business

import (
	"bufio"
	"fmt"
	"ipForbid/consts"
	"ipForbid/pkg/logrus"
	"ipForbid/pkg/types"
	"ipForbid/pkg/viper"
	"ipForbid/utils"
	"os"
	"strconv"
	"time"
)

/**
 * @note: 专门为了进行增量读取 并且不需要进行逐行阅读的效率方法
 * @auth: tongWz
 * @date: 2024/9/2 9:52
**/
func (i *IpForbidBus) IncrementReadingBuf(serviceName string) {
	// 30秒运行一次
	for {
		serviceMap := GetService(serviceName)
		if serviceMap == nil {
			logrus.LogrusObj.Fatalf("读取配置异常：serviceName:%s", serviceName)
			return
		}
		denyPath, nginxLogsPath, lineIndex, sizeIndex := GetServiceAttribute(serviceMap)

		// 读取已封禁ip列表
		// 一般长这样 deny 127.0.0.1;#2024-08-29 17:06:51
		denyIps := BlockNginxFile(denyPath)
		// fmt.Printf("封禁的ip:%#v \n", denyIps)

		// 读取当前文件指针的buf开始进行读取 - 超大文件可能时间较长 不过生产应该不会有那么多error日志
		readedIndex := i.IncrementalReading(sizeIndex, nginxLogsPath, denyIps)
		fmt.Printf("readedIndex = %d\n", readedIndex)

		needAddIps := []types.BlockIpType{}
		// 一分钟最大请求次数
		minMaxErrNum := viper.Viper.GetInt("app.one_min_error_num")
		// 先检测是否是我们的 cdn ip范围
		var cdnErrIps = make(map[string]string)
		for tmpIp, errMap := range i.forbidMaps {
			isIn := i.IsIpInCDN(tmpIp)
			// 在里面我们需要去掉这个ip进入黑名单
			if isIn {
				// 将cdn的异常访问提醒出来
				for minuteTmp, timesTmp := range errMap {
					cdnErrIps[tmpIp] = fmt.Sprintf("时间: %d, 异常访问次数: %d", minuteTmp, timesTmp)
				}
				delete(i.forbidMaps, tmpIp)
			}
		}
		// 将CDN异常情况 提醒到群里
		if len(cdnErrIps) > 0 {
			for tmpIp, errStr := range cdnErrIps {
				utils.RobotMsg("cdn服务器IP：%s, 异常访问 %s", tmpIp, errStr)
			}
		}

		for tmpIp, minMap := range i.forbidMaps {
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
				} else if nextMinTimes, ok := minMap[minuteInt64+1]; ok {
					// 如果第一次筛选数据后 还有单分钟没有触发的 我们搜索连续分钟是否有超过阈值的访问请求
					if (nextMinTimes + times) >= minMaxErrNum {
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
			}

			// 如果已经进入拉黑名单了 我们后续就不要做二次筛选确认是否拉黑了
			if isNeedDelMapKey == 1 {
				delete(i.forbidMaps, tmpIp)
			}
		}

		// 写入nginx的block的 配置conf文件
		if len(needAddIps) > 0 {
			WriteBlockNginxFile(needAddIps, denyPath)
		}

		if readedIndex != sizeIndex {
			// 将行数写入config.toml中去
			viper.Viper.Set(serviceName+".nginx_deny_conf_dir", denyPath)
			viper.Viper.Set(serviceName+".nginx_logs_dir", nginxLogsPath)
			viper.Viper.Set(serviceName+".start_line_index", lineIndex)
			viper.Viper.Set(serviceName+".start_size_index", readedIndex)
			viper.SetConfig()
		}
		// 增量运行30秒一次
		time.Sleep(time.Second * 30)
	}
}

// 2024年9月9日17:31:52 写入nginx的拉黑conf文件
func WriteBlockNginxFile(needAddIps []types.BlockIpType, denyPath string) {
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

func GetServiceAttribute(serviceMap map[string]string) (denyPath string, nginxLogsPath string, lineIndex int64, sizeIndex int64) {
	denyPath, ok := serviceMap["nginx_deny_conf_dir"]
	if !ok || denyPath == "" {
		fmt.Printf("denyPath异常:%s \n", denyPath)
		return
	}
	nginxLogsPath, ok = serviceMap["nginx_logs_dir"]
	if !ok || nginxLogsPath == "" {
		fmt.Printf("nginxLogsPath异常:%s \n", nginxLogsPath)
		return
	}
	// 这里的是通过行数来读取了
	nginxStartLineIndex, ok := serviceMap["start_line_index"]
	if !ok || nginxStartLineIndex == "" {
		fmt.Printf("nginxStartLineIndex异常:%s \n", nginxStartLineIndex)
		return
	}
	lineIndex, err := strconv.ParseInt(nginxStartLineIndex, 10, 64)
	if err != nil {
		logrus.LogrusObj.Errorf("转化文件大小异常：%s", err)
		return
	}

	// 这里是通过大小来锁定开始读取位置
	nginxStartSizeIndex, ok := serviceMap["start_size_index"]
	if !ok || nginxStartSizeIndex == "" {
		fmt.Printf("nginxStartSizeIndex异常:%s \n", nginxStartSizeIndex)
		return
	}
	sizeIndex, err = strconv.ParseInt(nginxStartSizeIndex, 10, 64)
	if err != nil {
		logrus.LogrusObj.Errorf("转化文件大小异常：%s", err)
		return
	}
	return
}

// 增量读取文件进行判断是否有新的ip需要加入黑名单
func (i *IpForbidBus) IncrementalReading(startSizeIndex int64, nginxFilePath string, blockMap map[string]types.BlockIpType) int64 {
	// 打开文件
	file, err := os.Open(nginxFilePath)
	defer file.Close()
	if err != nil {
		logrus.LogrusObj.Errorf("读取文件失败：%s", err.Error())
		return startSizeIndex
	}

	var realStartSize int64
	// 如果我们有开始文件大小读取位置我们需要跳过那个大小 也有可能从0开始
	realStartSize = startSizeIndex

	fmt.Printf("传入： realStartSize:%d \n", realStartSize)
	// 文件状态， 可以后续知道读取指针位置
	stat, _ := file.Stat()
	// 如果配置文件大小 大于  当前的 那么就用小的当前的
	fmt.Printf("stat.Size = %d \n", stat.Size())
	if stat.Size() < realStartSize {
		realStartSize = stat.Size()
	}
	fmt.Printf("实际上开始的文件大小指针是： realStartSize:%d \n", realStartSize)

	// 场景有两种 历史数据读取，实时新增读取

	// 为后续逐行读取做准备
	scanner := bufio.NewScanner(file)

	// 第一个key IP 第二个map key 分钟的整型 值 = 触发次数 - 最新的那一分钟

	// 检查文件大小是否增加
	newStat, err := file.Stat()
	if err != nil {
		logrus.LogrusObj.Errorf("获取文件信息错误: %v", err)
		time.Sleep(time.Second)
		return realStartSize
	}
	// 本机ip的 客户端也忽略
	localIp := GetLocalHostIp()

	if newStat.Size() > realStartSize {
		// 文件大小增加了，移动文件指针到上次的位置之后
		_, err := file.Seek(realStartSize, 0)
		if err != nil {
			logrus.LogrusObj.Errorf("追踪文件错误: %v", err)
			return realStartSize
		}

		// 逐行读取新的内容
		for scanner.Scan() {
			line := scanner.Text()
			// fmt.Println("New content:", line)

			isTrigger, triggerTime, ip, _ := MatchTriggerIp(line)
			// 不是我们的异常情况
			if !isTrigger {
				continue
			}
			// 已经在禁止名单里的 不进入黑名单
			if _, ok := blockMap[ip]; ok {
				continue
			}
			// 忽略本机ip的客户端以免出现访问异常
			if ip == localIp {
				continue
			}
			// 异常情况 我们统计异常数据
			timeMinute := triggerTime.Format(consts.TimeYmdHi)
			timeMinuteInt, _ := strconv.ParseInt(timeMinute, 10, 64)
			minMaps, ok := i.forbidMaps[ip]
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
			i.forbidMaps[ip] = minMaps
		}

		if err := scanner.Err(); err != nil {
			logrus.LogrusObj.Errorf("读取文件错误: %v", err)
			return realStartSize
		}
		realStartSize = newStat.Size()
	}

	fmt.Printf("检查完成：%s \n", time.Now().Format(consts.TimeYmdHis))
	return realStartSize
}
