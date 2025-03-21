-- fixed_window.lua

local function reset_reqs(key)
    redis.call("SET", key, 0)
    return 1
end

local function take(key, no_of_reqs)
    local reqs = tonumber(redis.call("GET", key) or 0)
    if reqs < no_of_reqs then
        redis.call("INCR", key)
        return 1
    else
        return 0
    end
end

local command = ARGV[1]

if command == "take" then
    return take(KEYS[1], tonumber(ARGV[2]))
elseif command == "core" then
    return reset_reqs(KEYS[1])
else
    return redis.error_reply("Invalid command")
end
