package viper

import (
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/pelletier/go-toml/v2"
	"ipForbid/pkg/basePath"
	"os"
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
	// 这里动态读取 只能读取到 我们自己修改了配置文件 而不能监控到我们通过命令行修改的操作
	config.OnConfigChange(func(in fsnotify.Event) {
		// log("配置文件已修改成功.修改的配置文件名为:%s", in.Name)
		fmt.Printf("配置文件已修改成功.修改的配置文件名为:%s \n", in.Name)
	})

	return config
}

func SetConfig() {
	// 获取所有的配置项
	configMap := Viper.AllSettings()

	// 序列化配置
	configTOML, err := toml.Marshal(configMap)
	if err != nil {
		fmt.Printf("将配置写成toml对象异常：%s, %s\n", configMap, err)
		return
	}

	// 将配置写回到文件中
	configFilePath := Viper.ConfigFileUsed()
	err = os.WriteFile(configFilePath, configTOML, 0644)
	if err != nil {
		fmt.Printf("将配置写成toml结果写入文件异常：%s, %s\n", configMap, err)
		return
	}
}
