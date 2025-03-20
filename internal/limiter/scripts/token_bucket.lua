-- token_bucket.lua

local function refill_tokens(key, capacity, refill)
    local tokens = tonumber(redis.call("GET", key) or 0)
    local newTokens = math.min(capacity, tokens + refill)
    redis.call("SET", key, newTokens)
    return newTokens
end

local function take(key)
    local tokens = tonumber(redis.call("GET", key) or 0)
    if tokens > 0 then
        redis.call("DECR", key)
        return 1
    else
        return 0
    end
end

local command = ARGV[1]

if command=="take" then
    return take(KEYS[1])
elseif command=="core" then
    return refill_tokens(KEYS[1],tonumber(ARGV[2]),tonumber(ARGV[3]))
else
    return redis.error_reply("Invalid command")
end