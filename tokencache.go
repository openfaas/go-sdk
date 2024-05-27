package sdk

import "sync"

type TokenCache struct {
	tokens map[string]*Token

	lock sync.RWMutex
}

func NewTokenCache() *TokenCache {
	return &TokenCache{
		tokens: map[string]*Token{},
	}
}

func (c *TokenCache) Set(key string, token *Token) {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.tokens[key] = token
}

func (c *TokenCache) Get(key string) (*Token, bool) {
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
