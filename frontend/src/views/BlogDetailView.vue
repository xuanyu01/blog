<template>
  <section class="page-block blog-detail-page">
    <div v-if="loading" class="empty-state">
      <h3>正在加载博客详情</h3>
      <p>请稍候，正在请求最新内容和评论。</p>
    </div>

    <div v-else-if="blog" class="detail-shell">
      <RouterLink to="/" class="back-link">返回首页</RouterLink>
      <article class="detail-card">
        <div class="detail-meta-row">
          <div class="detail-meta-group">
            <div v-if="createdAtText" class="detail-meta">{{ createdAtText }}</div>
            <div class="detail-author">作者：{{ blog.authorUsername || 'unknown' }}</div>
            <div class="detail-status-row">
              <span class="status-badge" :class="'status-' + blog.status">{{ statusText }}</span>
              <span v-if="blog.categoryName" class="category-badge">{{ blog.categoryName }}</span>
              <span v-if="blog.isTop" class="top-badge">置顶</span>
            </div>
          </div>
          <div v-if="canEdit || canDelete" class="detail-actions">
            <button v-if="canEdit" type="button" class="edit-btn" @click="handleEdit">编辑博客</button>
            <button v-if="canDelete" type="button" class="delete-btn" :disabled="deleting" @click="handleDelete">
              {{ deleting ? '删除中...' : '删除博客' }}
            </button>
          </div>
        </div>
        <div v-if="blog.tags.length" class="tag-row">
          <span v-for="item in blog.tags" :key="item.slug || item.name" class="tag-chip"># {{ item.name }}</span>
        </div>
        <h1 class="detail-title">{{ blog.title }}</h1>
        <div class="interaction-row">
          <div class="stats-row">
            <span>阅读 {{ blog.stats.viewCount }}</span>
            <span>点赞 {{ blog.stats.likeCount }}</span>
            <span>收藏 {{ blog.stats.favoriteCount }}</span>
            <span>评论 {{ blog.stats.commentCount }}</span>
          </div>
          <div class="interaction-actions">
            <button type="button" class="action-btn" :class="{ active: blog.liked }" :disabled="togglingLike" @click="handleToggleLike">
              {{ togglingLike ? '处理中...' : (blog.liked ? '取消点赞' : '点赞') }}
            </button>
            <button type="button" class="action-btn" :class="{ active: blog.favorited }" :disabled="togglingFavorite" @click="handleToggleFavorite">
              {{ togglingFavorite ? '处理中...' : (blog.favorited ? '取消收藏' : '收藏') }}
            </button>
          </div>
        </div>
        <div class="detail-content markdown-body" v-html="renderedContent"></div>
        <p v-if="message" :class="success ? 'feedback success' : 'feedback error'">{{ message }}</p>
      </article>
      <section class="comment-card">
        <div class="comment-header">
          <div>
            <h2>评论</h2>
            <p>{{ commentTotal }} 条评论</p>
          </div>
        </div>
        <div class="comment-editor">
          <div v-if="replyTarget" class="replying-bar">
            <span>正在回复 @{{ replyTarget.username }}</span>
            <button type="button" @click="clearReplyTarget">取消回复</button>
          </div>
          <textarea v-model="commentContent" class="comment-textarea" rows="4" maxlength="500" :disabled="submittingComment || !store.user.isLogin" :placeholder="commentPlaceholder" />
          <div class="comment-editor-footer">
            <span class="comment-tip">{{ store.user.isLogin ? '还可以输入 ' + commentRemaining + ' 字' : '登录后可发表评论' }}</span>
            <button type="button" class="comment-submit-btn" :disabled="!canSubmitComment" @click="handleCreateComment">{{ submittingComment ? '发布中...' : (replyTarget ? '发表回复' : '发表评论') }}</button>
          </div>
          <p v-if="commentMessage" :class="commentSuccess ? 'feedback success' : 'feedback error'">{{ commentMessage }}</p>
        </div>
        <div v-if="commentsLoading" class="comment-empty">正在加载评论...</div>
        <div v-else-if="comments.length === 0" class="comment-empty">还没有评论，来发表第一条吧。</div>
        <div v-else class="comment-list">
          <article v-for="item in comments" :key="item.id" class="comment-item">
            <div class="comment-item-head">
              <div><strong class="comment-name">{{ item.displayName || item.username }}</strong><span v-if="item.username" class="comment-username">@{{ item.username }}</span></div>
              <div class="comment-actions">
                <button v-if="store.user.isLogin" type="button" class="reply-btn" @click="handleStartReply(item)">回复</button>
                <button v-if="canDeleteComment(item)" type="button" class="comment-delete-btn" :disabled="deletingCommentId === item.id" @click="handleDeleteComment(item)">{{ deletingCommentId === item.id ? '删除中...' : '删除' }}</button>
              </div>
            </div>
            <div class="comment-time">{{ formatCommentTime(item.createdAt) }}</div>
            <p class="comment-content">{{ item.content }}</p>
            <div v-if="item.replies.length" class="comment-replies">
              <article v-for="reply in item.replies" :key="reply.id" class="comment-reply-item">
                <div class="comment-item-head">
                  <div><strong class="comment-name">{{ reply.displayName || reply.username }}</strong><span v-if="reply.username" class="comment-username">@{{ reply.username }}</span></div>
                  <div class="comment-actions">
                    <button v-if="store.user.isLogin" type="button" class="reply-btn" @click="handleStartReply(reply)">回复</button>
                    <button v-if="canDeleteComment(reply)" type="button" class="comment-delete-btn" :disabled="deletingCommentId === reply.id" @click="handleDeleteComment(reply)">{{ deletingCommentId === reply.id ? '删除中...' : '删除' }}</button>
                  </div>
                </div>
                <div class="comment-time">
                  <span>{{ formatCommentTime(reply.createdAt) }}</span>
                  <span v-if="reply.replyToUsername"> · 回复 @{{ reply.replyToUsername }}</span>
                </div>
                <p class="comment-content">{{ reply.content }}</p>
              </article>
            </div>
          </article>
        </div>
      </section>
    </div>
    <div v-else class="empty-state">
      <h3>{{ errorTitle }}</h3><p>{{ errorDescription }}</p><RouterLink to="/" class="back-link">返回首页</RouterLink>
    </div>
  </section>
</template>

<script setup>
import { computed, onMounted, ref, watch } from 'vue'
import { RouterLink, useRoute, useRouter } from 'vue-router'
import {
  createComment,
  deleteBlog,
  deleteComment,
  getBlogById,
  getCommentsByBlogId,
  toggleBlogFavorite,
  toggleBlogLike
} from '../api/client'
import { appStore as store, refreshAppState, refreshCurrentUser } from '../store/appStore'
import { renderMarkdown } from '../utils/markdown'

const route = useRoute()
const router = useRouter()
const loading = ref(true)
const deleting = ref(false)
const commentsLoading = ref(false)
const submittingComment = ref(false)
const togglingLike = ref(false)
const togglingFavorite = ref(false)
const deletingCommentId = ref(0)
const blog = ref(null)
const comments = ref([])
const commentContent = ref('')
const replyTarget = ref(null)
const message = ref('')
const commentMessage = ref('')
const success = ref(false)
const commentSuccess = ref(false)
const errorState = ref('')

const blogID = computed(() => Number.parseInt(route.params.id, 10))
const canManageAllBlogs = computed(
  () => store.user.permission === 'admin' || store.user.permission === 'user_admin'
)
const renderedContent = computed(() => renderMarkdown(blog.value?.content || ''))
const commentRemaining = computed(() => Math.max(0, 500 - commentContent.value.length))
const commentTotal = computed(() => comments.value.reduce((total, item) => total + 1 + item.replies.length, 0))
const commentPlaceholder = computed(() => {
  if (!store.user.isLogin) {
    return '登录后可发表评论'
  }
  if (replyTarget.value) {
    return `回复 @${replyTarget.value.username}，最多 500 字`
  }
  return '写下你的想法，最多 500 字'
})
const canSubmitComment = computed(
  () => Boolean(store.user.isLogin && commentContent.value.trim() && !submittingComment.value)
)

function normalizeBlog(data = {}) {
  return {
    id: data.id ?? data.ID ?? 0,
    title: data.title ?? data.Title ?? '',
    content: data.content ?? data.Content ?? '',
    authorUsername: data.authorUsername ?? data.AuthorUsername ?? '',
    status: data.status ?? data.Status ?? 'draft',
    isTop: Boolean(data.isTop ?? data.IsTop),
    categoryName: data.categoryName ?? data.CategoryName ?? '',
    createdAt: data.createdAt ?? data.CreatedAt ?? '',
    liked: Boolean(data.liked ?? data.Liked),
    favorited: Boolean(data.favorited ?? data.Favorited),
    tags: Array.isArray(data.tags ?? data.Tags) ? (data.tags ?? data.Tags) : [],
    stats: {
      viewCount: Number(data.stats?.viewCount ?? data.Stats?.ViewCount ?? data.viewCount ?? 0),
      likeCount: Number(data.stats?.likeCount ?? data.Stats?.LikeCount ?? data.likeCount ?? 0),
      favoriteCount: Number(data.stats?.favoriteCount ?? data.Stats?.FavoriteCount ?? data.favoriteCount ?? 0),
      commentCount: Number(data.stats?.commentCount ?? data.Stats?.CommentCount ?? data.commentCount ?? 0)
    }
  }
}

function normalizeComment(data = {}) {
  const replies = data.replies ?? data.Replies ?? []
  return {
    id: data.id ?? data.ID ?? 0,
    postId: data.postId ?? data.PostID ?? 0,
    userId: data.userId ?? data.UserID ?? 0,
    parentId: data.parentId ?? data.ParentID ?? null,
    rootId: data.rootId ?? data.RootID ?? null,
    username: data.username ?? data.Username ?? '',
    displayName: data.displayName ?? data.DisplayName ?? '',
    replyToUsername: data.replyToUsername ?? data.ReplyToUsername ?? '',
    content: data.content ?? data.Content ?? '',
    createdAt: data.createdAt ?? data.CreatedAt ?? '',
    replies: Array.isArray(replies) ? replies.map(normalizeComment) : []
  }
}

const createdAtText = computed(() => {
  if (!blog.value?.createdAt) {
    return ''
  }

  const date = new Date(blog.value.createdAt)
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

const statusText = computed(() => {
  switch (blog.value?.status) {
    case 'published':
      return '已发布'
    case 'hidden':
      return '已隐藏'
    default:
      return '草稿'
  }
})

const canDelete = computed(() => {
  if (!blog.value || !store.user.isLogin) {
    return false
  }
  if (canManageAllBlogs.value) {
    return true
  }
  return store.user.userName === blog.value.authorUsername
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
      return '这篇博客当前不是公开状态，只有作者或管理员可以查看。'
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
  await loadPage()
})

watch(blogID, async () => {
  await loadPage()
})

async function loadPage() {
  await loadBlog()
  if (blog.value) {
    await loadComments()
  } else {
    comments.value = []
  }
}

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
    blog.value = normalizeBlog(await getBlogById(blogID.value))
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

async function loadComments() {
  if (!blog.value) {
    comments.value = []
    return
  }

  commentsLoading.value = true
  try {
    const data = await getCommentsByBlogId(blog.value.id)
    comments.value = Array.isArray(data.items) ? data.items.map(normalizeComment) : []
    blog.value.stats.commentCount = commentTotal.value
  } catch (error) {
    comments.value = []
    if (!commentMessage.value) {
      commentSuccess.value = false
      commentMessage.value = error.message || '评论加载失败'
    }
  } finally {
    commentsLoading.value = false
  }
}

function handleEdit() {
  if (!blog.value || !canEdit.value) {
    return
  }
  router.push(`/blog/${blog.value.id}/edit`)
}

async function handleDelete() {
  if (!blog.value || !canDelete.value) {
    return
  }

  const confirmed = window.confirm(`确定删除《${blog.value.title}》吗？`)
  if (!confirmed) {
    return
  }

  deleting.value = true
  message.value = ''

  try {
    await deleteBlog(blog.value.id)
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

async function handleToggleLike() {
  if (!blog.value) {
    return
  }
  if (!store.user.isLogin) {
    message.value = '登录后才可以点赞'
    success.value = false
    return
  }

  togglingLike.value = true
  try {
    const data = await toggleBlogLike(blog.value.id)
    blog.value.liked = Boolean(data.active)
    blog.value.stats.likeCount = Number(data.likeCount || 0)
    blog.value.stats.favoriteCount = Number(data.favoriteCount ?? blog.value.stats.favoriteCount)
  } catch (error) {
    success.value = false
    message.value = error.message
  } finally {
    togglingLike.value = false
  }
}

async function handleToggleFavorite() {
  if (!blog.value) {
    return
  }
  if (!store.user.isLogin) {
    message.value = '登录后才可以收藏'
    success.value = false
    return
  }

  togglingFavorite.value = true
  try {
    const data = await toggleBlogFavorite(blog.value.id)
    blog.value.favorited = Boolean(data.active)
    blog.value.stats.likeCount = Number(data.likeCount ?? blog.value.stats.likeCount)
    blog.value.stats.favoriteCount = Number(data.favoriteCount || 0)
  } catch (error) {
    success.value = false
    message.value = error.message
  } finally {
    togglingFavorite.value = false
  }
}

async function handleCreateComment() {
  if (!blog.value || !canSubmitComment.value) {
    return
  }

  submittingComment.value = true
  commentMessage.value = ''

  try {
    await createComment(blog.value.id, {
      content: commentContent.value.trim(),
      parentId: replyTarget.value?.id || 0
    })
    commentSuccess.value = true
    commentMessage.value = replyTarget.value ? '回复发布成功' : '评论发布成功'
    commentContent.value = ''
    clearReplyTarget()
    await loadComments()
  } catch (error) {
    commentSuccess.value = false
    commentMessage.value = error.message
  } finally {
    submittingComment.value = false
  }
}

function handleStartReply(comment) {
  if (!store.user.isLogin) {
    commentSuccess.value = false
    commentMessage.value = '登录后可回复评论'
    return
  }
  replyTarget.value = comment
  commentMessage.value = ''
}

function clearReplyTarget() {
  replyTarget.value = null
}

function canDeleteComment(comment) {
  if (!store.user.isLogin) {
    return false
  }
  return canManageAllBlogs.value || comment.username === store.user.userName
}

async function handleDeleteComment(comment) {
  if (!canDeleteComment(comment)) {
    return
  }

  const confirmed = window.confirm('确定删除这条评论吗？')
  if (!confirmed) {
    return
  }

  deletingCommentId.value = comment.id
  commentMessage.value = ''

  try {
    await deleteComment(comment.id)
    commentSuccess.value = true
    commentMessage.value = '评论已删除'
    if (replyTarget.value?.id === comment.id) {
      clearReplyTarget()
    }
    await loadComments()
  } catch (error) {
    commentSuccess.value = false
    commentMessage.value = error.message
  } finally {
    deletingCommentId.value = 0
  }
}

function formatCommentTime(value) {
  const date = new Date(value)
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

.detail-card,
.comment-card {
  display: grid;
  gap: 20px;
  padding: 32px;
  border-radius: 26px;
  background: rgba(255, 255, 255, 0.92);
  box-shadow: 0 22px 46px rgba(40, 58, 80, 0.1);
}

.detail-meta-row,
.interaction-row,
.comment-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
}

.detail-meta-group {
  display: grid;
  gap: 8px;
}

.detail-status-row,
.detail-actions,
.tag-row,
.stats-row,
.interaction-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.detail-meta,
.detail-author,
.comment-header p,
.comment-time,
.comment-tip,
.comment-username,
.stats-row {
  color: #9c6a43;
  font-size: 14px;
  font-weight: 700;
}

.status-badge,
.top-badge,
.category-badge,
.tag-chip {
  display: inline-flex;
  align-items: center;
  padding: 4px 10px;
  border-radius: 999px;
  font-size: 12px;
  font-weight: 700;
}

.status-draft {
  background: #efe7d6;
  color: #7e5a28;
}

.status-published {
  background: #dcefe2;
  color: #25613a;
}

.status-hidden {
  background: #f3dfdf;
  color: #8e3434;
}

.category-badge {
  background: #e3efe7;
  color: #1f5a35;
}

.top-badge {
  background: #203040;
  color: #fff;
}

.tag-chip {
  background: #f3ecde;
  color: #7b5427;
}

.edit-btn,
.delete-btn,
.action-btn,
.comment-submit-btn,
.comment-delete-btn,
.reply-btn {
  border: none;
  border-radius: 14px;
  padding: 10px 14px;
  color: #fff;
  font-weight: 600;
  cursor: pointer;
}

.edit-btn,
.comment-submit-btn,
.action-btn,
.reply-btn {
  background: #203040;
}

.delete-btn,
.comment-delete-btn {
  background: #a53a3a;
}

.action-btn.active {
  background: #9c6a43;
}

.delete-btn:disabled,
.action-btn:disabled,
.comment-submit-btn:disabled,
.comment-delete-btn:disabled {
  opacity: 0.72;
  cursor: not-allowed;
}

.detail-title,
.comment-header h2 {
  margin: 0;
  color: #203040;
}

.detail-title {
  font-size: clamp(30px, 4vw, 42px);
  line-height: 1.2;
  word-break: break-word;
}

.detail-content {
  color: #415164;
  font-size: 16px;
  line-height: 1.9;
  word-break: break-word;
}

.detail-content :deep(h1),
.detail-content :deep(h2),
.detail-content :deep(h3) {
  color: #203040;
  line-height: 1.3;
}

.detail-content :deep(pre) {
  overflow-x: auto;
  padding: 14px;
  border-radius: 14px;
  background: #1d2732;
  color: #f5f7fa;
}

.detail-content :deep(code) {
  font-family: 'Consolas', 'Courier New', monospace;
}

.detail-content :deep(blockquote) {
  margin: 0;
  padding-left: 14px;
  border-left: 4px solid #d4b48a;
  color: #5f6f82;
}

.detail-content :deep(a) {
  color: #9c6a43;
}

.comment-editor {
  display: grid;
  gap: 12px;
}

.comment-textarea {
  width: 100%;
  resize: vertical;
  border: 1px solid rgba(32, 48, 64, 0.14);
  border-radius: 18px;
  padding: 14px 16px;
  font: inherit;
  line-height: 1.7;
  color: #203040;
  background: rgba(247, 244, 239, 0.75);
}

.comment-editor-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 12px;
}

.comment-list {
  display: grid;
  gap: 14px;
}

.replying-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  padding: 10px 14px;
  border-radius: 14px;
  background: rgba(156, 106, 67, 0.12);
  color: #7b5427;
  font-size: 14px;
  font-weight: 700;
}

.replying-bar button {
  border: none;
  background: transparent;
  color: #9c6a43;
  font: inherit;
  font-weight: 700;
  cursor: pointer;
}

.comment-item {
  display: grid;
  gap: 8px;
  padding: 18px 20px;
  border-radius: 20px;
  background: rgba(247, 244, 239, 0.8);
}

.comment-item-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.comment-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.comment-name {
  color: #203040;
  font-size: 16px;
}

.comment-username {
  margin-left: 8px;
}

.comment-content {
  margin: 0;
  color: #415164;
  line-height: 1.8;
  white-space: pre-wrap;
  word-break: break-word;
}

.comment-replies {
  display: grid;
  gap: 10px;
  margin-top: 8px;
  padding-left: 18px;
  border-left: 3px solid rgba(156, 106, 67, 0.18);
}

.comment-reply-item {
  display: grid;
  gap: 8px;
  padding: 14px 16px;
  border-radius: 16px;
  background: rgba(255, 255, 255, 0.62);
}

.comment-empty {
  padding: 18px 20px;
  border-radius: 18px;
  background: rgba(247, 244, 239, 0.8);
  color: #5f6f82;
}

@media (max-width: 640px) {
  .detail-meta-row,
  .detail-actions,
  .interaction-row,
  .comment-editor-footer,
  .comment-item-head,
  .replying-bar {
    flex-direction: column;
    align-items: stretch;
  }
}
</style>
