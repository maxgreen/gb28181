package gbs

import (
	"encoding/xml"
	"log/slog"

	"github.com/gowvp/gb28181/pkg/gbs/sip"
)

// MessageDeviceListResponse 设备明细列表返回结构
type MessageDeviceListResponse struct {
	XMLName  xml.Name   `xml:"Response"`
	CmdType  string     `xml:"CmdType"`
	SN       int        `xml:"SN"`
	DeviceID string     `xml:"DeviceID"`
	SumNum   int        `xml:"SumNum"`
	Item     []Channels `xml:"DeviceList>Item"`
}

func (g GB28181API) sipMessageCatalog(ctx *sip.Context) {
	var msg MessageDeviceListResponse
	if err := sip.XMLDecode(ctx.Request.Body(), &msg); err != nil {
		slog.Error("Message Unmarshal xml", "err", err)
		ctx.String(400, "xml err")
		return
	}
	if msg.SumNum < 0 {
		ctx.String(200, "OK")
		return
	}

	for _, d := range msg.Item {
		d.DeviceID = msg.DeviceID
		g.catalog.Write(&sip.CollectorMsg[Channels]{
			Key:   d.DeviceID,
			Data:  &d,
			Total: msg.SumNum,
		})

		// channel := Channels{ChannelID: d.ChannelID, DeviceID: message.DeviceID}
		// if err := db.Get(db.DBClient, &channel); err == nil {
		// 	channel.Active = time.Now().Unix()
		// 	channel.URIStr = fmt.Sprintf("sip:%s@%s", d.ChannelID, _sysinfo.Region)
		// 	channel.Status = transDeviceStatus(d.Status)
		// 	channel.Name = d.Name
		// 	channel.Manufacturer = d.Manufacturer
		// 	channel.Model = d.Model
		// 	channel.Owner = d.Owner
		// 	channel.CivilCode = d.CivilCode
		// 	// Address ip地址
		// 	channel.Address = d.Address
		// 	channel.Parental = d.Parental
		// 	channel.SafetyWay = d.SafetyWay
		// 	channel.RegisterWay = d.RegisterWay
		// 	channel.Secrecy = d.Secrecy
		// 	db.Save(db.DBClient, &channel)
		// 	go notify(notifyChannelsActive(channel))
		// } else {
		// 	// logrus.Infoln("deviceid not found,deviceid:", d.DeviceID, "pdid:", message.DeviceID, "err", err)
		// }
	}

	ctx.String(200, "OK")
}

// QueryCatalog 获取注册设备包含的列表
func (g GB28181API) QueryCatalog(ctx *sip.Context) {
	_, err := ctx.SendRequest(sip.MethodMessage, sip.GetCatalogXML(ctx.DeviceID))
	if err != nil {
		slog.Error("sipCatalog", "err", err)
		return
	}
	g.catalog.Run(ctx.DeviceID)
	g.catalog.Wait(ctx.DeviceID)
}
