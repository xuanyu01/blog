/*
该文件定义用户相关的领域模型和展示模型
*/
package model

// User 表示数据库中的用户实体
// 它承载账号 用户显示名 和头像路径等持久化字段
type User struct {
	Username    string
	DisplayName string
	Image       string
}

// UserView 表示前端页面使用的用户展示数据
// 它把登录态和展示字段组合成接口直接可返回的结构
type UserView struct {
	UserName    string `json:"userName"`
	DisplayName string `json:"displayName"`
	ImageRoute  string `json:"imageRoute"`
	IsLogin     bool   `json:"isLogin"`
}

// UserProfileUpdate 表示用户资料修改请求
type UserProfileUpdate struct {
	DisplayName string `json:"displayName"`
	ImageRoute  string `json:"imageRoute"`
}

// PasswordUpdate 表示密码修改请求
type PasswordUpdate struct {
	CurrentPassword string `json:"currentPassword"`
	NewPassword     string `json:"newPassword"`
}
