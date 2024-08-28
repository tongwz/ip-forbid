package crontab

import "time"

/**
 * @note: 进行接口定义
 * @auth: tongWz
 * @date: 2024/8/28 16:52
**/
type ITimer interface {
	Init()
	IsOpen() bool
	GetSpec() string
	Run()
}

// 定时任务基础结构
type BaseCron struct {
	isRunning      bool   // 是否在执行中
	name           string // 定时任务名称
	beginTimeStamp int64  // 开始执行的时间戳
}

// 设置定时任务名称
func (bc *BaseCron) setName(name string) {
	bc.name = name
}

// 获取定时任务名称
func (bc *BaseCron) getName() string {
	return bc.name
}

// 设置执行状态
func (bc *BaseCron) setRunningStatus(isRun bool) {
	bc.isRunning = isRun
	if isRun {
		bc.beginTimeStamp = time.Now().Unix()
	}
}

// 判断任务运行与否
func (bc *BaseCron) IsRunning() bool {
	return bc.isRunning
}
