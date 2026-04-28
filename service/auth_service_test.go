/*
auth_service_test.go 覆盖认证服务的核心业务测试。
*/
package service

import (
	"blog/model"
	"database/sql"
	"errors"
	"testing"
)

type fakeUserRepo struct {
	users map[string]model.User
	hash  map[string]string
	items []model.UserListItem
	total int
}

// Exists 模拟检查用户是否存在。
func (f *fakeUserRepo) Exists(username string) (bool, error) {
	_, ok := f.users[username]
	return ok, nil
}

// Create 模拟创建用户并记录密码哈希。
func (f *fakeUserRepo) Create(username, hashedPassword string) error {
	if f.users == nil {
		f.users = map[string]model.User{}
	}
	if f.hash == nil {
		f.hash = map[string]string{}
	}
	f.users[username] = model.User{Username: username, DisplayName: username, Permission: model.PermissionUser}
	f.hash[username] = hashedPassword
	return nil
}

// GetPasswordByUsername 模拟按用户名读取密码哈希。
func (f *fakeUserRepo) GetPasswordByUsername(username string) (string, error) {
	hash, ok := f.hash[username]
	if !ok {
		return "", sql.ErrNoRows
	}
	return hash, nil
}

// GetByUsername 模拟按用户名读取用户信息。
func (f *fakeUserRepo) GetByUsername(username string) (model.User, error) {
	user, ok := f.users[username]
	if !ok {
		return model.User{}, sql.ErrNoRows
	}
	return user, nil
}

// UpdateProfile 模拟更新用户资料。
func (f *fakeUserRepo) UpdateProfile(username, displayName, image string) error {
	user := f.users[username]
	user.DisplayName = displayName
	user.Image = image
	f.users[username] = user
	return nil
}

// UpdateImage 模拟更新头像地址。
func (f *fakeUserRepo) UpdateImage(username, image string) error {
	user := f.users[username]
	user.Image = image
	f.users[username] = user
	return nil
}

// UpdatePassword 模拟更新密码哈希。
func (f *fakeUserRepo) UpdatePassword(username, hashedPassword string) error {
	f.hash[username] = hashedPassword
	return nil
}

// UpdatePermission 模拟更新用户权限。
func (f *fakeUserRepo) UpdatePermission(username, permission string) error {
	user := f.users[username]
	user.Permission = permission
	f.users[username] = user
	return nil
}

// CountUsers 模拟返回用户总数。
func (f *fakeUserRepo) CountUsers() (int, error) { return f.total, nil }

// ListUsers 模拟返回用户分页列表。
func (f *fakeUserRepo) ListUsers(limit, offset int) ([]model.UserListItem, error) {
	return f.items, nil
}

// DeleteUser 模拟删除用户。
func (f *fakeUserRepo) DeleteUser(username string) error {
	delete(f.users, username)
	delete(f.hash, username)
	return nil
}

type fakeSessionStore struct {
	createdFor string
	sessionID  string
	current    map[string]string
}

// Create 模拟创建会话。
func (f *fakeSessionStore) Create(userID string) (string, error) {
	f.createdFor = userID
	if f.sessionID == "" {
		f.sessionID = "session-1"
	}
	if f.current == nil {
		f.current = map[string]string{}
	}
	f.current[f.sessionID] = userID
	return f.sessionID, nil
}

// Get 模拟读取会话。
func (f *fakeSessionStore) Get(sessionID string) (string, error) {
	userID, ok := f.current[sessionID]
	if !ok {
		return "", errors.New("session not found")
	}
	return userID, nil
}

// Update 模拟更新会话。
func (f *fakeSessionStore) Update(sessionID, userID string) error {
	if f.current == nil {
		f.current = map[string]string{}
	}
	f.current[sessionID] = userID
	return nil
}

// Delete 模拟删除会话。
func (f *fakeSessionStore) Delete(sessionID string) error {
	delete(f.current, sessionID)
	return nil
}

// TestAuthServiceRegisterAndLogin 验证注册和登录的基本流程。
func TestAuthServiceRegisterAndLogin(t *testing.T) {
	repo := &fakeUserRepo{}
	sessions := &fakeSessionStore{}
	service := NewAuthService(repo, sessions)

	if err := service.Register("alice", "secret123"); err != nil {
		t.Fatalf("Register returned error: %v", err)
	}

	if repo.hash["alice"] == "" {
		t.Fatal("expected hashed password to be stored")
	}

	sessionID, err := service.Login("alice", "secret123")
	if err != nil {
		t.Fatalf("Login returned error: %v", err)
	}

	if sessionID == "" || sessions.createdFor != "alice" {
		t.Fatalf("expected login to create session for alice, got session=%q user=%q", sessionID, sessions.createdFor)
	}
}

// TestAuthServiceUpdateUserPermissionRequiresAdmin 验证修改权限必须由管理员执行。
func TestAuthServiceUpdateUserPermissionRequiresAdmin(t *testing.T) {
	repo := &fakeUserRepo{
		users: map[string]model.User{
			"manager": {Username: "manager", Permission: model.PermissionUserAdmin},
			"target":  {Username: "target", Permission: model.PermissionUser},
		},
		hash: map[string]string{},
	}
	sessions := &fakeSessionStore{
		current: map[string]string{"session-1": "manager"},
	}
	service := NewAuthService(repo, sessions)

	err := service.UpdateUserPermission("session-1", model.UserPermissionUpdate{
		Username:   "target",
		Permission: model.PermissionUserAdmin,
	})
	if err == nil || err.Error() != "only admin can update user permission" {
		t.Fatalf("expected admin permission error, got %v", err)
	}
}

// TestAuthServiceDeleteUserPermissionRules 验证删除用户时的权限规则。
func TestAuthServiceDeleteUserPermissionRules(t *testing.T) {
	repo := &fakeUserRepo{
		users: map[string]model.User{
			"moderator": {Username: "moderator", Permission: model.PermissionUserAdmin},
			"admin":     {Username: "admin", Permission: model.PermissionAdmin},
		},
		hash: map[string]string{},
	}
	sessions := &fakeSessionStore{
		current: map[string]string{"session-2": "moderator"},
	}
	service := NewAuthService(repo, sessions)

	err := service.DeleteUser("session-2", "admin")
	if err == nil || err.Error() != "cannot delete admin user" {
		t.Fatalf("expected cannot delete admin user, got %v", err)
	}
}
