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
	ClientId  string
	FinalTick time.Duration

	MQTT []ConfigMQTT

	Tasks    map[string]ConfigTask
	Schedule ConfigSchedule
}

type ConfigTask struct {
	Qos byte
	Fn  string
	//	Fn  lua.LFunction
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
