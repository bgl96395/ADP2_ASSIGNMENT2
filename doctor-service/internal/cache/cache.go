package cache

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/bgl96395/ADP2_ASSIGNMENT2/doctor-service/internal/model"
	"github.com/redis/go-redis/v9"
)

type CacheRepository interface {
	GetDoctor(ctx context.Context, id string) (*model.Doctor, error)
	SetDoctor(ctx context.Context, id string, doctor *model.Doctor, ttl time.Duration) error
	GetDoctorList(ctx context.Context) ([]*model.Doctor, error)
	SetDoctorList(ctx context.Context, doctors []*model.Doctor, ttl time.Duration) error
	InvalidateDoctorList(ctx context.Context) error
}

type RedisCacheRepository struct {
	client *redis.Client
}

func NewRedisCacheRepository(client *redis.Client) *RedisCacheRepository {
	return &RedisCacheRepository{client: client}
}

func (red *RedisCacheRepository) GetDoctor(ctx context.Context, id string) (*model.Doctor, error) {
	key := "doctor:" + id
	val, err := red.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var doctor model.Doctor
	err = json.Unmarshal([]byte(val), &doctor)
	if err != nil {
		return nil, err
	}
	return &doctor, nil
}

func (red *RedisCacheRepository) SetDoctor(ctx context.Context, id string, doctor *model.Doctor, ttl time.Duration) error {
	key := "doctor:" + id
	data, err := json.Marshal(doctor)
	if err != nil {
		return err
	}
	return red.client.Set(ctx, key, data, ttl).Err()
}

func (red *RedisCacheRepository) GetDoctorList(ctx context.Context) ([]*model.Doctor, error) {
	val, err := red.client.Get(ctx, "doctors:list").Result()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var doctors []*model.Doctor
	err = json.Unmarshal([]byte(val), &doctors)
	if err != nil {
		return nil, err
	}
	return doctors, nil
}

func (r *RedisCacheRepository) SetDoctorList(ctx context.Context, doctors []*model.Doctor, ttl time.Duration) error {
	data, err := json.Marshal(doctors)
	if err != nil {
		return err
	}
	return r.client.Set(ctx, "doctors:list", data, ttl).Err()
}

func (red *RedisCacheRepository) InvalidateDoctorList(ctx context.Context) error {
	return red.client.Del(ctx, "doctors:list").Err()
}

type NoopCache struct{}

func (noopCache *NoopCache) GetDoctor(_ context.Context, _ string) (*model.Doctor, error) {
	return nil, nil
}

func (noopCache *NoopCache) SetDoctor(_ context.Context, _ string, _ *model.Doctor, _ time.Duration) error {
	return nil
}

func (noopCache *NoopCache) GetDoctorList(_ context.Context) ([]*model.Doctor, error) {
	return nil, nil
}

func (noopCache *NoopCache) SetDoctorList(_ context.Context, _ []*model.Doctor, _ time.Duration) error {
	return nil
}

func (noopCache *NoopCache) InvalidateDoctorList(_ context.Context) error {
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
	err = client.Ping(ctx).Err()
	if err != nil {
		log.Printf("WARN: Redis ping failed: %v", err)
		return nil, err
	}
	return client, nil
}
