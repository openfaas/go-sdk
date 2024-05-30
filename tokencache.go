package sdk

import "sync"

type TokenCache interface {
	Get(key string) (*Token, bool)
	Set(key string, token *Token)
}

type MemoryTokenCache struct {
	tokens map[string]*Token

	lock sync.RWMutex
}

func NewMemoryTokenCache() *MemoryTokenCache {
	return &MemoryTokenCache{
		tokens: map[string]*Token{},
	}
}

func (c *MemoryTokenCache) Set(key string, token *Token) {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.tokens[key] = token
}

func (c *MemoryTokenCache) Get(key string) (*Token, bool) {
	c.lock.RLock()
	token, ok := c.tokens[key]
	c.lock.RUnlock()

	if ok && token.Expired() {
		c.lock.Lock()
		delete(c.tokens, key)
		c.lock.Unlock()

		return nil, false
	}

	return token, ok
}
