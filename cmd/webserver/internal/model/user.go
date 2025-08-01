package model

import "time"

type User struct {
	Id              uint32     `json:"id" db:"id"`
	Uid             string     `json:"uid" db:"uid"`
	Username        string     `json:"username" db:"username"`
	Password        string     `json:"password" db:"password"`
	Nickname        string     `json:"nickname" db:"nickname"`
	Email           string     `json:"email" db:"email"`
	Phone           string     `json:"phone" db:"phone"`
	Avatar          string     `json:"avatar" db:"avatar"`
	CreateTime      time.Time  `json:"create_time" db:"create_time"`
	DeleteTime      time.Time  `json:"delete_time" db:"delete_time"`
	Status          uint32     `json:"status" db:"status"`               // 状态 0正常 1已删除
	VipStatus       uint8      `json:"vip_status" db:"vip_status"`       // VIP状态 0普通用户 1VIP用户
	VipExpireTime   *time.Time `json:"vip_expire_time" db:"vip_expire_time"` // VIP到期时间
	LinuxDoID       *int       `json:"linuxdo_id,omitempty" db:"linuxdo_id"`           // LinuxDo用户ID
	LinuxDoUsername *string    `json:"linuxdo_username,omitempty" db:"linuxdo_username"` // LinuxDo用户名

	Token string `json:"token"`
}

type RegisterInfo struct {
	Nickname  string `json:"nickname"`
	Username  string `json:"username"`
	Password  string `json:"password"`
	Email     string `json:"email"`
	EmailCode string `json:"emailCode"`
}
