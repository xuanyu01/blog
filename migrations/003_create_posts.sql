USE blog;

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
