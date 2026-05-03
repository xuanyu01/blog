<template>
  <RouterLink :to="detailLink" class="blog-link">
    <article class="blog">
      <div class="blog-meta">
        <span v-if="createdAtText">{{ createdAtText }}</span>
        <span class="blog-author">作者：{{ authorName }}</span>
        <span v-if="categoryName" class="blog-category">{{ categoryName }}</span>
      </div>

      <div class="blog-title">{{ title }}</div>
      <div class="blog-content">{{ summary }}</div>

      <div v-if="tags.length" class="tag-row">
        <span v-for="item in tags" :key="item.slug || item.name" class="tag-chip"># {{ item.name }}</span>
      </div>

      <div class="stats-row">
        <span>阅读 {{ stats.viewCount }}</span>
        <span>点赞 {{ stats.likeCount }}</span>
        <span>收藏 {{ stats.favoriteCount }}</span>
        <span>评论 {{ stats.commentCount }}</span>
      </div>
    </article>
  </RouterLink>
</template>

<script setup>
import { computed } from 'vue'
import { RouterLink } from 'vue-router'

const props = defineProps({
  blog: {
    type: Object,
    required: true
  }
})

const detailLink = computed(() => `/blog/${props.blog.ID ?? props.blog.id}`)
const title = computed(() => props.blog.Title ?? props.blog.title ?? '')
const summary = computed(() => props.blog.Summary ?? props.blog.summary ?? props.blog.Content ?? props.blog.content ?? '')
const authorName = computed(() => props.blog.AuthorUsername ?? props.blog.authorUsername ?? 'unknown')
const categoryName = computed(() => props.blog.CategoryName ?? props.blog.categoryName ?? '')
const tags = computed(() => props.blog.Tags ?? props.blog.tags ?? [])
const stats = computed(() => props.blog.Stats ?? props.blog.stats ?? {
  viewCount: 0,
  likeCount: 0,
  favoriteCount: 0,
  commentCount: 0
})

const createdAtText = computed(() => {
  const rawDate = props.blog.CreatedAt ?? props.blog.createdAt
  if (!rawDate) {
    return ''
  }

  const date = new Date(rawDate)
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
</script>

<style scoped>
.blog-link {
  display: block;
  color: inherit;
  text-decoration: none;
}

.blog {
  display: grid;
  gap: 12px;
  padding: 22px 24px;
  border-radius: 22px;
  background: rgba(255, 255, 255, 0.9);
  box-shadow: 0 18px 40px rgba(40, 58, 80, 0.08);
  transition: transform 0.18s ease, box-shadow 0.18s ease;
}

.blog-link:hover .blog,
.blog-link:focus-visible .blog {
  transform: translateY(-2px);
  box-shadow: 0 22px 46px rgba(40, 58, 80, 0.14);
}

.blog-meta,
.stats-row {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
  color: #9c6a43;
  font-size: 13px;
  font-weight: 700;
  letter-spacing: 0.04em;
}

.blog-author {
  color: #6d4d33;
}

.blog-category {
  color: #1f5a35;
}

.blog-title {
  font-size: 24px;
  font-weight: 700;
  line-height: 1.35;
  color: #203040;
  word-break: break-word;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
  text-overflow: ellipsis;
}

.blog-content {
  color: #5f6f82;
  line-height: 1.75;
  word-break: break-word;
  display: -webkit-box;
  -webkit-line-clamp: 3;
  -webkit-box-orient: vertical;
  overflow: hidden;
  text-overflow: ellipsis;
}

.tag-row {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.tag-chip {
  padding: 4px 10px;
  border-radius: 999px;
  background: #f3ecde;
  color: #7b5427;
  font-size: 12px;
  font-weight: 700;
}

.stats-row {
  color: #5f6f82;
  letter-spacing: 0;
}
</style>

