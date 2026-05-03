<template>
  <section v-if="ready" class="page-block user-profile-page">
    <div v-if="store.user.isLogin" class="container profile-shell">
      <div class="profile-head">
        <div class="profile-identity">
          <img
            v-if="store.user.imageRoute"
            :src="`/img/${store.user.imageRoute}`"
            :alt="displayNameForView"
            class="profile-avatar"
          />
          <div v-else class="profile-avatar profile-avatar-fallback">{{ initials }}</div>

          <div class="profile-copy">
            <h1>{{ displayNameForView }}</h1>
            <p>@{{ store.user.userName }}</p>
            <span>{{ permissionText }}</span>
          </div>
        </div>

        <div class="profile-actions">
          <RouterLink :to="editPath" class="secondary-link">编辑资料</RouterLink>
          <RouterLink to="/blog/create" class="primary-link">写博客</RouterLink>
        </div>
      </div>

      <div class="profile-tabs" role="tablist" aria-label="用户博客列表">
        <button
          v-for="tab in tabs"
          :key="tab.key"
          type="button"
          :class="{ active: activeTab === tab.key }"
          @click="activeTab = tab.key"
        >
          <strong>{{ tab.label }}</strong>
          <span>{{ tab.count }}</span>
        </button>
      </div>

      <div v-if="loading" class="empty-card">
        <h3>正在加载个人主页</h3>
        <p>请稍候，正在获取你的博客列表。</p>
      </div>

      <div v-else-if="activeItems.length" class="profile-list">
        <BlogCard
          v-for="blog in activeItems"
          :key="blog.id || blog.ID"
          :blog="blog"
        />
      </div>

      <div v-else class="empty-card">
        <h3>{{ emptyTitle }}</h3>
        <p>{{ emptyText }}</p>
      </div>
    </div>

    <div v-else class="empty-card">
      <h3>你还没有登录</h3>
      <p>请先登录后再查看用户主页。</p>
    </div>
  </section>
</template>

<script setup>
import { computed, onMounted, ref, watch } from 'vue'
import { RouterLink, useRoute, useRouter } from 'vue-router'
import BlogCard from '../components/BlogCard.vue'
import { getCurrentUserBlogs, getCurrentUserFavorites, getCurrentUserLikes } from '../api/client'
import { appStore as store, refreshCurrentUser } from '../store/appStore'

const route = useRoute()
const router = useRouter()
const ready = ref(false)
const loading = ref(false)
const activeTab = ref('published')
const publishedBlogs = ref([])
const likedBlogs = ref([])
const favoriteBlogs = ref([])

const displayNameForView = computed(() => store.user.displayName || store.user.userName || '用户')
const initials = computed(() => displayNameForView.value.slice(0, 1).toUpperCase())
const editPath = computed(() => `/user/${store.user.id}/edit`)

const permissionText = computed(() => {
  switch (store.user.permission) {
    case 'admin':
      return '系统管理员'
    case 'user_admin':
      return '用户管理员'
    default:
      return '普通用户'
  }
})

const tabs = computed(() => [
  { key: 'published', label: '我发布的', count: publishedBlogs.value.length },
  { key: 'liked', label: '我点赞的', count: likedBlogs.value.length },
  { key: 'favorites', label: '我收藏的', count: favoriteBlogs.value.length }
])

const activeItems = computed(() => {
  switch (activeTab.value) {
    case 'liked':
      return likedBlogs.value
    case 'favorites':
      return favoriteBlogs.value
    default:
      return publishedBlogs.value
  }
})

const emptyTitle = computed(() => {
  switch (activeTab.value) {
    case 'liked':
      return '你还没有点赞博客'
    case 'favorites':
      return '你还没有收藏博客'
    default:
      return '你还没有发布博客'
  }
})

const emptyText = computed(() => {
  switch (activeTab.value) {
    case 'liked':
      return '看到喜欢的文章时点一下点赞，这里会保存对应列表。'
    case 'favorites':
      return '收藏感兴趣的文章后，可以在这里快速找回。'
    default:
      return '发布后的博客会出现在这里，草稿仍然保留在草稿箱。'
  }
})

function routeUserID() {
  return Number.parseInt(route.params.id, 10)
}

async function ensureOwnProfile() {
  const user = await refreshCurrentUser()
  if (!user.isLogin) {
    ready.value = true
    return false
  }

  if (routeUserID() !== user.id) {
    await router.replace(`/user/${user.id}`)
    return false
  }

  return true
}

async function loadProfileLists() {
  loading.value = true
  try {
    const [publishedData, likedData, favoriteData] = await Promise.all([
      getCurrentUserBlogs({ page: 1, pageSize: 50, status: 'published' }),
      getCurrentUserLikes({ page: 1, pageSize: 50 }),
      getCurrentUserFavorites({ page: 1, pageSize: 50 })
    ])

    publishedBlogs.value = publishedData.items || []
    likedBlogs.value = likedData.items || []
    favoriteBlogs.value = favoriteData.items || []
  } finally {
    loading.value = false
  }
}

onMounted(async () => {
  const canLoad = await ensureOwnProfile()
  ready.value = true
  if (canLoad) {
    await loadProfileLists()
  }
})

watch(() => route.params.id, async () => {
  if (!ready.value) {
    return
  }

  const canLoad = await ensureOwnProfile()
  if (canLoad) {
    await loadProfileLists()
  }
})
</script>

<style scoped>
.profile-shell {
  display: grid;
  gap: 20px;
}

.profile-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 20px;
  padding: 24px;
  border-radius: 24px;
  background: rgba(255, 255, 255, 0.88);
  box-shadow: 0 18px 40px rgba(40, 58, 80, 0.08);
}

.profile-identity {
  display: flex;
  align-items: center;
  gap: 18px;
  min-width: 0;
}

.profile-avatar {
  width: 84px;
  height: 84px;
  border-radius: 50%;
  object-fit: cover;
  flex: 0 0 auto;
}

.profile-avatar-fallback {
  display: grid;
  place-items: center;
  background: #203040;
  color: #fff;
  font-size: 30px;
  font-weight: 700;
}

.profile-copy {
  display: grid;
  gap: 6px;
  min-width: 0;
}

.profile-copy h1,
.profile-copy p {
  margin: 0;
}

.profile-copy h1 {
  color: #203040;
  font-size: 32px;
  line-height: 1.2;
  word-break: break-word;
}

.profile-copy p {
  color: #5f6f82;
}

.profile-copy span {
  color: #8a5b36;
  font-size: 13px;
  font-weight: 700;
}

.profile-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
  justify-content: flex-end;
}

.profile-tabs {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 12px;
}

.profile-tabs button {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  min-height: 58px;
  padding: 14px 16px;
  border: 1px solid #dce4ec;
  border-radius: 16px;
  background: rgba(255, 255, 255, 0.82);
  color: #203040;
  cursor: pointer;
}

.profile-tabs button.active {
  border-color: #203040;
  background: #203040;
  color: #fff;
}

.profile-tabs span {
  min-width: 28px;
  text-align: right;
  font-weight: 700;
}

.profile-list {
  display: grid;
  gap: 18px;
}

.primary-link,
.secondary-link {
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

@media (max-width: 720px) {
  .profile-head,
  .profile-identity {
    align-items: flex-start;
  }

  .profile-head {
    flex-direction: column;
  }

  .profile-actions {
    justify-content: flex-start;
  }

  .profile-tabs {
    grid-template-columns: 1fr;
  }
}
</style>
