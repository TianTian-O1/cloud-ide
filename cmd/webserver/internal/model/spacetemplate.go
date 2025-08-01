package model

import "time"

// SpaceTemplate 云开发空间模板
type SpaceTemplate struct {
	Id         uint32    `json:"id" db:"id"`
	KindId     uint32    `json:"kind_id" db:"kind_id"` // 类别Id
	Name       string    `json:"name" db:"name"`       // 空间模板名称
	Desc       string    `json:"desc" db:"desc"`       // 描述
	Tags       string    `json:"tags" db:"tags"`       // 标签，使用|隔开
	Image      string    `json:"image" db:"image"`     // 镜像
	Status     uint32    `json:"status" db:"status"`   // 0可用 1 已删除
	Avatar     string    `json:"avatar" db:"avatar"`
	CreateTime time.Time `json:"create_time" db:"create_time"`
	DeleteTime time.Time `json:"delete_time" db:"delete_time"`
}

type TmplKind struct {
	Id   uint32 `json:"id" db:"id"`
	Name string `json:"name" db:"name"`
}

// Space的Status
const (
	SpaceStatusDeleted = iota
	SpaceStatusAvailable
	SpaceStatusUncreated
)

const (
	RunningStatusStop = iota
	RunningStatusRunning
)

// Space 用户根据模板创建的空间
type Space struct {
	Id            uint32        `json:"id" db:"id"`
	UserId        uint32        `json:"user_id" db:"user_id"` // 所属用户的id
	TmplId        uint32        `json:"tmpl_id" db:"tmpl_id"` // 模板的id
	SpecId        uint32        `json:"spec_id" db:"spec_id"` // 规格id
	Spec          SpaceSpec     `json:"spec"`
	Sid           string        `json:"sid" db:"sid"`   // 工作空间Id，用于访问时的url中
	Name          string        `json:"name" db:"name"` // 名称
	Status        uint32        `json:"-" db:"status"`  // 0 已删除  1 可用 2 未创建
	RunningStatus uint32        `json:"running_status"` // 0 停止  1 正在运行
	GitRepository string        `json:"git_repository" db:"git_repository"`
	CreateTime    time.Time     `json:"create_time" db:"create_time"`
	DeleteTime    time.Time     `json:"delete_time" db:"delete_time"`
	StopTime      time.Time     `json:"stop_time" db:"stop_time"`   // 停止时间
	TotalTime     time.Duration `json:"total_time" db:"total_time"` // 总运行时间
	Environment   string        `json:"environment"`
	Avatar        string        `json:"avatar"`
}

// SpaceSpec 云空间的配置
type SpaceSpec struct {
	Id          uint32 `json:"id" db:"id"`
	CpuSpec     string `json:"cpu_spec" db:"cpu_spec"`         // CPU规格
	MemSpec     string `json:"mem_spec" db:"mem_spec"`         // 内存规格
	StorageSpec string `json:"storage_spec" db:"storage_spec"` // 存储规格
	Name        string `json:"name" db:"name"`
	Desc        string `json:"desc" db:"desc"`
}
