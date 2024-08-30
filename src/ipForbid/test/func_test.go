package test

import (
	"ipForbid/business"
	"ipForbid/pkg/logrus"
	"testing"
)

// 测试读取nginx的静止文件
// go test -v -run TestBlockNginxFile test/func_test.go -count=1
func TestBlockNginxFile(t *testing.T) {
	logrus.InitLog()
	denyIpList := business.BlockNginxFile("/etc/nginx/conf.d/tongwz_block.conf")

	t.Logf("获取结果是：%#v", denyIpList)
}

// GetAllService
// go test -v -run TestAllService test/func_test.go -count=1
func TestAllService(t *testing.T) {
	logrus.InitLog()
	allService := business.GetAllService()

	t.Logf("获取结果是：%+v", allService)
}

// go test -v -run TestTriggerString test/func_test.go -count=1
func TestTriggerString(t *testing.T) {
	logrus.InitLog()
	logStr := `2024/08/29 17:52:23 [error] 9851#9851: *412 FastCGI sent in stderr: "Primary script unknown" while reading response header from upstream, client: 192.168.56.1, server: www.tongwz.com, request: "GET /cccc/cccc.php HTTP/1.1", upstream: "fastcgi://127.0.0.1:9000", host: "www.tongwz.com"`
	isTrigger, triggerTime, ip, request := business.MatchTriggerIp(logStr)

	t.Logf("获取结果是：%+v, %+v,%+v,%+v", isTrigger, triggerTime, ip, request)
}
