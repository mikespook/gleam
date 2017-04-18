package gleam

import (
	"fmt"
	"log"
	"os"
	"syscall"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/mikespook/golib/signal"
)

type Gleam struct {
	lua        *luaEnv
	config     Config
	mqttClient mqtt.Client
}

func NewGleam(root string) *Gleam {
	return &Gleam{
		lua: newLuaEnv(root),
	}
}

func (g *Gleam) Init() error {
	g.config.Brokers = []string{
		"tcp://iot.eclipse.org:1883",
	}
	if err := g.lua.Init(&g.config); err != nil {
		return err
	}

	log.Printf("%+v", g.config)

	if err := g.initMQTT(); err != nil {
		return err
	}
	return nil
}

func (g *Gleam) initMQTT() error {
	opts := mqtt.NewClientOptions()
	for _, broker := range g.config.Brokers {
		opts.AddBroker(broker)
	}
	opts.SetClientID(g.config.ClientId)
	opts.SetDefaultPublishHandler(g.lua.defaultHandler)
	opts.SetAutoReconnect(true)
	g.mqttClient = mqtt.NewClient(opts)
	if token := g.mqttClient.Connect(); token.Wait() && token.Error() != nil {
		return token.Error()
	}
	for name, qos := range g.config.Tasks {
		g.mqttClient.Subscribe(fmt.Sprintf("%s/%s", g.config.Prefix, name), qos, g.lua.newHandler(name))
	}
	return nil
}

func (g *Gleam) Serve() error {
	sh := signal.New(nil)
	sh.Bind(os.Interrupt, func() uint {
		return signal.BreakExit
	})
	sh.Bind(syscall.SIGHUP, func() uint {
		for name, qos := range g.config.Tasks {
			topic := fmt.Sprintf("%s/%s", g.config.Prefix, name)
			g.mqttClient.Unsubscribe(topic)
			g.mqttClient.Subscribe(topic, qos, g.lua.newHandler(name))
		}
		return signal.Continue
	})
	sh.Wait()
	return nil
}

func (g *Gleam) Final() error {
	for name := range g.config.Tasks {
		g.mqttClient.Unsubscribe(fmt.Sprintf("%s/%s", g.config.Prefix, name))
	}
	g.mqttClient.Disconnect(500)
	return g.lua.Final()
}
