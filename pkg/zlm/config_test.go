package zlm

import (
	"fmt"
	"testing"
)

func TestEngine_GetServerConfig(t *testing.T) {
	const url = "http://127.0.0.1:8080"
	e := NewEngine().SetConfig(Config{URL: url, Secret: "OHvo86N9Ww6V8mHPWMisxNgkb8dvqAV420241107"})
	out, err := e.GetServerConfig()
	if err != nil {
		t.Errorf("Engine.GetServerConfig() error = %v", err)
		return
	}
	fmt.Printf("%+v", out)
}
