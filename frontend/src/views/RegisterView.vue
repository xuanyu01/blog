<!--
/*
	该文件定义注册页面组件
*/
-->
<template>
  <section class="page-block auth-page">
    <div class="register-container">
      <h2>Register</h2>

      <div v-if="redirecting" class="auth-form">
        <p class="feedback success">{{ message }}</p>
      </div>

      <template v-else>
        <form class="auth-form" @submit.prevent="handleSubmit">
          <div class="form-group">
            <label for="register-username">Username:</label>
            <input
              id="register-username"
              v-model.trim="form.username"
              type="text"
              placeholder="Enter your username"
              required
            />
          </div>

          <div class="form-group">
            <label for="register-password">Password:</label>
            <input
              id="register-password"
              v-model="form.password"
              type="password"
              placeholder="Enter your password"
              required
            />
          </div>

          <div class="form-group">
            <button type="submit" :disabled="loading">
              {{ loading ? 'Loading...' : 'Register' }}
            </button>
          </div>

          <a class="switch-link" @click.prevent="router.push('/login')" href="/login">
            点击此处跳转到登录界面
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
	该页面负责收集注册信息并提交到后端
*/
import { onMounted, reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { register } from '../api/client'
import { refreshCurrentUser } from '../store/appStore'

const router = useRouter()
const loading = ref(false)
const message = ref('')
const success = ref(false)
const redirecting = ref(false)

// form 保存注册表单输入
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
  message.value = '你已经登录，3 秒后自动返回首页'

  setTimeout(() => {
    router.push('/')
  }, 3000)
})

// handleSubmit 提交注册表单
async function handleSubmit() {
  loading.value = true
  message.value = ''

  try {
    await register(form)
    success.value = true
    message.value = '注册成功，2 秒后跳转到登录页'

    // 给用户一个明确的成功反馈 再跳转到登录页
    setTimeout(() => {
      router.push('/login')
    }, 2000)
  } catch (error) {
    success.value = false
    message.value = error.message
  } finally {
    loading.value = false
  }
}
</script>
