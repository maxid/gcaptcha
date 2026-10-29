package gcaptcha

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
)

// takeScript atomically reads and deletes a challenge (one-time use).
var takeScript = redis.NewScript(`
local v = redis.call('GET', KEYS[1])
if v then
  redis.call('DEL', KEYS[1])
end
return v
`)

// RedisStore keeps challenges in Redis with a TTL.
type RedisStore struct {
	client redis.UniversalClient
	prefix string
}

// NewRedisStore creates a Redis-backed store. An empty prefix defaults to
// "captcha:".
func NewRedisStore(addr, password string, db int, prefix string) *RedisStore {
	if prefix == "" {
		prefix = "captcha:"
	}
	return &RedisStore{
		client: redis.NewClient(&redis.Options{Addr: addr, Password: password, DB: db}),
		prefix: prefix,
	}
}

// Close closes the Redis client.
func (s *RedisStore) Close() error { return s.client.Close() }

// Save implements Store.
func (s *RedisStore) Save(ctx context.Context, id string, data ValidData, ttl time.Duration) error {
	raw, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return s.client.Set(ctx, s.prefix+id, raw, ttl).Err()
}

// Take implements Store.
func (s *RedisStore) Take(ctx context.Context, id string) (ValidData, bool, error) {
	raw, err := takeScript.Run(ctx, s.client, []string{s.prefix + id}).Text()
	if errors.Is(err, redis.Nil) {
		return ValidData{}, false, nil
	}
	if err != nil {
		return ValidData{}, false, err
	}
	var data ValidData
	if err := json.Unmarshal([]byte(raw), &data); err != nil {
		return ValidData{}, false, err
	}
	return data, true, nil
}
