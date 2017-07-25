package gleam

import (
	"crypto/tls"
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

	g.initMQTT()
	g.initSchedule()

	return g.lua.onEvent(AfterInitFunc, g.mqttClient)
}

func (g *Gleam) initSchedule() {
	if g.config.Schedule.Tick == 0 {
		log.Printf("Scheduler not init")
		return
	}
	g.scheduler = schego.New(g.config.Schedule.Tick * time.Millisecond)
	g.scheduler.ErrorFunc = g.lua.onError
	for name, interval := range g.config.Schedule.Tasks {
		f := g.lua.newOnSchedule(name, &g.mqttClient)
		g.scheduler.Add(name, time.Now(), interval*time.Millisecond, schego.ForEver, f)
		fn := "Skip"
		if f != nil {
			fn = "Serve"
		}
		log.Printf("Schedule: %s (%d) => %s", name, interval, fn)
	}
}

func (g *Gleam) initMQTT() {
	if g.config.ClientId == "" {
		log.Printf("MQTT not init")
		return
	}

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
	opts.SetDefaultPublishHandler(g.lua.newOnMessage(MessageFunc))
	opts.SetAutoReconnect(true)
	if g.config.NotVerifyTLS {
		opts.SetTLSConfig(&tls.Config{InsecureSkipVerify: true})
	}
	opts.SetOnConnectHandler(g.newOnConnect())
	opts.SetConnectionLostHandler(g.newOnLostConnect())
	g.mqttClient = mqtt.NewClient(opts)
	for {
		token := g.mqttClient.Connect()
		if !token.Wait() || token.Error() == nil {
			break
		}
		log.Printf("Conn Error: %s", token.Error())
		time.Sleep(5 * time.Second)
	}
}

func (g *Gleam) newOnConnect() mqtt.OnConnectHandler {
	return func(client mqtt.Client) {
		for name, task := range g.config.Tasks {
			for topic, qos := range task {
				f := g.lua.newOnMessage(name)
				if f == nil {
					name = "{default}"
				}
				log.Printf("Subscribe: %s (%d) => %s", topic, qos, name)
				if token := client.Subscribe(topic, qos, f); token.Wait() && token.Error() != nil {
					log.Printf("Subscribe Error: %s %s", topic, token.Error())
				}
			}
		}
	}
}

func (g *Gleam) newOnLostConnect() mqtt.ConnectionLostHandler {
	return func(client mqtt.Client, err error) {
		log.Printf("Conn Error: %s", err)
	}
}

func (g *Gleam) Serve() {
	if g.scheduler != nil {
		go g.scheduler.Serve()
	}

	sh := signal.New(nil)
	sh.Bind(os.Interrupt, func() uint {
		return signal.BreakExit
	})
	sh.Bind(syscall.SIGTERM, func() uint {
		return signal.BreakExit
	})
	sh.Wait()
}

func (g *Gleam) Final() error {
	if g.scheduler != nil {
		if err := g.scheduler.Close(); err != nil {
			return err
		}
	}

	for topic, _ := range g.config.Tasks {
		if token := g.mqttClient.Unsubscribe(topic); token.Wait() && token.Error() != nil {
			log.Printf("Unsubscribe error: %s => [%s]", topic, token.Error())
		}
		log.Printf("Unsubscribe: %s", topic)
	}

	if err := g.lua.onEvent(BeforeFinalizeFunc, g.mqttClient); err != nil {
		log.Printf("BeforeFinalize: %s", err)
	}
	if g.config.FinalTick != 0 {
		time.Sleep(g.config.FinalTick * time.Millisecond)
	}
	if g.mqttClient != nil {
		g.mqttClient.Disconnect(500)
	}
	g.lua.Final()
	return nil
}
