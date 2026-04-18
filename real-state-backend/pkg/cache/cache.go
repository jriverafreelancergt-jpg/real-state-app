package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"real-state-backend/internal/core/domain"

	"github.com/redis/go-redis/v9"
)

// Cache define una interfaz genérica para cacheo
type Cache interface {
	// Get obtiene un valor del cache
	Get(ctx context.Context, key string, dest interface{}) error
	// Set almacena un valor en el cache
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
	// Delete elimina un valor del cache
	Delete(ctx context.Context, key string) error
	// Flush limpia todo el cache
	Flush(ctx context.Context) error
}

// RedisCache implementa Cache usando Redis
type RedisCache struct {
	client *redis.Client
}

// NewRedisCache crea una instancia de RedisCache
func NewRedisCache(addr string) (*RedisCache, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: "",
		DB:       0,
	})

	// Verificar conexión
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	slog.Info("Redis cache connected", "addr", addr)
	return &RedisCache{client: client}, nil
}

// Get obtiene un valor del cache
func (rc *RedisCache) Get(ctx context.Context, key string, dest interface{}) error {
	val, err := rc.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return fmt.Errorf("cache miss: key=%s", key)
	}
	if err != nil {
		return fmt.Errorf("cache error: %w", err)
	}

	// Deserializar JSON
	if err := json.Unmarshal([]byte(val), dest); err != nil {
		return fmt.Errorf("failed to unmarshal cache value: %w", err)
	}

	return nil
}

// Set almacena un valor en el cache
func (rc *RedisCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	// Serializar JSON
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}

	if err := rc.client.Set(ctx, key, data, ttl).Err(); err != nil {
		return fmt.Errorf("cache set error: %w", err)
	}

	return nil
}

// Delete elimina un valor del cache
func (rc *RedisCache) Delete(ctx context.Context, key string) error {
	if err := rc.client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("cache delete error: %w", err)
	}
	return nil
}

// Flush limpia todo el cache
func (rc *RedisCache) Flush(ctx context.Context) error {
	if err := rc.client.FlushDB(ctx).Err(); err != nil {
		return fmt.Errorf("cache flush error: %w", err)
	}
	return nil
}

// Close cierra la conexión a Redis
func (rc *RedisCache) Close() error {
	return rc.client.Close()
}

// CachePermissions almacena los permisos de un usuario en el cache
func (rc *RedisCache) CachePermissions(ctx context.Context, userID string, permissions []domain.Permission, ttl time.Duration) error {
	key := fmt.Sprintf("user_permissions:%s", userID)
	return rc.Set(ctx, key, permissions, ttl)
}

// GetCachedPermissions obtiene los permisos de un usuario del cache
func (rc *RedisCache) GetCachedPermissions(ctx context.Context, userID string) ([]domain.Permission, error) {
	key := fmt.Sprintf("user_permissions:%s", userID)
	var permissions []domain.Permission
	if err := rc.Get(ctx, key, &permissions); err != nil {
		return nil, err
	}
	return permissions, nil
}

// InvalidatePermissions invalida el cache de permisos de un usuario
func (rc *RedisCache) InvalidatePermissions(ctx context.Context, userID string) error {
	key := fmt.Sprintf("user_permissions:%s", userID)
	return rc.Delete(ctx, key)
}

// InvalidateAllPermissions invalida el cache de permisos de todos los usuarios
func (rc *RedisCache) InvalidateAllPermissions(ctx context.Context) error {
	// Nota: Esta es una operación costosa. En producción usar patrón de keys y delete por batch
	iter := rc.client.Scan(ctx, 0, "user_permissions:*", 0).Iterator()
	var keys []string
	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}
	if err := iter.Err(); err != nil {
		return err
	}

	if len(keys) > 0 {
		if err := rc.client.Del(ctx, keys...).Err(); err != nil {
			return fmt.Errorf("failed to invalidate permissions: %w", err)
		}
	}
	return nil
}
