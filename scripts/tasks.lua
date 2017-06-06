local _M = {}

function _M.test1(client, msg)
	logf("Test1: %s", msg.Payload)
	token = Client:Publish("foobar", 0, false, "foobar")
	if token:Wait() and token:Error() ~= nil then
		log(token:Error())	
	end
end

function _M.heartbeat(client)
	logf("Heartbeat: %s", os.date())
end

function _M.fireOnMsg(client, msg)
	log(config.Test .. "ABC")
end

function _M.fireOnSchedule(client)
	log(config.Test .. "ABC")
end

return _M
