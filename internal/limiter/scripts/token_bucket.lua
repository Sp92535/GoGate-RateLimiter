-- token_bucket.lua

-- function to refill tokens in the bucket
local function refill_tokens(key, capacity, refill)
    -- getting current tokens in bucket
    local tokens = tonumber(redis.call("GET", key) or 0)

    -- refilling with new tokens
    local newTokens = math.min(capacity, tokens + refill)
    redis.call("SET", key, newTokens)

    return newTokens
end

-- function to permit request
local function take(key)
    -- getting current tokens in bucket
    local tokens = tonumber(redis.call("GET", key) or 0)

    -- take the token if bucket is not empty
    if tokens > 0 then
        redis.call("DECR", key)
        return 1
    else
        return 0
    end
end

local command = ARGV[1]
local key = KEYS[1]
if command == "take" then
    return take(key)
elseif command == "core" then
    local capacity = tonumber(ARGV[2])
    local refill = tonumber(ARGV[3])
    return refill_tokens(key, capacity, refill)
else
    return redis.error_reply("Invalid command")
end
