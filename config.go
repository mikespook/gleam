package gleam

import (
	"time"
)

const (
	TopicAll        = ""
	TopicIndividual = "i"
	TopicBroadcast  = "b"
)

type Config struct {
	MQTT      []ConfigMQTT
	Prefix    string
	ClientId  string
	Tasks     map[string]ConfigTask
	Schedule  ConfigSchedule
	FinalTick time.Duration
}

type ConfigTask struct {
	Qos   byte
	Topic string
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
