/*
permission.go 定义权限常量和权限判断函数。
*/
package model

const (
	PermissionUser      = "user"
	PermissionUserAdmin = "user_admin"
	PermissionAdmin     = "admin"
)

// CanManageAllBlogs 。жϵ。ǰȨ。。。Ƿ。。。Թ。。。ȫ。。。。。͡。
func CanManageAllBlogs(permission string) bool {
	return permission == PermissionAdmin || permission == PermissionUserAdmin
}

// IsAssignablePermission 。жϹ。。。ӿ。。Ƿ。。。。。。。。。Ȩ。ޡ。
func IsAssignablePermission(permission string) bool {
	return permission == PermissionUser || permission == PermissionUserAdmin
}

