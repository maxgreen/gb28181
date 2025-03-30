package conf

import (
	"time"

	"github.com/ixugo/goweb/pkg/orm"
)

func DefaultConfig() Bootstrap {
	return Bootstrap{
		Server: Server{
			RTMPSecret: "123",
			HTTP: ServerHTTP{
				Port:      15123,
				Timeout:   Duration(60 * time.Second),
				JwtSecret: orm.GenerateRandomString(24),
				PProf: ServerPPROF{
					Enabled:   true,
					AccessIps: []string{"::1", "127.0.0.1"},
				},
			},
		},
		Data: Data{
			Database: Database{
				Dsn:             "./configs/data.db",
				MaxIdleConns:    10,
				MaxOpenConns:    50,
				ConnMaxLifetime: Duration(6 * time.Hour),
				SlowThreshold:   Duration(200 * time.Millisecond),
			},
		},
		Sip: SIP{
			Port:     15060,
			ID:       "3402000000200000001",
			Domain:   "3402000000",
			Password: "",
		},
		Media: Media{
			IP:           "127.0.0.1",
			HTTPPort:     8080,
			Secret:       "",
			WebHookIP:    "127.0.0.1",
			SDPIP:        "127.0.0.1",
			RTPPortRange: "20000-20300",
		},
		Log: Log{
			Dir:          "./logs",
			Level:        "info",
			MaxAge:       Duration(14 * 24 * time.Hour),
			RotationTime: Duration(8 * time.Hour),
			RotationSize: 50,
		},
	}
}
