package utils

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"ipForbid/http/req"
	"ipForbid/http/rsp"
	"ipForbid/pkg/logrus"
	"ipForbid/pkg/viper"
	"net"
	"net/http"
	"os/exec"
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
		Content: req.RobotContent{
			Text: sendContent,
		},
	}
	for _, url := range urls {
		_ = ApiPost(url, &rspInfo, robotMsg, time.Second*3)
	}
}

// 2024年8月30日15:20:24 进行nginx重启
func ReloadNginx() error {
	// 指定要执行的命令及其参数
	cmd := exec.Command("nginx", "-t")

	// 执行命令并获取输出
	out, err := cmd.CombinedOutput()
	if err != nil {
		logrus.LogrusObj.Errorf("Error nginx -t: %v, Output: %s", err, out)
		return err
	}
	cmd = exec.Command("nginx", "-s", "reload")
	out, err = cmd.CombinedOutput()
	if err != nil {
		logrus.LogrusObj.Errorf("Error nginx -s reload: %v, Output: %s", err, out)
		return err
	}
	return nil
}

// 带子网掩码的读取 例如 173.245.48.0/20
func IsIpInSlice(ipStr string, cloudIps []string) (bool, error) {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return false, errors.New("Invalid IP address" + ipStr)
	}

	for _, rangeStr := range cloudIps {
		_, ipNet, err := net.ParseCIDR(rangeStr)
		if err != nil {
			continue
		}
		if ipNet.Contains(ip) {
			return true, nil
		}
	}
	return false, nil
}

// 2024年8月30日17:43:52 获取非标准请求相应的字符串
func GetNoJsonReturnReq(url string) ([]byte, error) {
	// 以get方法请求
	// https://api.ipify.cn/?format=json
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		logrus.LogrusObj.Error("请求url：", "cip.cc", "，错误内容：1", err.Error())
		return []byte{}, err
	}

	timeout := 5 * time.Second
	// 发送请求
	client := &http.Client{
		Timeout: timeout,
	}
	resp, err := client.Do(req)
	if err != nil {
		logrus.LogrusObj.Error("请求url：", "cip.cc", "，错误内容：2", err.Error())
		return []byte{}, err
	}

	defer resp.Body.Close()

	outByte, err := io.ReadAll(resp.Body)
	if err != nil {
		logrus.LogrusObj.Error("请求url：", "cip.cc", "，错误内容：3", err.Error())
		return []byte{}, err
	}

	return outByte, nil
}
