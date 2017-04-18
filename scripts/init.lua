config = {}
config.Brokers = {"tcp://iot.eclipse.org:1883"}
config.Prefix = "gleam"
config.ClientId = "testing"
config.StateUpdate = 30
config.Tasks = {}
config.Tasks["test1"] = 0
config.Tasks["test2"] = 0
config.Tasks["test3"] = 0

function DefaultPublishHandler(c,m) 
	print(m)
end

