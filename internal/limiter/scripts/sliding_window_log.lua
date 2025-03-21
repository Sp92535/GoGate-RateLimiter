-- sliding_window_log.lua

local function remove_logs(key)
    local res = redis.call("LPOP", key)
    return tonumber(res) or 0
end

local function take(key, no_of_reqs)
    local reqs = redis.call("LLEN", key)
    if reqs < no_of_reqs then
        local time_data = redis.call("TIME")
        local curr_time = time_data[1] * 1000 + math.floor(time_data[2] / 1000)
        redis.call("LPUSH", key, tonumber(curr_time))
        return 1
    else
        return 0
    end
end

local command = ARGV[1]

if command == "take" then
    return take(KEYS[1], tonumber(ARGV[2]))
elseif command == "core" then
    return remove_logs(KEYS[1])
else
    return redis.error_reply("Invalid command")
end
