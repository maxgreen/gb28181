package gbs

import (
	"github.com/gowvp/gb28181/internal/core/gb28181"
	"github.com/gowvp/gb28181/pkg/gbs/sip"
)

// 获取设备信息（注册设备）
func (g GB28181API) QueryDeviceInfo(ctx *sip.Context) {
	tx, err := ctx.SendRequest(sip.MethodMessage, sip.GetDeviceInfoXML(ctx.DeviceID))
	if err != nil {
		ctx.Log.Error("sipDeviceInfo", "err", err)
		return
	}
	if _, err := sipResponse(tx); err != nil {
		ctx.Log.Error("sipResponse", "err", err)
		return
	}
}

// MessageDeviceInfoResponse 主设备明细返回结构
type MessageDeviceInfoResponse struct {
	CmdType      string `xml:"CmdType"`
	SN           int    `xml:"SN"`
	DeviceID     string `xml:"DeviceID"`
	DeviceName   string `xml:"DeviceName"` // 设备名
	DeviceType   string `xml:"DeviceType"`
	Manufacturer string `xml:"Manufacturer"` // 生产商
	Model        string `xml:"Model"`        // 设备型号
	Firmware     string `xml:"Firmware"`     // 固件版本
}

func (g GB28181API) sipMessageDeviceInfo(ctx *sip.Context) {
	var msg MessageDeviceInfoResponse
	if err := sip.XMLDecode(ctx.Request.Body(), &msg); err != nil {
		ctx.Log.Error("sipMessageDeviceInfo", "err", err)
		ctx.String(400, ErrXMLDecode.Error())
		return
	}

	if err := g.store.Edit(ctx.DeviceID, func(d *gb28181.Device) {
		d.Ext.Firmware = msg.Firmware
		d.Ext.Manufacturer = msg.Manufacturer
		d.Ext.Model = msg.Model
		d.Ext.Name = msg.DeviceName
	}); err != nil {
		ctx.Log.Error("Edit", "err", err)
		ctx.String(500, ErrDatabase.Error())
		return
	}

	ctx.String(200, "OK")

	// db.UpdateAll(db.DBClient, new(Devices), db.M{"deviceid=?": u.DeviceID}, Devices{
	// 	Model:        message.Model,
	// 	DeviceType:   message.DeviceType,
	// 	Firmware:     message.Firmware,
	// 	Manufacturer: message.Manufacturer,
	// })
}
