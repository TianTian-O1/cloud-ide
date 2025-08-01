-- 判断method
local method = ngx.req.get_method()
if method ~= "POST" and method ~= "DELETE" then
    return ngx.exit(ngx.HTTP_BAD_REQUEST) 
end

-- 验证Token
local token = ngx.req.get_headers()["token"]
if not token then 
    ngx.exit(ngx.HTTP_UNAUTHORIZED)
end

if token ~= ngx.var.token then
    ngx.exit(ngx.HTTP_UNAUTHORIZED)
end

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

if method == "POST" then
    if not req.sid or not req.endpoint then
        return ngx.exit(ngx.HTTP_BAD_REQUEST)
    end    

    local success, err = eps:set(req.sid, req.endpoint)
    if not success then
        ngx.log(ngx.ERR, "Failed to save data to shared memory:", err)
        return ngx.exit(ngx.HTTP_BAD_REQUEST)
    end
elseif method == "DELETE" then    
    if not req.sid then
        return ngx.exit(ngx.HTTP_BAD_REQUEST)
    end  
    eps:delete(req.sid)
end
