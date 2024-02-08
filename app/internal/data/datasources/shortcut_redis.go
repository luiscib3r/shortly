package datasources

import (
	"context"
	"os"

	"github.com/luiscib3r/shortly/app/internal/domain/entities"
	"github.com/redis/go-redis/v9"
)

type ShortcutRedis struct {
	rdb *redis.Client
}

func NewShortcutRedis() (*ShortcutRedis, error) {
	var redisAddr string
	redisAddr = os.Getenv("RedisEndpoint")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}

	rdb := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})

	_, err := rdb.Ping(context.TODO()).Result()

	if err != nil {
		return nil, err
	}

	return &ShortcutRedis{
		rdb: rdb,
	}, nil
}

func (s ShortcutRedis) Save(entity entities.Shortcut) (bool, error) {
	_, err := s.rdb.Set(context.Background(), entity.Id(), entity.Url(), 0).Result()

	if err != nil {
		return false, err
	}

	return true, nil
}

func (s ShortcutRedis) FindById(id string) (string, error) {
	result, err := s.rdb.Get(context.TODO(), id).Result()

	if err != nil {
		return "", err
	}

	return result, nil
}
