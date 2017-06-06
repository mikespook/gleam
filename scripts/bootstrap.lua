Logf("Initialisation")

-- include libraries
tasks = require("tasks")
utils = require("utils")
json = require("json")

-- global variables
Prefix = "gleam"

-- config
config.MQTT = {
	{
		Addr = "tcp://test.mosquitto.org:1883",
		Username = "",
		Password = ""
	}
}
config.ClientId = "testing"
-- topic: {fn, qos}
config.Tasks = {}
config.Tasks[Prefix .. ":test1"] = {
	Fn = "tasks.test1",
	Qos = 0
}
config.Tasks[Prefix .. ":test2"] = {} -- no fn will lead to call defaultTask

config.Schedule.Tick = 1000
-- fn: tick
config.Schedule.Tasks = {}
config.Schedule.Tasks["tasks.heartbeat"] = 5000
config.Schedule.Tasks["non-exist"] = 3000

-- functions
function defaultTask(client, msg) 
	Logf("Default: %s", msg.Payload)
end

function afterInit(Client)
	Log("After Initialisation")
	-- register the client to hub
	data = {
		M = "init",
		ID = config.ClientId,
		Hostname = utils.getHostname(),
		TS = os.time()
	}
	token = Client:Publish(Prefix .. ":hub", 0, true, json.encode(data))
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
	token = Client:Publish(Prefix .. ":hub", 0, true, json.encode(data))
	if token:Wait() and token:Error() ~= nil then
		Logf("onAfterInit: %s", token:Error())
	end
	Log("Before Finalisation")
end

function onError(event, err)
	Log("%v: %s", event, err)
end
