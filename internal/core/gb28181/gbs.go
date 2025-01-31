package gb28181

import (
	"context"

	"github.com/gowvp/gb28181/internal/core/bz"
	"github.com/gowvp/gb28181/internal/core/uniqueid"
	"github.com/ixugo/goweb/pkg/orm"
)

type GB28181 struct {
	deviceStore  DeviceStorer
	channelStore ChannelStorer
	uni          uniqueid.Core
}

func NewGB28181(ds DeviceStorer, cs ChannelStorer, uni uniqueid.Core) GB28181 {
	return GB28181{
		deviceStore:  ds,
		channelStore: cs,
		uni:          uni,
	}
}

func (g GB28181) GetDeviceByDeviceID(deviceID string) (*Device, error) {
	ctx := context.TODO()
	var d Device
	if err := g.deviceStore.Get(ctx, &d, orm.Where("device_id=?", deviceID)); err != nil {
		if !orm.IsErrRecordNotFound(err) {
			return nil, err
		}
		d.init(g.uni.UniqueID(bz.IDPrefixGB), deviceID)
		if err := g.deviceStore.Add(ctx, &d); err != nil {
			return nil, err
		}
	}
	return &d, nil
}

func (g GB28181) Logout(deviceID string, changeFn func(*Device)) error {
	var d Device
	if err := g.deviceStore.Edit(context.TODO(), &d, func(d *Device) {
		changeFn(d)
	}, orm.Where("device_id=?", deviceID)); err != nil {
		return err
	}

	return nil
}

func (g GB28181) Login(deviceID string, changeFn func(*Device)) error {
	var d Device
	if err := g.deviceStore.Edit(context.TODO(), &d, func(d *Device) {
		changeFn(d)
	}, orm.Where("device_id=?", deviceID)); err != nil {
		return err
	}

	return nil
}

func (g GB28181) Edit(deviceID string, changeFn func(*Device)) error {
	var d Device
	if err := g.deviceStore.Edit(context.TODO(), &d, func(d *Device) {
		changeFn(d)
	}, orm.Where("device_id=?", deviceID)); err != nil {
		return err
	}

	return nil
}

func (g GB28181) SaveChannels(channels []*Channel) error {
	if len(channels) <= 0 {
		return nil
	}
	var dev Device
	g.deviceStore.Edit(context.TODO(), &dev, func(d *Device) {
		d.Channels = len(channels)
	}, orm.Where("device_id=?", channels[0].DeviceID))

	for _, channel := range channels {
		var ch Channel
		if err := g.channelStore.Edit(context.TODO(), &ch, func(c *Channel) {
			c.IsOnline = channel.IsOnline
		}, orm.Where("device_id = ? AND channel_id = ?", channel.DeviceID, channel.ChannelID)); err != nil {
			channel.ID = g.uni.UniqueID(bz.IDPrefixGBChannel)
			g.channelStore.Add(context.TODO(), channel)
		}
	}
	return nil
}
