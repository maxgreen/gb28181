package gbs

import (
	"fmt"
	"net"
	"testing"

	"github.com/gowvp/gb28181/pkg/gbs/sip"
)

func TestDevices(t *testing.T) {
	cli := NewClient()

	{
		a := Device{
			source: &net.UDPAddr{},
			to:     &sip.Address{},
		}
		cli.Store("123", &a)

		dev, ok := cli.Load("123")
		dev.channels.Store("123", &Channel{
			ChannelID: "123",
			device:    dev,
		})
		if ok {
			fmt.Printf("1: %p\n", dev)
		}
	}

	{
		a := Device{
			source: &net.UDPAddr{},
			to:     &sip.Address{},
		}
		cli.Store("123", &a)

		dev, ok := cli.Load("123")
		if ok {
			fmt.Printf("2: %p\n", dev)
		}
	}
}
