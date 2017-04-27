package gleam

import (
	"time"
)

type Config struct {
	Brokers  []string
	Prefix   string
	ClientId string
	Tasks    map[string]byte
	Schedule ConfigSchedule
}

type ConfigSchedule struct {
	Tick  time.Duration
	Tasks map[string]time.Duration
}
