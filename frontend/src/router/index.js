/*
	这个文件定义前端页面路由。
*/
import { createRouter, createWebHistory } from 'vue-router'
import HomeView from '../views/HomeView.vue'
import LoginView from '../views/LoginView.vue'
import RegisterView from '../views/RegisterView.vue'
import UserView from '../views/UserView.vue'
import UserProfileView from '../views/UserProfileView.vue'
import AdminView from '../views/AdminView.vue'
import AvatarUploadView from '../views/AvatarUploadView.vue'
import CreateBlogView from '../views/CreateBlogView.vue'
import BlogDetailView from '../views/BlogDetailView.vue'
import { appStore, refreshCurrentUser } from '../store/appStore'

function redirectToCurrentUser() {
  if (!appStore.user.isLogin) {
    return { name: 'login' }
  }

  return {
    name: 'user-profile',
    params: { id: String(appStore.user.id) }
  }
}
// routes 描述路径和页面组件的映射关系
const routes = [
  { path: '/', name: 'home', component: HomeView },
  { path: '/login', name: 'login', component: LoginView, meta: { guestOnly: true } },
  { path: '/register', name: 'register', component: RegisterView, meta: { guestOnly: true } },
  { path: '/user', name: 'user-root', beforeEnter: redirectToCurrentUser, meta: { requiresAuth: true } },
  { path: '/user/drafts', name: 'user-drafts', beforeEnter: redirectToCurrentUser, meta: { requiresAuth: true } },
  { path: '/user/favorites', name: 'user-favorites', beforeEnter: redirectToCurrentUser, meta: { requiresAuth: true } },
  { path: '/user/:id(\\d+)/edit', name: 'user-edit', component: UserView, meta: { requiresAuth: true } },
  { path: '/user/:id(\\d+)', name: 'user-profile', component: UserProfileView, meta: { requiresAuth: true } },
  { path: '/admin', name: 'admin', component: AdminView, meta: { requiresManager: true } },
  { path: '/user/avatar', name: 'user-avatar', component: AvatarUploadView, meta: { requiresAuth: true } },
  { path: '/blog/:id', name: 'blog-detail', component: BlogDetailView },
  { path: '/blog/create', name: 'blog-create', component: CreateBlogView, meta: { requiresAuth: true } },
  { path: '/blog/:id/edit', name: 'blog-edit', component: CreateBlogView, meta: { requiresAuth: true } }
]

// router 是前端路由实例
const router = createRouter({
  history: createWebHistory(),
  routes,
  scrollBehavior() {
    return { top: 0 }
  }
})

router.beforeEach(async (to) => {
  const requiresUserState = to.matched.some((record) => (
    record.meta.requiresAuth ||
    record.meta.requiresManager ||
    record.meta.guestOnly
  ))

  if (requiresUserState) {
    await refreshCurrentUser()
  }

  if (to.matched.some((record) => record.meta.guestOnly) && appStore.user.isLogin) {
    return { name: 'home' }
  }

  if (to.matched.some((record) => record.meta.requiresAuth) && !appStore.user.isLogin) {
    return {
      name: 'login',
      query: { redirect: to.fullPath }
    }
  }

  if (to.matched.some((record) => record.meta.requiresManager)) {
    const canManage = appStore.user.permission === 'admin' || appStore.user.permission === 'user_admin'
    if (!appStore.user.isLogin) {
      return {
        name: 'login',
        query: { redirect: to.fullPath }
      }
    }
    if (!canManage) {
      return { name: 'home' }
    }
  }

  return true
})

export default router

