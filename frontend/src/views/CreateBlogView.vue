<template>
  <section class="page-block user-page" v-if="ready">
    <div class="container" v-if="user.isLogin">
      <div class="editor-card">
        <div class="editor-head">
          <div>
            <h2>{{ pageTitle }}</h2>
            <p>{{ pageDescription }}</p>
          </div>
          <div class="editor-head-actions">
            <button type="button" class="secondary-btn" @click="router.push(`/user/${store.user.id}`)">返回用户主页</button>
          </div>
        </div>

        <div v-if="isEditMode && loadingBlog" class="feedback">正在加载博客内容...</div>
        <div v-else-if="isEditMode && !canEdit" class="feedback error">
          {{ message || '你没有权限编辑这篇博客。' }}
        </div>

        <form v-else class="editor-form" @submit.prevent>
          <label class="field">
            <span>标题</span>
            <input v-model.trim="form.title" type="text" maxlength="100" placeholder="请输入博客标题" />
          </label>

          <label class="field">
            <span>分类</span>
            <select v-model.number="form.categoryId">
              <option :value="0">未分类</option>
              <option v-for="item in store.taxonomy.categories" :key="item.id" :value="item.id">
                {{ item.name }}
              </option>
            </select>
          </label>

          <label class="field">
            <span>标签</span>
            <input v-model.trim="form.tags" type="text" placeholder="多个标签用逗号分隔" />
          </label>

          <label class="field">
            <span>内容</span>
            <textarea v-model.trim="form.content" rows="14" placeholder="请使用 Markdown 编写博客内容"></textarea>
          </label>

          <label class="field compact-field">
            <span>状态</span>
            <select v-model="form.status">
              <option value="draft">草稿</option>
              <option value="published">发布</option>
            </select>
          </label>

          <label v-if="canManageAllBlogs" class="check-field">
            <input v-model="form.isTop" type="checkbox" />
            <span>置顶文章</span>
          </label>

          <div class="editor-actions">
            <button type="button" class="secondary-btn" :disabled="submitting" @click="handleSubmit('draft')">
              {{ submitting && pendingAction === 'draft' ? '保存中...' : draftButtonText }}
            </button>
            <button type="button" class="primary-btn" :disabled="submitting" @click="handleSubmit('published')">
              {{ submitting && pendingAction === 'published' ? '发布中...' : publishButtonText }}
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
      <p>请先登录后再发布博客。</p>
    </div>
  </section>
</template>

<script setup>
import { computed, onMounted, reactive, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { createBlog, getBlogById, updateBlog } from '../api/client'
import { appStore as store, refreshCurrentUser, refreshTaxonomy } from '../store/appStore'

const route = useRoute()
const router = useRouter()
const ready = ref(false)
const loadingBlog = ref(false)
const submitting = ref(false)
const pendingAction = ref('')
const message = ref('')
const success = ref(false)
const currentBlog = ref(null)

const form = reactive({
  title: '',
  content: '',
  status: 'draft',
  isTop: false,
  categoryId: 0,
  tags: ''
})

const user = computed(() => store.user)
const isEditMode = computed(() => route.name === 'blog-edit')
const editingBlogID = computed(() => Number.parseInt(route.params.id, 10))
const canManageAllBlogs = computed(() => store.user.permission === 'admin' || store.user.permission === 'user_admin')
const canEdit = computed(() => {
  if (!isEditMode.value) {
    return true
  }
  if (!currentBlog.value || !store.user.isLogin) {
    return false
  }
  return canManageAllBlogs.value || store.user.userName === currentBlog.value.authorUsername
})

const pageTitle = computed(() => (isEditMode.value ? '编辑博客' : '创作博客'))
const pageDescription = computed(() => (isEditMode.value ? '修改正文、分类、标签或发布状态。' : '填写标题和内容后即可创建新的博客。'))
const draftButtonText = computed(() => (isEditMode.value ? '保存为草稿' : '先存草稿'))
const publishButtonText = computed(() => (isEditMode.value ? '确认并发布' : '发布博客'))

function normalizeBlog(blog = {}) {
  return {
    id: blog.id ?? blog.ID ?? 0,
    title: blog.title ?? blog.Title ?? '',
    content: blog.content ?? blog.Content ?? '',
    status: blog.status ?? blog.Status ?? 'draft',
    isTop: Boolean(blog.isTop ?? blog.IsTop),
    categoryId: Number(blog.categoryId ?? blog.CategoryID ?? 0),
    authorUsername: blog.authorUsername ?? blog.AuthorUsername ?? '',
    tags: Array.isArray(blog.tags ?? blog.Tags) ? (blog.tags ?? blog.Tags) : []
  }
}

function syncForm(blog) {
  form.title = blog.title
  form.content = blog.content
  form.status = blog.status === 'published' ? 'published' : 'draft'
  form.isTop = canManageAllBlogs.value ? blog.isTop : false
  form.categoryId = blog.categoryId
  form.tags = blog.tags.map((item) => item.name || item.Name).filter(Boolean).join(', ')
}

async function loadEditingBlog() {
  if (!isEditMode.value) {
    return
  }
  if (!Number.isInteger(editingBlogID.value) || editingBlogID.value <= 0) {
    message.value = '无效的博客编号。'
    return
  }

  loadingBlog.value = true
  try {
    const blog = normalizeBlog(await getBlogById(editingBlogID.value))
    currentBlog.value = blog
    if (!canEdit.value) {
      message.value = '你没有权限编辑这篇博客。'
      return
    }
    syncForm(blog)
  } catch (error) {
    message.value = error.status === 404 ? '未找到要编辑的博客。' : error.message
  } finally {
    loadingBlog.value = false
  }
}

onMounted(async () => {
  const currentUser = await refreshCurrentUser()
  ready.value = true
  if (!currentUser.isLogin) {
    setTimeout(() => router.push('/login'), 1200)
    return
  }

  await refreshTaxonomy()
  await loadEditingBlog()
})

async function handleSubmit(targetStatus) {
  if (isEditMode.value && !canEdit.value) {
    message.value = '你没有权限编辑这篇博客。'
    return
  }

  submitting.value = true
  pendingAction.value = targetStatus
  message.value = ''

  const payload = {
    title: form.title,
    content: form.content,
    status: targetStatus,
    isTop: canManageAllBlogs.value && form.isTop ? 'true' : 'false',
    categoryId: String(form.categoryId || 0),
    tags: form.tags
  }

  try {
    const result = isEditMode.value
      ? await updateBlog(editingBlogID.value, payload)
      : await createBlog(payload)

    success.value = true
    message.value = targetStatus === 'published' ? '博客已发布。' : '草稿已保存。'

    setTimeout(() => {
      if (targetStatus === 'draft') {
        router.push(`/user/${store.user.id}`)
        return
      }
      const nextID = isEditMode.value ? editingBlogID.value : (result?.id || result?.ID)
      router.push(nextID ? `/blog/${nextID}` : '/')
    }, 900)
  } catch (error) {
    success.value = false
    message.value = error.message
  } finally {
    submitting.value = false
    pendingAction.value = ''
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

.editor-head-actions,
.editor-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
  justify-content: flex-end;
}

.editor-form {
  display: grid;
  gap: 16px;
}

.field {
  display: grid;
  gap: 8px;
}

.field span,
.check-field span {
  font-weight: 600;
  color: #203040;
}

.field input,
.field select,
.field textarea {
  width: 100%;
  padding: 12px 14px;
  border: 1px solid #d5dee8;
  border-radius: 14px;
  font-size: 14px;
  resize: vertical;
}

.compact-field {
  max-width: 280px;
}

.check-field {
  display: flex;
  align-items: center;
  gap: 10px;
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

.primary-btn:disabled,
.secondary-btn:disabled {
  cursor: wait;
  opacity: 0.72;
}

@media (max-width: 700px) {
  .editor-head {
    flex-direction: column;
  }

  .editor-head-actions,
  .editor-actions {
    justify-content: flex-start;
  }
}
</style>
