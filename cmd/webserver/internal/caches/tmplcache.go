package caches

import (
	"time"

	"github.com/mangohow/cloud-ide/cmd/webserver/internal/dao"
	"github.com/mangohow/cloud-ide/cmd/webserver/internal/model"
	"github.com/mangohow/cloud-ide/pkg/cache"
)

// TmplCache 加载mysql中的SpaceTemplate到内存中，数据量不大
type TmplCache struct {
	cache *cache.Cache
	dao   *dao.SpaceTemplateDao
}

func newTmplCache(dao *dao.SpaceTemplateDao) *TmplCache {
	t := &TmplCache{
		cache: cache.New("spaceTmpl"),
		dao:   dao,
	}

	// 每隔一分钟刷新一次缓存
	go func() {
		ticker := time.NewTicker(time.Minute)
		defer ticker.Stop()
		for {
			<-ticker.C
			t.LoadCache()
		}
	}()

	return t
}

const (
	TmplsKey = "tmpls"
	KindsKey = "kinds"
)

func (t *TmplCache) LoadCache() {
	tmpls, err := t.dao.GetAllUsingTmpl()
	if err != nil {
		panic(err)
	}

	tpls := make(map[uint32]*model.SpaceTemplate, len(tmpls))

	for i := 0; i < len(tmpls); i++ {
		tp := tmpls[i]
		tpls[tp.Id] = &tp
	}
	t.cache.Set(TmplsKey, tpls)

	kinds, err := t.dao.GetAllTmplKind()
	if err != nil {
		panic(err)
	}
	kds := make(map[uint32]*model.TmplKind, len(kinds))
	for i := 0; i < len(kinds); i++ {
		kd := kinds[i]
		kds[kd.Id] = &kd
	}
	t.cache.Set(KindsKey, kds)
}

func (t *TmplCache) GetTmpl(key uint32) *model.SpaceTemplate {
	item, ok := t.cache.Get(TmplsKey)
	if !ok {
		return nil
	}

	tps, ok := item.(map[uint32]*model.SpaceTemplate)
	if !ok {
		return nil
	}

	tmpl := *(tps[key])
	return &tmpl
}

func (t *TmplCache) GetAllTmpl() []*model.SpaceTemplate {
	get, ok := t.cache.Get(TmplsKey)
	if !ok {
		return nil
	}

	items := get.(map[uint32]*model.SpaceTemplate)

	// 拷贝一份
	tmpls := make([]*model.SpaceTemplate, len(items))
	i := 0
	for _, v := range items {
		p := *v
		tmpls[i] = &p
		i++
	}

	return tmpls
}

func (t *TmplCache) GetAllKinds() []*model.TmplKind {
	get, ok := t.cache.Get(KindsKey)
	if !ok {
		return nil
	}
	items := get.(map[uint32]*model.TmplKind)

	kinds := make([]*model.TmplKind, len(items))
	i := 0
	for _, v := range items {
		k := *v
		kinds[i] = &k
	}

	return kinds
}
