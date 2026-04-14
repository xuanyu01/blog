<!--
/*
	该文件定义顶部导航组件
*/
-->
<template>
  <header class="nav">
    <div class="nav-main">
      <ul class="nav-list nav-list-desktop">
        <li class="link"><RouterLink to="/" class="link_href">首页</RouterLink></li>
        <li class="link"><a href="#" class="link_href">说说</a></li>
        <li class="link"><a href="#" class="link_href">动漫</a></li>
        <li class="link nav-item-with-menu">
          <a href="#" class="link_block">关于我</a>
          <ul class="droplist">
            <li class="link2"><a href="#" class="link_href">个人成就</a></li>
          </ul>
        </li>
        <li class="link nav-item-with-menu">
          <a href="#" class="link_block">CTF</a>
          <ul class="droplist">
            <li class="link2"><a href="#" class="link_href">Write Up</a></li>
          </ul>
        </li>
      </ul>

      <div class="nav-collapsed">
        <button class="menu-trigger" type="button">菜单</button>
        <ul class="collapsed-menu">
          <li><RouterLink to="/">首页</RouterLink></li>
          <li><a href="#">说说</a></li>
          <li><a href="#">动漫</a></li>
          <li><a href="#">关于我</a></li>
          <li><a href="#">CTF</a></li>
        </ul>
      </div>
    </div>

    <div class="user-area">
      <template v-if="store.user.isLogin">
        <RouterLink to="/user" class="avatar-link">
          <img
            v-if="store.user.imageRoute"
            :src="`/img/${store.user.imageRoute}`"
            :alt="displayNameForView"
            class="avatar"
          />
          <div v-else class="avatar avatar-fallback">{{ initials }}</div>
        </RouterLink>
        <button class="login-btn" @click="handleLogout">Logout</button>
      </template>
      <template v-else>
        <RouterLink to="/login" class="login-btn">Login</RouterLink>
        <RouterLink to="/register" class="login-btn">Register</RouterLink>
      </template>
    </div>
  </header>
</template>

<script setup>
/*
	该组件负责展示站点导航和登录状态入口
*/
import { computed, onMounted } from 'vue'
import { RouterLink, useRouter } from 'vue-router'
import { appStore as store, logoutAndClear, refreshCurrentUser } from '../store/appStore'

const router = useRouter()

// displayNameForView 用显示名称优先展示 当前没有时回退到账号
const displayNameForView = computed(() => store.user.displayName || store.user.userName || 'U')

// initials 根据显示名称生成默认头像字母
const initials = computed(() => displayNameForView.value.slice(0, 1).toUpperCase())

// handleLogout 处理退出登录动作
async function handleLogout() {
  await logoutAndClear()
  router.push('/')
}

// 组件挂载时尝试同步登录态 避免刷新页面后导航状态不一致
onMounted(() => {
  if (!store.user.isLogin) {
    refreshCurrentUser()
  }
})
</script>
