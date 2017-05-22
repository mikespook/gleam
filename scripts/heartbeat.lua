-- register the client to hub
data = {
	ID = config.ClientId,
	TS = os.time()		
}
token = Client:Publish(config.Prefix .. ":hub", 0, true, json.encode(data))
if token:Wait() and token:Error() ~= nil then
	Logf("onHeartbeat: %s", token:Error())
end
