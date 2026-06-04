-- MySQL 初始化脚本
-- 创建复制用户、应用数据库和示例表

-- 创建复制用户
CREATE USER IF NOT EXISTS 'repl'@'%' IDENTIFIED BY 'repl123';
GRANT REPLICATION SLAVE ON *.* TO 'repl'@'%';
FLUSH PRIVILEGES;

-- 使用应用数据库
USE app;

-- 创建示例表
CREATE TABLE IF NOT EXISTS users (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    username VARCHAR(64) NOT NULL UNIQUE,
    email VARCHAR(128) NOT NULL UNIQUE,
    password VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS conversations (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(128) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS messages (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    conversation_id BIGINT NOT NULL,
    sender_id BIGINT NOT NULL,
    content TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_conversation_created (conversation_id, created_at),
    FOREIGN KEY (conversation_id) REFERENCES conversations(id),
    FOREIGN KEY (sender_id) REFERENCES users(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 插入示例数据
INSERT INTO users (username, email, password) VALUES
    ('admin', 'admin@example.com', 'hashed_password'),
    ('user1', 'user1@example.com', 'hashed_password');

INSERT INTO conversations (name) VALUES ('General Chat');

SELECT 'MySQL 初始化完成' AS message;
