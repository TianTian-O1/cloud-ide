-- 获取body
ngx.req.read_body()
local body = ngx.req.get_body_data()
if not body then
    return ngx.exit(ngx.HTTP_BAD_REQUEST) 
end

-- 保存到共享内存中
local cjson = require("cjson")
local req = cjson.decode(body)

local eps = ngx.shared.endpoints

local ep, flags = eps:get(req.sid)
if not ep then
    return ngx.exit(ngx.HTTP_BAD_REQUEST)
end

ngx.say(ep)