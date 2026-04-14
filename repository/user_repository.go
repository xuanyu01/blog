/*
该文件实现用户数据的数据库访问逻辑
*/
package repository

import (
	"blog/model"
	"database/sql"
	"errors"
)

// UserRepository 负责读取和写入用户数据
type UserRepository struct {
	db *sql.DB
}

// NewUserRepository 创建用户仓储实例
func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

// Exists 判断用户名是否已经存在
func (r *UserRepository) Exists(username string) (bool, error) {
	var storedUsername string
	err := r.db.QueryRow("SELECT username FROM user WHERE username=?", username).Scan(&storedUsername)
	if err == nil {
		return true, nil
	}

	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	return false, err
}

// Create 创建新用户
func (r *UserRepository) Create(username, hashedPassword string) error {
	_, err := r.db.Exec(
		"INSERT INTO user (username, display_name, password) VALUES (?, ?, ?)",
		username,
		username,
		hashedPassword,
	)
	return err
}

// GetPasswordByUsername 根据用户名查询密码哈希
func (r *UserRepository) GetPasswordByUsername(username string) (string, error) {
	var hashedPassword string
	err := r.db.QueryRow("SELECT password FROM user WHERE username=?", username).Scan(&hashedPassword)
	if err != nil {
		return "", err
	}
	return hashedPassword, nil
}

// GetByUsername 根据用户名查询基础用户信息
func (r *UserRepository) GetByUsername(username string) (model.User, error) {
	var user model.User
	err := r.db.QueryRow(
		"SELECT username, display_name, image FROM user WHERE username=?",
		username,
	).Scan(&user.Username, &user.DisplayName, &user.Image)
	if err != nil {
		return model.User{}, err
	}
	return user, nil
}

// UpdateProfile 更新用户显示名和头像路径
func (r *UserRepository) UpdateProfile(username, displayName, image string) error {
	_, err := r.db.Exec(
		"UPDATE user SET display_name=?, image=? WHERE username=?",
		displayName,
		image,
		username,
	)
	return err
}

// UpdateImage 只更新用户头像路径
func (r *UserRepository) UpdateImage(username, image string) error {
	_, err := r.db.Exec(
		"UPDATE user SET image=? WHERE username=?",
		image,
		username,
	)
	return err
}

// UpdatePassword 更新用户密码哈希
func (r *UserRepository) UpdatePassword(username, hashedPassword string) error {
	_, err := r.db.Exec(
		"UPDATE user SET password=? WHERE username=?",
		hashedPassword,
		username,
	)
	return err
}
