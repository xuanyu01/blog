/*
该文件实现注册 登录 当前用户查询和用户资料修改相关的业务逻辑
*/
package service

import (
	"blog/model"
	"blog/repository"
	"blog/session"
	"database/sql"
	"errors"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

// ErrInvalidCredentials 表示登录凭证无效
var ErrInvalidCredentials = errors.New("username or password is invalid")

// AuthService 负责认证和用户资料相关业务
// 它把用户仓储和会话存储组合成完整的登录流程
type AuthService struct {
	userRepo     *repository.UserRepository
	sessionStore session.Store
}

// NewAuthService 创建认证业务服务
func NewAuthService(userRepo *repository.UserRepository, sessionStore session.Store) *AuthService {
	return &AuthService{
		userRepo:     userRepo,
		sessionStore: sessionStore,
	}
}

// Register 注册新用户
func (s *AuthService) Register(username, password string) error {
	username = strings.TrimSpace(username)
	if username == "" || password == "" {
		return errors.New("username and password are required")
	}

	exists, err := s.userRepo.Exists(username)
	if err != nil {
		return err
	}
	if exists {
		return errors.New("user already exists")
	}

	// 密码哈希放在服务层处理 这样安全规则不会泄漏到更底层的数据访问代码里
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	return s.userRepo.Create(username, string(hashedPassword))
}

// Login 校验凭证并创建会话
func (s *AuthService) Login(username, password string) (string, error) {
	hashedPassword, err := s.userRepo.GetPasswordByUsername(username)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", ErrInvalidCredentials
		}
		return "", err
	}

	// 无论是用户名不存在还是密码错误 都统一返回同一种错误 减少信息泄露
	if err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password)); err != nil {
		return "", ErrInvalidCredentials
	}

	return s.sessionStore.Create(username)
}

// Logout 删除当前会话
func (s *AuthService) Logout(sessionID string) error {
	if sessionID == "" {
		return nil
	}
	return s.sessionStore.Delete(sessionID)
}

// CurrentUser 根据 sessionID 解析当前用户信息
func (s *AuthService) CurrentUser(sessionID string) (model.UserView, error) {
	if sessionID == "" {
		return model.UserView{}, nil
	}

	username, err := s.sessionStore.Get(sessionID)
	if err != nil {
		return model.UserView{}, err
	}

	return s.userViewByUsername(username)
}

// UpdateProfile 更新用户显示名和头像路径
func (s *AuthService) UpdateProfile(sessionID string, payload model.UserProfileUpdate) (model.UserView, error) {
	if sessionID == "" {
		return model.UserView{}, errors.New("unauthorized")
	}

	username, err := s.sessionStore.Get(sessionID)
	if err != nil {
		return model.UserView{}, err
	}

	displayName := strings.TrimSpace(payload.DisplayName)
	if displayName == "" {
		return model.UserView{}, errors.New("display name is required")
	}

	if err := s.userRepo.UpdateProfile(username, displayName, strings.TrimSpace(payload.ImageRoute)); err != nil {
		return model.UserView{}, err
	}

	return s.userViewByUsername(username)
}

// UpdateAvatar 更新当前用户头像路径
func (s *AuthService) UpdateAvatar(sessionID, imageRoute string) (model.UserView, error) {
	if sessionID == "" {
		return model.UserView{}, errors.New("unauthorized")
	}

	username, err := s.sessionStore.Get(sessionID)
	if err != nil {
		return model.UserView{}, err
	}

	if strings.TrimSpace(imageRoute) == "" {
		return model.UserView{}, errors.New("image route is required")
	}

	if err := s.userRepo.UpdateImage(username, strings.TrimSpace(imageRoute)); err != nil {
		return model.UserView{}, err
	}

	return s.userViewByUsername(username)
}

// UpdatePassword 更新当前用户密码
func (s *AuthService) UpdatePassword(sessionID string, payload model.PasswordUpdate) error {
	if sessionID == "" {
		return errors.New("unauthorized")
	}

	username, err := s.sessionStore.Get(sessionID)
	if err != nil {
		return err
	}

	if payload.CurrentPassword == "" || payload.NewPassword == "" {
		return errors.New("current password and new password are required")
	}

	hashedPassword, err := s.userRepo.GetPasswordByUsername(username)
	if err != nil {
		return err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(payload.CurrentPassword)); err != nil {
		return errors.New("current password is invalid")
	}

	newHashedPassword, err := bcrypt.GenerateFromPassword([]byte(payload.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	return s.userRepo.UpdatePassword(username, string(newHashedPassword))
}

// userViewByUsername 根据用户名加载页面展示需要的用户信息
func (s *AuthService) userViewByUsername(username string) (model.UserView, error) {
	user, err := s.userRepo.GetByUsername(username)
	if err != nil {
		return model.UserView{}, err
	}

	return model.UserView{
		UserName:    user.Username,
		DisplayName: user.DisplayName,
		ImageRoute:  user.Image,
		IsLogin:     true,
	}, nil
}
