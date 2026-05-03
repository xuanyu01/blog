<template>
  <section v-if="ready" class="page-block user-page">
    <div v-if="store.user.isLogin" class="container favorites-page">
      <div class="favorites-head"><div><h2>我的草稿</h2><p>查看自己保存的草稿，也可以继续编辑。</p></div><div class="favorites-head-actions"><RouterLink to="/blog/create" class="primary-link">新建博客</RouterLink><RouterLink :to="`/user/${store.user.id}`" class="secondary-link">返回用户主页</RouterLink></div></div>
      <div v-if="loading" class="empty-card"><h3>正在加载草稿</h3><p>请稍候，正在获取你保存的草稿列表。</p></div>
      <div v-else-if="drafts.length" class="favorite-grid"><article v-for="item in drafts" :key="item.id" class="favorite-card"><h3>{{ item.title || '未命名草稿' }}</h3><p>{{ item.summary || '这篇草稿还没有摘要。' }}</p><div class="favorite-actions"><RouterLink :to="`/blog/${item.id}/edit`" class="primary-link">继续编辑</RouterLink></div></article></div>
      <div v-else class="empty-card"><h3>你还没有草稿</h3><p>去创作页写点内容，保存为草稿后就会显示在这里。</p></div>
    </div>
    <div v-else class="empty-card"><h3>你还没有登录</h3><p>请先登录后再查看自己的草稿。</p></div>
  </section>
</template>
<script setup>
import { onMounted, ref } from 'vue'
import { RouterLink, useRouter } from 'vue-router'
import { getCurrentUserBlogs } from '../api/client'
import { appStore as store, refreshCurrentUser } from '../store/appStore'
const router = useRouter(); const ready = ref(false); const loading = ref(false); const drafts = ref([])
function normalizeBlog(item = {}) { return { id: item.id ?? item.ID ?? 0, title: item.title ?? item.Title ?? '', summary: item.summary ?? item.Summary ?? '' } }
onMounted(async () => { const user = await refreshCurrentUser(); ready.value = true; if (!user.isLogin) { setTimeout(() => router.push('/login'), 1200); return } loading.value = true; try { const data = await getCurrentUserBlogs({ page: 1, pageSize: 50, status: 'draft' }); drafts.value = (data.items || []).map(normalizeBlog) } finally { loading.value = false } })
</script>
<style scoped>
.drafts-page {
  display: grid;
  gap: 20px;
}

.drafts-head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
  padding: 24px;
  border-radius: 24px;
  background: rgba(255, 255, 255, 0.88);
  box-shadow: 0 18px 40px rgba(40, 58, 80, 0.08);
}

.drafts-head h2,
.drafts-head p {
  margin: 0;
}

.drafts-head p {
  margin-top: 8px;
  color: #5f6f82;
}

.drafts-head-actions,
.draft-actions {
  display: flex;
  gap: 12px;
  flex-wrap: wrap;
}

.draft-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(260px, 1fr));
  gap: 18px;
}

.draft-card {
  display: grid;
  gap: 12px;
  padding: 22px;
  border-radius: 22px;
  background: rgba(255, 255, 255, 0.9);
  box-shadow: 0 18px 40px rgba(40, 58, 80, 0.08);
}

.draft-card h3,
.draft-card p {
  margin: 0;
}

.draft-card p {
  color: #5f6f82;
  line-height: 1.7;
}

.draft-meta {
  display: flex;
  justify-content: space-between;
  gap: 10px;
  flex-wrap: wrap;
  color: #7a8797;
  font-size: 13px;
}

.draft-status {
  color: #8a5b36;
  font-weight: 700;
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

@media (max-width: 700px) {
  .drafts-head {
    flex-direction: column;
  }
}
</style>


