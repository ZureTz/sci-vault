local limit = tonumber(ARGV[1])
local window = tonumber(ARGV[2])

-- Try to get current value of main key (IP)
local main_count = tonumber(redis.call("GET", KEYS[1]) or "0")
local email_count = tonumber(redis.call("GET", KEYS[2]) or "0")
local composite_count = tonumber(redis.call("GET", KEYS[3]) or "0")

-- If any key is at or above limit, reject
if main_count >= limit or email_count >= limit or composite_count >= limit then
    local ttl = redis.call("PTTL", KEYS[1])
    if ttl < 0 then ttl = window end
    return {main_count + 1, ttl, "blocked"}
end

-- Increment all keys atomically
redis.call("INCR", KEYS[1])
redis.call("INCR", KEYS[2])
redis.call("INCR", KEYS[3])

-- Set expiration on first increment
if main_count == 0 then
    redis.call("PEXPIRE", KEYS[1], window)
end
if email_count == 0 then
    redis.call("PEXPIRE", KEYS[2], window)
end
if composite_count == 0 then
    redis.call("PEXPIRE", KEYS[3], window)
end

local ttl = redis.call("PTTL", KEYS[1])
return {main_count + 1, ttl, "ok"}

