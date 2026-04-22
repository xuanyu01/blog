<!--
/*
	这个文件定义博客详情页面组件
*/
-->
<template>
  <section class="page-block blog-detail-page">
    <div v-if="blog" class="detail-shell">
      <RouterLink to="/" class="back-link">返回首页</RouterLink>

      <article class="detail-card">
        <div class="detail-meta-row">
          <div class="detail-meta-group">
            <div class="detail-meta" v-if="createdAtText">{{ createdAtText }}</div>
            <div class="detail-author">作者 {{ blog.AuthorUsername || 'unknown' }}</div>
          </div>

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

        <h1 class="detail-title">{{ blog.Title }}</h1>
        <div class="detail-content">{{ blog.Content }}</div>

        <p v-if="message" :class="success ? 'feedback success' : 'feedback error'">
          {{ message }}
        </p>
      </article>
    </div>

    <div v-else class="empty-state">
      <h3>没有找到这篇博客</h3>
      <p>这篇博客可能不存在 或者当前列表还没有加载成功</p>
      <RouterLink to="/" class="back-link">返回首页</RouterLink>
    </div>
  </section>
</template>

<script setup>
/*
	这个页面负责根据路由参数展示单篇博客的完整内容
*/
import { computed, onMounted, ref } from 'vue'
import { RouterLink, useRoute, useRouter } from 'vue-router'
import { deleteBlog } from '../api/client'
import { appStore as store, refreshAppState, refreshCurrentUser } from '../store/appStore'

const route = useRoute()
const router = useRouter()
const deleting = ref(false)
const message = ref('')
const success = ref(false)

onMounted(async () => {
  await refreshCurrentUser()

  if (!store.blogs.length) {
    await refreshAppState()
  }
})

const blogID = computed(() => Number.parseInt(route.params.id, 10))

const blog = computed(() => {
  if (!Number.isInteger(blogID.value) || blogID.value <= 0) {
    return null
  }

  return store.blogs.find((item) => item.ID === blogID.value) || null
})

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

async function handleDelete() {
  if (!blog.value || !canDelete.value) {
    return
  }

  const confirmed = window.confirm(`确定删除《${blog.value.Title}》吗`)
  if (!confirmed) {
    return
  }

  deleting.value = true
  message.value = ''

  try {
    await deleteBlog(blog.value.ID)
    await refreshAppState()
    success.value = true
    message.value = '博客已删除 即将返回首页'

    setTimeout(() => {
      router.push('/')
    }, 800)
  } catch (error) {
    success.value = false
    message.value = error.message
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

.detail-meta,
.detail-author {
  color: #9c6a43;
  font-size: 14px;
  font-weight: 700;
}

.delete-btn {
  border: none;
  border-radius: 14px;
  padding: 10px 14px;
  background: #a53a3a;
  color: #fff;
  font-weight: 600;
  cursor: pointer;
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
}
</style>
