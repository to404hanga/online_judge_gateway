CREATE DATABASE IF NOT EXISTS online_judge;

USE online_judge; 

CREATE TABLE IF NOT EXISTS user (
    id bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '用户ID',
    username varchar(50) NOT NULL COMMENT '用户名(学号)',
    realname varchar(50) NOT NULL COMMENT '真实姓名',
    password varchar(255) NOT NULL COMMENT '密码',
    role tinyint NOT NULL DEFAULT 0 COMMENT '角色(0:普通用户,1:管理员)',
    status tinyint NOT NULL DEFAULT 0 COMMENT '状态(0:正常,1:禁用)',
    created_at datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    PRIMARY KEY (id),
    UNIQUE KEY uk_username (username)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci COMMENT = '用户表';

INSERT INTO user (username, realname, password, role, status)
VALUES
    ('admin1234567890123', 'to404hanga', '$2a$10$AAIzgJi/SZjPKOdJ2hliF./nzymKYUHaAlqbl0ugRy72m4cF2n1Pi', 1, 0), -- 密码 123456
    ('1234567890123', 'to404hanga', '$2a$10$AAIzgJi/SZjPKOdJ2hliF./nzymKYUHaAlqbl0ugRy72m4cF2n1Pi', 0, 0); -- 密码 123456
