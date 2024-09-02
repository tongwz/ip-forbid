package crontab

import (
	"fmt"
	"ipForbid/business"
	"ipForbid/consts"
	"ipForbid/pkg/logrus"
	"sync"
	"time"
)

type IpForbidTimer struct {
	BaseCron
}

// 是否使用这个定时任务 这个项目在多服务器中执行 因此按服务器判断定时任务执行情况
func (cto *IpForbidTimer) IsOpen() bool {
	return true
}

// 初始化准备
func (cto *IpForbidTimer) Init() {
	// 设置定时任务名称
	cto.setName("IpForbidTimer")
}

// 设置定时任务 执行时间  ------ 秒 分 时 几号 月份 周几
func (cto *IpForbidTimer) GetSpec() string {
	return "*/20 * * * * *"
}

/**
 * @note: 同步ip黑名单逻辑
 * @auth: tongWz
 * @date: 2024年8月28日17:03:29
**/
func (cto *IpForbidTimer) Run() {
	startTime := time.Now()
	if cto.IsRunning() {
		logrus.LogrusObj.Errorf("定时任务正在执行中：%s, 执行开始时间：%s", cto.getName(), time.Unix(cto.beginTimeStamp, 0).Format(consts.TimeYmdHis))
		return
	}
	// 设置开始运行状态
	cto.setRunningStatus(true)
	// 设置完成时 运行状态为结束
	defer cto.setRunningStatus(false)

	// 运行逻辑
	// 读取需要进行封禁Ip的项目名称，通过项目名称找到日志位置和
	serviceMap := business.GetAllService()
	if serviceMap == nil {
		return
	}
	// fmt.Printf("获取到所有需要监控的服务map:%s \n", serviceMap)
	// 通过每个service拿到它的日志相关配置 和 读取逻辑 异步进行
	waitG := sync.WaitGroup{}
	for serviceName, serInfo := range serviceMap {
		if len(serInfo) == 0 {
			continue
		}
		waitG.Add(1)
		go business.GetIpForbidBusObj().ServiceErrorWatch(serviceName, serInfo, &waitG)
	}
	waitG.Wait()
	fmt.Printf("我们执行任务完成:开始时间：%s 当前时间是：%s \n", startTime.Format(consts.TimeYmdHis), time.Now().Format(consts.TimeYmdHis))
}
