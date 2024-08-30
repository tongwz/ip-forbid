package req

type RobotMsg struct {
	Msgtype string       `json:"msg_type"`
	Content RobotContent `json:"content"`
}

type RobotContent struct {
	Text string `json:"text"`
}
