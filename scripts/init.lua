foobar = require("foobar/foobar")
Logf("Initialisation")

config.MQTT = {
	{
		Addr = "tcp://iot.eclipse.org:1883",
		Username = "",
		Password = ""
	}
}
config.Prefix = "gleam"
config.ClientId = "testing"
config.Tasks = {}
config.Tasks["test1"] = 0
config.Tasks["test2"] = 0
config.Tasks["test3"] = 0
config.Schedule.Tick = 1000
config.Schedule.Tasks = {}
config.Schedule.Tasks["task1"] = 5000

function DefaultPublishHandler(c,m) 
	Log(m.Payload)
end
