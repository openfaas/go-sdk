package sdk

import (
	"reflect"
	"testing"
	"time"
)

func Test_TokenCache(t *testing.T) {
	cache := NewTokenCache()

	t.Run("Cache hit for token", func(t *testing.T) {
		token := &Token{
			IDToken: "token1",
			Scope:   []string{"function"},
		}

		cache.Set("token1", token)

		got, ok := cache.Get("token1")

		if !ok {
			t.Errorf("Want cache hit")
		}

		if !reflect.DeepEqual(token, got) {
			t.Errorf("Want cached token: %v, got: %v", token, got)
		}
	})

	t.Run("No cache hit for missing key", func(t *testing.T) {
		got, ok := cache.Get("token2")

		if ok {
			t.Errorf("Want cache miss, got: %v", got)
		}
	})

	t.Run("No cache hit for expired token", func(t *testing.T) {
		token := &Token{
			IDToken: "token3",
			Expiry:  time.Now().Add(time.Minute * -10),
			Scope:   []string{"function"},
		}

		cache.Set("token3", token)

		got, ok := cache.Get("token3")

		if ok {
			t.Errorf("Want cache miss, got: %v", got)
		}
	})
}
