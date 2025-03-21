-- leaky_bucket.lua

local function drip_reqs(key, no_of_reqs)
    local res = redis.call("LRANGE", key, 0, no_of_reqs - 1)
    redis.call("LTRIM", key, no_of_reqs, -1)
    return res
end

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

if command == "take" then
    return take(KEYS[1], tostring(ARGV[2]), tonumber(ARGV[3]))
elseif command == "core" then
    return drip_reqs(KEYS[1], tonumber(ARGV[2]))
else
    return redis.error_reply("Invalid command")
end
