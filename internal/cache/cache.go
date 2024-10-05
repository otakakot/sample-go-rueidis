package cache

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log/slog"

	hc "github.com/Code-Hex/go-generics-cache"
	"github.com/redis/go-redis/v9"
	"github.com/redis/rueidis"
)

var ErrNotFound = errors.New("Not Found")

type Cache[T any] interface {
	Get(ctx context.Context, key string) (*T, error)
	Set(ctx context.Context, key string, val T) error
	Del(ctx context.Context, key string) error
}

var _ Cache[any] = (*InMemory[any])(nil)

type InMemory[T any] struct {
	client *redis.Client
	local  *hc.Cache[string, T]
}

func NewInMemory[T any](client *redis.Client) *InMemory[T] {
	return &InMemory[T]{
		client: client,
		local:  hc.New[string, T](),
	}
}

func (im *InMemory[T]) Get(
	ctx context.Context,
	key string,
) (*T, error) {
	var val T

	if v, ok := im.local.Get("key"); ok {
		slog.InfoContext(ctx, "local cache hit")

		return &v, nil
	} else {
		slog.InfoContext(ctx, "local cache miss")
	}

	v, err := im.client.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, ErrNotFound
		}

		return nil, err
	}

	if err := json.NewDecoder(bytes.NewBufferString(v)).Decode(&val); err != nil {
		return nil, err
	}

	im.local.Set(key, val)

	return &val, nil
}

func (im *InMemory[T]) Set(
	ctx context.Context,
	key string,
	val T,
) error {
	buf := bytes.NewBuffer(nil)

	if err := json.NewEncoder(buf).Encode(val); err != nil {
		return err
	}

	if err := im.client.Set(ctx, key, buf.String(), 0).Err(); err != nil {
		return err
	}

	return nil
}

func (im *InMemory[T]) Del(
	ctx context.Context,
	key string,
) error {
	if _, err := im.client.Del(ctx, key).Result(); err != nil {
		return err
	}

	im.local.Delete(key)

	return nil
}

var _ Cache[any] = (*Rueidis[any])(nil)

type Rueidis[T any] struct {
	client rueidis.Client
}

func NewRueidis[T any](client rueidis.Client) *Rueidis[T] {
	return &Rueidis[T]{
		client: client,
	}
}

func (rue *Rueidis[T]) Get(
	ctx context.Context,
	key string,
) (*T, error) {
	cmd := rue.client.B().Get().Key(key).Build()

	v, err := rue.client.Do(ctx, cmd).ToString()
	if err != nil {
		if errors.Is(err, rueidis.Nil) {
			return nil, ErrNotFound
		}

		return nil, err
	}

	var val T

	if err := json.NewDecoder(bytes.NewBufferString(v)).Decode(&val); err != nil {
		return nil, err
	}

	return &val, nil
}

func (rue *Rueidis[T]) Set(
	ctx context.Context,
	key string,
	val T,
) error {
	buf := bytes.NewBuffer(nil)

	if err := json.NewEncoder(buf).Encode(val); err != nil {
		return err
	}

	cmd := rue.client.B().Set().Key(key).Value(buf.String()).Build()

	if err := rue.client.Do(ctx, cmd).Error(); err != nil {
		return err
	}

	return nil
}

func (rue *Rueidis[T]) Del(
	ctx context.Context,
	key string,
) error {
	cmd := rue.client.B().Del().Key(key).Build()

	if err := rue.client.Do(ctx, cmd).Error(); err != nil {
		return err
	}

	return nil
}
