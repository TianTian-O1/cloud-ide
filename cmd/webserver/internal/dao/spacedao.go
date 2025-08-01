package dao

import (
	"github.com/jmoiron/sqlx"
	"github.com/mangohow/cloud-ide/cmd/webserver/internal/dao/db"
	"github.com/mangohow/cloud-ide/cmd/webserver/internal/model"
)

type SpaceDao struct {
	db *sqlx.DB
}

func NewSpaceDao() *SpaceDao {
	return &SpaceDao{
		db: db.DB(),
	}
}

func (d *SpaceDao) Insert(space *model.Space) (uint32, error) {
	sql := `INSERT INTO t_space 
(user_id, tmpl_id, spec_id, sid, name, status, create_time, delete_time, stop_time, total_time, git_repository)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	res, err := d.db.Exec(sql, space.UserId, space.TmplId, space.SpecId, space.Sid, space.Name,
		space.Status, space.CreateTime, space.DeleteTime, space.StopTime, space.TotalTime, space.GitRepository)
	if err != nil {
		return 0, err
	}
	id, err := res.LastInsertId()

	return uint32(id), err
}

// FindByUserIdAndName TODO 增加联合索引 idx_userid_name
// 根据userid和name查询, 用于查询某个用户下的space名称是否重复
func (d *SpaceDao) FindByUserIdAndName(userId uint32, name string) error {
	sql := `SELECT id FROM t_space WHERE user_id = ? AND name = ? AND status != ?`
	var id uint32
	return d.db.Get(&id, sql, userId, name, model.SpaceStatusDeleted)
}

func (d *SpaceDao) FindCountByUserId(userId uint32) (count uint32, err error) {
	sql := `SELECT COUNT(*) FROM t_space WHERE user_id = ? AND status != ?`
	err = d.db.Get(&count, sql, userId, model.SpaceStatusDeleted)

	return
}

func (d *SpaceDao) FindAllSpaceByUserId(userId uint32) (spaces []model.Space, err error) {
	sql := `SELECT id, user_id, tmpl_id, spec_id, sid, name, create_time, stop_time, total_time, git_repository FROM t_space WHERE status != ? AND user_id = ?`
	err = d.db.Select(&spaces, sql, model.SpaceStatusDeleted, userId)
	return
}

func (d *SpaceDao) DeleteSpaceById(id uint32) error {
	// 不真正的删除，给其状态设置为已删除，待以后再删除
	sql := `UPDATE t_space SET status = ? WHERE id = ?`
	_, err := d.db.Exec(sql, model.SpaceStatusDeleted, id)

	return err
}

func (d *SpaceDao) FindByIdAndUserId(id, userId uint32) (space *model.Space, err error) {
	sql := `SELECT tmpl_id, spec_id, sid, name, status, git_repository FROM t_space WHERE id = ? AND user_id = ?;`
	space = &model.Space{}
	err = d.db.Get(space, sql, id, userId)
	return
}

func (d *SpaceDao) UpdateStatusById(id, status uint32) error {
	sql := `UPDATE t_space SET status = ? WHERE id = ?`
	_, err := d.db.Exec(sql, status, id)
	return err
}

func (d *SpaceDao) UpdateNameById(name string, id uint32) error {
	sql := `UPDATE t_space SET name = ? WHERE id = ?`
	_, err := d.db.Exec(sql, name, id)
	return err
}
