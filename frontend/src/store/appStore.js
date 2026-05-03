/*
前端全局状态存储。
*/
import { reactive } from 'vue'
import {
  getAppState,
  getArchives,
  getBlogList,
  getCategories,
  getCurrentUser,
  logout as logoutRequest,
  updateUserPassword,
  updateUserProfile,
  uploadUserAvatar
} from '../api/client'

export const appStore = reactive({
  blogs: [],
  blogList: {
    page: 1,
    pageSize: 10,
    total: 0,
    keyword: '',
    categoryId: 0,
    archive: ''
  },
  taxonomy: {
    categories: [],
    archives: []
  },
  user: {
    id: 0,
    userName: '',
    displayName: '',
    imageRoute: '',
    permission: '',
    isLogin: false
  },
  loading: false
})

function emptyUser() {
  return {
    id: 0,
    userName: '',
    displayName: '',
    imageRoute: '',
    permission: '',
    isLogin: false
  }
}

function normalizeUser(user = {}) {
  return {
    id: Number(user.id || user.ID || 0),
    userName: user.userName || user.UserName || '',
    displayName: user.displayName || user.DisplayName || '',
    imageRoute: user.imageRoute || user.ImageRoute || '',
    permission: user.permission || user.Permission || '',
    isLogin: Boolean(user.isLogin ?? user.IsLogin)
  }
}

export async function refreshAppState(params = {}) {
  appStore.loading = true
  try {
    const data = await getAppState()
    appStore.user = normalizeUser(data.user)
    await Promise.all([
      refreshBlogList(params),
      refreshTaxonomy()
    ])
  } finally {
    appStore.loading = false
  }
}

export async function refreshBlogList(params = {}) {
  const nextPage = params.page ?? appStore.blogList.page
  const nextPageSize = params.pageSize ?? appStore.blogList.pageSize
  const nextKeyword = params.keyword ?? appStore.blogList.keyword
  const nextCategoryId = params.categoryId ?? appStore.blogList.categoryId
  const nextArchive = params.archive ?? appStore.blogList.archive

  const data = await getBlogList({
    page: nextPage,
    pageSize: nextPageSize,
    keyword: nextKeyword,
    categoryId: nextCategoryId,
    archive: nextArchive
  })

  appStore.blogs = data.items || []
  appStore.blogList = {
    page: data.page || nextPage,
    pageSize: data.pageSize || nextPageSize,
    total: data.total || 0,
    keyword: data.keyword || '',
    categoryId: Number(data.categoryId || 0),
    archive: data.archive || ''
  }

  return appStore.blogList
}

export async function refreshTaxonomy() {
  const [categories, archives] = await Promise.all([
    getCategories(),
    getArchives()
  ])

  appStore.taxonomy.categories = categories.items || []
  appStore.taxonomy.archives = archives.items || []
  return appStore.taxonomy
}

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

export async function saveUserProfile(payload) {
  const user = normalizeUser(await updateUserProfile(payload))
  appStore.user = user
  return user
}

export async function uploadAvatarAndSync(file) {
  const user = normalizeUser(await uploadUserAvatar(file))
  appStore.user = user
  return user
}

export async function changeUserPassword(payload) {
  return updateUserPassword(payload)
}

export async function logoutAndClear() {
  await logoutRequest()
  appStore.user = emptyUser()
}


