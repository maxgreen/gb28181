package media

import "github.com/ixugo/goweb/pkg/orm"

const StatusPushing = "PUSHING" // 推流中状态

type StreamPush struct {
	ID            int
	App           string    // 应用名
	CreatedAt     orm.Time  // 创建时间
	UpdatedAt     orm.Time  // 更新时间
	PushedAt      *orm.Time // 最后一次推流时间
	StoppedAt     *orm.Time // 最后一次停止时间
	Stream        string    // 流 ID
	MediaServerID string    // 媒体服务器 ID
	ServerID      string    // 服务器 ID
	Status        string    // 推流状态(PUSHING)
}
