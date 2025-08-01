package caches

import (
	"strconv"
	"time"

	"github.com/mangohow/cloud-ide/cmd/webserver/internal/dao"
	"github.com/mangohow/cloud-ide/cmd/webserver/internal/model"
	"github.com/mangohow/cloud-ide/pkg/cache"
)

// 加载mysql中的SpaceSpec到内存中，数据量不大

type SpecCache struct {
	cache *cache.Cache
	dao   *dao.SpaceTemplateDao
}

func newSpecCache(dao *dao.SpaceTemplateDao) *SpecCache {
	s := &SpecCache{
		cache: cache.New("space-spec"),
		dao:   dao,
	}

	// 每隔一分钟刷新一次
	go func() {
		ticker := time.NewTicker(time.Minute)
		defer ticker.Stop()
		for {
			<-ticker.C
			s.refresh()
		}
	}()

	return s
}

func (c *SpecCache) refresh() {
	specs, err := c.dao.GetAllSpec()
	if err != nil {
		return
	}
	m := make(map[string]interface{}, len(specs))
	for i, _ := range specs {
		m[strconv.Itoa(int(specs[i].Id))] = &specs[i]
	}
	c.cache.Replace(m)
}

func (c *SpecCache) LoadCache() {
	specs, err := c.dao.GetAllSpec()
	if err != nil {
		panic(err)
	}

	// 巨坑
	for i, _ := range specs {
		spec := specs[i]
		c.cache.Set(strconv.Itoa(int(spec.Id)), &spec)
	}
}

func (c *SpecCache) Get(key uint32) *model.SpaceSpec {
	item, ok := c.cache.GetByInt(int(key))
	if !ok {
		return nil
	}
	// 复制一份，防止外面把缓存数据给修改了
	s := item.(*model.SpaceSpec)
	spec := *s
	return &spec
}

func (c *SpecCache) GetAll() []*model.SpaceSpec {
	all := c.cache.GetAll()
	items := make([]*model.SpaceSpec, len(all))
	for i := 0; i < len(all); i++ {
		item := all[i].(*model.SpaceSpec)
		s := *item
		items[i] = &s
	}

	return items
}
