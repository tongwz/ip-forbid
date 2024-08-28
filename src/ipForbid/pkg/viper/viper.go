package viper

import (
	"fmt"
	"github.com/fsnotify/fsnotify"
	"ipForbid/pkg/basePath"
	"path/filepath"

	"github.com/spf13/viper"
)

// 配置文件的变量
var Viper = getConfig()

func getConfig() *viper.Viper {
	config := viper.New()

	// 设置配置位置
	config.AddConfigPath(filepath.Join(basePath.BasePath, "config"))
	config.SetConfigName("config")
	config.SetConfigType("toml")

	if err := config.ReadInConfig(); err != nil {
		panic("读取配置文件失败!,原因:" + err.Error())
	}

	// 动态读取配置文件,不需要重启
	config.WatchConfig()
	config.OnConfigChange(func(in fsnotify.Event) {
		// log("配置文件已修改成功.修改的配置文件名为:%s", in.Name)
		fmt.Printf("配置文件已修改成功.修改的配置文件名为:%s \n", in.Name)
	})

	return config
}
