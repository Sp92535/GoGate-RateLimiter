-- leaky_bucket.lua

-- function to get specified no of oldest requests
local function drip_reqs(key, no_of_reqs)
    local res = redis.call("LRANGE", key, 0, no_of_reqs - 1)
    redis.call("LTRIM", key, no_of_reqs, -1)
    return res
end

-- function to permit request if bucket is not full
local function take(key, id, capacity)
    local reqs = redis.call("LLEN", key)
    if reqs < capacity then
        redis.call("LPUSH", key, id)
        return 1
    else
        return 0
    end
end

local command = ARGV[1]
local key = KEYS[1]
if command == "take" then
    local id = tostring(ARGV[2])
    local capacity = tonumber(ARGV[3])
    return take(key, id, capacity)
elseif command == "core" then
    local no_of_reqs = tonumber(ARGV[2])
    return drip_reqs(key, no_of_reqs)
else
    return redis.error_reply("Invalid command")
end
