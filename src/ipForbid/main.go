package main

import (
	"github.com/spf13/cobra"
	"ipForbid/command"
	"ipForbid/pkg/logrus"
)

func main() {
	// 初始化日志系统
	logrus.InitLog()
	// 我们需要初始化
	var cmd = &cobra.Command{Use: "comm"}

	cmd.AddCommand(command.CrontabCmd)

	_ = cmd.Execute()

	// 想到了一个更好的 先全量读取拿到读取指针位置再 增量读取，避免了 定时任务的 前面行数跳过的流程专注于后续增量读取

}
