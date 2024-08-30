package rsp

type CloudFlareRsp struct {
	Success  bool         `json:"success"`
	Errors   any          `json:"errors"`
	Messages any          `json:"messages"`
	Result   ResultStruct `json:"result"`
}

type ResultStruct struct {
	Ipv4Cidrs []string `json:"ipv4_cidrs"`
	Etag      string   `json:"etag"`
}
