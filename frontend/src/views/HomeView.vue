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
              placeholder="搜索标题、标签、内容或作者"
            />
            <button type="button" class="filter-toggle-btn" @click="showFilters = !showFilters">
              筛选
            </button>
            <button type="submit" class="search-btn">搜索</button>
            <button
              v-if="hasActiveFilters"
              type="button"
              class="clear-btn"
              @click="handleClearFilters"
            >
              清除
            </button>
          </form>

          <div v-if="showFilters" class="filter-panel">
            <label class="filter-field">
              <span>分类</span>
              <select v-model.number="draftFilters.categoryId">
                <option :value="0">全部分类</option>
                <option v-for="item in store.taxonomy.categories" :key="item.id" :value="item.id">
                  {{ item.name }}
                </option>
              </select>
            </label>

            <label class="filter-field">
              <span>时间</span>
              <select v-model="draftFilters.archive">
                <option value="">全部时间</option>
                <option v-for="item in store.taxonomy.archives" :key="item.archive" :value="item.archive">
                  {{ item.archive }}（{{ item.count }}）
                </option>
              </select>
            </label>

            <div class="filter-actions">
              <button type="button" class="secondary-btn" @click="applyFilters">应用筛选</button>
            </div>
          </div>

          <div class="search-meta">
            <span>共 {{ store.blogList.total }} 篇</span>
            <span>第 {{ currentPage }} / {{ totalPages }} 页</span>
          </div>

          <div class="active-filters">
            <span v-if="activeCategoryName" class="filter-chip">分类：{{ activeCategoryName }}</span>
            <span v-if="store.blogList.archive" class="filter-chip">时间：{{ store.blogList.archive }}</span>
          </div>
        </div>

        <div v-if="store.loading" class="empty-state">
          <h3>正在加载博客列表</h3>
          <p>请稍候，正在获取最新内容。</p>
        </div>

        <template v-else>
          <div class="blog-list" v-if="store.blogs.length">
            <BlogCard
              v-for="blog in store.blogs"
              :key="blog.id || blog.ID"
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
    </div>
  </section>
</template>

<script setup>
import { computed, onMounted, reactive, ref } from 'vue'
import BlogCard from '../components/BlogCard.vue'
import { appStore as store, refreshAppState, refreshBlogList } from '../store/appStore'

const keywordInput = ref('')
const showFilters = ref(false)
const draftFilters = reactive({
  categoryId: 0,
  archive: ''
})

const currentPage = computed(() => store.blogList.page || 1)
const totalPages = computed(() => {
  const total = store.blogList.total || 0
  const pageSize = store.blogList.pageSize || 10
  return Math.max(1, Math.ceil(total / pageSize))
})

const hasActiveFilters = computed(() => (
  Boolean(store.blogList.keyword || store.blogList.categoryId || store.blogList.archive)
))

const activeCategoryName = computed(() => {
  if (!store.blogList.categoryId) {
    return ''
  }
  const target = store.taxonomy.categories.find((item) => Number(item.id) === Number(store.blogList.categoryId))
  return target?.name || ''
})

const emptyTitle = computed(() => (
  hasActiveFilters.value ? '没有找到匹配的博客' : '还没有内容'
))

const emptyDescription = computed(() => (
  hasActiveFilters.value
    ? '当前搜索或筛选条件下没有结果，可以尝试更换关键词、分类或时间。'
    : '当前还没有可展示的博客内容。'
))

onMounted(async () => {
  keywordInput.value = store.blogList.keyword || ''
  draftFilters.categoryId = Number(store.blogList.categoryId || 0)
  draftFilters.archive = store.blogList.archive || ''
  await refreshAppState({ page: 1 })
})

async function handleSearch() {
  await refreshBlogList({
    page: 1,
    keyword: keywordInput.value,
    categoryId: draftFilters.categoryId,
    archive: draftFilters.archive
  })
}

async function applyFilters() {
  await refreshBlogList({
    page: 1,
    keyword: keywordInput.value,
    categoryId: draftFilters.categoryId,
    archive: draftFilters.archive
  })
}

async function handleClearFilters() {
  keywordInput.value = ''
  draftFilters.categoryId = 0
  draftFilters.archive = ''
  await refreshBlogList({
    page: 1,
    keyword: '',
    categoryId: 0,
    archive: ''
  })
}

async function changePage(page) {
  if (page < 1 || page > totalPages.value || page === currentPage.value) {
    return
  }

  await refreshBlogList({ page })
}
</script>

<style scoped>
.home-layout {
  display: grid;
  grid-template-columns: minmax(0, 1fr);
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
  align-items: center;
}

.search-input {
  flex: 1;
  min-width: 0;
  padding: 12px 14px;
  border: 1px solid #d5dee8;
  border-radius: 14px;
  font-size: 14px;
}

.filter-panel {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 12px;
  padding: 14px;
  border-radius: 16px;
  background: #f7f3ec;
}

.filter-field {
  display: grid;
  gap: 6px;
}

.filter-field span {
  color: #5f6f82;
  font-size: 13px;
  font-weight: 700;
}

.filter-field select {
  width: 100%;
  padding: 10px 12px;
  border: 1px solid #d5dee8;
  border-radius: 12px;
  background: #fff;
}

.search-btn,
.clear-btn,
.pager-btn,
.filter-toggle-btn,
.secondary-btn {
  border: none;
  border-radius: 14px;
  padding: 12px 16px;
  font-weight: 600;
  cursor: pointer;
}

.search-btn,
.pager-btn,
.filter-toggle-btn {
  background: #203040;
  color: #fff;
}

.clear-btn,
.secondary-btn {
  background: #e8edf2;
  color: #203040;
}

.search-meta,
.active-filters {
  display: flex;
  justify-content: space-between;
  gap: 12px;
  color: #5f6f82;
  font-size: 14px;
  flex-wrap: wrap;
}

.filter-chip {
  padding: 6px 10px;
  border-radius: 999px;
  background: #f3ecde;
  color: #7b5427;
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

@media (max-width: 760px) {
  .filter-panel {
    grid-template-columns: 1fr;
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
