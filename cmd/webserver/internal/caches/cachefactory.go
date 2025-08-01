package caches

import (
	"reflect"
	"sync"

	"github.com/mangohow/cloud-ide/cmd/webserver/internal/dao"
)

type cacheFactory struct {
	caches map[reflect.Type]interface{}
	lock   sync.Mutex
}

var (
	factory = &cacheFactory{caches: make(map[reflect.Type]interface{})}
)

func CacheFactory() *cacheFactory {
	return factory
}

func (f *cacheFactory) TmplCache(dao *dao.SpaceTemplateDao) *TmplCache {
	t := reflect.TypeOf(&TmplCache{})
	f.lock.Lock()
	defer f.lock.Unlock()
	if c, ok := f.caches[t]; ok {
		return c.(*TmplCache)
	}

	cache := newTmplCache(dao)
	f.caches[t] = cache

	cache.LoadCache()

	return cache
}

func (f *cacheFactory) SpecCache(dao *dao.SpaceTemplateDao) *SpecCache {
	t := reflect.TypeOf(&SpecCache{})
	f.lock.Lock()
	defer f.lock.Unlock()
	if c, ok := f.caches[t]; ok {
		return c.(*SpecCache)
	}

	cache := newSpecCache(dao)
	f.caches[t] = cache
	cache.LoadCache()

	return cache
}
