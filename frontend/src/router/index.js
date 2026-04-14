/*
	这个文件定义前端页面路由
*/
import { createRouter, createWebHistory } from 'vue-router'
import HomeView from '../views/HomeView.vue'
import LoginView from '../views/LoginView.vue'
import RegisterView from '../views/RegisterView.vue'
import UserView from '../views/UserView.vue'
import AvatarUploadView from '../views/AvatarUploadView.vue'

// routes 描述路径和页面组件的映射关系
const routes = [
  { path: '/', name: 'home', component: HomeView },
  { path: '/login', name: 'login', component: LoginView },
  { path: '/register', name: 'register', component: RegisterView },
  { path: '/user', name: 'user', component: UserView },
  { path: '/user/avatar', name: 'user-avatar', component: AvatarUploadView }
]

// router 是前端路由实例
const router = createRouter({
  history: createWebHistory(),
  routes,
  scrollBehavior() {
    return { top: 0 }
  }
})

export default router
