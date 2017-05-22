package gleam

import (
	"fmt"
	"log"
	"os"
	"syscall"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/mikespook/golib/signal"
	"github.com/mikespook/schego"
)

type Gleam struct {
	lua        *luaEnv
	config     Config
	mqttClient mqtt.Client
	scheduler  *schego.Scheduler
}

func NewGleam(root string) *Gleam {
	return &Gleam{
		lua: newLuaEnv(root),
	}
}

func (g *Gleam) Init() error {
	if err := g.lua.Init(&g.config); err != nil {
		return err
	}
	if err := g.initMQTT(); err != nil {
		return err
	}
	if err := g.initSchedule(&g.config); err != nil {
		return err
	}
	return nil
}

func (g *Gleam) initSchedule(config *Config) error {
	g.scheduler = schego.New(config.Schedule.Tick * time.Millisecond)
	g.scheduler.ErrorFunc = g.lua.errorFunc
	for name, interval := range config.Schedule.Tasks {
		g.scheduler.Add(name, time.Now(), interval*time.Millisecond, schego.ForEver, g.lua.newExecFunc(name))
	}
	return nil
}

func (g *Gleam) initMQTT() error {
	opts := mqtt.NewClientOptions()
	for _, broker := range g.config.MQTT {
		opts.AddBroker(broker.Addr)
		log.Printf("Add Broker: %s@%s", broker.Username, broker.Addr)
		if broker.Username != "" {
			opts.SetUsername(broker.Username).SetPassword(broker.Password)
		}
	}
	opts.SetClientID(g.config.ClientId)
	log.Printf("ClientId: %s", g.config.ClientId)
	opts.SetDefaultPublishHandler(g.lua.defaultMQTTHandler)
	opts.SetAutoReconnect(true)
	g.mqttClient = mqtt.NewClient(opts)
	if token := g.mqttClient.Connect(); token.Wait() && token.Error() != nil {
		return token.Error()
	}
	for name, qos := range g.config.Tasks {
		topic := fmt.Sprintf("%s/%s", g.config.Prefix, name)
		if token := g.mqttClient.Subscribe(topic, qos, g.lua.newMQTTHandler(name)); token.Wait() && token.Error() != nil {
			return token.Error()
		}
		log.Printf("Subscribe: %s = %d", topic, qos)
	}
	return nil
}

func (g *Gleam) Serve() error {
	go g.scheduler.Serve()

	sh := signal.New(nil)
	sh.Bind(os.Interrupt, func() uint {
		return signal.BreakExit
	})
	sh.Bind(syscall.SIGHUP, func() uint {
		log.Printf("Reloading scripts")
		for name, qos := range g.config.Tasks {
			topic := fmt.Sprintf("%s/%s", g.config.Prefix, name)
			if token := g.mqttClient.Unsubscribe(topic); token.Wait() && token.Error() != nil {
				log.Printf("Unsubscribe error: %s", token.Error())
			}
			log.Printf("Unsubscribe: %s", topic)
			if token := g.mqttClient.Subscribe(topic, qos, g.lua.newMQTTHandler(name)); token.Wait() && token.Error() != nil {
				log.Printf("Subscribe error: %s", token.Error())
			}
			log.Printf("Subscribe: %s = %d", topic, qos)
		}
		return signal.Continue
	})
	sh.Wait()
	return nil
}

func (g *Gleam) Final() error {
	if err := g.scheduler.Close(); err != nil {
		return err
	}

	for name := range g.config.Tasks {
		topic := fmt.Sprintf("%s/%s", g.config.Prefix, name)
		if token := g.mqttClient.Unsubscribe(topic); token.Wait() && token.Error() != nil {
			log.Printf("Unsubscribe error: %s", token.Error())
		}
		log.Printf("Unsubscribe: %s", topic)
	}
	g.mqttClient.Disconnect(500)
	return g.lua.Final()
}
