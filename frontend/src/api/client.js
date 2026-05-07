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
    const error = new Error(data?.message || 'request failed')
    error.status = response.status
    throw error
  }

  return data
}

export function getAppState() {
  return request('/api/state')
}

export function getBlogList(params = {}) {
  const searchParams = new URLSearchParams()
  searchParams.set('page', String(params.page || 1))
  searchParams.set('pageSize', String(params.pageSize || 10))

  if (params.keyword) {
    searchParams.set('keyword', params.keyword)
  }

  if (params.categoryId) {
    searchParams.set('categoryId', String(params.categoryId))
  }

  if (params.archive) {
    searchParams.set('archive', params.archive)
  }

  return request(`/api/blogs?${searchParams.toString()}`)
}

export function getCategories() {
  return request('/api/categories')
}

export function getManageCategories() {
  return request('/api/admin/categories')
}

export function createCategory(payload) {
  return request('/api/admin/categories', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json'
    },
    body: JSON.stringify(payload)
  })
}

export function updateCategory(categoryID, payload) {
  return request(`/api/admin/categories/${categoryID}`, {
    method: 'PUT',
    headers: {
      'Content-Type': 'application/json'
    },
    body: JSON.stringify(payload)
  })
}

export function deleteCategory(categoryID) {
  return request(`/api/admin/categories/${categoryID}`, {
    method: 'DELETE'
  })
}

export function getTags() {
  return request('/api/tags')
}

export function getArchives() {
  return request('/api/archives')
}

export function createBlog(payload) {
  const body = new URLSearchParams(payload)
  return request('/api/blogs', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/x-www-form-urlencoded'
    },
    body
  })
}

export function getBlogById(blogID) {
  return request(`/api/blogs/${blogID}`)
}

export function toggleBlogLike(blogID) {
  return request(`/api/blogs/${blogID}/like`, {
    method: 'POST'
  })
}

export function toggleBlogFavorite(blogID) {
  return request(`/api/blogs/${blogID}/favorite`, {
    method: 'POST'
  })
}

export function getCommentsByBlogId(blogID) {
  return request(`/api/blogs/${blogID}/comments`)
}

export function createComment(blogID, payload) {
  return request(`/api/blogs/${blogID}/comments`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json'
    },
    body: JSON.stringify(payload)
  })
}

export function deleteComment(commentID) {
  return request(`/api/comments/${commentID}`, {
    method: 'DELETE'
  })
}

export function updateBlog(blogID, payload) {
  const body = new URLSearchParams(payload)
  return request(`/api/blogs/${blogID}`, {
    method: 'PUT',
    headers: {
      'Content-Type': 'application/x-www-form-urlencoded'
    },
    body
  })
}

export function deleteBlog(blogID) {
  return request(`/api/blogs/${blogID}`, {
    method: 'DELETE'
  })
}

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

export function logout() {
  return request('/api/logout', {
    method: 'POST'
  })
}

export function getCurrentUser() {
  return request('/api/me')
}

export function updateUserProfile(payload) {
  return request('/api/user/profile', {
    method: 'PUT',
    headers: {
      'Content-Type': 'application/json'
    },
    body: JSON.stringify(payload)
  })
}

export function uploadUserAvatar(file) {
  const formData = new FormData()
  formData.append('avatar', file)

  return request('/api/user/avatar', {
    method: 'POST',
    body: formData
  })
}

export function updateUserPassword(payload) {
  return request('/api/user/password', {
    method: 'PUT',
    headers: {
      'Content-Type': 'application/json'
    },
    body: JSON.stringify(payload)
  })
}

export function updateUserPermission(payload) {
  return request('/api/user/permission', {
    method: 'PUT',
    headers: {
      'Content-Type': 'application/json'
    },
    body: JSON.stringify(payload)
  })
}

export function getAdminUsers(page = 1, pageSize = 10) {
  return request(`/api/admin/users?page=${page}&pageSize=${pageSize}`)
}

export function getCurrentUserBlogs(params = {}) {
  const searchParams = new URLSearchParams()
  searchParams.set('page', String(params.page || 1))
  searchParams.set('pageSize', String(params.pageSize || 10))

  if (params.status) {
    searchParams.set('status', params.status)
  }

  return request(`/api/user/blogs?${searchParams.toString()}`)
}

export function getCurrentUserFavorites(params = {}) {
  const searchParams = new URLSearchParams()
  searchParams.set('page', String(params.page || 1))
  searchParams.set('pageSize', String(params.pageSize || 10))
  return request(`/api/user/favorites?${searchParams.toString()}`)
}


export function getCurrentUserLikes(params = {}) {
  const searchParams = new URLSearchParams()
  searchParams.set('page', String(params.page || 1))
  searchParams.set('pageSize', String(params.pageSize || 10))
  return request(`/api/user/likes?${searchParams.toString()}`)
}
export function getAdminBlogs(params = {}) {
  const searchParams = new URLSearchParams()
  searchParams.set('page', String(params.page || 1))
  searchParams.set('pageSize', String(params.pageSize || 10))

  if (params.keyword) {
    searchParams.set('keyword', params.keyword)
  }

  if (params.author) {
    searchParams.set('author', params.author)
  }

  if (params.status) {
    searchParams.set('status', params.status)
  }

  return request(`/api/admin/blogs?${searchParams.toString()}`)
}

export function reviewAdminBlog(blogID, payload) {
  return request(`/api/admin/blogs/${blogID}/review`, {
    method: 'PUT',
    headers: {
      'Content-Type': 'application/json'
    },
    body: JSON.stringify(payload)
  })
}

export function deleteManagedUser(username) {
  return request(`/api/admin/users/${encodeURIComponent(username)}`, {
    method: 'DELETE'
  })
}

