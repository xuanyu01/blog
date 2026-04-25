<!--
/*
  这个文件定义首页页面组件
*/
-->
<template>
  <section class="page-block home-page">
    <div class="home-layout">
      <div class="home-main">
        <div class="toolbar-card">
          <form class="search-form" @submit.prevent="handleSearch">
            <input
              v-model.trim="keywordInput"
              type="search"
              class="search-input"
              placeholder="搜索标题、内容或作者"
            />
            <button type="submit" class="search-btn">搜索</button>
            <button
              v-if="store.blogList.keyword"
              type="button"
              class="clear-btn"
              @click="handleClearSearch"
            >
              清除
            </button>
          </form>

          <div class="search-meta">
            <span>共 {{ store.blogList.total }} 篇</span>
            <span>第 {{ currentPage }} / {{ totalPages }} 页</span>
          </div>
        </div>

        <div v-if="store.loading" class="empty-state">
          <h3>正在加载博客列表</h3>
          <p>请稍候，正在获取最新内容</p>
        </div>

        <template v-else>
          <div class="blog-list" v-if="store.blogs.length">
            <BlogCard
              v-for="blog in store.blogs"
              :key="blog.ID"
              :blog="blog"
            />
          </div>

          <div class="empty-state" v-else>
            <h3>{{ emptyTitle }}</h3>
            <p>{{ emptyDescription }}</p>
          </div>

          <div class="pager" v-if="store.blogList.total > 0">
            <button
              type="button"
              class="pager-btn"
              :disabled="currentPage <= 1 || store.loading"
              @click="changePage(currentPage - 1)"
            >
              上一页
            </button>

            <div class="pager-status">
              <span>当前第 {{ currentPage }} 页</span>
              <span>每页 {{ store.blogList.pageSize }} 篇</span>
            </div>

            <button
              type="button"
              class="pager-btn"
              :disabled="currentPage >= totalPages || store.loading"
              @click="changePage(currentPage + 1)"
            >
              下一页
            </button>
          </div>
        </template>
      </div>

      <aside class="home-side">
        <div class="create-module">
          <p class="create-kicker">Blog Studio</p>
          <h3>创作</h3>
          <p class="create-text">写下新的标题和内容，发布一篇新的博客</p>

          <RouterLink
            v-if="store.user.isLogin"
            to="/blog/create"
            class="create-action"
          >
            进入创作
          </RouterLink>

          <RouterLink
            v-else
            to="/login"
            class="create-action"
          >
            登录后创作
          </RouterLink>
        </div>
      </aside>
    </div>
  </section>
</template>

<script setup>
/*
  这个页面负责加载并展示带分页和搜索的博客列表
*/
import { computed, onMounted, ref } from 'vue'
import { RouterLink } from 'vue-router'
import BlogCard from '../components/BlogCard.vue'
import { appStore as store, refreshAppState, refreshBlogList } from '../store/appStore'

const keywordInput = ref('')

const currentPage = computed(() => store.blogList.page || 1)
const totalPages = computed(() => {
  const total = store.blogList.total || 0
  const pageSize = store.blogList.pageSize || 10
  return Math.max(1, Math.ceil(total / pageSize))
})

const emptyTitle = computed(() => (
  store.blogList.keyword ? '没有找到匹配的博客' : '还没有内容'
))

const emptyDescription = computed(() => (
  store.blogList.keyword
    ? `没有找到和“${store.blogList.keyword}”相关的内容，试试换个关键词。`
    : '当前还没有可展示的博客内容。'
))

onMounted(async () => {
  keywordInput.value = store.blogList.keyword || ''
  await refreshAppState({ page: 1 })
})

async function handleSearch() {
  await refreshBlogList({
    page: 1,
    keyword: keywordInput.value
  })
}

async function handleClearSearch() {
  keywordInput.value = ''
  await refreshBlogList({
    page: 1,
    keyword: ''
  })
}

async function changePage(page) {
  if (page < 1 || page > totalPages.value || page === currentPage.value) {
    return
  }

  await refreshBlogList({
    page
  })
}
</script>

<style scoped>
.home-layout {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 280px;
  gap: 24px;
  align-items: start;
}

.home-main {
  min-width: 0;
  display: grid;
  gap: 20px;
}

.toolbar-card {
  display: grid;
  gap: 14px;
  padding: 20px 22px;
  border-radius: 22px;
  background: rgba(255, 255, 255, 0.9);
  box-shadow: 0 18px 40px rgba(40, 58, 80, 0.08);
}

.search-form {
  display: flex;
  gap: 12px;
}

.search-input {
  flex: 1;
  min-width: 0;
  padding: 12px 14px;
  border: 1px solid #d5dee8;
  border-radius: 14px;
  font-size: 14px;
}

.search-btn,
.clear-btn,
.pager-btn {
  border: none;
  border-radius: 14px;
  padding: 12px 16px;
  font-weight: 600;
  cursor: pointer;
}

.search-btn,
.pager-btn {
  background: #203040;
  color: #fff;
}

.clear-btn {
  background: #e8edf2;
  color: #203040;
}

.pager-btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.search-meta {
  display: flex;
  justify-content: space-between;
  gap: 12px;
  color: #5f6f82;
  font-size: 14px;
}

.blog-list {
  display: grid;
  gap: 20px;
}

.pager {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
}

.pager-status {
  display: grid;
  gap: 4px;
  text-align: center;
  color: #5f6f82;
  font-size: 14px;
}

.home-side {
  position: sticky;
  top: 16px;
}

.create-module {
  display: grid;
  gap: 12px;
  padding: 24px;
  border-radius: 24px;
  background: rgba(255, 255, 255, 0.88);
  box-shadow: 0 18px 40px rgba(40, 58, 80, 0.08);
}

.create-kicker {
  margin: 0;
  color: #9c6a43;
  font-size: 12px;
  font-weight: 700;
  letter-spacing: 0.08em;
  text-transform: uppercase;
}

.create-module h3 {
  margin: 0;
  font-size: 28px;
}

.create-text {
  margin: 0;
  color: #5f6f82;
  line-height: 1.7;
}

.create-action {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  padding: 12px 16px;
  border-radius: 14px;
  background: #203040;
  color: #fff;
  font-weight: 600;
}

@media (max-width: 900px) {
  .home-layout {
    grid-template-columns: 1fr;
  }

  .home-side {
    position: static;
  }
}

@media (max-width: 640px) {
  .search-form,
  .pager,
  .search-meta {
    flex-direction: column;
  }

  .pager-status {
    text-align: left;
  }
}
</style>
