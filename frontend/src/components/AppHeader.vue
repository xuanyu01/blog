<!--
/*
  这个文件定义顶部导航组件
*/
-->
<template>
  <header class="nav">
    <div class="nav-main">
      <ul class="nav-list nav-list-desktop">
        <li class="link"><RouterLink to="/" class="link_href">首页</RouterLink></li>
        <li v-if="store.user.isLogin" class="link">
          <RouterLink to="/blog/create" class="link_href">创作</RouterLink>
        </li>
        <li v-if="store.user.isLogin" class="link">
          <RouterLink to="/user/favorites" class="link_href">我的收藏</RouterLink>
        </li>
      </ul>
    </div>

    <div class="user-area">
      <template v-if="store.user.isLogin">
        <RouterLink
          v-if="canOpenAdmin"
          to="/admin"
          class="admin-entry"
        >
          管理员界面
        </RouterLink>

        <RouterLink to="/user" class="avatar-link">
          <img
            v-if="store.user.imageRoute"
            :src="`/img/${store.user.imageRoute}`"
            :alt="displayNameForView"
            class="avatar"
          />
          <div v-else class="avatar avatar-fallback">{{ initials }}</div>
        </RouterLink>

        <button class="login-btn" @click="handleLogout">退出登录</button>
      </template>
      <template v-else>
        <RouterLink to="/login" class="login-btn">登录</RouterLink>
        <RouterLink to="/register" class="login-btn">注册</RouterLink>
      </template>
    </div>
  </header>
</template>

<script setup>
/*
  这个组件负责展示站点导航和登录状态入口
*/
import { computed, onMounted } from 'vue'
import { RouterLink, useRouter } from 'vue-router'
import { appStore as store, logoutAndClear, refreshCurrentUser } from '../store/appStore'

const router = useRouter()

// displayNameForView 用显示名称优先展示
const displayNameForView = computed(() => store.user.displayName || store.user.userName || 'U')

// initials 根据显示名称生成默认头像字母
const initials = computed(() => displayNameForView.value.slice(0, 1).toUpperCase())

// canOpenAdmin 控制管理员界面的入口显示
const canOpenAdmin = computed(() => {
  return store.user.permission === 'admin' || store.user.permission === 'user_admin'
})

// handleLogout 处理退出登录动作
async function handleLogout() {
  await logoutAndClear()
  router.push('/')
}

// 组件挂载时尝试同步登录态
onMounted(() => {
  if (!store.user.isLogin) {
    refreshCurrentUser()
  }
})
</script>

<style scoped>
.user-area {
  display: flex;
  align-items: center;
  gap: 12px;
}

.admin-entry {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  padding: 10px 14px;
  border-radius: 14px;
  background: #e8edf2;
  color: #203040;
  font-weight: 600;
  text-decoration: none;
}
</style>
