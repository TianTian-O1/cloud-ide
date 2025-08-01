package dao

import (
	"errors"
	"strings"
	"github.com/jmoiron/sqlx"
	"github.com/mangohow/cloud-ide/cmd/webserver/internal/dao/db"
	"github.com/mangohow/cloud-ide/cmd/webserver/internal/model"
)

type UserDao struct {
	db *sqlx.DB
}

func NewUserDao() *UserDao {
	return &UserDao{
		db: db.DB(),
	}
}

func (u *UserDao) FindByUsernameDetailed(username string) (user *model.User, _ error) {
	sql := `SELECT id, uid, username, password, nickname, email, avatar, status, linuxdo_id, linuxdo_username FROM t_user WHERE username = ? AND delete_time > NOW()`
	user = &model.User{}
	err := u.db.Get(user, sql, username)
	return user, err
}

func (u *UserDao) FindByUsername(username string) error {
	sql := "SELECT 1 FROM t_user WHERE username = ?"
	var n int
	return u.db.Get(&n, sql, username)
}

func (u *UserDao) FindByEmail(email string) error {
	sql := `SELECT 1 FROM t_user WHERE email = ?`
	var n int
	return u.db.Get(&n, sql, email)
}

func (u *UserDao) FindByEmailDetailed(email string) (user *model.User, _ error) {
	sql := `SELECT id, uid, username, password, nickname, email, avatar, status, linuxdo_id, linuxdo_username FROM t_user WHERE email = ? AND delete_time > NOW()`
	user = &model.User{}
	err := u.db.Get(user, sql, email)
	return user, err
}

func (u *UserDao) AddUser(user *model.User) error {
	sql := `Insert into t_user (uid, username, password, nickname, email, create_time, delete_time, status, linuxdo_id, linuxdo_username) values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := u.db.Exec(sql, user.Uid, user.Username, user.Password, user.Nickname, user.Email, user.CreateTime, user.DeleteTime, user.Status, user.LinuxDoID, user.LinuxDoUsername)
	return err
}

// UpdatePasswordByEmail 根据邮箱更新用户密码
func (u *UserDao) UpdatePasswordByEmail(email, newPassword string) error {
	sql := `UPDATE t_user SET password = ? WHERE email = ? AND delete_time > NOW()`
	result, err := u.db.Exec(sql, newPassword, email)
	if err != nil {
		return err
	}
	
	// 检查是否实际更新了记录
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	
	if rowsAffected == 0 {
		return errors.New("no user found with the given email")
	}
	
	return nil
}

// FindByLinuxDoID 根据LinuxDo ID查找用户
func (u *UserDao) FindByLinuxDoID(linuxdoID int) (user *model.User, _ error) {
	sql := `SELECT id, uid, username, password, nickname, email, avatar, status, linuxdo_id, linuxdo_username FROM t_user WHERE linuxdo_id = ? AND delete_time > NOW()`
	user = &model.User{}
	err := u.db.Get(user, sql, linuxdoID)
	return user, err
}

// UpdateUser 更新用户信息
func (u *UserDao) UpdateUser(userID uint32, updates map[string]interface{}) error {
	if len(updates) == 0 {
		return nil
	}
	
	setParts := make([]string, 0, len(updates))
	args := make([]interface{}, 0, len(updates)+1)
	
	for field, value := range updates {
		setParts = append(setParts, field+" = ?")
		args = append(args, value)
	}
	
	args = append(args, userID)
	sql := `UPDATE t_user SET ` + strings.Join(setParts, ", ") + ` WHERE id = ?`
	
	_, err := u.db.Exec(sql, args...)
	return err
}
