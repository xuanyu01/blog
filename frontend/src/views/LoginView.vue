<!--
/*
	该文件定义登录页面组件
*/
-->
<template>
  <section class="page-block auth-page">
    <div class="register-container">
      <h2>Login</h2>

      <div v-if="redirecting" class="auth-form">
        <p class="feedback success">{{ message }}</p>
      </div>

      <template v-else>
        <form class="auth-form" @submit.prevent="handleSubmit">
          <div class="form-group">
            <label for="login-username">Username:</label>
            <input
              id="login-username"
              v-model.trim="form.username"
              type="text"
              placeholder="Enter your username"
              required
            />
          </div>

          <div class="form-group">
            <label for="login-password">Password:</label>
            <input
              id="login-password"
              v-model="form.password"
              type="password"
              placeholder="Enter your password"
              required
            />
          </div>

          <div class="form-group">
            <button type="submit" :disabled="loading">
              {{ loading ? 'Loading...' : 'Login' }}
            </button>
          </div>

          <a class="switch-link" @click.prevent="router.push('/register')" href="/register">
            点击此处跳转到注册界面
          </a>
        </form>

        <p v-if="message" :class="success ? 'feedback success' : 'feedback error'">
          {{ message }}
        </p>
      </template>
    </div>
  </section>
</template>

<script setup>
/*
	该页面负责收集登录信息并发起登录请求
*/
import { onMounted, reactive, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { login } from '../api/client'
import { refreshAppState, refreshCurrentUser } from '../store/appStore'

const router = useRouter()
const route = useRoute()
const loading = ref(false)
const message = ref('')
const success = ref(false)
const redirecting = ref(false)

// form 保存登录表单输入
const form = reactive({
  username: '',
  password: ''
})

// 页面打开时先检查当前登录态
onMounted(async () => {
  const user = await refreshCurrentUser()
  if (!user.isLogin) {
    return
  }

  redirecting.value = true
  success.value = true
  message.value = '你已登录 3 秒后自动返回首页'

  setTimeout(() => {
    router.push(resolveRedirectPath())
  }, 3000)
})

// handleSubmit 提交登录表单
async function handleSubmit() {
  loading.value = true
  message.value = ''

  try {
    await login(form)
    await refreshAppState()
    success.value = true
    message.value = '登录成功 3 秒后自动跳转到首页'

    // 先显示成功提示 再跳转回首页 让反馈更自然
    setTimeout(() => {
      router.push(resolveRedirectPath())
    }, 3000)
  } catch (error) {
    success.value = false
    message.value = error.message
  } finally {
    loading.value = false
  }
}

function resolveRedirectPath() {
  const redirect = typeof route.query.redirect === 'string' ? route.query.redirect : ''
  return redirect || '/'
}
</script>
