package command

import (
	"github.com/robfig/cron"
	"github.com/spf13/cobra"
	"ipForbid/crontab"
	"ipForbid/pkg/logrus"
)

var CrontabCmd = &cobra.Command{
	Use:   "timer",
	Short: "定时任务服务",
	Args:  cobra.NoArgs, // 没有参数
	// 命令执行
	Run: func(cmd *cobra.Command, args []string) {
		// 定时任务列表
		var crontabList = []crontab.ITimer{
			// 定时加入封禁IP
			new(crontab.IpForbidTimer),
			// ...
		}
		// 初始化定时器
		cronJob := cron.New()
		// 定时器添加任务
		for _, task := range crontabList {
			if !task.IsOpen() {
				continue
			}
			task.Init()
			err := cronJob.AddFunc(task.GetSpec(), task.Run)
			if err != nil {
				logrus.LogrusObj.Error("定时任务初始化失败")
				return
			}
		}
		cronJob.Start()
		select {}
	},
}
