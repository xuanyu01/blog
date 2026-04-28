/*
permission.go 定义权限常量和权限判断函数。
*/
package model

const (
	PermissionUser      = "user"
	PermissionUserAdmin = "user_admin"
	PermissionAdmin     = "admin"
)

// CanManageAllBlogs 判断当前权限是否可以管理全部博客。
func CanManageAllBlogs(permission string) bool {
	return permission == PermissionAdmin || permission == PermissionUserAdmin
}

// IsAssignablePermission 判断管理接口是否允许分配该权限。
func IsAssignablePermission(permission string) bool {
	return permission == PermissionUser || permission == PermissionUserAdmin
}
