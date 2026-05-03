/*
user_repository.go 负责用户数据的查询和写入逻辑。
*/
package repository

import (
	"blog/model"
	"database/sql"
	"errors"

	"gorm.io/gorm"
)

const defaultUserPermission = model.PermissionUser

// UserRepository 负责用户数据读写。
type UserRepository struct {
	db *gorm.DB
}

// NewUserRepository 创建用户仓储。
func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

// Exists 判断用户名是否已存在。
func (r *UserRepository) Exists(username string) (bool, error) {
	var count int64
	err := r.db.Table("users").
		Where("username = ? AND deleted_at IS NULL", username).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// Create 创建新用户。
func (r *UserRepository) Create(username, hashedPassword string) error {
	return r.db.Exec(
		"INSERT INTO users (username, display_name, email, password_hash, permission) VALUES (?, ?, NULL, ?, ?)",
		username,
		username,
		hashedPassword,
		defaultUserPermission,
	).Error
}

// GetPasswordByUsername 按用户名读取密码哈希。
func (r *UserRepository) GetPasswordByUsername(username string) (string, error) {
	var hashedPassword string
	err := r.db.Table("users").
		Select("password_hash").
		Where("username = ? AND deleted_at IS NULL", username).
		Row().
		Scan(&hashedPassword)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", sql.ErrNoRows
		}
		return "", err
	}
	return hashedPassword, nil
}

// GetByUsername 按用户名读取用户信息。
func (r *UserRepository) GetByUsername(username string) (model.User, error) {
	var user model.User
	err := r.db.Raw(
		"SELECT id, username, COALESCE(display_name, ''), COALESCE(avatar_url, ''), COALESCE(permission, ''), COALESCE(status, '') FROM users WHERE username=? AND deleted_at IS NULL",
		username,
	).Row().Scan(&user.ID, &user.Username, &user.DisplayName, &user.Image, &user.Permission, &user.Status)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return model.User{}, sql.ErrNoRows
		}
		return model.User{}, err
	}

	return user, nil
}

// UpdateProfile 更新显示名和头像。
func (r *UserRepository) UpdateProfile(username, displayName, image string) error {
	return r.db.Exec(
		"UPDATE users SET display_name=?, avatar_url=? WHERE username=? AND deleted_at IS NULL",
		displayName,
		image,
		username,
	).Error
}

// UpdateImage 更新头像路径。
func (r *UserRepository) UpdateImage(username, image string) error {
	return r.db.Exec(
		"UPDATE users SET avatar_url=? WHERE username=? AND deleted_at IS NULL",
		image,
		username,
	).Error
}

// UpdatePassword 更新密码哈希。
func (r *UserRepository) UpdatePassword(username, hashedPassword string) error {
	return r.db.Exec(
		"UPDATE users SET password_hash=? WHERE username=? AND deleted_at IS NULL",
		hashedPassword,
		username,
	).Error
}

// UpdatePermission 更新用户权限。
func (r *UserRepository) UpdatePermission(username, permission string) error {
	return r.db.Exec(
		"UPDATE users SET permission=? WHERE username=? AND deleted_at IS NULL",
		permission,
		username,
	).Error
}

// CountByPermission 统计指定权限的用户数量。
func (r *UserRepository) CountByPermission(permission string) (int, error) {
	var count int64
	err := r.db.Table("users").
		Where("permission = ? AND deleted_at IS NULL", permission).
		Count(&count).Error
	if err != nil {
		return 0, err
	}
	return int(count), nil
}

// ListUsers 分页查询用户列表。
func (r *UserRepository) ListUsers(limit, offset int) ([]model.UserListItem, error) {
	rows, err := r.db.Raw(`
		SELECT username, COALESCE(display_name, ''), COALESCE(permission, '')
		FROM users
		WHERE deleted_at IS NULL
		ORDER BY id DESC
		LIMIT ? OFFSET ?
	`, limit, offset).Rows()
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

	return users, rows.Err()
}

// CountUsers 统计用户总数。
func (r *UserRepository) CountUsers() (int, error) {
	var count int64
	err := r.db.Table("users").
		Where("deleted_at IS NULL").
		Count(&count).Error
	if err != nil {
		return 0, err
	}
	return int(count), nil
}

// DeleteUser 软删除用户。
func (r *UserRepository) DeleteUser(username string) error {
	return r.db.Exec("UPDATE users SET deleted_at=NOW(), status='deleted' WHERE username=? AND deleted_at IS NULL", username).Error
}

