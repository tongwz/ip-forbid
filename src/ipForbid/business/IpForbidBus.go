package business

import (
	"bufio"
	"io"
	"ipForbid/pkg/logrus"
	"ipForbid/pkg/types"
	"ipForbid/pkg/viper"
	"os"
	"strconv"
	"strings"
)

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

// nginx_logs_dir = "xxxx"
// nginx_deny_conf_dir = "yyyy"
// start_line_index = 0

func ServiceErrorWatch(serviceMap map[string]string) {
	if serviceMap == nil {
		return
	}
	denyPath, ok := serviceMap["nginx_deny_conf_dir"]
	if !ok || denyPath == "" {
		return
	}
	nginxLogsPath, ok := serviceMap["nginx_logs_dir"]
	if !ok || nginxLogsPath == "" {
		return
	}
	nginxStartLineIndex, ok := serviceMap["start_line_index"]
	if !ok || nginxStartLineIndex == "" {
		return
	}

	lineIndex, err := strconv.Atoi(nginxStartLineIndex)
	if err != nil {
		logrus.LogrusObj.Errorf("转化行数异常：%s", err)
		return
	}

	// 读取已封禁ip列表
	// 一般长这样 deny 127.0.0.1;#2024-8-29 17:06:51
	denyIps := BlockNginxFile(denyPath)

	// 读取当前文件index后100行 或者一行一行的读取
}

func BlockNginxFile(path string) []types.BlockIpType {
	// 打开文件
	file, err := os.Open(path)
	defer file.Close()
	if err != nil {
		logrus.LogrusObj.Errorf("读取文件失败：%s", err.Error())
		return nil
	}
	// 创建一个带缓冲的读取器
	reader := bufio.NewReader(file)

	denyIps := []types.BlockIpType{}
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
			logrus.LogrusObj.Errorf("黑名单ip格式异常 原始字符串是：%s", line)
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

		denyIps = append(denyIps, line)
	}
	return denyIps
}
