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
    2、从共享内存中根据sid查询后端ip和端口
    注意：在跳转网页时 一定是 http://ip:port/ws/sid/    最后面一定要有'/'
--]]


local eps = ngx.shared.endpoints
local ep, flags = eps:get(sid)
if not ep then
    return ngx.exit(ngx.HTTP_BAD_GATEWAY)
end

ngx.log(ngx.INFO, 'sid:'..sid..', host:'..ep)

-- 设置backend
ngx.var.backend = ep
ngx.log(ngx.NOTICE, "other_path: "..other_path)

