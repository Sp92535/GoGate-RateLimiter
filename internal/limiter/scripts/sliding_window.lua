-- sliding_window.lua

-- function to reset requests in preious and current window
local function reset_reqs(key)

    -- intitializing all keys
    local curr_key = key .. ":curr"
    local prev_key = key .. ":prev"
    local time_key = key .. ":timeStamp"

    -- modifying all keys as per sliding window rules
    local reqs = tonumber(redis.call("GET", curr_key) or 0)
    redis.call("SET", curr_key, 0)
    redis.call("SET", prev_key, reqs)

    local time = redis.call("TIME")
    redis.call("SET", time_key, time[1])

    return 1
end

-- function to permit requests
local function take(key, no_of_reqs, interval)

    -- intitializing all keys
    local curr_key = key .. ":curr"
    local prev_key = key .. ":prev"
    local time_key = key .. ":timeStamp"

    -- calculation of weight to check request in current dynamic window
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
local key = KEYS[1]
if command == "take" then
    local no_of_reqs = tonumber(ARGV[2])
    local interval = tonumber(ARGV[3])
    return take(key, no_of_reqs, interval)
elseif command == "core" then
    return reset_reqs(key)
else
    return redis.error_reply("Invalid command")
end
