<template>
  <section v-if="ready" class="page-block admin-page">
    <div v-if="canEnter" class="admin-layout">
      <aside class="admin-sidebar">
        <div class="admin-sidebar-title">管理中心</div>
        <button type="button" class="admin-nav-item" :class="{ active: activeTab === 'users' }" @click="switchTab('users')">
          用户管理
        </button>
        <button type="button" class="admin-nav-item" :class="{ active: activeTab === 'blogs' }" @click="switchTab('blogs')">
          博客管理
        </button>
        <button type="button" class="admin-nav-item" :class="{ active: activeTab === 'categories' }" @click="switchTab('categories')">
          分类管理
        </button>
      </aside>

      <div class="admin-content">
        <section v-if="activeTab === 'users'" class="panel">
          <div class="panel-head">
            <div>
              <h2>用户管理</h2>
              <p>分页查看站内用户，并执行权限调整或删除操作。</p>
            </div>
          </div>

          <div v-if="users.length" class="table-wrap">
            <table class="admin-table">
              <thead>
                <tr>
                  <th>用户名</th>
                  <th>显示名称</th>
                  <th>权限</th>
                  <th>操作</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="item in users" :key="item.username">
                  <td>{{ item.username }}</td>
                  <td>{{ item.displayName || '-' }}</td>
                  <td>
                    <template v-if="isAdmin && item.username !== store.user.userName">
                      <select
                        :value="permissionDrafts[item.username] || item.permission"
                        class="field-select compact-select"
                        @change="setPermissionDraft(item.username, $event.target.value)"
                      >
                        <option value="user">普通用户</option>
                        <option value="user_admin">用户管理员</option>
                        <option v-if="item.permission === 'admin'" value="admin">系统管理员</option>
                      </select>
                    </template>
                    <template v-else>
                      {{ permissionLabel(item.permission) }}
                    </template>
                  </td>
                  <td>
                    <div class="action-row">
                      <button
                        v-if="canEditPermission(item)"
                        type="button"
                        class="secondary-btn"
                        :disabled="permissionUpdatingFor === item.username"
                        @click="handleUpdatePermission(item)"
                      >
                        {{ permissionUpdatingFor === item.username ? '保存中...' : '保存权限' }}
                      </button>
                      <button
                        v-if="canDeleteUser(item)"
                        type="button"
                        class="danger-btn"
                        :disabled="deletingUserFor === item.username"
                        @click="handleDeleteUser(item)"
                      >
                        {{ deletingUserFor === item.username ? '删除中...' : '删除用户' }}
                      </button>
                    </div>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>

          <div v-else class="empty-card admin-empty">
            <h3>暂无用户信息</h3>
            <p>当前没有可展示的用户信息。</p>
          </div>

          <div class="pager">
            <button type="button" class="secondary-btn" :disabled="userPage <= 1 || userLoading" @click="goToUserPage(userPage - 1)">上一页</button>
            <span>第 {{ userPage }} / {{ userTotalPages }} 页</span>
            <button type="button" class="secondary-btn" :disabled="userPage >= userTotalPages || userLoading" @click="goToUserPage(userPage + 1)">下一页</button>
          </div>
        </section>

        <section v-else-if="activeTab === 'blogs'" class="panel">
          <div class="panel-head">
            <div>
              <h2>博客管理</h2>
              <p>按关键词、作者、状态筛选文章，并执行删除和审查操作。</p>
            </div>
          </div>

          <form class="blog-filter-bar" @submit.prevent="applyBlogFilters">
            <input v-model.trim="blogFilters.keyword" class="field-input" type="text" placeholder="搜索标题、摘要、正文、作者" />
            <input v-model.trim="blogFilters.author" class="field-input" type="text" placeholder="按作者筛选" />
            <select v-model="blogFilters.status" class="field-select">
              <option value="">全部状态</option>
              <option value="draft">草稿</option>
              <option value="published">已发布</option>
              <option value="hidden">已隐藏</option>
            </select>
            <button type="submit" class="primary-btn" :disabled="blogLoading">搜索</button>
            <button type="button" class="secondary-btn" :disabled="blogLoading" @click="resetBlogFilters">重置</button>
          </form>

          <div v-if="blogs.length" class="table-wrap">
            <table class="admin-table">
              <thead>
                <tr>
                  <th>标题</th>
                  <th>作者</th>
                  <th>状态</th>
                  <th>置顶</th>
                  <th>发布时间</th>
                  <th>操作</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="item in blogs" :key="item.id">
                  <td>
                    <div class="blog-title-cell">
                      <strong>{{ item.title }}</strong>
                      <span class="blog-summary">{{ item.summary || '暂无摘要' }}</span>
                    </div>
                  </td>
                  <td>{{ item.authorUsername || '-' }}</td>
                  <td>
                    <select
                      :value="blogDraft(item.id).status"
                      class="field-select compact-select"
                      @change="setBlogDraft(item.id, 'status', $event.target.value)"
                    >
                      <option value="draft">草稿</option>
                      <option value="published">已发布</option>
                      <option value="hidden">已隐藏</option>
                    </select>
                  </td>
                  <td>
                    <label class="checkbox-label">
                      <input
                        type="checkbox"
                        :checked="blogDraft(item.id).isTop"
                        @change="setBlogDraft(item.id, 'isTop', $event.target.checked)"
                      />
                      <span>{{ blogDraft(item.id).isTop ? '是' : '否' }}</span>
                    </label>
                  </td>
                  <td>{{ formatDateTime(item.publishedAt) }}</td>
                  <td>
                    <div class="action-row">
                      <button type="button" class="secondary-btn" :disabled="reviewingBlogFor === item.id" @click="handleReviewBlog(item)">
                        {{ reviewingBlogFor === item.id ? '保存中...' : '保存审查' }}
                      </button>
                      <button type="button" class="danger-btn" :disabled="deletingBlogFor === item.id" @click="handleDeleteBlog(item)">
                        {{ deletingBlogFor === item.id ? '删除中...' : '删除文章' }}
                      </button>
                    </div>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>

          <div v-else class="empty-card admin-empty">
            <h3>暂无博客内容</h3>
            <p>当前筛选条件下没有可管理的文章。</p>
          </div>

          <div class="pager">
            <button type="button" class="secondary-btn" :disabled="blogPage <= 1 || blogLoading" @click="goToBlogPage(blogPage - 1)">上一页</button>
            <span>第 {{ blogPage }} / {{ blogTotalPages }} 页</span>
            <button type="button" class="secondary-btn" :disabled="blogPage >= blogTotalPages || blogLoading" @click="goToBlogPage(blogPage + 1)">下一页</button>
          </div>
        </section>

        <section v-else class="panel">
          <div class="panel-head">
            <div>
              <h2>分类管理</h2>
              <p>管理员和用户管理员可以维护分类，方便文章归档和筛选。</p>
            </div>
          </div>

          <form class="category-form" @submit.prevent="handleCreateCategory">
            <input v-model.trim="newCategoryName" class="field-input" type="text" maxlength="50" placeholder="请输入新的分类名称" />
            <button type="submit" class="primary-btn" :disabled="categorySaving">新增分类</button>
          </form>

          <div v-if="categories.length" class="table-wrap">
            <table class="admin-table">
              <thead>
                <tr>
                  <th>名称</th>
                  <th>状态</th>
                  <th>文章数</th>
                  <th>操作</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="item in categories" :key="item.id">
                  <td>
                    <input
                      v-model.trim="categoryDrafts[item.id]"
                      class="field-input"
                      type="text"
                      maxlength="50"
                    />
                  </td>
                  <td>{{ item.status === 'active' ? '可用' : '已隐藏' }}</td>
                  <td>{{ item.postCount }}</td>
                  <td>
                    <div class="action-row">
                      <button type="button" class="secondary-btn" :disabled="categorySavingId === item.id" @click="handleUpdateCategory(item)">
                        {{ categorySavingId === item.id ? '保存中...' : '保存' }}
                      </button>
                      <button
                        v-if="item.status === 'active'"
                        type="button"
                        class="danger-btn"
                        :disabled="categoryDeletingId === item.id"
                        @click="handleDeleteCategory(item)"
                      >
                        {{ categoryDeletingId === item.id ? '隐藏中...' : '隐藏' }}
                      </button>
                    </div>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </section>

        <p v-if="message" :class="success ? 'feedback success' : 'feedback error'">
          {{ message }}
        </p>
      </div>
    </div>

    <div v-else class="empty-card">
      <h3>无权使用后台</h3>
      <p>只有用户管理员或系统管理员可以进入这里。</p>
    </div>
  </section>
</template>

<script setup>
import { computed, onMounted, reactive, ref } from 'vue'
import {
  createCategory,
  deleteBlog,
  deleteCategory,
  deleteManagedUser,
  getAdminBlogs,
  getAdminUsers,
  getManageCategories,
  reviewAdminBlog,
  updateCategory,
  updateUserPermission
} from '../api/client'
import { appStore as store, refreshCurrentUser, refreshTaxonomy } from '../store/appStore'

const ready = ref(false)
const activeTab = ref('users')
const message = ref('')
const success = ref(false)

const userLoading = ref(false)
const users = ref([])
const userPage = ref(1)
const userPageSize = ref(8)
const userTotal = ref(0)
const deletingUserFor = ref('')
const permissionUpdatingFor = ref('')
const permissionDrafts = reactive({})

const blogLoading = ref(false)
const blogs = ref([])
const blogPage = ref(1)
const blogPageSize = ref(8)
const blogTotal = ref(0)
const deletingBlogFor = ref('')
const reviewingBlogFor = ref('')
const blogFilters = reactive({
  keyword: '',
  author: '',
  status: ''
})
const blogReviewDrafts = reactive({})

const categories = ref([])
const newCategoryName = ref('')
const categorySaving = ref(false)
const categorySavingId = ref(0)
const categoryDeletingId = ref(0)
const categoryDrafts = reactive({})

const canEnter = computed(() => store.user.permission === 'admin' || store.user.permission === 'user_admin')
const isAdmin = computed(() => store.user.permission === 'admin')
const userTotalPages = computed(() => Math.max(1, Math.ceil(userTotal.value / userPageSize.value)))
const blogTotalPages = computed(() => Math.max(1, Math.ceil(blogTotal.value / blogPageSize.value)))

function switchTab(tab) {
  activeTab.value = tab
  clearMessage()
}

function clearMessage() {
  message.value = ''
}

function showMessage(text, isSuccess) {
  message.value = text
  success.value = isSuccess
}

function permissionLabel(permission) {
  switch (permission) {
    case 'admin':
      return '系统管理员'
    case 'user_admin':
      return '用户管理员'
    default:
      return '普通用户'
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

function setPermissionDraft(username, permission) {
  permissionDrafts[username] = permission
}

function canEditPermission(item) {
  if (!isAdmin.value || item.username === store.user.userName) {
    return false
  }
  return item.permission !== 'admin'
}

function canDeleteUser(item) {
  if (item.username === store.user.userName) {
    return false
  }
  if (store.user.permission === 'admin') {
    return item.permission !== 'admin'
  }
  if (store.user.permission === 'user_admin') {
    return item.permission === 'user'
  }
  return false
}

function ensureBlogDraft(item) {
  if (!blogReviewDrafts[item.id]) {
    blogReviewDrafts[item.id] = {
      status: item.status || 'draft',
      isTop: Boolean(item.isTop)
    }
  }
  return blogReviewDrafts[item.id]
}

function blogDraft(blogID) {
  return blogReviewDrafts[blogID] || { status: 'draft', isTop: false }
}

function setBlogDraft(blogID, field, value) {
  if (!blogReviewDrafts[blogID]) {
    blogReviewDrafts[blogID] = { status: 'draft', isTop: false }
  }
  blogReviewDrafts[blogID][field] = value
}

function normalizeBlog(item) {
  return {
    id: item.id ?? item.ID ?? 0,
    title: item.title ?? item.Title ?? '',
    summary: item.summary ?? item.Summary ?? '',
    authorUsername: item.authorUsername ?? item.AuthorUsername ?? '',
    status: item.status ?? item.Status ?? 'draft',
    isTop: Boolean(item.isTop ?? item.IsTop),
    publishedAt: item.publishedAt ?? item.PublishedAt ?? ''
  }
}

function formatDateTime(value) {
  if (!value) {
    return '-'
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

async function loadUsers(targetPage = userPage.value) {
  userLoading.value = true
  try {
    const data = await getAdminUsers(targetPage, userPageSize.value)
    users.value = data.items || []
    userPage.value = data.page || targetPage
    userPageSize.value = data.pageSize || userPageSize.value
    userTotal.value = data.total || 0
    for (const item of users.value) {
      permissionDrafts[item.username] = item.permission
    }
  } finally {
    userLoading.value = false
  }
}

async function loadBlogs(targetPage = blogPage.value) {
  blogLoading.value = true
  try {
    const data = await getAdminBlogs({
      page: targetPage,
      pageSize: blogPageSize.value,
      keyword: blogFilters.keyword,
      author: blogFilters.author,
      status: blogFilters.status
    })

    blogs.value = (data.items || []).map(normalizeBlog)
    blogPage.value = data.page || targetPage
    blogPageSize.value = data.pageSize || blogPageSize.value
    blogTotal.value = data.total || 0
    for (const item of blogs.value) {
      ensureBlogDraft(item)
    }
  } finally {
    blogLoading.value = false
  }
}

async function loadCategories() {
  const data = await getManageCategories()
  categories.value = data.items || []
  for (const item of categories.value) {
    categoryDrafts[item.id] = item.name
  }
}

async function goToUserPage(targetPage) {
  clearMessage()
  await loadUsers(targetPage)
}

async function goToBlogPage(targetPage) {
  clearMessage()
  await loadBlogs(targetPage)
}

async function applyBlogFilters() {
  clearMessage()
  await loadBlogs(1)
}

async function resetBlogFilters() {
  blogFilters.keyword = ''
  blogFilters.author = ''
  blogFilters.status = ''
  clearMessage()
  await loadBlogs(1)
}

async function handleUpdatePermission(item) {
  if (!canEditPermission(item)) {
    return
  }
  permissionUpdatingFor.value = item.username
  clearMessage()
  try {
    await updateUserPermission({
      username: item.username,
      permission: permissionDrafts[item.username] || item.permission
    })
    showMessage(`已更新 ${item.username} 的权限`, true)
    await loadUsers(userPage.value)
  } catch (error) {
    showMessage(error.message, false)
  } finally {
    permissionUpdatingFor.value = ''
  }
}

async function handleDeleteUser(item) {
  if (!canDeleteUser(item) || !window.confirm(`确定删除用户 ${item.username} 吗？`)) {
    return
  }
  deletingUserFor.value = item.username
  clearMessage()
  try {
    await deleteManagedUser(item.username)
    showMessage(`已删除用户 ${item.username}`, true)
    const remaining = users.value.length - 1
    const nextPage = remaining <= 0 && userPage.value > 1 ? userPage.value - 1 : userPage.value
    await loadUsers(nextPage)
  } catch (error) {
    showMessage(error.message, false)
  } finally {
    deletingUserFor.value = ''
  }
}

async function handleReviewBlog(item) {
  reviewingBlogFor.value = item.id
  clearMessage()
  try {
    const draft = blogDraft(item.id)
    await reviewAdminBlog(item.id, {
      status: draft.status,
      isTop: draft.isTop
    })
    showMessage(`已保存《${item.title}》的审查结果：${statusLabel(draft.status)}`, true)
    await loadBlogs(blogPage.value)
  } catch (error) {
    showMessage(error.message, false)
  } finally {
    reviewingBlogFor.value = ''
  }
}

async function handleDeleteBlog(item) {
  if (!window.confirm(`确定删除文章《${item.title}》吗？`)) {
    return
  }
  deletingBlogFor.value = item.id
  clearMessage()
  try {
    await deleteBlog(item.id)
    showMessage(`已删除文章《${item.title}》`, true)
    const remaining = blogs.value.length - 1
    const nextPage = remaining <= 0 && blogPage.value > 1 ? blogPage.value - 1 : blogPage.value
    await loadBlogs(nextPage)
  } catch (error) {
    showMessage(error.message, false)
  } finally {
    deletingBlogFor.value = ''
  }
}

async function handleCreateCategory() {
  if (!newCategoryName.value) {
    showMessage('请输入分类名称', false)
    return
  }
  categorySaving.value = true
  clearMessage()
  try {
    await createCategory({ name: newCategoryName.value })
    newCategoryName.value = ''
    showMessage('分类已创建', true)
    await Promise.all([loadCategories(), refreshTaxonomy()])
  } catch (error) {
    showMessage(error.message, false)
  } finally {
    categorySaving.value = false
  }
}

async function handleUpdateCategory(item) {
  categorySavingId.value = item.id
  clearMessage()
  try {
    await updateCategory(item.id, { name: categoryDrafts[item.id] || item.name })
    showMessage(`已更新分类 ${item.name}`, true)
    await Promise.all([loadCategories(), refreshTaxonomy()])
  } catch (error) {
    showMessage(error.message, false)
  } finally {
    categorySavingId.value = 0
  }
}

async function handleDeleteCategory(item) {
  if (!window.confirm(`确定隐藏分类 ${item.name} 吗？`)) {
    return
  }
  categoryDeletingId.value = item.id
  clearMessage()
  try {
    await deleteCategory(item.id)
    showMessage(`已隐藏分类 ${item.name}`, true)
    await Promise.all([loadCategories(), refreshTaxonomy()])
  } catch (error) {
    showMessage(error.message, false)
  } finally {
    categoryDeletingId.value = 0
  }
}

onMounted(async () => {
  await refreshCurrentUser()
  ready.value = true
  if (!canEnter.value) {
    return
  }
  await Promise.all([loadUsers(1), loadBlogs(1), loadCategories()])
})
</script>

<style scoped>
.admin-layout {
  display: grid;
  grid-template-columns: 220px minmax(0, 1fr);
  gap: 24px;
  align-items: start;
}

.admin-sidebar {
  display: grid;
  gap: 12px;
  padding: 20px;
  border-radius: 24px;
  background: rgba(255, 255, 255, 0.88);
  box-shadow: 0 18px 40px rgba(40, 58, 80, 0.08);
  position: sticky;
  top: 16px;
}

.admin-sidebar-title {
  color: #8a5b36;
  font-size: 12px;
  font-weight: 700;
  letter-spacing: 0.08em;
  text-transform: uppercase;
}

.admin-nav-item {
  border: none;
  border-radius: 16px;
  padding: 14px 16px;
  background: #eef2f6;
  color: #203040;
  font-weight: 600;
  text-align: left;
  cursor: pointer;
}

.admin-nav-item.active {
  background: #203040;
  color: #fff;
}

.panel {
  display: grid;
  gap: 20px;
  padding: 24px;
  border-radius: 24px;
  background: rgba(255, 255, 255, 0.88);
  box-shadow: 0 18px 40px rgba(40, 58, 80, 0.08);
}

.panel-head h2,
.panel-head p {
  margin: 0;
}

.panel-head p {
  color: #5f6f82;
  line-height: 1.7;
}

.blog-filter-bar,
.category-form {
  display: grid;
  grid-template-columns: minmax(220px, 2fr) minmax(160px, 1fr) minmax(140px, 0.8fr) auto auto;
  gap: 12px;
}

.category-form {
  grid-template-columns: 1fr auto;
}

.table-wrap {
  overflow-x: auto;
}

.admin-table {
  width: 100%;
  border-collapse: collapse;
}

.admin-table th,
.admin-table td {
  padding: 14px 12px;
  border-bottom: 1px solid #e6edf3;
  text-align: left;
  vertical-align: middle;
}

.field-input,
.field-select {
  width: 100%;
  padding: 10px 12px;
  border: 1px solid #d5dee8;
  border-radius: 12px;
  background: #fff;
  font-size: 14px;
}

.compact-select {
  min-width: 140px;
}

.checkbox-label,
.action-row,
.pager {
  display: flex;
  align-items: center;
  gap: 8px;
}

.pager {
  justify-content: flex-end;
  gap: 12px;
}

.primary-btn,
.secondary-btn,
.danger-btn {
  border: none;
  border-radius: 12px;
  padding: 10px 14px;
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

.danger-btn {
  background: #a53a3a;
  color: #fff;
}

.primary-btn:disabled,
.secondary-btn:disabled,
.danger-btn:disabled {
  opacity: 0.72;
  cursor: wait;
}

.blog-title-cell {
  display: grid;
  gap: 6px;
}

.blog-summary {
  color: #5f6f82;
  font-size: 13px;
}

@media (max-width: 1100px) {
  .blog-filter-bar {
    grid-template-columns: 1fr 1fr 1fr;
  }
}

@media (max-width: 900px) {
  .admin-layout {
    grid-template-columns: 1fr;
  }

  .admin-sidebar {
    position: static;
  }

  .blog-filter-bar,
  .category-form {
    grid-template-columns: 1fr;
  }
}
</style>
