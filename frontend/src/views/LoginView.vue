<!--
登录页面组件。
-->
<template>
  <section class="page-block auth-page">
    <div class="register-container">
      <h2>用户登录</h2>

      <div v-if="redirecting" class="auth-form">
        <p class="feedback success">{{ message }}</p>
      </div>

      <template v-else>
        <form class="auth-form" @submit.prevent="handleSubmit">
          <div class="form-group">
            <label for="login-username">用户名：</label>
            <input id="login-username" v-model.trim="form.username" type="text" placeholder="请输入用户名" required />
          </div>

          <div class="form-group">
            <label for="login-password">密码：</label>
            <input id="login-password" v-model="form.password" type="password" placeholder="请输入密码" required />
          </div>

          <div class="form-group">
            <button type="submit" :disabled="loading">{{ loading ? '登录中...' : '登录' }}</button>
          </div>

          <a class="switch-link" @click.prevent="router.push('/register')" href="/register">还没有账号？去注册</a>
        </form>

        <p v-if="message" :class="success ? 'feedback success' : 'feedback error'">{{ message }}</p>
      </template>
    </div>
  </section>
</template>

<script setup>
import { onMounted, reactive, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { login } from '../api/client'
import { appStore as store, refreshAppState, refreshCurrentUser } from '../store/appStore'

const router = useRouter()
const route = useRoute()
const loading = ref(false)
const message = ref('')
const success = ref(false)
const redirecting = ref(false)

const form = reactive({ username: '', password: '' })

onMounted(async () => {
  const user = await refreshCurrentUser()
  if (!user.isLogin) {
    return
  }

  redirecting.value = true
  success.value = true
  message.value = user.mustChangePassword ? '你需要先修改密码，即将跳转。' : '你已经登录，3 秒后自动返回首页。'

  setTimeout(() => {
    router.push(user.mustChangePassword ? passwordChangePath(user.id) : resolveRedirectPath())
  }, 3000)
})

async function handleSubmit() {
  loading.value = true
  message.value = ''

  try {
    await login(form)
    await refreshAppState()
    success.value = true
    message.value = store.user.mustChangePassword ? '登录成功，首次登录需要先修改密码。' : '登录成功，3 秒后自动跳转。'

    setTimeout(() => {
      router.push(store.user.mustChangePassword ? passwordChangePath(store.user.id) : resolveRedirectPath())
    }, 3000)
  } catch (error) {
    success.value = false
    message.value = translateAuthMessage(error.message)
  } finally {
    loading.value = false
  }
}

function resolveRedirectPath() {
  const redirect = typeof route.query.redirect === 'string' ? route.query.redirect : ''
  return redirect || '/'
}

function passwordChangePath(userID) {
  return '/user/' + userID + '/edit?forcePassword=1'
}

function translateAuthMessage(raw = '') {
  const messages = {
    'username and password are required': '请输入用户名和密码。',
    'username or password is invalid': '用户名或密码错误。',
    'too many login attempts, please try again later': '登录失败次数过多，请稍后再试。',
    'login is temporarily unavailable': '登录服务暂时不可用，请稍后再试。',
    'request failed': '请求失败，请稍后再试。'
  }
  return messages[raw] || '操作失败，请稍后再试。'
}
</script>
