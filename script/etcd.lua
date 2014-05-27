gleam.Debugf("%s", gleam.Data)

err = gleam.Delete("/foobar")
if err ~= nil then
	gleam.Errorf("%s", err)
end
err = gleam.Set("/foobar/config1", gleam.Data, 0)
if err ~= nil then
	gleam.Errorf("%s", err)
end

c, err = gleam.Get("/foobar/config1")
if err ~= nil then
	gleam.Errorf("%s", err)
else
	gleam.Message(c)
end
