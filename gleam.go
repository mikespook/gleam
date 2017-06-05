package gleam

import (
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

func NewGleam(workdir string) *Gleam {
	return &Gleam{
		lua: newLuaEnv(workdir),
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
	return g.lua.onEvent("afterInit", g.mqttClient)
}

func (g *Gleam) initSchedule(config *Config) error {
	g.scheduler = schego.New(config.Schedule.Tick * time.Millisecond)
	g.scheduler.ErrorFunc = g.lua.errorFunc
	for name, interval := range config.Schedule.Tasks {
		g.scheduler.Add(name, time.Now(), interval*time.Millisecond, schego.ForEver, g.lua.newScheduleFunc(name, g.mqttClient))
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
	opts.SetDefaultPublishHandler(g.lua.newMQTTFunc("defaultTask"))
	opts.SetAutoReconnect(true)
	g.mqttClient = mqtt.NewClient(opts)
	if token := g.mqttClient.Connect(); token.Wait() && token.Error() != nil {
		return token.Error()
	}
	for topic, task := range g.config.Tasks {
		qos := task.Qos
		if token := g.mqttClient.Subscribe(topic, qos, g.lua.newMQTTFunc(task.Fn)); token.Wait() && token.Error() != nil {
			return token.Error()
		}
		fn := "{default}"
		if task.Fn != "" {
			fn = task.Fn
		}
		log.Printf("Subscribe: %s (%d) => %s", topic, qos, fn)
	}
	return nil
}

func (g *Gleam) Serve() error {
	go g.scheduler.Serve()

	sh := signal.New(nil)
	sh.Bind(os.Interrupt, func() uint {
		return signal.BreakExit
	})
	sh.Bind(syscall.SIGTERM, func() uint {
		return signal.BreakExit
	})
	sh.Wait()
	return nil
}

func (g *Gleam) Final() error {
	if err := g.scheduler.Close(); err != nil {
		return err
	}

	for topic, _ := range g.config.Tasks {
		if token := g.mqttClient.Unsubscribe(topic); token.Wait() && token.Error() != nil {
			log.Printf("Unsubscribe error: %s => [%s]", topic, token.Error())
		}
		log.Printf("Unsubscribe: %s", topic)
	}

	if err := g.lua.onEvent("beforeFinalize", g.mqttClient); err != nil {
		log.Printf("BeforeFinalize: %s", err)
	}
	if g.config.FinalTick != 0 {
		time.Sleep(g.config.FinalTick * time.Millisecond)
	}
	g.mqttClient.Disconnect(500)
	return g.lua.Final()
}
