package cache

import (
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/storage/memory/v2"
)

// ResilientStorage wraps a primary storage (like Redis) and falls back to memory if it fails.
type ResilientStorage struct {
	primary   fiber.Storage
	fallback  fiber.Storage
	isHealthy bool
	lastCheck time.Time
}

func NewResilientStorage(primary fiber.Storage) *ResilientStorage {
	return &ResilientStorage{
		primary:   primary,
		fallback:  memory.New(),
		isHealthy: true,
	}
}

func (s *ResilientStorage) checkHealth() {
	if !s.isHealthy && time.Since(s.lastCheck) > 30*time.Second {
		log.Printf("[ResilientStorage] Attempting to reconnect to primary storage...")
		s.isHealthy = true
		s.lastCheck = time.Now()
	}
}


func (s *ResilientStorage) Get(key string) ([]byte, error) {
	s.checkHealth()

	if s.isHealthy {
		val, err := s.primary.Get(key)
		if err == nil {
			return val, nil
		}

		log.Printf("[ResilientStorage] Primary storage error on Get: %v. Falling back to memory.", err)
		s.isHealthy = false
		s.lastCheck = time.Now()
	}

	return s.fallback.Get(key)
}

func (s *ResilientStorage) Set(key string, val []byte, exp time.Duration) error {
	s.checkHealth()

	if s.isHealthy {
		err := s.primary.Set(key, val, exp)
		if err == nil {
			return nil
		}

		log.Printf("[ResilientStorage] Primary storage error on Set: %v. Falling back to memory.", err)
		s.isHealthy = false
		s.lastCheck = time.Now()
	}

	return s.fallback.Set(key, val, exp)
}

func (s *ResilientStorage) Delete(key string) error {
	if s.isHealthy {
		err := s.primary.Delete(key)
		if err == nil {
			return nil
		}
		s.isHealthy = false
	}
	return s.fallback.Delete(key)
}

func (s *ResilientStorage) Reset() error {
	_ = s.fallback.Reset()
	if s.isHealthy {
		return s.primary.Reset()
	}
	return nil
}

func (s *ResilientStorage) Close() error {
	_ = s.fallback.Close()
	return s.primary.Close()
}
