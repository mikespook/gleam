package gleam

import (
	"time"
)

type Config struct {
	MQTT      []ConfigMQTT
	Prefix    string
	ClientId  string
	Tasks     map[string]byte
	Schedule  ConfigSchedule
	FinalTick time.Duration
}

type ConfigMQTT struct {
	Addr     string
	Username string
	Password string
}

type ConfigSchedule struct {
	Tick  time.Duration
	Tasks map[string]time.Duration
}
