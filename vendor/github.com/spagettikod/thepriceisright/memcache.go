package thepriceisright

import (
	"sync"
	"time"
)

type MemCache struct {
	baseCache
	mu *sync.Mutex
}

func NewMemCache(areaCode string) *MemCache {
	return &MemCache{
		mu:        &sync.Mutex{},
		baseCache: baseCache{areaCode: areaCode},
	}
}

func (mc MemCache) AreaCode() string {
	return mc.areaCode
}

func (mc MemCache) Expired() bool {
	if !mc.pricesCache.IsValid() {
		return true
	}
	return mc.pricesCache.IsExpired(time.Now())
}

func (mc MemCache) TodaysPrices() TodaysPrices {
	return mc.pricesCache
}

func (mc *MemCache) Update() error {
	tp, err := fetch(mc.areaCode)
	if err != nil {
		return err
	}
	mc.mu.Lock()
	defer mc.mu.Unlock()
	mc.pricesCache = tp
	return nil
}
