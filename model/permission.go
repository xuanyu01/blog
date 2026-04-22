/*
这个文件定义用户权限相关的常量和判断函数
*/
package model

const (
	PermissionUser      = "user"
	PermissionUserAdmin = "user_admin"
	PermissionAdmin     = "admin"
)

// CanManageAllBlogs 判断当前权限是否可以管理全部博客
func CanManageAllBlogs(permission string) bool {
	return permission == PermissionAdmin || permission == PermissionUserAdmin
}

// IsAssignablePermission 判断是否允许通过管理接口分配该权限
func IsAssignablePermission(permission string) bool {
	return permission == PermissionUser || permission == PermissionUserAdmin
}
