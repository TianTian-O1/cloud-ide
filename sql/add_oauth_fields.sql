-- 添加OAuth相关字段到用户表
-- 这个脚本用于为现有的t_user表添加LinuxDo OAuth支持

-- 添加LinuxDo ID字段（可为空，因为现有用户可能没有LinuxDo账号）
ALTER TABLE t_user 
ADD COLUMN linuxdo_id INT NULL COMMENT 'LinuxDo用户ID',
ADD COLUMN linuxdo_username VARCHAR(100) NULL COMMENT 'LinuxDo用户名';

-- 为LinuxDo ID添加唯一索引，但允许NULL值
CREATE UNIQUE INDEX idx_linuxdo_id ON t_user (linuxdo_id);

-- 为LinuxDo用户名添加索引
CREATE INDEX idx_linuxdo_username ON t_user (linuxdo_username);

-- 显示表结构确认更改
DESCRIBE t_user;