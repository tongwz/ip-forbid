package req

type RobotMsg struct {
	Msgtype string       `json:"msgtype"`
	Text    RobotContent `json:"text"`
}

type RobotContent struct {
	Content string `json:"content"`
}
