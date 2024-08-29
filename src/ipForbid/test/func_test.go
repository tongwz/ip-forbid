package test

import (
	"ipForbid/business"
	"ipForbid/pkg/logrus"
	"ipForbid/utils"
	"testing"
)

// 2024年8月29日17:28:43 测试查询自身服务器ip 这个方法弃用
// go test -v -run TestSelfIp test/func_test.go -count=1
func TestSelfIp(t *testing.T) {
	logrus.InitLog()
	resp, err := utils.SelfIp()

	if err != nil {
		t.Logf("获取返回信息异常：%v", err)
		return
	}
	t.Logf("获取结果是：%s", resp)
}

// 测试读取nginx的静止文件
// go test -v -run TestBlockNginxFile test/func_test.go -count=1
func TestBlockNginxFile(t *testing.T) {
	logrus.InitLog()
	business.BlockNginxFile("/etc/nginx/conf.d/tongwz_block.conf")

	t.Logf("获取结果是：%s", "结束")
}

// GetAllService
// go test -v -run TestAllService test/func_test.go -count=1
func TestAllService(t *testing.T) {
	logrus.InitLog()
	allService := business.GetAllService()

	t.Logf("获取结果是：%+v", allService)
}
