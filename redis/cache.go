package redis

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

var keyReg = regexp.MustCompile("^(\\d+):")

var (
	ErrInvalidCacheKey = errors.New("invalid cache key")
)

type cacheItem struct {
	Expiration time.Time
	Value      string
}

type Cache struct {
	Redis *Redis
	store map[string]cacheItem
	mutex sync.RWMutex
}

func NewCache(connURL string) (*Cache, error) {
	r, err := NewRedis(connURL)
	if err != nil {
		return nil, err
	}
	return &Cache{
		Redis: r,
		store: make(map[string]cacheItem),
	}, nil
}

func (c *Cache) Close() {
	c.Redis.Close()
}

func (c *Cache) Get(key string) (string, bool, error) {
	c.mutex.RLock()
	x, ok := c.store[key]
	c.mutex.RUnlock()
	if ok {
		if !x.Expiration.IsZero() && x.Expiration.Before(time.Now()) {
			return "", false, nil
		}
		return x.Value, true, nil
	}

	data, ok, err := c.Redis.Get(key)
	if err != nil || !ok {
		return "", ok, err
	}

	// Parse the actual stored data from our custom format
	str := string(data)
	nanoStr := keyReg.FindString(str)
	if nanoStr == "" {
		return "", false, ErrInvalidCacheKey
	}
	nano, _ := strconv.ParseInt(nanoStr[:len(nanoStr)-1], 10, 64)
	var t time.Time
	if nano > 0 {
		t = time.Unix(nano, 0)
	}
	str = str[strings.Index(str, ":")+1:]
	c.mutex.Lock()
	c.store[key] = cacheItem{
		Expiration: t,
		Value:      str,
	}
	c.mutex.Unlock()

	return str, true, nil
}

func (c *Cache) Set(key, value string, expire time.Duration) error {
	future := time.Now().Add(expire)
	var ex int64
	if expire > 0 {
		ex = future.Unix()
	} else {
		expire = 0
	}
	// We store expiration before value in redis and parse it out later
	if err := c.Redis.Set(key, fmt.Sprintf("%d:%s", ex, value), expire); err != nil {
		return err
	}
	c.mutex.Lock()
	if c.store == nil {
		c.store = map[string]cacheItem{}
	}
	c.store[key] = cacheItem{
		Value:      value,
		Expiration: future,
	}
	c.mutex.Unlock()
	return nil
}

func (c *Cache) del(key string) {
	c.mutex.Lock()
	delete(c.store, key)
	c.mutex.Unlock()
}

func (c *Cache) Del(keys ...string) error {
	if err := c.Redis.Del(keys...); err != nil {
		return err
	}
	for _, key := range keys {
		c.del(key)
	}
	return nil
}

func (c *Cache) SearchRedis(match string) ([]string, error) {
	return c.Redis.ScanAll(match)
}

func (c *Cache) DeleteKeyMatch(match string) (int, error) {
	return c.Redis.DeleteKeyMatchFn(match, c.del)
}

func (c *Cache) WipeLocal() {
	c.mutex.Lock()
	c.store = map[string]cacheItem{}
	c.mutex.Unlock()
}
