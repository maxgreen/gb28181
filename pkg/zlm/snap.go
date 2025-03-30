package zlm

const (
	getSnapshot = `/index/api/getSnap`
)

type GetSnapRequest struct {
	URL        string `json:"url"`
	TimeoutSec int    `json:"timeout_sec"` // 截图失败超时时间
	ExpireSec  int    `json:"expire_sec"`  // 截图过期时间
}

// GetSnap 获取截图或生成实时截图并返回
// https://docs.zlmediakit.com/zh/guide/media_server/restful_api.html#_23%E3%80%81-index-api-getsnap
func (e *Engine) GetSnap(in GetSnapRequest) ([]byte, error) {
	body, err := struct2map(in)
	if err != nil {
		return nil, err
	}
	return e.post2(getSnapshot, body)
}
