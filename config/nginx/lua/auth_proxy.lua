local http = require "resty.http"
local cjson = require "cjson"

local function split(str,reps)
    local resultStrList = {}
    string.gsub(str,'[^'..reps..']+',function (w)
        table.insert(resultStrList,w)
    end)
    return resultStrList
end

--[[
    1、解析出路径中的sid和其它路径
--]]

-- 获取请求的路径
local request_uri = ngx.var.request_uri
-- 分割路径
local data = split(request_uri, '/')

-- 请求路径为 /ws/sid/... , 因此至少为2个
if #data < 2 then
    return
end

-- lua中数组下标从1开始，sid为第二个
local ws = data[1]
if ws ~= "ws" then
    return ngx.exit(404)
end

local sid = data[2]
local sid_index = string.find(request_uri, sid)
local other_path_indx = sid_index + string.len(sid)
-- 获取到sid后面的路径
local other_path = string.sub(request_uri, other_path_indx + 1)

if other_path == '/' then
    other_path = ''
end

-- 设置nginx.conf中的变量
ngx.var.pth = other_path

--[[
    2、验证用户权限 - 检查JWT Token和sid所有权
--]]

-- 获取Authorization header
local auth_header = ngx.var.http_authorization
if not auth_header then
    ngx.log(ngx.ERR, "Missing Authorization header for sid: " .. sid)
    return ngx.exit(ngx.HTTP_UNAUTHORIZED)
end

-- 检查Bearer token格式
if not string.match(auth_header, "^Bearer ") then
    ngx.log(ngx.ERR, "Invalid Authorization format for sid: " .. sid)
    return ngx.exit(ngx.HTTP_UNAUTHORIZED)
end

local token = string.sub(auth_header, 8) -- 去掉"Bearer "

-- 调用webserver验证token和权限
local httpc = http.new()
httpc:set_timeout(5000) -- 5秒超时

local res, err = httpc:request_uri("http://cloud-ide-web-svc.cloud-ide.svc.cluster.local:8088/api/internal/verify-workspace-access", {
    method = "POST",
    headers = {
        ["Content-Type"] = "application/json",
        ["Authorization"] = "Bearer " .. token
    },
    body = cjson.encode({
        sid = sid
    })
})

if not res then
    ngx.log(ngx.ERR, "Failed to verify workspace access: " .. (err or "unknown error"))
    return ngx.exit(ngx.HTTP_INTERNAL_SERVER_ERROR)
end

if res.status ~= 200 then
    ngx.log(ngx.WARN, "Access denied for sid: " .. sid .. ", status: " .. res.status)
    return ngx.exit(ngx.HTTP_FORBIDDEN)
end

--[[
    3、从共享内存中根据sid查询后端ip和端口
--]]

local eps = ngx.shared.endpoints
local ep, flags = eps:get(sid)
if not ep then
    ngx.log(ngx.WARN, "Workspace not found for sid: " .. sid)
    return ngx.exit(ngx.HTTP_BAD_GATEWAY)
end

ngx.log(ngx.INFO, 'Authorized access - sid:' .. sid .. ', host:' .. ep)

-- 设置backend
ngx.var.backend = ep
ngx.log(ngx.NOTICE, "other_path: " .. other_path)