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

}
