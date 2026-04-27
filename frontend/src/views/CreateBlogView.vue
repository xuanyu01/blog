<template>
  <section v-if="ready" class="page-block user-page">
    <div v-if="user.isLogin" class="container">
      <div class="editor-card">
        <div class="editor-head">
          <div>
            <h2>{{ pageTitle }}</h2>
            <p>{{ pageDescription }}</p>
          </div>
          <div class="head-actions">
            <button
              type="button"
              class="drafts-list-btn"
              @click="router.push('/user/drafts')"
            >
              查看我的草稿
            </button>
            <button
              v-if="showDraftEntryButton"
              type="button"
              class="draft-entry-btn"
              @click="goToDraft"
            >
              进入草稿
            </button>
            <button type="button" class="secondary-btn" @click="router.push('/user')">返回用户中心</button>
          </div>
        </div>

        <div v-if="isEditMode && loadingBlog" class="feedback">正在加载要编辑的博客内容...</div>

        <div v-else-if="isEditMode && !canEdit" class="feedback error">
          {{ message || '你没有权限编辑这篇博客。' }}
        </div>

        <form v-else class="editor-form" @submit.prevent>
          <label class="field">
            <span>标题</span>
            <input
              v-model.trim="form.title"
              type="text"
              maxlength="100"
              placeholder="请输入博客标题"
            />
          </label>

          <div class="meta-grid">
            <label class="field">
              <span>分类</span>
              <select v-model.number="form.categoryId">
                <option :value="0">未分类</option>
                <option
                  v-for="item in store.taxonomy.categories"
                  :key="item.id"
                  :value="item.id"
                >
                  {{ item.name }}
                </option>
              </select>
            </label>

            <label class="field">
              <span>标签</span>
              <input
                v-model.trim="form.tagsText"
                type="text"
                placeholder="多个标签用逗号分隔"
              />
            </label>
          </div>

          <div class="markdown-toolbar">
            <div class="toolbar-tip">支持标题、列表、代码块、引用、链接等常用 Markdown 语法。</div>
            <div class="toolbar-actions">
              <button type="button" class="chip-btn" @click="insertSnippet('heading')">标题</button>
              <button type="button" class="chip-btn" @click="insertSnippet('bold')">加粗</button>
              <button type="button" class="chip-btn" @click="insertSnippet('quote')">引用</button>
              <button type="button" class="chip-btn" @click="insertSnippet('code')">代码块</button>
              <button type="button" class="chip-btn" @click="insertSnippet('link')">链接</button>
            </div>
          </div>

          <div class="editor-grid">
            <label class="field">
              <span>Markdown 内容</span>
              <textarea
                ref="editorRef"
                v-model="form.content"
                class="markdown-input"
                rows="18"
                placeholder="请使用 Markdown 编写博客内容"
              ></textarea>
            </label>

            <section class="preview-panel">
              <div class="preview-head">
                <span>实时预览</span>
                <span class="preview-meta">{{ previewMeta }}</span>
              </div>
              <div v-if="form.content.trim()" class="markdown-preview prose" v-html="previewHtml"></div>
              <div v-else class="preview-empty">输入 Markdown 后，这里会显示过滤后的预览效果。</div>
            </section>
          </div>

          <label class="field">
            <span>状态</span>
            <select v-model="form.status">
              <option value="draft">草稿</option>
              <option value="published">发布</option>
            </select>
          </label>

          <label v-if="canManageAllBlogs" class="toggle-field">
            <input v-model="form.isTop" type="checkbox" />
            <span>置顶文章</span>
          </label>

          <div class="editor-actions">
            <button
              type="button"
              class="secondary-btn"
              :disabled="submitting || loadingBlog"
              @click="handleSubmit('draft')"
            >
              {{ submitting && pendingAction === 'draft' ? '保存中...' : draftButtonText }}
            </button>
            <button
              type="button"
              class="primary-btn"
              :disabled="submitting || loadingBlog"
              @click="handleSubmit('published')"
            >
              {{ submitting && pendingAction === 'published' ? '发布中...' : publishButtonText }}
            </button>
          </div>

          <p class="status-tip">
            当前选择：{{ form.status === 'draft' ? '草稿' : '发布' }}。发布前会再次确认，草稿不会出现在前台列表。
          </p>

          <p v-if="message" :class="success ? 'feedback success' : 'feedback error'">
            {{ message }}
          </p>
        </form>
      </div>
    </div>

    <div v-else class="empty-card">
      <h3>你还没有登录</h3>
      <p>请先登录后再创建或编辑博客。</p>
    </div>
  </section>
</template>

<script setup>
import { computed, nextTick, onMounted, reactive, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { createBlog, getBlogById, updateBlog } from '../api/client'
import { appStore as store, refreshAppState, refreshCurrentUser, refreshTaxonomy } from '../store/appStore'
import { renderMarkdown } from '../utils/markdown'

const route = useRoute()
const router = useRouter()
const editorRef = ref(null)
const ready = ref(false)
const loadingBlog = ref(false)
const submitting = ref(false)
const pendingAction = ref('')
const message = ref('')
const success = ref(false)
const currentBlog = ref(null)
const draftBlogID = ref(null)

const form = reactive({
  title: '',
  content: '',
  status: 'draft',
  isTop: false,
  categoryId: 0,
  tagsText: ''
})

const editingBlogID = computed(() => Number.parseInt(route.params.id, 10))
const isEditMode = computed(() => route.name === 'blog-edit')
const user = computed(() => store.user)
const canManageAllBlogs = computed(
  () => store.user.permission === 'admin' || store.user.permission === 'user_admin'
)

const canEdit = computed(() => {
  if (!isEditMode.value) {
    return true
  }
  if (!currentBlog.value || !store.user.isLogin) {
    return false
  }
  if (canManageAllBlogs.value) {
    return true
  }
  return store.user.userName === currentBlog.value.authorUsername
})

const pageTitle = computed(() => (isEditMode.value ? '编辑博客' : '创作博客'))
const pageDescription = computed(() => (
  isEditMode.value
    ? '继续完善正文、分类和标签，保存草稿或确认发布都可以。'
    : '使用 Markdown 写内容，并补充分类与标签，让文章更容易被检索。'
))
const draftButtonText = computed(() => (isEditMode.value ? '保存为草稿' : '先存草稿'))
const publishButtonText = computed(() => (isEditMode.value ? '确认并发布' : '发布博客'))
const previewHtml = computed(() => renderMarkdown(form.content))
const previewMeta = computed(() => `${form.content.length} 字符`)
const showDraftEntryButton = computed(() => {
  if (draftBlogID.value) {
    return true
  }
  return isEditMode.value && currentBlog.value && currentBlog.value.status !== 'published'
})

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

onMounted(async () => {
  const currentUser = await refreshCurrentUser()
  if (!currentUser.isLogin) {
    ready.value = true
    setTimeout(() => router.push('/login'), 1200)
    return
  }

  await refreshTaxonomy()

  if (isEditMode.value) {
    await loadEditableBlog()
  }

  ready.value = true
})

async function loadEditableBlog() {
  if (!Number.isInteger(editingBlogID.value) || editingBlogID.value <= 0) {
    success.value = false
    message.value = '无效的博客编号。'
    return
  }

  loadingBlog.value = true
  try {
    currentBlog.value = normalizeBlog(await getBlogById(editingBlogID.value))

    if (!canEdit.value) {
      success.value = false
      message.value = '你没有权限编辑这篇博客。'
      return
    }

    form.title = currentBlog.value.title
    form.content = currentBlog.value.content
    form.status = currentBlog.value.status === 'published' ? 'published' : 'draft'
    form.isTop = canManageAllBlogs.value ? currentBlog.value.isTop : false
    form.categoryId = currentBlog.value.categoryId
    form.tagsText = currentBlog.value.tags.map((item) => item.name).join(', ')
    if (currentBlog.value.status !== 'published') {
      draftBlogID.value = currentBlog.value.id
    }
  } catch (error) {
    success.value = false
    if (error.status === 404) {
      message.value = '未找到要编辑的博客。'
    } else if (error.status === 403) {
      message.value = '你没有权限编辑这篇博客。'
    } else {
      message.value = error.message
    }
  } finally {
    loadingBlog.value = false
  }
}

async function handleSubmit(targetStatus) {
  if (isEditMode.value && (!currentBlog.value || !canEdit.value)) {
    success.value = false
    message.value = '你没有权限编辑这篇博客。'
    return
  }

  form.status = targetStatus
  if (targetStatus === 'published') {
    const targetTitle = form.title || currentBlog.value?.title || '未命名博客'
    const confirmed = window.confirm(`确认发布《${targetTitle}》吗？`)
    if (!confirmed) {
      return
    }
  }

  submitting.value = true
  pendingAction.value = targetStatus
  message.value = ''

  try {
    const payload = {
      title: form.title,
      content: form.content,
      status: targetStatus,
      isTop: canManageAllBlogs.value && form.isTop ? 'true' : 'false',
      categoryId: String(form.categoryId || 0),
      tags: form.tagsText
    }

    let result
    if (isEditMode.value) {
      await updateBlog(editingBlogID.value, payload)
      draftBlogID.value = targetStatus === 'draft' ? editingBlogID.value : null
    } else {
      result = await createBlog(payload)
      const createdID = result?.id || result?.ID
      draftBlogID.value = targetStatus === 'draft' ? createdID : null
    }

    await refreshAppState({ page: 1 })
    success.value = true
    message.value = targetStatus === 'published'
      ? (isEditMode.value ? '博客已更新并发布。' : '博客已创建并发布。')
      : (isEditMode.value ? '草稿已更新。' : '草稿已保存。')

    if (!isEditMode.value) {
      form.title = ''
      form.content = ''
      form.status = 'draft'
      form.isTop = false
      form.categoryId = 0
      form.tagsText = ''
    }

    setTimeout(() => {
      if (isEditMode.value) {
        router.push(`/blog/${editingBlogID.value}`)
        return
      }

      const createdID = result?.id || result?.ID
      if (createdID) {
        router.push(`/blog/${createdID}`)
        return
      }

      router.push(targetStatus === 'draft' ? '/user' : '/')
    }, 900)
  } catch (error) {
    success.value = false
    message.value = error.message
  } finally {
    submitting.value = false
    pendingAction.value = ''
  }
}

async function insertSnippet(type) {
  const snippets = {
    heading: '# 一级标题\n\n',
    bold: '**加粗内容**',
    quote: '> 这里是一段引用\n',
    code: '```go\nfmt.Println("hello markdown")\n```\n',
    link: '[链接标题](https://example.com)'
  }

  form.content += snippets[type] || ''
  await nextTick()
  editorRef.value?.focus()
}

function goToDraft() {
  if (!showDraftEntryButton.value) {
    return
  }
  const targetID = draftBlogID.value || editingBlogID.value
  if (!targetID) {
    return
  }
  router.push(`/blog/${targetID}`)
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

.head-actions {
  display: flex;
  gap: 12px;
  align-items: center;
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
.field textarea,
.field select {
  width: 100%;
  padding: 12px 14px;
  border: 1px solid #d5dee8;
  border-radius: 14px;
  font-size: 14px;
  resize: vertical;
}

.meta-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 16px;
}

.markdown-toolbar {
  display: grid;
  gap: 10px;
  padding: 16px;
  border-radius: 18px;
  background: #f6f1e8;
}

.toolbar-tip {
  color: #5f6f82;
  font-size: 14px;
}

.toolbar-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.chip-btn {
  border: none;
  border-radius: 999px;
  padding: 8px 12px;
  background: #fff;
  color: #203040;
  font-weight: 600;
  cursor: pointer;
}

.editor-grid {
  display: grid;
  grid-template-columns: minmax(0, 1fr) minmax(0, 1fr);
  gap: 18px;
  align-items: start;
}

.markdown-input {
  min-height: 420px;
  font-family: 'Consolas', 'Courier New', monospace;
  line-height: 1.7;
}

.preview-panel {
  display: grid;
  gap: 12px;
  min-height: 100%;
  padding: 16px;
  border: 1px solid #e4e8ef;
  border-radius: 18px;
  background: #fcfcfd;
}

.preview-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  color: #203040;
  font-weight: 700;
}

.preview-meta {
  color: #7a8797;
  font-size: 13px;
}

.preview-empty {
  color: #7a8797;
  line-height: 1.8;
}

.markdown-preview {
  color: #2b3744;
  line-height: 1.8;
  word-break: break-word;
}

.markdown-preview :deep(h1),
.markdown-preview :deep(h2),
.markdown-preview :deep(h3) {
  color: #203040;
  line-height: 1.3;
}

.markdown-preview :deep(pre) {
  overflow-x: auto;
  padding: 14px;
  border-radius: 14px;
  background: #1d2732;
  color: #f5f7fa;
}

.markdown-preview :deep(code) {
  font-family: 'Consolas', 'Courier New', monospace;
}

.markdown-preview :deep(blockquote) {
  margin: 0;
  padding-left: 14px;
  border-left: 4px solid #d4b48a;
  color: #5f6f82;
}

.markdown-preview :deep(a) {
  color: #9c6a43;
}

.toggle-field {
  display: inline-flex;
  align-items: center;
  gap: 10px;
  color: #203040;
  font-weight: 600;
}

.editor-actions {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
}

.status-tip {
  margin: 0;
  color: #5f6f82;
  font-size: 14px;
}

.primary-btn,
.secondary-btn,
.draft-entry-btn,
.drafts-list-btn {
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

.draft-entry-btn {
  background: #d8eadf;
  color: #1f5a35;
}

.drafts-list-btn {
  background: #efe4d3;
  color: #7b5427;
}

.primary-btn:disabled,
.secondary-btn:disabled,
.draft-entry-btn:disabled,
.drafts-list-btn:disabled {
  cursor: wait;
  opacity: 0.72;
}

@media (max-width: 980px) {
  .editor-grid {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 700px) {
  .editor-head,
  .editor-actions,
  .meta-grid {
    flex-direction: column;
    grid-template-columns: 1fr;
  }

  .head-actions {
    width: 100%;
    flex-direction: column;
  }

  .editor-actions button,
  .head-actions button {
    width: 100%;
  }
}
</style>
