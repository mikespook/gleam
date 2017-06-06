logf("Initialisation")

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
config.ClientId = utils.getClientId()
-- fn:topic:qos
config.Tasks = {}
config.Tasks["tasks.test1"] = {}
config.Tasks["tasks.test1"][Prefix .. ":test1"] = 0
config.Tasks["tasks.test1"][Prefix .. ":test1:" .. config.ClientId] = 0

config.Tasks["tasks.test2"] = {} -- no fn will lead to call defaultTask
config.Tasks["tasks.test2"][Prefix .. ":test2"] = 0
-- set fire
config.Tasks["tasks.fireOnMsg"] = {}
config.Tasks["tasks.fireOnMsg"][Prefix .. ":fire"] = 0

config.Schedule.Tick = 1000
-- fn: tick
config.Schedule.Tasks = {}
config.Schedule.Tasks["tasks.heartbeat"] = 5000
config.Schedule.Tasks["non-exist"] = 3000
-- set fire
config.Schedule.Tasks["tasks.fireOnSchedule"] = 10000

-- functions
function onDefaultMessage(client, msg) 
	logf("Default: %s", msg.Payload)
end

function afterInit(Client)
	log("After Initialisation")
	-- register the client to hub
	data = {
		M = "init",
		ID = config.ClientId,
		Hostname = utils.getHostname(),
		TS = os.time()
	}
	token = Client:Publish(Prefix .. ":hub", 0, true, json.encode(data))
	if token:Wait() and token:Error() ~= nil then
		logf("onAfterInit: %s", token:Error())
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
		logf("onAfterInit: %s", token:Error())
	end
	log("Before Finalisation")
end

function onError(ctx, err)
	logf("%s: %s", ctx.Id, err)
end
