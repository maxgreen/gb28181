package api

import (
	"github.com/gin-gonic/gin"
	"github.com/gowvp/gb28181/internal/core/sms"
	"github.com/ixugo/goweb/pkg/web"
)

type WebHookAPI struct {
	core sms.Core
}

func NewWebHookAPI(core sms.Core) WebHookAPI {
	return WebHookAPI{core: core}
}

func registerZLMWebhook(r gin.IRouter, api WebHookAPI, handler ...gin.HandlerFunc) {
	{
		group := r.Group("/webhook", handler...)
		group.POST("/on_server_keepalive", web.WarpH(api.onServerKeepalive))
		group.POST("/on_stream_changed", web.WarpH(api.onStreamChanged))
	}
}

// onServerKeepalive 服务器定时上报时间，上报间隔可配置，默认 10s 上报一次
// https://docs.zlmediakit.com/zh/guide/media_server/web_hook_api.html#_16%E3%80%81on-server-keepalive
func (w WebHookAPI) onServerKeepalive(_ *gin.Context, in *onServerKeepaliveInput) (gin.H, error) {
	w.core.Keepalive(in.MediaServerID)
	return gin.H{}, nil
}

// onStreamChanged rtsp/rtmp 流注册或注销时触发此事件；此事件对回复不敏感。
// https://docs.zlmediakit.com/zh/guide/media_server/web_hook_api.html#_12%E3%80%81on-stream-changed
func (w WebHookAPI) onStreamChanged(_ *gin.Context, in *onStreamChangedInput) (gin.H, error) {
	return gin.H{}, nil
}
