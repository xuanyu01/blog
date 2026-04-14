/*
	这个文件封装前端访问后端 API 的请求方法
*/

// request 统一处理请求发送 响应解析和错误抛出
async function request(url, options = {}) {
  const headers = {
    ...(options.headers || {})
  }

  const response = await fetch(url, {
    credentials: 'include',
    ...options,
    headers
  })

  const contentType = response.headers.get('content-type') || ''
  const data = contentType.includes('application/json') ? await response.json() : null

  if (!response.ok) {
    throw new Error(data?.message || 'request failed')
  }

  return data
}

// getAppState 获取首页聚合状态
export function getAppState() {
  return request('/api/state')
}

// login 发送登录请求
export function login(payload) {
  const body = new URLSearchParams(payload)
  return request('/api/login', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/x-www-form-urlencoded'
    },
    body
  })
}

// register 发送注册请求
export function register(payload) {
  const body = new URLSearchParams(payload)
  return request('/api/register', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/x-www-form-urlencoded'
    },
    body
  })
}

// logout 发送退出登录请求
export function logout() {
  return request('/api/logout', {
    method: 'POST'
  })
}

// getCurrentUser 获取当前登录用户
export function getCurrentUser() {
  return request('/api/me')
}

// updateUserProfile 更新用户资料
export function updateUserProfile(payload) {
  return request('/api/user/profile', {
    method: 'PUT',
    headers: {
      'Content-Type': 'application/json'
    },
    body: JSON.stringify(payload)
  })
}

// uploadUserAvatar 上传用户头像
export function uploadUserAvatar(file) {
  const formData = new FormData()
  formData.append('avatar', file)

  return request('/api/user/avatar', {
    method: 'POST',
    body: formData
  })
}

// updateUserPassword 更新当前用户密码
export function updateUserPassword(payload) {
  return request('/api/user/password', {
    method: 'PUT',
    headers: {
      'Content-Type': 'application/json'
    },
    body: JSON.stringify(payload)
  })
}
