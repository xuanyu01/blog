/*
	该文件实现前端全局状态存储
*/
import { reactive } from 'vue'
import {
  getAppState,
  getCurrentUser,
  logout as logoutRequest,
  updateUserPassword,
  updateUserProfile,
  uploadUserAvatar
} from '../api/client'

// appStore 保存博客列表 当前用户和加载状态
// 页面组件通过它共享基础数据 而不必各自重复请求
export const appStore = reactive({
  blogs: [],
  user: {
    userName: '',
    displayName: '',
    imageRoute: '',
    isLogin: false
  },
  loading: false
})

function emptyUser() {
  return {
    userName: '',
    displayName: '',
    imageRoute: '',
    isLogin: false
  }
}

// normalizeUser 兼容后端字段名并统一成前端使用的格式
function normalizeUser(user = {}) {
  return {
    userName: user.userName || user.UserName || '',
    displayName: user.displayName || user.DisplayName || '',
    imageRoute: user.imageRoute || user.ImageRoute || '',
    isLogin: Boolean(user.isLogin ?? user.IsLogin)
  }
}

// refreshAppState 刷新首页所需的聚合状态
export async function refreshAppState() {
  appStore.loading = true
  try {
    const data = await getAppState()
    appStore.blogs = data.blogs || []
    appStore.user = normalizeUser(data.user)
  } finally {
    appStore.loading = false
  }
}

// refreshCurrentUser 刷新当前用户状态
export async function refreshCurrentUser() {
  try {
    const user = normalizeUser(await getCurrentUser())
    appStore.user = user
    return user
  } catch {
    appStore.user = emptyUser()
    return appStore.user
  }
}

// saveUserProfile 保存用户资料并同步全局状态
export async function saveUserProfile(payload) {
  const user = normalizeUser(await updateUserProfile(payload))
  appStore.user = user
  return user
}

// uploadAvatarAndSync 上传头像并同步全局状态
export async function uploadAvatarAndSync(file) {
  const user = normalizeUser(await uploadUserAvatar(file))
  appStore.user = user
  return user
}

// changeUserPassword 修改当前用户密码
export async function changeUserPassword(payload) {
  return updateUserPassword(payload)
}

// logoutAndClear 退出登录并清理本地用户状态
export async function logoutAndClear() {
  await logoutRequest()
  appStore.user = emptyUser()
}
