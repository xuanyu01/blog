-- blog 数据库完整初始化脚本。
-- 一次性执行完整建库、建表和初始化数据。

-- =========================================================
-- 第 001 段：创建数据库
-- =========================================================
-- 创建 blog 数据库并切换到该库。
CREATE DATABASE IF NOT EXISTS blog
  DEFAULT CHARACTER SET utf8mb4
  DEFAULT COLLATE utf8mb4_unicode_ci;

USE blog;
SET NAMES utf8mb4;

-- =========================================================
-- 第 002 段：创建用户相关表
-- =========================================================
USE blog;
SET NAMES utf8mb4;

CREATE TABLE IF NOT EXISTS users (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    username VARCHAR(64) NOT NULL,
    display_name VARCHAR(128) NOT NULL DEFAULT '',
    email VARCHAR(191) NULL DEFAULT NULL,
    password_hash VARCHAR(255) NOT NULL,
    avatar_url VARCHAR(255) NOT NULL DEFAULT '',
    permission VARCHAR(32) NOT NULL DEFAULT 'user',
    status VARCHAR(32) NOT NULL DEFAULT 'active',
    must_change_password TINYINT(1) NOT NULL DEFAULT 0,
    bio VARCHAR(500) NOT NULL DEFAULT '',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at DATETIME NULL DEFAULT NULL,
    PRIMARY KEY (id),
    UNIQUE KEY uk_users_username (username),
    UNIQUE KEY uk_users_email (email),
    KEY idx_users_permission (permission),
    KEY idx_users_status (status),
    KEY idx_users_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- =========================================================
-- 第 003 段：创建分类和标签表
-- =========================================================
USE blog;
SET NAMES utf8mb4;

CREATE TABLE IF NOT EXISTS categories (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    parent_id BIGINT UNSIGNED NULL DEFAULT NULL,
    name VARCHAR(100) NOT NULL,
    slug VARCHAR(120) NOT NULL,
    sort_order INT NOT NULL DEFAULT 0,
    status VARCHAR(32) NOT NULL DEFAULT 'active',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    UNIQUE KEY uk_categories_slug (slug),
    UNIQUE KEY uk_categories_name (name),
    KEY idx_categories_parent_id (parent_id),
    KEY idx_categories_status_sort (status, sort_order, id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS tags (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    name VARCHAR(64) NOT NULL,
    slug VARCHAR(80) NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    UNIQUE KEY uk_tags_name (name),
    UNIQUE KEY uk_tags_slug (slug)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- =========================================================
-- 第 004 段：创建文章相关表
-- =========================================================
USE blog;
SET NAMES utf8mb4;

CREATE TABLE IF NOT EXISTS posts (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    author_id BIGINT UNSIGNED NOT NULL,
    category_id BIGINT UNSIGNED NULL DEFAULT NULL,
    title VARCHAR(200) NOT NULL,
    slug VARCHAR(220) NOT NULL,
    summary VARCHAR(500) NOT NULL DEFAULT '',
    cover_url VARCHAR(255) NOT NULL DEFAULT '',
    status VARCHAR(32) NOT NULL DEFAULT 'draft',
    visibility VARCHAR(32) NOT NULL DEFAULT 'public',
    is_top TINYINT(1) NOT NULL DEFAULT 0,
    allow_comment TINYINT(1) NOT NULL DEFAULT 1,
    published_at DATETIME NULL DEFAULT NULL,
    last_commented_at DATETIME NULL DEFAULT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at DATETIME NULL DEFAULT NULL,
    PRIMARY KEY (id),
    UNIQUE KEY uk_posts_slug (slug),
    KEY idx_posts_author_id (author_id),
    KEY idx_posts_category_id (category_id),
    KEY idx_posts_status_pub_id (status, published_at DESC, id DESC),
    KEY idx_posts_author_status_created (author_id, status, created_at DESC, id DESC),
    KEY idx_posts_visibility_status_pub (visibility, status, published_at DESC, id DESC),
    KEY idx_posts_top_status_pub (is_top, status, published_at DESC, id DESC),
    CONSTRAINT fk_posts_author_id FOREIGN KEY (author_id) REFERENCES users(id),
    CONSTRAINT fk_posts_category_id FOREIGN KEY (category_id) REFERENCES categories(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS post_contents (
    post_id BIGINT UNSIGNED NOT NULL,
    content_markdown MEDIUMTEXT NOT NULL,
    content_html MEDIUMTEXT NULL,
    content_text MEDIUMTEXT NULL,
    word_count INT NOT NULL DEFAULT 0,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (post_id),
    FULLTEXT KEY ftx_post_contents_text (content_markdown, content_text),
    CONSTRAINT fk_post_contents_post_id FOREIGN KEY (post_id) REFERENCES posts(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS post_stats (
    post_id BIGINT UNSIGNED NOT NULL,
    view_count BIGINT UNSIGNED NOT NULL DEFAULT 0,
    like_count BIGINT UNSIGNED NOT NULL DEFAULT 0,
    favorite_count BIGINT UNSIGNED NOT NULL DEFAULT 0,
    comment_count BIGINT UNSIGNED NOT NULL DEFAULT 0,
    share_count BIGINT UNSIGNED NOT NULL DEFAULT 0,
    score DECIMAL(12,4) NOT NULL DEFAULT 0,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (post_id),
    KEY idx_post_stats_score (score DESC),
    CONSTRAINT fk_post_stats_post_id FOREIGN KEY (post_id) REFERENCES posts(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS post_tags (
    post_id BIGINT UNSIGNED NOT NULL,
    tag_id BIGINT UNSIGNED NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (post_id, tag_id),
    KEY idx_post_tags_tag_id_post_id (tag_id, post_id),
    CONSTRAINT fk_post_tags_post_id FOREIGN KEY (post_id) REFERENCES posts(id) ON DELETE CASCADE,
    CONSTRAINT fk_post_tags_tag_id FOREIGN KEY (tag_id) REFERENCES tags(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- =========================================================
-- 第 005 段：创建评论表
-- =========================================================
USE blog;
SET NAMES utf8mb4;

CREATE TABLE IF NOT EXISTS comments (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    post_id BIGINT UNSIGNED NOT NULL,
    user_id BIGINT UNSIGNED NOT NULL,
    parent_id BIGINT UNSIGNED NULL DEFAULT NULL,
    root_id BIGINT UNSIGNED NULL DEFAULT NULL,
    content VARCHAR(2000) NOT NULL,
    status VARCHAR(32) NOT NULL DEFAULT 'published',
    like_count BIGINT UNSIGNED NOT NULL DEFAULT 0,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at DATETIME NULL DEFAULT NULL,
    PRIMARY KEY (id),
    KEY idx_comments_post_status_created (post_id, status, created_at ASC, id ASC),
    KEY idx_comments_user_id_created (user_id, created_at DESC, id DESC),
    KEY idx_comments_parent_id (parent_id),
    KEY idx_comments_root_id (root_id),
    CONSTRAINT fk_comments_post_id FOREIGN KEY (post_id) REFERENCES posts(id) ON DELETE CASCADE,
    CONSTRAINT fk_comments_user_id FOREIGN KEY (user_id) REFERENCES users(id),
    CONSTRAINT fk_comments_parent_id FOREIGN KEY (parent_id) REFERENCES comments(id) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- =========================================================
-- 第 006 段：创建互动统计表
-- =========================================================
USE blog;
SET NAMES utf8mb4;

CREATE TABLE IF NOT EXISTS post_likes (
    post_id BIGINT UNSIGNED NOT NULL,
    user_id BIGINT UNSIGNED NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (post_id, user_id),
    KEY idx_post_likes_user_id (user_id, created_at DESC),
    CONSTRAINT fk_post_likes_post_id FOREIGN KEY (post_id) REFERENCES posts(id) ON DELETE CASCADE,
    CONSTRAINT fk_post_likes_user_id FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS post_favorites (
    post_id BIGINT UNSIGNED NOT NULL,
    user_id BIGINT UNSIGNED NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (post_id, user_id),
    KEY idx_post_favorites_user_id (user_id, created_at DESC),
    CONSTRAINT fk_post_favorites_post_id FOREIGN KEY (post_id) REFERENCES posts(id) ON DELETE CASCADE,
    CONSTRAINT fk_post_favorites_user_id FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS user_follows (
    follower_id BIGINT UNSIGNED NOT NULL,
    followee_id BIGINT UNSIGNED NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (follower_id, followee_id),
    KEY idx_user_follows_followee_id (followee_id, follower_id),
    CONSTRAINT fk_user_follows_follower_id FOREIGN KEY (follower_id) REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT fk_user_follows_followee_id FOREIGN KEY (followee_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- =========================================================
-- 第 007 段：写入初始数据
-- =========================================================
USE blog;
SET NAMES utf8mb4;

INSERT INTO users (username, display_name, email, password_hash, permission, status, must_change_password)
VALUES ('admin', 'admin', 'admin@example.com', '$2a$10$yFAlJrwC4LVItEyQBnHid.EPU3Vf3qXLFPiHLenCGD0tKwPjWnoV6', 'admin', 'active', 1)
ON DUPLICATE KEY UPDATE
    display_name = VALUES(display_name),
    permission = VALUES(permission),
    status = VALUES(status),
    must_change_password = VALUES(must_change_password);

INSERT INTO categories (name, slug, sort_order, status)
VALUES ('默认分类', 'default', 0, 'active')
ON DUPLICATE KEY UPDATE
    sort_order = VALUES(sort_order),
    status = VALUES(status);

INSERT INTO tags (name, slug)
VALUES ('gin', 'gin'), ('go', 'go')
ON DUPLICATE KEY UPDATE
    slug = VALUES(slug);

-- =========================================================
-- 第 008 段：强制修改密码
-- =========================================================
USE blog;
SET NAMES utf8mb4;

ALTER TABLE users
    ADD COLUMN must_change_password TINYINT(1) NOT NULL DEFAULT 0 AFTER status;

UPDATE users
SET must_change_password = 1
WHERE username = 'admin'
  AND permission = 'admin'
  AND deleted_at IS NULL
  AND status <> 'deleted';
