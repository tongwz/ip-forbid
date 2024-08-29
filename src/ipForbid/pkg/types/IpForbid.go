package types

import "time"

// 加入黑名单时间
type BlockIpType struct {
	Ip      string
	AddTime *time.Time
}
