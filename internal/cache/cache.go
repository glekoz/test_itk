package cache

import (
	"time"

	"github.com/glekoz/cache"
)

type Cache struct {
	c   *cache.Cache[string, int]
	ttl time.Duration
}

func New(ttl int) (*Cache, error) {
	c, err := cache.New[string, int]()
	if err != nil {
		return nil, err
	}
	return &Cache{
		c:   c,
		ttl: time.Duration(ttl) * time.Second,
	}, nil
}

func (c *Cache) Add(walletID string, balance int) error {
	return c.c.Add(walletID, balance, c.ttl)
}

func (c *Cache) Get(walletID string) (int, bool) {
	return c.c.Get(walletID)
}

func (c *Cache) Delete(walletID string) {
	c.c.Delete(walletID)
}
