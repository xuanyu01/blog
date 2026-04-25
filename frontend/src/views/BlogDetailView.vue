<!--
/*
  这个文件定义博客详情页面组件
*/
-->
<template>
  <section class="page-block blog-detail-page">
    <div v-if="loading" class="empty-state">
      <h3>正在加载博客详情</h3>
      <p>请稍候，正在根据文章编号请求最新内容</p>
    </div>

    <div v-else-if="blog" class="detail-shell">
      <RouterLink to="/" class="back-link">返回首页</RouterLink>

      <article class="detail-card">
        <div class="detail-meta-row">
          <div class="detail-meta-group">
            <div class="detail-meta" v-if="createdAtText">{{ createdAtText }}</div>
            <div class="detail-author">作者 {{ blog.AuthorUsername || 'unknown' }}</div>
          </div>

          <div class="detail-actions" v-if="canEdit || canDelete">
            <button
              v-if="canEdit"
              type="button"
              class="edit-btn"
              @click="handleEdit"
            >
              编辑博客
            </button>

            <button
              v-if="canDelete"
              type="button"
              class="delete-btn"
              :disabled="deleting"
              @click="handleDelete"
            >
              {{ deleting ? '删除中...' : '删除博客' }}
            </button>
          </div>
        </div>

        <h1 class="detail-title">{{ blog.Title }}</h1>
        <div class="detail-content">{{ blog.Content }}</div>

        <p v-if="message" :class="success ? 'feedback success' : 'feedback error'">
          {{ message }}
        </p>
      </article>
    </div>

    <div v-else class="empty-state">
      <h3>{{ errorTitle }}</h3>
      <p>{{ errorDescription }}</p>
      <RouterLink to="/" class="back-link">返回首页</RouterLink>
    </div>
  </section>
</template>

<script setup>
/*
  这个页面负责根据路由参数按 id 请求单篇博客的完整内容
*/
import { computed, onMounted, ref, watch } from 'vue'
import { RouterLink, useRoute, useRouter } from 'vue-router'
import { deleteBlog, getBlogById } from '../api/client'
import { appStore as store, refreshAppState, refreshCurrentUser } from '../store/appStore'

const route = useRoute()
const router = useRouter()
const loading = ref(true)
const deleting = ref(false)
const blog = ref(null)
const message = ref('')
const success = ref(false)
const errorState = ref('')

const blogID = computed(() => Number.parseInt(route.params.id, 10))

const createdAtText = computed(() => {
  if (!blog.value?.CreatedAt) {
    return ''
  }

  const date = new Date(blog.value.CreatedAt)
  if (Number.isNaN(date.getTime())) {
    return ''
  }

  return date.toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit'
  })
})

const canDelete = computed(() => {
  if (!blog.value || !store.user.isLogin) {
    return false
  }

  if (store.user.permission === 'admin' || store.user.permission === 'user_admin') {
    return true
  }

  return store.user.userName === blog.value.AuthorUsername
})

const canEdit = computed(() => canDelete.value)

const errorTitle = computed(() => {
  switch (errorState.value) {
    case 'invalid':
      return '无效的博客编号'
    case 'forbidden':
      return '你没有权限查看这篇博客'
    case 'not-found':
      return '没有找到这篇博客'
    case 'deleted':
      return '这篇博客已被删除'
    default:
      return '加载博客详情失败'
  }
})

const errorDescription = computed(() => {
  switch (errorState.value) {
    case 'invalid':
      return '当前链接中的博客编号不正确，请返回列表重新进入。'
    case 'forbidden':
      return '这篇博客当前不可访问，可能需要更高权限才能查看。'
    case 'not-found':
      return '这篇博客可能不存在，或者已经被删除。'
    case 'deleted':
      return '删除操作已经完成，这篇博客不再可访问。'
    default:
      return '请稍后重试，或者返回首页查看其他博客。'
  }
})

onMounted(async () => {
  await refreshCurrentUser()
  await loadBlog()
})

watch(blogID, async () => {
  await loadBlog()
})

async function loadBlog() {
  message.value = ''
  success.value = false
  blog.value = null

  if (!Number.isInteger(blogID.value) || blogID.value <= 0) {
    loading.value = false
    errorState.value = 'invalid'
    return
  }

  loading.value = true
  errorState.value = ''

  try {
    blog.value = await getBlogById(blogID.value)
  } catch (error) {
    if (error.status === 403) {
      errorState.value = 'forbidden'
    } else if (error.status === 404) {
      errorState.value = message.value === '博客已删除，即将返回首页' ? 'deleted' : 'not-found'
    } else {
      errorState.value = 'unknown'
    }
  } finally {
    loading.value = false
  }
}

function handleEdit() {
  if (!blog.value || !canEdit.value) {
    return
  }

  router.push(`/blog/${blog.value.ID}/edit`)
}

async function handleDelete() {
  if (!blog.value || !canDelete.value) {
    return
  }

  const confirmed = window.confirm(`确定删除《${blog.value.Title}》吗？`)
  if (!confirmed) {
    return
  }

  deleting.value = true
  message.value = ''

  try {
    await deleteBlog(blog.value.ID)
    await refreshAppState()
    success.value = true
    message.value = '博客已删除，即将返回首页'

    blog.value = null
    errorState.value = 'deleted'

    setTimeout(() => {
      router.push('/')
    }, 800)
  } catch (error) {
    success.value = false
    if (error.status === 404) {
      errorState.value = 'deleted'
      message.value = '这篇博客已经不存在了'
    } else if (error.status === 403) {
      message.value = '你没有权限删除这篇博客'
    } else {
      message.value = error.message
    }
  } finally {
    deleting.value = false
  }
}
</script>

<style scoped>
.detail-shell {
  display: grid;
  gap: 18px;
}

.back-link {
  width: fit-content;
  color: #9c6a43;
  font-weight: 700;
}

.detail-card {
  display: grid;
  gap: 20px;
  padding: 32px;
  border-radius: 26px;
  background: rgba(255, 255, 255, 0.92);
  box-shadow: 0 22px 46px rgba(40, 58, 80, 0.1);
}

.detail-meta-row {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
}

.detail-meta-group {
  display: grid;
  gap: 8px;
}

.detail-actions {
  display: flex;
  gap: 12px;
}

.detail-meta,
.detail-author {
  color: #9c6a43;
  font-size: 14px;
  font-weight: 700;
}

.edit-btn,
.delete-btn {
  border: none;
  border-radius: 14px;
  padding: 10px 14px;
  color: #fff;
  font-weight: 600;
  cursor: pointer;
}

.edit-btn {
  background: #203040;
}

.delete-btn {
  background: #a53a3a;
}

.delete-btn:disabled {
  opacity: 0.72;
  cursor: wait;
}

.detail-title {
  margin: 0;
  color: #203040;
  font-size: clamp(30px, 4vw, 42px);
  line-height: 1.2;
  word-break: break-word;
}

.detail-content {
  color: #415164;
  font-size: 16px;
  line-height: 1.9;
  white-space: pre-wrap;
  word-break: break-word;
}

@media (max-width: 640px) {
  .detail-meta-row {
    flex-direction: column;
  }

  .detail-actions {
    width: 100%;
    flex-direction: column;
  }
}
</style>
