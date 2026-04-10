local limit = tonumber(ARGV[1])
local window = tonumber(ARGV[2])

-- Check all keys first
for i, key in ipairs(KEYS) do
    local current = tonumber(redis.call("GET", key) or "0")
    if current >= limit then
        local ttl = redis.call("PTTL", key)
        return {current + 1, ttl, key}
    end
end

-- Increment all keys
for i, key in ipairs(KEYS) do
    local count = redis.call("INCR", key)
    if count == 1 then
        redis.call("PEXPIRE", key, window)
    end
end

local main_ttl = redis.call("PTTL", KEYS[1])
local main_count = tonumber(redis.call("GET", KEYS[1]))
return {main_count, main_ttl, ""}
