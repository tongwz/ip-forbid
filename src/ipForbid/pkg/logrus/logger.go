package logrus

import (
	"ipForbid/consts"
	"ipForbid/pkg/basePath"
	"log"
	"os"
	"path"
	"time"

	"github.com/sirupsen/logrus"
)

// 日志对象
var LogrusObj *logrus.Logger

// 条件编译 -ldflags='-X InfoMiner/pkg/logrus.IsUseElasticSearch=true'
// var IsUseElasticSearch string

// 最后更新时间
var lastTime time.Time

func InitLog() {
	if LogrusObj != nil {
		src, _ := setOutputFile()
		// 设置输出
		LogrusObj.Out = src
		return
	}
	// 实例化
	logger := logrus.New()
	src, _ := setOutputFile()
	// 设置输出
	logger.Out = src
	// 设置日志级别
	logger.SetLevel(logrus.DebugLevel)
	// 设置日志格式
	logger.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: consts.TimeYmdHis,
	})
	// 加入执行日志的代码源
	logger.SetReportCaller(true)
	LogrusObj = logger
	lastTime = time.Now()
	go LogTick()
}

// 按时间更新日志文件
func LogTick() {
	nextDay := time.Now().AddDate(0, 0, 1)
	// 获取第二天凌晨的数据
	newNextDay, _ := time.ParseInLocation(consts.TimeYmd, nextDay.Format(consts.TimeYmd), time.Local)

	timeDuration := newNextDay.Sub(lastTime)

	timer := time.NewTimer(timeDuration + time.Second + 1)
	for {
		select {
		case t := <-timer.C:
			src, _ := setOutputFile()
			LogrusObj.Out = src
			lastTime = time.Now()
			go LogTick()
			LogrusObj.Infof("文件名发生变化，我们替换文件名：%s", t.Format(consts.TimeYmdHis))
			return
		}
	}
}

func setOutputFile() (*os.File, error) {
	now := time.Now()
	logFilePath := ""
	logFilePath = basePath.BasePath + "/logs/"
	_, err := os.Stat(logFilePath)
	if os.IsNotExist(err) {
		if err := os.MkdirAll(logFilePath, 0777); err != nil {
			log.Println(err.Error())
			return nil, err
		}
	}
	logFileName := "logger" + now.Format(consts.TimeYmd) + ".log"
	// 日志文件
	fileName := path.Join(logFilePath, logFileName)
	if _, err := os.Stat(fileName); err != nil {
		if _, err := os.Create(fileName); err != nil {
			log.Println(err.Error())
			return nil, err
		}
	}
	// 写入文件
	src, err := os.OpenFile(fileName, os.O_APPEND|os.O_WRONLY, os.ModeAppend)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return src, nil
}
