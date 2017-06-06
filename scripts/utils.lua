local _M = {}

function _M.getHostname()
    local f = io.popen ("/bin/hostname")
    local hostname = f:read("*a") or ""
    f:close()
    hostname =string.gsub(hostname, "\n$", "")
    return hostname
end

function _M.getRandom(l) 
	math.randomseed(os.time())
	local str="0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	if l < 1 then return nil end -- Check for l < 1
	local s = "" -- Start string
	for i = 1, l do
 		s = s .. string.char(str:byte(math.random(1, #str)))
	end
	return s -- Return string
end

function _M.getClientId()
	return _M.getHostname() .. "_" .. _M.getRandom(6)
end

return _M
