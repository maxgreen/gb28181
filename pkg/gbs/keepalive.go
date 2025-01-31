package gbs

import (
	"github.com/gowvp/gb28181/internal/core/gb28181"
	"github.com/gowvp/gb28181/pkg/gbs/sip"
	"github.com/ixugo/goweb/pkg/orm"
	// "github.com/panjjo/gosip/db"
)

// MessageNotify 心跳包xml结构
type MessageNotify struct {
	CmdType  string `xml:"CmdType"`
	SN       int    `xml:"SN"`
	DeviceID string `xml:"DeviceID"`
	Status   string `xml:"Status"`
	Info     string `xml:"Info"`
}

func (g GB28181API) sipMessageKeepalive(ctx *sip.Context) {
	var msg MessageNotify
	if err := sip.XMLDecode(ctx.Request.Body(), &msg); err != nil {
		ctx.Log.Error("Message Unmarshal xml err", "err", err)
		return
	}

	// device, ok := _activeDevices.Get(ctx.DeviceID)
	// if !ok {
	// device = Devices{DeviceID: ctx.DeviceID}
	// if err := db.Get(db.DBClient, &device); err != nil {
	// logrus.Warnln("Device Keepalive not found ", u.DeviceID, err)
	// }
	// }

	if err := g.store.Edit(ctx.DeviceID, func(d *gb28181.Device) {
		d.KeepaliveAt = orm.Now()
		d.IsOnline = msg.Status == "OK"
	}); err != nil {
		ctx.Log.Error("keepalive", "err", err)
	}

	// _activeDevices.Store(u.DeviceID, u)
	// go notify(notifyDevicesAcitve(u.DeviceID, message.Status))
	// _, err := db.UpdateAll(db.DBClient, new(Devices), map[string]interface{}{"deviceid=?": u.DeviceID}, Devices{
	// 	Host:     u.Host,
	// 	Port:     u.Port,
	// 	Rport:    u.Rport,
	// 	RAddr:    u.RAddr,
	// 	Source:   u.Source,
	// 	URIStr:   u.URIStr,
	// 	ActiveAt: device.ActiveAt,
	// })
	// return err
	// return nil
	ctx.String(200, "OK")
}
