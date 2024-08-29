package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"ipForbid/http/req"
	"ipForbid/http/rsp"
	"ipForbid/pkg/logrus"
	"ipForbid/pkg/viper"
	"net/http"
	"time"
)

// 结构体转化成json
func CommonStruct2Json[T any](data T) []byte {
	jsByte, _ := json.Marshal(data)
	return jsByte
}

//

func ApiPost(url string, out interface{}, postBody interface{}, time time.Duration) error {
	// interface{}转化成[]byte
	postData, err := json.Marshal(postBody)
	if err != nil {
		logrus.LogrusObj.Error("转化成json出错，请求url：", url, "，请求参数：", postBody, err.Error())
		return err
	}

	// 以post方法请求
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(postData))
	if err != nil {
		logrus.LogrusObj.Error("请求url：", url, "请求参数：", string(postData), "，错误内容：", err.Error())
		return err
	}

	// 以json格式请求
	req.Header.Set("Content-Type", "application/json")

	// 发送请求
	client := &http.Client{
		Timeout: time,
	}
	resp, err := client.Do(req)
	if err != nil {
		logrus.LogrusObj.Error("请求url：", url, "请求参数：", string(postData), "，错误内容：", err.Error())
		return err
	}

	defer resp.Body.Close()

	outByte, err := io.ReadAll(resp.Body)

	// byte[]转化成interface{}
	err = json.Unmarshal(outByte, out)
	if err != nil {
		logrus.LogrusObj.Error("byte[]转化成interface{}出错，请求url：", url, "响应：", string(outByte), "，错误内容：", err.Error())
		return err
	}
	// logrus.LogrusObj.Info("POST请求：",url ,"，请求参数：",string(postData) ,", 返回结果：", string(outByte))

	return nil
}

func ApiGet(url string, out interface{}, time time.Duration) error {
	// 以get方法请求
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		logrus.LogrusObj.Error("请求url：", url, "，错误内容：", err.Error())
		return err
	}

	// 发送请求
	client := &http.Client{
		Timeout: time,
	}
	resp, err := client.Do(req)
	if err != nil {
		logrus.LogrusObj.Error("请求url：", url, "，错误内容：", err.Error())
		return err
	}

	defer resp.Body.Close()

	outByte, err := io.ReadAll(resp.Body)

	// byte[]转化成interface{}
	err = json.Unmarshal(outByte, out)
	if err != nil {
		logrus.LogrusObj.Error("byte[]转化成interface{}出错，请求url：", url, "响应：", string(outByte), "，错误内容：", err.Error())
		return err
	}

	return nil
}

// 2024年8月29日10:53:45 机器消息
func RobotMsg(content string, v ...any) {
	sendContent := fmt.Sprintf(content, v...)
	urls := viper.Viper.GetStringSlice("robot.urls")
	rspInfo := rsp.CommonServerRsp{}
	robotMsg := req.RobotMsg{
		Msgtype: "text",
		Text: req.RobotContent{
			Content: sendContent,
		},
	}
	for _, url := range urls {
		_ = ApiPost(url, &rspInfo, robotMsg, time.Second*3)
	}
}

func SelfIp() (string, error) {
	// 以get方法请求
	req, err := http.NewRequest("GET", "https://api.ipify.cn/?format=json", nil)
	if err != nil {
		logrus.LogrusObj.Error("请求url：", "cip.cc", "，错误内容：1", err.Error())
		return "", err
	}

	timeout := 5 * time.Second
	// 发送请求
	client := &http.Client{
		Timeout: timeout,
	}
	resp, err := client.Do(req)
	if err != nil {
		logrus.LogrusObj.Error("请求url：", "cip.cc", "，错误内容：2", err.Error())
		return "", err
	}

	defer resp.Body.Close()

	outByte, err := io.ReadAll(resp.Body)
	if err != nil {
		logrus.LogrusObj.Error("请求url：", "cip.cc", "，错误内容：3", err.Error())
		return "", err
	}

	return string(outByte), nil
}
