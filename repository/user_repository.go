/*
实现用户数据访问逻辑。
*/
package repository

import (
	"blog/model"
	"database/sql"
	"errors"
)

const defaultUserPermission = model.PermissionUser

// UserRepository 负责用户数据读写。
type UserRepository struct {
	db *sql.DB
}

// NewUserRepository 创建用户仓储。
func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

// Exists 判断用户名是否已存在。
func (r *UserRepository) Exists(username string) (bool, error) {
	var storedUsername string
	err := r.db.QueryRow("SELECT username FROM users WHERE username=? AND deleted_at IS NULL", username).Scan(&storedUsername)
	if err == nil {
		return true, nil
	}

	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	return false, err
}

// Create 创建新用户。
func (r *UserRepository) Create(username, hashedPassword string) error {
	_, err := r.db.Exec(
		"INSERT INTO users (username, display_name, email, password_hash, permission) VALUES (?, ?, NULL, ?, ?)",
		username,
		username,
		hashedPassword,
		defaultUserPermission,
	)
	return err
}

// GetPasswordByUsername 按用户名读取密码哈希。
func (r *UserRepository) GetPasswordByUsername(username string) (string, error) {
	var hashedPassword string
	err := r.db.QueryRow("SELECT password_hash FROM users WHERE username=? AND deleted_at IS NULL", username).Scan(&hashedPassword)
	if err != nil {
		return "", err
	}
	return hashedPassword, nil
}

// GetByUsername 按用户名读取用户信息。
func (r *UserRepository) GetByUsername(username string) (model.User, error) {
	var user model.User
	err := r.db.QueryRow(
		"SELECT id, username, COALESCE(display_name, ''), COALESCE(avatar_url, ''), COALESCE(permission, ''), COALESCE(status, '') FROM users WHERE username=? AND deleted_at IS NULL",
		username,
	).Scan(&user.ID, &user.Username, &user.DisplayName, &user.Image, &user.Permission, &user.Status)
	if err != nil {
		return model.User{}, err
	}

	return user, nil
}

// UpdateProfile 更新显示名和头像。
func (r *UserRepository) UpdateProfile(username, displayName, image string) error {
	_, err := r.db.Exec(
		"UPDATE users SET display_name=?, avatar_url=? WHERE username=? AND deleted_at IS NULL",
		displayName,
		image,
		username,
	)
	return err
}

// UpdateImage 更新头像路径。
func (r *UserRepository) UpdateImage(username, image string) error {
	_, err := r.db.Exec(
		"UPDATE users SET avatar_url=? WHERE username=? AND deleted_at IS NULL",
		image,
		username,
	)
	return err
}

// UpdatePassword 更新密码哈希。
func (r *UserRepository) UpdatePassword(username, hashedPassword string) error {
	_, err := r.db.Exec(
		"UPDATE users SET password_hash=? WHERE username=? AND deleted_at IS NULL",
		hashedPassword,
		username,
	)
	return err
}

// UpdatePermission 更新用户权限。
func (r *UserRepository) UpdatePermission(username, permission string) error {
	_, err := r.db.Exec(
		"UPDATE users SET permission=? WHERE username=? AND deleted_at IS NULL",
		permission,
		username,
	)
	return err
}

// CountByPermission 统计指定权限的用户数量。
func (r *UserRepository) CountByPermission(permission string) (int, error) {
	var count int
	err := r.db.QueryRow("SELECT COUNT(*) FROM users WHERE permission=? AND deleted_at IS NULL", permission).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

// ListUsers 分页查询用户列表。
func (r *UserRepository) ListUsers(limit, offset int) ([]model.UserListItem, error) {
	rows, err := r.db.Query(`
		SELECT username, COALESCE(display_name, ''), COALESCE(permission, '')
		FROM users
		WHERE deleted_at IS NULL
		ORDER BY id DESC
		LIMIT ? OFFSET ?
	`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []model.UserListItem
	for rows.Next() {
		var user model.UserListItem
		if err := rows.Scan(&user.Username, &user.DisplayName, &user.Permission); err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

// CountUsers 统计用户总数。
func (r *UserRepository) CountUsers() (int, error) {
	var count int
	err := r.db.QueryRow("SELECT COUNT(*) FROM users WHERE deleted_at IS NULL").Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

// DeleteUser 软删除用户。
func (r *UserRepository) DeleteUser(username string) error {
	_, err := r.db.Exec("UPDATE users SET deleted_at=NOW(), status='deleted' WHERE username=? AND deleted_at IS NULL", username)
	return err
}
