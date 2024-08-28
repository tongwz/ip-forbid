package basePath

import (
	"os"
	"path/filepath"
)

var BasePath = myBasePath()

func myBasePath() string {
	var path string
	var ok bool
	if path, ok = os.LookupEnv("INFO_MINER"); ok {
		return path
	}

	path, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		panic("根目录地址查询失败：" + err.Error())
	}
	return path
}
