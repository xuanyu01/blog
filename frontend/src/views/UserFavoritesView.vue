<template>
  <section v-if="ready" class="page-block user-page">
    <div v-if="store.user.isLogin" class="container favorites-page">
      <div class="favorites-head">
        <div>
          <h2>我的收藏</h2>
          <p>查看自己收藏过的博客，随时回到感兴趣的内容。</p>
        </div>
        <div class="favorites-head-actions">
          <RouterLink to="/" class="secondary-link">返回首页</RouterLink>
          <RouterLink to="/user" class="secondary-link">用户中心</RouterLink>
        </div>
      </div>

      <div v-if="loading" class="empty-card">
        <h3>正在加载收藏</h3>
        <p>请稍候，正在获取你的收藏列表。</p>
      </div>

      <div v-else-if="favorites.length" class="favorite-grid">
        <article v-for="item in favorites" :key="item.id" class="favorite-card">
          <div class="favorite-meta">
            <span class="favorite-status">{{ statusLabel(item.status) }}</span>
            <span>{{ formatDate(item.publishedAt || item.updatedAt || item.createdAt) }}</span>
          </div>

          <h3>{{ item.title || '未命名博客' }}</h3>
          <p>{{ item.summary || '这篇博客还没有摘要。' }}</p>

          <div class="favorite-extra">
            <span>作者：{{ item.authorUsername || 'unknown' }}</span>
            <span>收藏 {{ item.stats.favoriteCount }}</span>
          </div>

          <div class="favorite-actions">
            <RouterLink :to="`/blog/${item.id}`" class="primary-link">查看详情</RouterLink>
            <button
              type="button"
              class="danger-btn"
              :disabled="cancelingID === item.id"
              @click="handleUnfavorite(item)"
            >
              {{ cancelingID === item.id ? '取消中...' : '取消收藏' }}
            </button>
          </div>
        </article>
      </div>

      <div v-else class="empty-card">
        <h3>你还没有收藏博客</h3>
        <p>看到感兴趣的文章时点一下“收藏”，这里就会出现对应内容。</p>
      </div>
    </div>

    <div v-else class="empty-card">
      <h3>你还没有登录</h3>
      <p>请先登录后再查看自己的收藏。</p>
    </div>
  </section>
</template>

<script setup>
import { onMounted, ref } from 'vue'
import { RouterLink, useRouter } from 'vue-router'
import { getCurrentUserFavorites, toggleBlogFavorite } from '../api/client'
import { appStore as store, refreshCurrentUser } from '../store/appStore'

const router = useRouter()
const ready = ref(false)
const loading = ref(false)
const cancelingID = ref(0)
const favorites = ref([])

function normalizeBlog(item = {}) {
  return {
    id: item.id ?? item.ID ?? 0,
    title: item.title ?? item.Title ?? '',
    summary: item.summary ?? item.Summary ?? '',
    status: item.status ?? item.Status ?? 'draft',
    authorUsername: item.authorUsername ?? item.AuthorUsername ?? '',
    createdAt: item.createdAt ?? item.CreatedAt ?? '',
    updatedAt: item.updatedAt ?? item.UpdatedAt ?? '',
    publishedAt: item.publishedAt ?? item.PublishedAt ?? '',
    stats: {
      favoriteCount: Number(item.stats?.favoriteCount ?? item.Stats?.FavoriteCount ?? item.favoriteCount ?? 0)
    }
  }
}

function statusLabel(status) {
  switch (status) {
    case 'published':
      return '已发布'
    case 'hidden':
      return '已隐藏'
    default:
      return '草稿'
  }
}

function formatDate(value) {
  if (!value) {
    return '刚刚更新'
  }

  const date = new Date(value)
  if (Number.isNaN(date.getTime())) {
    return String(value)
  }

  return date.toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit'
  })
}

onMounted(async () => {
  const user = await refreshCurrentUser()
  ready.value = true

  if (!user.isLogin) {
    setTimeout(() => router.push('/login'), 1200)
    return
  }

  loading.value = true
  try {
    const data = await getCurrentUserFavorites({
      page: 1,
      pageSize: 50
    })
    favorites.value = (data.items || []).map(normalizeBlog)
  } finally {
    loading.value = false
  }
})

async function handleUnfavorite(item) {
  if (!item?.id || cancelingID.value) {
    return
  }

  cancelingID.value = item.id
  try {
    const result = await toggleBlogFavorite(item.id)
    if (!result.active) {
      favorites.value = favorites.value.filter((favorite) => favorite.id !== item.id)
      return
    }

    favorites.value = favorites.value.map((favorite) => {
      if (favorite.id !== item.id) {
        return favorite
      }
      return {
        ...favorite,
        stats: {
          ...favorite.stats,
          favoriteCount: Number(result.favoriteCount ?? favorite.stats.favoriteCount)
        }
      }
    })
  } finally {
    cancelingID.value = 0
  }
}
</script>

<style scoped>
.favorites-page {
  display: grid;
  gap: 20px;
}

.favorites-head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
  padding: 24px;
  border-radius: 24px;
  background: rgba(255, 255, 255, 0.88);
  box-shadow: 0 18px 40px rgba(40, 58, 80, 0.08);
}

.favorites-head h2,
.favorites-head p {
  margin: 0;
}

.favorites-head p {
  margin-top: 8px;
  color: #5f6f82;
}

.favorites-head-actions,
.favorite-actions {
  display: flex;
  gap: 12px;
  flex-wrap: wrap;
}

.favorite-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(260px, 1fr));
  gap: 18px;
}

.favorite-card {
  display: grid;
  gap: 12px;
  padding: 22px;
  border-radius: 22px;
  background: rgba(255, 255, 255, 0.9);
  box-shadow: 0 18px 40px rgba(40, 58, 80, 0.08);
}

.favorite-card h3,
.favorite-card p {
  margin: 0;
}

.favorite-card p,
.favorite-extra {
  color: #5f6f82;
  line-height: 1.7;
}

.favorite-meta,
.favorite-extra {
  display: flex;
  justify-content: space-between;
  gap: 10px;
  flex-wrap: wrap;
  font-size: 13px;
}

.favorite-status {
  color: #8a5b36;
  font-weight: 700;
}

.primary-link,
.secondary-link,
.danger-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  padding: 10px 14px;
  border-radius: 14px;
  font-weight: 600;
}

.primary-link {
  background: #203040;
  color: #fff;
}

.secondary-link {
  background: #e8edf2;
  color: #203040;
}

.danger-btn {
  border: none;
  background: #a53a3a;
  color: #fff;
  cursor: pointer;
}

.danger-btn:disabled {
  opacity: 0.72;
  cursor: not-allowed;
}

@media (max-width: 700px) {
  .favorites-head {
    flex-direction: column;
  }
}
</style>
