USE blog;

INSERT INTO users (username, display_name, email, password_hash, permission, status)
VALUES ('admin', 'admin', 'admin@example.com', '$2a$10$yFAlJrwC4LVItEyQBnHid.EPU3Vf3qXLFPiHLenCGD0tKwPjWnoV6', 'admin', 'active')
ON DUPLICATE KEY UPDATE
    display_name = VALUES(display_name),
    permission = VALUES(permission),
    status = VALUES(status);

INSERT INTO categories (name, slug, sort_order, status)
VALUES ('默认分类', 'default', 0, 'active')
ON DUPLICATE KEY UPDATE
    sort_order = VALUES(sort_order),
    status = VALUES(status);

INSERT INTO tags (name, slug)
VALUES ('gin', 'gin'), ('go', 'go')
ON DUPLICATE KEY UPDATE
    slug = VALUES(slug);
