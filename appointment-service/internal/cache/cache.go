package cache

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/bgl96395/ADP2_ASSIGNMENT2/appointment-service/internal/model"
	"github.com/redis/go-redis/v9"
)

type CacheRepository interface {
	GetAppointment(ctx context.Context, id string) (*model.Appointment, error)
	SetAppointment(ctx context.Context, id string, a *model.Appointment, ttl time.Duration) error
	GetAppointmentList(ctx context.Context) ([]*model.Appointment, error)
	SetAppointmentList(ctx context.Context, list []*model.Appointment, ttl time.Duration) error
	InvalidateAppointment(ctx context.Context, id string) error
	InvalidateAppointmentList(ctx context.Context) error
}

type RedisCacheRepository struct {
	client *redis.Client
}

func NewRedisCacheRepository(client *redis.Client) *RedisCacheRepository {
	return &RedisCacheRepository{client: client}
}

func (red *RedisCacheRepository) GetAppointment(ctx context.Context, id string) (*model.Appointment, error) {
	key := "appointment:" + id
	val, err := red.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var a model.Appointment
	err = json.Unmarshal([]byte(val), &a)
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (red *RedisCacheRepository) SetAppointment(ctx context.Context, id string, a *model.Appointment, ttl time.Duration) error {
	key := "appointment:" + id
	data, err := json.Marshal(a)
	if err != nil {
		return err
	}
	return red.client.Set(ctx, key, data, ttl).Err()
}

func (red *RedisCacheRepository) GetAppointmentList(ctx context.Context) ([]*model.Appointment, error) {
	val, err := red.client.Get(ctx, "appointments:list").Result()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var list []*model.Appointment
	err = json.Unmarshal([]byte(val), &list)
	if err != nil {
		return nil, err
	}
	return list, nil
}

func (red *RedisCacheRepository) SetAppointmentList(ctx context.Context, list []*model.Appointment, ttl time.Duration) error {
	data, err := json.Marshal(list)
	if err != nil {
		return err
	}
	return red.client.Set(ctx, "appointments:list", data, ttl).Err()
}

func (red *RedisCacheRepository) InvalidateAppointment(ctx context.Context, id string) error {
	return red.client.Del(ctx, "appointment:"+id).Err()
}

func (red *RedisCacheRepository) InvalidateAppointmentList(ctx context.Context) error {
	return red.client.Del(ctx, "appointments:list").Err()
}

type NoopCache struct{}

func (noopCache *NoopCache) GetAppointment(_ context.Context, _ string) (*model.Appointment, error) {
	return nil, nil
}
func (noopCache *NoopCache) SetAppointment(_ context.Context, _ string, _ *model.Appointment, _ time.Duration) error {
	return nil
}
func (noopCache *NoopCache) GetAppointmentList(_ context.Context) ([]*model.Appointment, error) {
	return nil, nil
}
func (noopCache *NoopCache) SetAppointmentList(_ context.Context, _ []*model.Appointment, _ time.Duration) error {
	return nil
}
func (noopCache *NoopCache) InvalidateAppointment(_ context.Context, _ string) error {
	return nil
}
func (noopCache *NoopCache) InvalidateAppointmentList(_ context.Context) error {
	return nil
}

func NewRedisClient(redisURL string) (*redis.Client, error) {
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, err
	}
	client := redis.NewClient(opts)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		log.Printf("WARN: Redis ping failed: %v", err)
		return nil, err
	}
	return client, nil
}
