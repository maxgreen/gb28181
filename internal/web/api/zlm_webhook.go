package api

import (
	"github.com/gin-gonic/gin"
	"github.com/gowvp/gb28181/internal/core/media"
	"github.com/gowvp/gb28181/internal/core/sms"
	"github.com/ixugo/goweb/pkg/web"
)

type WebHookAPI struct {
	smsCore   sms.Core
	mediaCore media.Core
}

func NewWebHookAPI(core sms.Core, mediaCore media.Core) WebHookAPI {
	return WebHookAPI{smsCore: core, mediaCore: mediaCore}
}

func registerZLMWebhookAPI(r gin.IRouter, api WebHookAPI, handler ...gin.HandlerFunc) {
	{
		group := r.Group("/webhook", handler...)
		group.POST("/on_server_keepalive", web.WarpH(api.onServerKeepalive))
		group.POST("/on_stream_changed", web.WarpH(api.onStreamChanged))
		group.POST("/on_publish", web.WarpH(api.onPublish))
		group.POST("/on_play", web.WarpH(api.onPlay))
	}
}

// onServerKeepalive 服务器定时上报时间，上报间隔可配置，默认 10s 上报一次
// https://docs.zlmediakit.com/zh/guide/media_server/web_hook_api.html#_16%E3%80%81on-server-keepalive
func (w WebHookAPI) onServerKeepalive(_ *gin.Context, in *onServerKeepaliveInput) (gin.H, error) {
	w.smsCore.Keepalive(in.MediaServerID)
	return gin.H{}, nil
}

// onPublish rtsp/rtmp/rtp 推流鉴权事件。
// https://docs.zlmediakit.com/zh/guide/media_server/web_hook_api.html#_7%E3%80%81on-publish
func (w WebHookAPI) onPublish(c *gin.Context, in *onPublishInput) (*onPublishOutput, error) {
	// TODO: 待完善，鉴权推流
	// TODO: 待重构，封装 publish 接口
	if err := w.mediaCore.Publish(c.Request.Context(), in.App, in.Stream, in.MediaServerID); err != nil {
		return &onPublishOutput{DefaultOutput: DefaultOutput{Code: 1, Msg: err.Error()}}, nil
	}
	return &onPublishOutput{
		DefaultOutput: newDefaultOutputOK(),
	}, nil
}

// onStreamChanged rtsp/rtmp 流注册或注销时触发此事件；此事件对回复不敏感。
// https://docs.zlmediakit.com/zh/guide/media_server/web_hook_api.html#_12%E3%80%81on-stream-changed
func (w WebHookAPI) onStreamChanged(_ *gin.Context, in *struct{}) (DefaultOutput, error) {
	return newDefaultOutputOK(), nil
}

// onPlay rtsp/rtmp/http-flv/ws-flv/hls 播放触发播放器身份验证事件。
// 播放流时会触发此事件。如果流不存在，则首先触发 on_play 事件，然后触发 on_stream_not_found 事件。
// 播放rtsp流时，如果该流开启了rtsp专用认证（on_rtsp_realm），则不会触发on_play事件。
// https://docs.zlmediakit.com/guide/media_server/web_hook_api.html#_6-on-play
func (w WebHookAPI) onPlay(_ *gin.Context, in *onPublishInput) (DefaultOutput, error) {
	return newDefaultOutputOK(), nil
}
