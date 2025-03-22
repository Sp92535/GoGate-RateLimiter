-- fixed_window.lua

-- function to reset request in new window to 0
local function reset_reqs(key)
    redis.call("SET", key, 0)
    return 1
end

-- function to allow request if still space in window
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
local key = KEYS[1]
if command == "take" then
    local no_of_reqs = tonumber(ARGV[2])
    return take(key, no_of_reqs)
elseif command == "core" then
    return reset_reqs(key)
else
    return redis.error_reply("Invalid command")
end
