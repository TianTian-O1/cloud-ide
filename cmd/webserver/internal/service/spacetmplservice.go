package service

import (
	"github.com/mangohow/cloud-ide/cmd/webserver/internal/caches"
	"github.com/mangohow/cloud-ide/cmd/webserver/internal/dao"
	"github.com/mangohow/cloud-ide/cmd/webserver/internal/model"
)

type SpaceTmplService struct {
	dao       *dao.SpaceTemplateDao
	tmplCache *caches.TmplCache
	specCache *caches.SpecCache
}

func NewSpaceTmplService() *SpaceTmplService {
	d := dao.NewSpaceTemplateDao()
	return &SpaceTmplService{
		dao:       d,
		tmplCache: caches.CacheFactory().TmplCache(d),
		specCache: caches.CacheFactory().SpecCache(d),
	}
}

func (s *SpaceTmplService) GetAllUsingTmpl() ([]*model.SpaceTemplate, []*model.TmplKind, error) {
	tmpls := s.tmplCache.GetAllTmpl()
	kinds := s.tmplCache.GetAllKinds()
	for i := 0; i < len(tmpls); i++ {
		tmpls[i].Image = ""
	}
	return tmpls, kinds, nil
}

func (s *SpaceTmplService) GetAllSpec() ([]*model.SpaceSpec, error) {
	return s.specCache.GetAll(), nil
}
