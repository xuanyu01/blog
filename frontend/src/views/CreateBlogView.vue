<!--
/*
	这个文件定义博客发布页面组件
*/
-->
<template>
  <section class="page-block user-page" v-if="ready">
    <div class="container" v-if="user.isLogin">
      <div class="editor-card">
        <div class="editor-head">
          <div>
            <h2>发布博客</h2>
            <p>填写标题和内容后即可创建新的博客</p>
          </div>
          <button type="button" class="secondary-btn" @click="router.push('/user')">返回用户中心</button>
        </div>

        <form class="editor-form" @submit.prevent="handleSubmit">
          <label class="field">
            <span>标题</span>
            <input
              v-model.trim="form.title"
              type="text"
              maxlength="100"
              placeholder="请输入博客标题"
            />
          </label>

          <label class="field">
            <span>内容</span>
            <textarea
              v-model.trim="form.content"
              rows="12"
              placeholder="请输入博客内容"
            ></textarea>
          </label>

          <div class="editor-actions">
            <button type="submit" class="primary-btn" :disabled="submitting">
              {{ submitting ? '发布中...' : '发布博客' }}
            </button>
          </div>

          <p v-if="message" :class="success ? 'feedback success' : 'feedback error'">
            {{ message }}
          </p>
        </form>
      </div>
    </div>

    <div class="empty-card" v-else>
      <h3>你还没有登录</h3>
      <p>请先登录后再发布博客</p>
    </div>
  </section>
</template>

<script setup>
/*
	这个页面负责创建新的博客内容
*/
import { computed, onMounted, reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { createBlog } from '../api/client'
import { appStore as store, refreshAppState, refreshCurrentUser } from '../store/appStore'

const router = useRouter()
const ready = ref(false)
const submitting = ref(false)
const message = ref('')
const success = ref(false)

const form = reactive({
  title: '',
  content: ''
})

// user 是当前用户状态的计算属性别名
const user = computed(() => store.user)

// 页面挂载后同步当前用户状态
onMounted(async () => {
  const currentUser = await refreshCurrentUser()
  ready.value = true

  if (!currentUser.isLogin) {
    setTimeout(() => router.push('/login'), 1200)
  }
})

// handleSubmit 提交博客创建请求
async function handleSubmit() {
  submitting.value = true
  message.value = ''

  try {
    await createBlog({
      title: form.title,
      content: form.content
    })

    await refreshAppState()
    success.value = true
    message.value = '博客创建成功 1 秒后返回首页'
    form.title = ''
    form.content = ''

    setTimeout(() => {
      router.push('/')
    }, 1000)
  } catch (error) {
    success.value = false
    message.value = error.message
  } finally {
    submitting.value = false
  }
}
</script>

<style scoped>
.editor-card {
  display: grid;
  gap: 20px;
  padding: 28px;
  border-radius: 24px;
  background: rgba(255, 255, 255, 0.88);
  box-shadow: 0 18px 40px rgba(40, 58, 80, 0.08);
}

.editor-head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
}

.editor-head h2 {
  margin: 0 0 8px;
}

.editor-head p {
  margin: 0;
  color: #5f6f82;
}

.editor-form {
  display: grid;
  gap: 16px;
}

.field {
  display: grid;
  gap: 8px;
}

.field span {
  font-weight: 600;
  color: #203040;
}

.field input,
.field textarea {
  width: 100%;
  padding: 12px 14px;
  border: 1px solid #d5dee8;
  border-radius: 14px;
  font-size: 14px;
  resize: vertical;
}

.editor-actions {
  display: flex;
  justify-content: flex-end;
}

.primary-btn,
.secondary-btn {
  border: none;
  border-radius: 14px;
  padding: 12px 16px;
  font-weight: 600;
  cursor: pointer;
}

.primary-btn {
  background: #203040;
  color: #fff;
}

.secondary-btn {
  background: #e8edf2;
  color: #203040;
}

.primary-btn:disabled {
  cursor: wait;
  opacity: 0.72;
}
</style>
