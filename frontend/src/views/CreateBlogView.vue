<!--
/*
  这个文件定义博客发布和编辑页面组件
*/
-->
<template>
  <section class="page-block user-page" v-if="ready">
    <div class="container" v-if="user.isLogin">
      <div class="editor-card">
        <div class="editor-head">
          <div>
            <h2>{{ pageTitle }}</h2>
            <p>{{ pageDescription }}</p>
          </div>
          <button type="button" class="secondary-btn" @click="router.push('/user')">返回用户中心</button>
        </div>

        <div v-if="isEditMode && loadingBlog" class="feedback">正在加载要编辑的博客内容...</div>

        <div v-else-if="isEditMode && !canEdit" class="feedback error">
          {{ message || '你没有权限编辑这篇博客' }}
        </div>

        <form v-else class="editor-form" @submit.prevent="handleSubmit">
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
            <button type="submit" class="primary-btn" :disabled="submitting || loadingBlog">
              {{ submitText }}
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
      <p>请先登录后再发布或编辑博客</p>
    </div>
  </section>
</template>

<script setup>
/*
  这个页面负责创建新博客，也复用为博客编辑页
*/
import { computed, onMounted, reactive, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { createBlog, getBlogById, updateBlog } from '../api/client'
import { appStore as store, refreshAppState, refreshCurrentUser } from '../store/appStore'

const route = useRoute()
const router = useRouter()
const ready = ref(false)
const loadingBlog = ref(false)
const submitting = ref(false)
const message = ref('')
const success = ref(false)
const currentBlog = ref(null)

const form = reactive({
  title: '',
  content: ''
})

const editingBlogID = computed(() => Number.parseInt(route.params.id, 10))
const isEditMode = computed(() => route.name === 'blog-edit')
const user = computed(() => store.user)

const canEdit = computed(() => {
  if (!isEditMode.value) {
    return true
  }

  if (!currentBlog.value || !store.user.isLogin) {
    return false
  }

  if (store.user.permission === 'admin' || store.user.permission === 'user_admin') {
    return true
  }

  return store.user.userName === currentBlog.value.AuthorUsername
})

const pageTitle = computed(() => (isEditMode.value ? '编辑博客' : '发布博客'))
const pageDescription = computed(() => (
  isEditMode.value
    ? '系统会先带出原有内容，修改后需要再次确认发布'
    : '填写标题和内容后即可创建新的博客'
))

const submitText = computed(() => {
  if (submitting.value) {
    return isEditMode.value ? '更新中...' : '发布中...'
  }

  return isEditMode.value ? '确认并重新发布' : '发布博客'
})

onMounted(async () => {
  const currentUser = await refreshCurrentUser()
  if (!currentUser.isLogin) {
    ready.value = true
    setTimeout(() => router.push('/login'), 1200)
    return
  }

  if (isEditMode.value) {
    await loadEditableBlog()
  }

  ready.value = true
})

async function loadEditableBlog() {
  if (!Number.isInteger(editingBlogID.value) || editingBlogID.value <= 0) {
    success.value = false
    message.value = '无效的博客编号'
    return
  }

  loadingBlog.value = true
  try {
    currentBlog.value = await getBlogById(editingBlogID.value)

    if (!canEdit.value) {
      success.value = false
      message.value = '你没有权限编辑这篇博客'
      return
    }

    form.title = currentBlog.value.Title || ''
    form.content = currentBlog.value.Content || ''
  } catch (error) {
    success.value = false
    if (error.status === 404) {
      message.value = '未找到要编辑的博客'
    } else if (error.status === 403) {
      message.value = '你没有权限编辑这篇博客'
    } else {
      message.value = error.message
    }
  } finally {
    loadingBlog.value = false
  }
}

async function handleSubmit() {
  if (isEditMode.value && (!currentBlog.value || !canEdit.value)) {
    success.value = false
    message.value = '你没有权限编辑这篇博客'
    return
  }

  if (isEditMode.value) {
    const targetTitle = form.title || currentBlog.value?.Title || '未命名博客'
    const confirmed = window.confirm(`确认重新发布《${targetTitle}》吗？`)
    if (!confirmed) {
      return
    }
  }

  submitting.value = true
  message.value = ''

  try {
    if (isEditMode.value) {
      await updateBlog(editingBlogID.value, {
        title: form.title,
        content: form.content
      })
    } else {
      await createBlog({
        title: form.title,
        content: form.content
      })
    }

    await refreshAppState({ page: 1 })
    success.value = true
    message.value = isEditMode.value ? '博客更新成功，1 秒后跳转到详情页' : '博客创建成功，1 秒后返回首页'

    if (!isEditMode.value) {
      form.title = ''
      form.content = ''
    }

    setTimeout(() => {
      if (isEditMode.value) {
        router.push(`/blog/${editingBlogID.value}`)
        return
      }

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
