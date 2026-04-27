/*
定义用户相关模型。
*/
package model

// User 表示数据库中的用户记录。
type User struct {
	ID          int64
	Username    string
	DisplayName string
	Image       string
	Permission  string
	Status      string
}

// UserView 表示前端使用的用户信息。
type UserView struct {
	UserName    string `json:"userName"`
	DisplayName string `json:"displayName"`
	ImageRoute  string `json:"imageRoute"`
	Permission  string `json:"permission"`
	IsLogin     bool   `json:"isLogin"`
}

// UserProfileUpdate 表示资料修改请求。
type UserProfileUpdate struct {
	DisplayName string `json:"displayName"`
	ImageRoute  string `json:"imageRoute"`
}

// UserPermissionUpdate 表示权限修改请求。
type UserPermissionUpdate struct {
	Username   string `json:"username"`
	Permission string `json:"permission"`
}

// UserListItem 表示后台用户列表中的单项数据。
type UserListItem struct {
	Username    string `json:"username"`
	DisplayName string `json:"displayName"`
	Permission  string `json:"permission"`
}

// PasswordUpdate 表示密码修改请求。
type PasswordUpdate struct {
	CurrentPassword string `json:"currentPassword"`
	NewPassword     string `json:"newPassword"`
}
