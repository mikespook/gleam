gleam.Debugf("%s", gleam.Data)
c, err = gleam.Get(gleam.Data)
if err ~= nil then
	gleam.Errorf("%s", err)
else
	gleam.Debug(c)
	file = io.open(gleam.Data, "w")
	file:write(c)
	file:close()
end
