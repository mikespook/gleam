Logf("Initialisation")

foobar = require("foobar/foobar")
utils = require("utils")
json = require("json")

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
config.Tasks["test1"] = {
	Qos = 0
}
config.Tasks["non-exist"] = {
	Topic = "i"
}
config.Schedule.Tick = 1000
config.Schedule.Tasks = {}
config.Schedule.Tasks["heartbeat"] = 5000
config.Schedule.Tasks["non-exist"] = 3000

function MQTTDefaultHandler(Client, Message) 
	Logf("Default: %s", Message.Payload)
end

function ScheduleDefaultFunc(ctx, Client)
	Log("Scheduler")
end

function afterInit(Client)
	-- register the client to hub
	data = {
		M = "init",
		ID = config.ClientId,
		Hostname = utils.getHostname(),
		TS = os.time()
	}
	token = Client:Publish(config.Prefix .. ":hub", 0, true, json.encode(data))
	if token:Wait() and token:Error() ~= nil then
		Logf("onAfterInit: %s", token:Error())
	end
end

function beforeFinalize(Client)
	-- register the client to hub
	data = {
		M = "finalize",
		ID = config.ClientId,
		TS = os.time()		
	}
	token = Client:Publish(config.Prefix .. ":hub", 0, true, json.encode(data))
	if token:Wait() and token:Error() ~= nil then
		Logf("onAfterInit: %s", token:Error())
	end
end
