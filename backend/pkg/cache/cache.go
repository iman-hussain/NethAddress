package cache

import (
	"bytes"
	"context"
	"encoding/gob"
	"fmt"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

// CacheService provides caching functionality for API responses
type CacheService struct {
	client *redis.Client
	ctx    context.Context
}

// NewCacheService creates a new cache service instance
func NewCacheService(redisURL string) (*CacheService, error) {
	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Redis URL: %w", err)
	}

	client := redis.NewClient(opt)
	ctx := context.Background()

	// Test connection
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &CacheService{
		client: client,
		ctx:    ctx,
	}, nil
}

// Get retrieves a value from cache
func (cs *CacheService) Get(key string, dest interface{}) error {
	val, err := cs.client.Get(cs.ctx, key).Result()
	if err == redis.Nil {
		return fmt.Errorf("key not found: %s", key)
	}
	if err != nil {
		return fmt.Errorf("cache get error: %w", err)
	}

	b := bytes.NewBuffer([]byte(val))
	if err := gob.NewDecoder(b).Decode(dest); err != nil {
		return fmt.Errorf("failed to decode cached value (gob): %w", err)
	}

	return nil
}

// Set stores a value in cache with TTL
func (cs *CacheService) Set(key string, value interface{}, ttl time.Duration) error {
	var b bytes.Buffer
	if err := gob.NewEncoder(&b).Encode(value); err != nil {
		return fmt.Errorf("failed to encode value (gob): %w", err)
	}

	if err := cs.client.Set(cs.ctx, key, b.Bytes(), ttl).Err(); err != nil {
		return fmt.Errorf("cache set error: %w", err)
	}

	return nil
}

// Delete removes a key from cache
func (cs *CacheService) Delete(key string) error {
	if err := cs.client.Del(cs.ctx, key).Err(); err != nil {
		return fmt.Errorf("cache delete error: %w", err)
	}
	return nil
}

// Exists checks if a key exists in cache
func (cs *CacheService) Exists(key string) (bool, error) {
	count, err := cs.client.Exists(cs.ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("cache exists error: %w", err)
	}
	return count > 0, nil
}

// Close closes the Redis connection
func (cs *CacheService) Close() error {
	return cs.client.Close()
}

// FlushAll clears all keys from the cache
func (cs *CacheService) FlushAll() error {
	if err := cs.client.FlushAll(cs.ctx).Err(); err != nil {
		return fmt.Errorf("cache flush error: %w", err)
	}
	return nil
}

// Cache TTL constants for different data types
const (
	// Property data - changes infrequently
	PropertyDataTTL = 24 * time.Hour

	// Valuation data - update daily
	ValuationDataTTL = 24 * time.Hour

	// Transaction history - rarely changes
	TransactionDataTTL = 7 * 24 * time.Hour

	// Weather data - update every 30 minutes
	WeatherDataTTL = 30 * time.Minute

	// Traffic data - update every 5 minutes
	TrafficDataTTL = 5 * time.Minute

	// Demographics - changes annually
	DemographicsDataTTL = 30 * 24 * time.Hour

	// Air quality - update hourly
	AirQualityDataTTL = 1 * time.Hour

	// Static data (soil, elevation) - cache for 90 days
	StaticDataTTL = 90 * 24 * time.Hour
)

// CacheKey generates consistent cache keys
type CacheKey struct{}

// PropertyKey generates a cache key for property data
func (ck CacheKey) PropertyKey(bagID string) string {
	return fmt.Sprintf("property:%s", bagID)
}

// ValuationKey generates a cache key for valuation data
func (ck CacheKey) ValuationKey(bagID string) string {
	return fmt.Sprintf("valuation:%s", bagID)
}

// TransactionKey generates a cache key for transaction history
func (ck CacheKey) TransactionKey(bagID string) string {
	return fmt.Sprintf("transactions:%s", bagID)
}

// WeatherKey generates a cache key for weather data
func (ck CacheKey) WeatherKey(lat, lon float64) string {
	return fmt.Sprintf("weather:%.4f:%.4f", lat, lon)
}

// TrafficKey generates a cache key for traffic data
func (ck CacheKey) TrafficKey(lat, lon float64, radius int) string {
	return fmt.Sprintf("traffic:%.4f:%.4f:%d", lat, lon, radius)
}

// DemographicsKey generates a cache key for demographics data
func (ck CacheKey) DemographicsKey(regionCode string) string {
	return fmt.Sprintf("demographics:%s", regionCode)
}

// AirQualityKey generates a cache key for air quality data
func (ck CacheKey) AirQualityKey(lat, lon float64) string {
	return fmt.Sprintf("airquality:%.4f:%.4f", lat, lon)
}

// SoilKey generates a cache key for soil data
func (ck CacheKey) SoilKey(lat, lon float64) string {
	return fmt.Sprintf("soil:%.4f:%.4f", lat, lon)
}

// ElevationKey generates a cache key for elevation data
func (ck CacheKey) ElevationKey(lat, lon float64) string {
	return fmt.Sprintf("elevation:%.4f:%.4f", lat, lon)
}

// AggregatedKey generates a cache key for aggregated property data
func (ck CacheKey) AggregatedKey(postcode, houseNumber string) string {
	normalizedPostcode := strings.ToUpper(strings.ReplaceAll(postcode, " ", ""))
	normalizedHouseNumber := strings.TrimSpace(houseNumber)
	return fmt.Sprintf("aggregated:%s:%s", normalizedPostcode, normalizedHouseNumber)
}

// ScoresKey generates a cache key for calculated scores
func (ck CacheKey) ScoresKey(bagID string) string {
	return fmt.Sprintf("scores:%s", bagID)
}
