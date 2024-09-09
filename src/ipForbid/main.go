package main

import (
	"ipForbid/business"
	"ipForbid/pkg/logrus"
)

func main() {
	// 初始化日志系统
	logrus.InitLog()
	// 我们需要初始化
	// var cmd = &cobra.Command{Use: "comm"}
	//
	// cmd.AddCommand(command.CrontabCmd)
	//
	// _ = cmd.Execute()

	// 想到了一个更好的 先全量读取拿到读取指针位置再 增量读取，避免了 定时任务的 前面行数跳过的流程专注于后续增量读取
	// 读取需要进行封禁Ip的项目名称，通过项目名称找到日志位置和
	serviceMap := business.GetAllService()
	if serviceMap == nil {
		return
	}
	// fmt.Printf("获取到所有需要监控的服务map:%s \n", serviceMap)
	// 通过每个service拿到它的日志相关配置 和 读取逻辑 异步进行
	for serviceName, serInfo := range serviceMap {
		if len(serInfo) == 0 {
			continue
		}
		go business.GetIpForbidBusObj().IncrementReadingBuf(serviceName)
	}
	select {}
}
