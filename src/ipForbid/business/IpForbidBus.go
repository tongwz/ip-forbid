package business

import (
	"ipForbid/pkg/logrus"
	"ipForbid/pkg/viper"
)

func GetAllService() map[string]map[string]string {
	allServiceSlice := viper.Viper.GetStringSlice("app.deny_service")
	if len(allServiceSlice) == 0 {
		logrus.LogrusObj.Infof("没有查询到需要进行ip检测的服务")
		return nil
	}
	serviceMap := make(map[string]map[string]string)
	for _, service := range allServiceSlice {
		serviceMap[service] = viper.Viper.GetStringMapString(service)
	}
	return serviceMap

}
