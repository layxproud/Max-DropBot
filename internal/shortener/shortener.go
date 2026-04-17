package shortener

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	TTL        = 24 * time.Hour
	tokenBytes = 6
)

type Shortener struct {
	rdb    *redis.Client
	prefix string
}

func NewShortener(rdb *redis.Client, publicPrefix string) *Shortener {
	return &Shortener{rdb: rdb, prefix: publicPrefix}
}
func (s *Shortener) Shorten(ctx context.Context, longURL string) (string, error) {
	b := make([]byte, tokenBytes)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	token := base64.RawURLEncoding.EncodeToString(b)

	if err := s.rdb.Set(ctx, "short:"+token, longURL, TTL).Err(); err != nil {
		return "", err
	}
	return fmt.Sprintf("%s%s", s.prefix, token), nil
}

func (s *Shortener) Resolve(ctx context.Context, token string) (string, error) {
	return s.rdb.Get(ctx, "short:"+token).Result()
}
