-- sliding_window.lua
local function reset_reqs(key)
    local curr_key = key .. ":curr"
    local prev_key = key .. ":prev"
    local time_key = key .. ":timeStamp"

    local reqs = tonumber(redis.call("GET", curr_key) or 0)
    redis.call("SET", curr_key, 0)
    redis.call("SET", prev_key, reqs)

    local time = redis.call("TIME")
    redis.call("SET", time_key, time[1])

    return 1
end

local function take(key, no_of_reqs, interval)
    local curr_key = key .. ":curr"
    local prev_key = key .. ":prev"
    local time_key = key .. ":timeStamp"

    local curr = tonumber(redis.call("GET", curr_key) or 0)
    local prev = tonumber(redis.call("GET", prev_key) or 0)
    local timeStamp = tonumber(redis.call("GET", time_key) or 0)

    local time = redis.call("TIME")
    local now = tonumber(time[1]) -- Current timestamp from Redis

    local elapsed = now - timeStamp
    local weight = (interval - elapsed) / interval

    if weight < 0 then
        weight = 0
    end

    local reqsInCurrSlidingWindow = prev * weight + curr

    if reqsInCurrSlidingWindow < no_of_reqs then
        redis.call("INCR", curr_key)
        return 1
    else
        return 0
    end
end

local command = ARGV[1]

if command == "take" then
    return take(KEYS[1], tonumber(ARGV[2]), tonumber(ARGV[3]))
elseif command == "core" then
    return reset_reqs(KEYS[1])
else
    return redis.error_reply("Invalid command")
end
