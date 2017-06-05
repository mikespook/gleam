local _M = {}

function _M.test1(Client, Message)
	Logf("Test1: %s", Message.Payload)
	token = Client:Publish("foobar", 0, false, "foobar")
	if token:Wait() and token:Error() ~= nil then
		Log(token:Error())	
	end
end

return _M
