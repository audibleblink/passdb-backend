package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"go.etcd.io/bbolt"
)

type CacheConfig struct {
	Enabled    bool
	DefaultTTL time.Duration
	RouteTTLs  map[string]time.Duration
	DBPath     string
}

type CacheEntry struct {
	StatusCode int               `json:"status_code"`
	Headers    map[string]string `json:"headers"`
	Body       []byte            `json:"body"`
	Timestamp  time.Time         `json:"timestamp"`
	TTL        time.Duration     `json:"ttl"`
}

type responseCapture struct {
	http.ResponseWriter
	statusCode int
	body       *bytes.Buffer
	headers    http.Header
}

func newResponseCapture(w http.ResponseWriter) *responseCapture {
	return &responseCapture{
		ResponseWriter: w,
		statusCode:     http.StatusOK,
		body:          new(bytes.Buffer),
		headers:       make(http.Header),
	}
}

func (rc *responseCapture) WriteHeader(statusCode int) {
	rc.statusCode = statusCode
	rc.ResponseWriter.WriteHeader(statusCode)
}

func (rc *responseCapture) Write(data []byte) (int, error) {
	rc.body.Write(data)
	return rc.ResponseWriter.Write(data)
}

func (rc *responseCapture) Header() http.Header {
	return rc.ResponseWriter.Header()
}

func LoadCacheConfig() CacheConfig {
	config := CacheConfig{
		Enabled:    getEnvBool("CACHE_ENABLED", true),
		DefaultTTL: getEnvDuration("CACHE_DEFAULT_TTL", 720*time.Hour), // 30 days
		DBPath:     getEnv("CACHE_DB_PATH", "./cache.db"),
		RouteTTLs:  make(map[string]time.Duration),
	}

	config.RouteTTLs["/breaches/"] = getEnvDuration("CACHE_TTL_BREACHES", 168*time.Hour)   // 7 days
	config.RouteTTLs["/usernames/"] = getEnvDuration("CACHE_TTL_USERNAMES", 720*time.Hour) // 30 days
	config.RouteTTLs["/passwords/"] = getEnvDuration("CACHE_TTL_PASSWORDS", 720*time.Hour) // 30 days
	config.RouteTTLs["/domains/"] = getEnvDuration("CACHE_TTL_DOMAINS", 720*time.Hour)     // 30 days
	config.RouteTTLs["/emails/"] = getEnvDuration("CACHE_TTL_EMAILS", 720*time.Hour)       // 30 days

	return config
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.ParseBool(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}

func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if parsed, err := time.ParseDuration(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}

func CacheMiddleware(config CacheConfig) func(http.Handler) http.Handler {
	if !config.Enabled {
		return func(next http.Handler) http.Handler {
			return next
		}
	}

	db, err := bbolt.Open(config.DBPath, 0600, &bbolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		log.Printf("Failed to open cache database: %v", err)
		return func(next http.Handler) http.Handler {
			return next
		}
	}

	err = db.Update(func(tx *bbolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("cache"))
		return err
	})
	if err != nil {
		log.Printf("Failed to create cache bucket: %v", err)
		db.Close()
		return func(next http.Handler) http.Handler {
			return next
		}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cacheKey := fmt.Sprintf("%s:%s", r.Method, r.URL.Path)
			
			var cached *CacheEntry
			err := db.View(func(tx *bbolt.Tx) error {
				bucket := tx.Bucket([]byte("cache"))
				data := bucket.Get([]byte(cacheKey))
				if data == nil {
					return nil
				}
				
				cached = &CacheEntry{}
				return json.Unmarshal(data, cached)
			})

			if err == nil && cached != nil {
				if time.Since(cached.Timestamp) < cached.TTL {
					for key, value := range cached.Headers {
						w.Header().Set(key, value)
					}
					w.Header().Set("X-Cache", "HIT")
					w.WriteHeader(cached.StatusCode)
					w.Write(cached.Body)
					log.Printf("CACHE HIT: %s", cacheKey)
					return
				}
			}

			rc := newResponseCapture(w)
			next.ServeHTTP(rc, r)

			if rc.statusCode >= 200 && rc.statusCode < 300 {
				ttl := getTTLForPath(r.URL.Path, config)
				
				entry := CacheEntry{
					StatusCode: rc.statusCode,
					Headers:    make(map[string]string),
					Body:       rc.body.Bytes(),
					Timestamp:  time.Now(),
					TTL:        ttl,
				}

				for key, values := range rc.Header() {
					if len(values) > 0 {
						entry.Headers[key] = values[0]
					}
				}

				data, err := json.Marshal(entry)
				if err == nil {
					db.Update(func(tx *bbolt.Tx) error {
						bucket := tx.Bucket([]byte("cache"))
						return bucket.Put([]byte(cacheKey), data)
					})
				}
				log.Printf("CACHE MISS: %s (cached for %v)", cacheKey, ttl)
			} else {
				log.Printf("CACHE MISS: %s (not cached - status %d)", cacheKey, rc.statusCode)
			}

			w.Header().Set("X-Cache", "MISS")
		})
	}
}

func getTTLForPath(path string, config CacheConfig) time.Duration {
	for routePattern, ttl := range config.RouteTTLs {
		if strings.HasPrefix(path, routePattern) {
			return ttl
		}
	}
	return config.DefaultTTL
}