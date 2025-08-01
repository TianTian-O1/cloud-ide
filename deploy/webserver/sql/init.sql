/*
 Navicat MySQL Data Transfer

 Source Server         : cloud-ide-mysql
 Source Server Type    : MySQL
 Source Server Version : 80027
 Source Host           : 192.168.44.100:30306
 Source Schema         : cloudcode

 Target Server Type    : MySQL
 Target Server Version : 80027
 File Encoding         : 65001

 Date: 30/10/2023 22:50:15
*/

CREATE DATABASE cloudide CHARACTER SET utf8mb4;
USE cloudide;


SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

-- ----------------------------
-- Table structure for t_space
-- ----------------------------
DROP TABLE IF EXISTS `t_space`;
CREATE TABLE `t_space`  (
  `id` int(0) UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键id',
  `user_id` int(0) UNSIGNED NOT NULL COMMENT '用户id',
  `tmpl_id` int(0) UNSIGNED NOT NULL COMMENT '模板id',
  `spec_id` int(0) UNSIGNED NOT NULL COMMENT '空间规格id',
  `sid` char(24) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT 'space id',
  `name` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '空间名称',
  `git_repository` varchar(128) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '要克隆的git仓库',
  `status` int(0) NOT NULL COMMENT '空间状态 0 已删除 1 可用 2 未创建',
  `create_time` datetime(0) NOT NULL COMMENT '创建时间',
  `delete_time` datetime(0) NOT NULL COMMENT '删除时间',
  `stop_time` datetime(0) NOT NULL COMMENT '停止时间',
  `total_time` bigint(0) NOT NULL COMMENT '总运行时间',
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE INDEX `idx_id_user_id`(`id`, `user_id`) USING BTREE COMMENT '空间id和用户id联合索引'
) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of t_space
-- ----------------------------

-- ----------------------------
-- Table structure for t_space_template
-- ----------------------------
DROP TABLE IF EXISTS `t_space_template`;
CREATE TABLE `t_space_template`  (
  `id` int(0) UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键id',
  `kind_id` int(0) UNSIGNED NOT NULL COMMENT '类别id',
  `name` varchar(128) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '空间模板名称',
  `desc` varchar(256) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '描述',
  `tags` varchar(128) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '标签，使用|隔开',
  `image` varchar(128) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '镜像名称',
  `status` int(0) NOT NULL DEFAULT 0 COMMENT '状态 0可用 1已删除',
  `avatar` varchar(128) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT '' COMMENT '头像',
  `create_time` datetime(0) NOT NULL COMMENT '创建时间',
  `delete_time` datetime(0) NOT NULL COMMENT '删除时间',
  PRIMARY KEY (`id`) USING BTREE
) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of t_space_template
-- ----------------------------
INSERT INTO `t_space_template` VALUES (1, 1, 'Go', 'go workspace with go 1.21.3, make', 'Go,Make,Git', 'registry.cn-hangzhou.aliyuncs.com/k8s-cloud-ide/code-server-go:v1.21', 0, 'images/go.png', '2022-12-08 16:53:45', '2022-12-08 16:53:47');
INSERT INTO `t_space_template` VALUES (2, 1, 'Node.js', 'js workspace', 'Node.js', 'node.js', 0, 'images/nodejs.png', '2022-12-11 21:18:22', '2022-12-11 21:18:24');
INSERT INTO `t_space_template` VALUES (3, 1, 'C/C++', 'c/c++ workspace with gcc g++ make cmake git', 'C,CPP,Make,Git', 'registry.cn-hangzhou.aliyuncs.com/k8s-cloud-ide/code-server-cxx:v1.0', 0, 'images/cpp.png', '2022-12-11 22:40:28', '2022-12-11 22:40:30');
INSERT INTO `t_space_template` VALUES (4, 1, 'Java', 'java workspace', 'Java', 'java', 0, 'images/java.png', '2023-02-26 16:56:43', '2023-02-26 16:57:33');
INSERT INTO `t_space_template` VALUES (5, 1, 'Vue', 'Vue workspace', 'Vue,Yarn', 'Vue', 0, 'images/vue.png', '2023-02-26 17:05:18', '2023-02-26 17:05:20');
INSERT INTO `t_space_template` VALUES (6, 1, 'Python', 'python workspace', 'Python', 'Python', 0, 'images/python.png', '2023-02-26 17:05:45', '2023-02-26 17:05:48');

-- ----------------------------
-- Table structure for t_spacespec
-- ----------------------------
DROP TABLE IF EXISTS `t_spacespec`;
CREATE TABLE `t_spacespec`  (
  `id` int(0) UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键id',
  `cpu_spec` varchar(16) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT 'cpu规格',
  `mem_spec` varchar(16) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '内存规格',
  `storage_spec` varchar(16) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '存储规格',
  `name` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '名称',
  `desc` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '描述',
  PRIMARY KEY (`id`) USING BTREE
) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of t_spacespec
-- ----------------------------
INSERT INTO `t_spacespec` VALUES (1, '2', '4Gi', '8Gi', '标准型', '标准型 2CPU 4GB / 8GB存储 ');
INSERT INTO `t_spacespec` VALUES (2, '4', '8Gi', '16Gi', '增强型', '计算型 4CPU 4GB内存 / 16GB存储');
INSERT INTO `t_spacespec` VALUES (3, '8', '16Gi', '32Gi', '专业型', '专业型 8CPU 16GB内存 / 32GB存储');
INSERT INTO `t_spacespec` VALUES (4, '2', '2Gi', '4Gi', '测试型', '测试型 2CPU 2GB内存 / 4GB存储');

-- ----------------------------
-- Table structure for t_template_kind
-- ----------------------------
DROP TABLE IF EXISTS `t_template_kind`;
CREATE TABLE `t_template_kind`  (
  `id` int(0) UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键id',
  `name` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '类别名称',
  PRIMARY KEY (`id`) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 2 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of t_template_kind
-- ----------------------------
INSERT INTO `t_template_kind` VALUES (1, '编程语言');

-- ----------------------------
-- Table structure for t_user
-- ----------------------------
DROP TABLE IF EXISTS `t_user`;
CREATE TABLE `t_user`  (
  `id` int(0) UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键Id',
  `uid` char(24) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT 'uid,使用mongodb的_id方式生成',
  `username` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '用户名',
  `password` varchar(48) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '密码',
  `nickname` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '昵称',
  `email` varchar(128) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '邮箱',
  `phone` char(11) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT '' COMMENT '手机号',
  `avatar` varchar(128) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT '' COMMENT '头像',
  `create_time` datetime(0) NOT NULL COMMENT '创建时间',
  `delete_time` datetime(0) NOT NULL COMMENT '删除时间',
  `status` int(0) NOT NULL COMMENT '状态 0 可用 1 已注销',
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE INDEX `idx_username`(`username`) USING BTREE COMMENT '用户名索引',
  UNIQUE INDEX `idx_email`(`email`) USING BTREE COMMENT '邮箱索引',
  UNIQUE INDEX `idx_uid`(`uid`) USING BTREE COMMENT 'uid索引'
) ENGINE = InnoDB AUTO_INCREMENT = 8 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of t_user
-- ----------------------------

SET FOREIGN_KEY_CHECKS = 1;
