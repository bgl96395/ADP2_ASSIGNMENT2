package middleware

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

func NewRateLimiterInterceptor(rdb *redis.Client, rpm int) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		if rdb == nil {
			return handler(ctx, req)
		}

		ip := "unknown"
		p, ok := peer.FromContext(ctx)
		if ok {
			ip = p.Addr.String()
		}

		window := time.Now().UTC().Format("200601021504")
		key := fmt.Sprintf("ratelimit:doctor:%s:%s", ip, window)

		pipe := rdb.Pipeline()
		incr := pipe.Incr(ctx, key)
		pipe.Expire(ctx, key, 2*time.Minute)
		_, err := pipe.Exec(ctx)
		if err != nil {
			log.Printf("WARN: rate limiter Redis error: %v", err)
			return handler(ctx, req)
		}

		count := incr.Val()
		if count > int64(rpm) {
			retryAfter := 60 - time.Now().Second()
			return nil, status.Errorf(
				codes.ResourceExhausted,
				"rate limit exceeded: %d/%d req/min. Retry after %d seconds",
				count, rpm, retryAfter,
			)
		}

		_ = strconv.Itoa(rpm)
		return handler(ctx, req)
	}
}
