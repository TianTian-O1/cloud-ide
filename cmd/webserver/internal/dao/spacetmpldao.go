package dao

import (
	"github.com/jmoiron/sqlx"
	"github.com/mangohow/cloud-ide/cmd/webserver/internal/dao/db"
	"github.com/mangohow/cloud-ide/cmd/webserver/internal/model"
)

type SpaceTemplateDao struct {
	db *sqlx.DB
}

func NewSpaceTemplateDao() *SpaceTemplateDao {
	return &SpaceTemplateDao{
		db: db.DB(),
	}
}

const (
	TmplUsing = iota
	TmplDeleted
)

func (s *SpaceTemplateDao) GetAllTmplKind() (kinds []model.TmplKind, err error) {
	sql := `SELECT id, name FROM t_template_kind`
	err = s.db.Select(&kinds, sql)
	return
}

func (s *SpaceTemplateDao) GetAllUsingTmpl() (tmpls []model.SpaceTemplate, err error) {
	sql := "SELECT id, kind_id, name, `desc`, tags, image, avatar FROM t_space_template WHERE status = ?"
	err = s.db.Select(&tmpls, sql, TmplUsing)

	return
}

func (s *SpaceTemplateDao) GetAllTmpl() (tmpls []model.SpaceTemplate, err error) {
	sql := "SELECT id, kind_id, name, `desc`, tags, image, avatar FROM t_space_template"
	err = s.db.Select(&tmpls, sql)

	return
}

func (s *SpaceTemplateDao) GetAllSpec() (specs []model.SpaceSpec, err error) {
	sql := "SELECT id, cpu_spec, mem_spec, storage_spec, name, `desc` FROM t_spacespec"
	err = s.db.Select(&specs, sql)

	return
}
