package test

import (
	"ipForbid/pkg/logrus"
	"log"
	"os/exec"
	"testing"
)

// go test -v -run TestCmdNginx test/cmd_test.go -count=1
func TestCmdNginx(t *testing.T) {
	logrus.InitLog()
	// 指定要执行的命令及其参数
	cmd := exec.Command("nginx", "-t")

	// 执行命令并获取输出
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Error nginx: %v, Output: %s", err, out)
	}

	log.Println("Nginx reloaded successfully.", string(out))
}
