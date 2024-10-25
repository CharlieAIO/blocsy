package cache

import (
	"defi-intel/internal/types"
	"time"

	"github.com/patrickmn/go-cache"
)

type CacheEntry struct {
	Token       *types.Token
	Pair        *types.Pair
	LastUpdated time.Time
}

type Cache struct {
	cache *cache.Cache
}

func NewCache() *Cache {
	return &Cache{
		cache: cache.New(20*time.Second, 40*time.Second),
	}
}

func (c *Cache) PutToken(tokenAddress string, tokenData types.Token) {
	c.cache.Set(tokenAddress, tokenData, cache.DefaultExpiration)
}

func (c *Cache) GetToken(tokenAddress string) (*types.Token, bool) {
	data, exists := c.cache.Get(tokenAddress)
	if !exists {
		return nil, false
	}

	token, ok := data.(types.Token)
	if !ok {
		return nil, false
	}

	return &token, true
}

func (c *Cache) PutPair(pairAddress string, pairData types.Pair) {
	c.cache.Set(pairAddress, pairData, cache.DefaultExpiration)
}

func (c *Cache) GetPair(pairAddress string) (*types.Pair, bool) {
	data, exists := c.cache.Get(pairAddress)
	if !exists {
		return nil, false
	}

	pair, ok := data.(types.Pair)
	if !ok {
		return nil, false
	}

	return &pair, true
}
