token = Client:Publish("foobar", 0, false, "foobar")
if token:Wait() and token:Error() ~= nil then
	Log(token.Error())	
end
Log(Message.Payload)
