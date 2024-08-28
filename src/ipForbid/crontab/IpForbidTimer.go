package crontab

import (
	"ipForbid/consts"
	"ipForbid/pkg/logrus"
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
	return "*/6 15 3 * * *"
}

/**
 * @note: 同步ip黑名单逻辑
 * @auth: tongWz
 * @date: 2024年8月28日17:03:29
**/
func (cto *IpForbidTimer) Run() {
	if cto.IsRunning() {
		logrus.LogrusObj.Errorf("定时任务正在执行中：%s, 执行开始时间：%s", cto.getName(), time.Unix(cto.beginTimeStamp, 0).Format(consts.TimeYmdHis))
		return
	}
	// 设置开始运行状态
	cto.setRunningStatus(true)
	// 设置完成时 运行状态为结束
	defer cto.setRunningStatus(false)

	// 运行逻辑
	// fmt.Printf("我们执行的文件名是：%s 当前时间是：%s \n", middleFile, time.Now().Format(utils.TimeFormatLocal))
}
