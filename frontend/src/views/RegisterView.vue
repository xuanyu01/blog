<!--
注册页面组件。
-->
<template>
  <section class="page-block auth-page">
    <div class="register-container">
      <h2>用户注册</h2>

      <div v-if="redirecting" class="auth-form">
        <p class="feedback success">{{ message }}</p>
      </div>

      <template v-else>
        <form class="auth-form" @submit.prevent="handleSubmit">
          <div class="form-group">
            <label for="register-username">用户名：</label>
            <input id="register-username" v-model.trim="form.username" type="text" placeholder="请输入用户名" required />
          </div>

          <div class="form-group">
            <label for="register-password">密码：</label>
            <input id="register-password" v-model="form.password" type="password" placeholder="请输入密码" required />
          </div>

          <div class="form-group">
            <button type="submit" :disabled="loading">{{ loading ? '注册中...' : '注册' }}</button>
          </div>

          <a class="switch-link" @click.prevent="router.push('/login')" href="/login">已有账号？去登录</a>
        </form>

        <p v-if="message" :class="success ? 'feedback success' : 'feedback error'">{{ message }}</p>
      </template>
    </div>
  </section>
</template>

<script setup>
import { onMounted, reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { register } from '../api/client'
import { refreshCurrentUser } from '../store/appStore'

const router = useRouter()
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
  message.value = '你已经登录，3 秒后自动返回首页。'

  setTimeout(() => {
    router.push('/')
  }, 3000)
})

async function handleSubmit() {
  loading.value = true
  message.value = ''

  try {
    await register(form)
    success.value = true
    message.value = '注册成功，2 秒后跳转到登录页。'

    setTimeout(() => {
      router.push('/login')
    }, 2000)
  } catch (error) {
    success.value = false
    message.value = translateAuthMessage(error.message)
  } finally {
    loading.value = false
  }
}

function translateAuthMessage(raw = '') {
  const messages = {
    'username and password are required': '请输入用户名和密码。',
    'user already exists': '该用户名已被注册。',
    'request failed': '请求失败，请稍后再试。'
  }
  return messages[raw] || '操作失败，请稍后再试。'
}
</script>
