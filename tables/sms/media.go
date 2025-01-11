package sms

import "github.com/ixugo/goweb/pkg/orm"

type MediaServer struct {
	ID                string
	IP                string
	CreatedAt         orm.Time
	UpdatedAt         orm.Time
	HookIP            string
	SDPIP             string
	StreamIP          string
	Ports             MediaServerPorts
	AutoConfig        bool
	Secret            string
	HookAliveInterval int
	RTPEnable         bool
	Status            bool
	RTPPortRange      string
	SendRTPPortRange  string
	RecordAssistPort  int
	LastKeepaliveAt   orm.Time
	IsDefault         bool
	RecordDay         int
	RecordPath        string
	Type              string
	TranscodeSuffix   string
}

type MediaServerPorts struct {
	HTTP     int `json:"http"`
	HTTPS    int `json:"https"`
	RTMP     int `json:"rtmp"`
	FLV      int `json:"flv"`
	FLVs     int `json:"fl_vs"`
	WsFLV    int `json:"ws_flv"`
	WsFLVs   int `json:"ws_fl_vs"`
	RTMPS    int `json:"rtmps"`
	RTPPorxy int `json:"rtp_porxy"`
	RTSP     int `json:"rtsp"`
	RTSPs    int `json:"rts_ps"`
}
