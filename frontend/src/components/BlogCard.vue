<!--
/*
	这个文件定义博客列表中的单条博客组件
*/
-->
<template>
  <RouterLink :to="detailLink" class="blog-link">
    <article class="blog">
      <div class="blog_meta">
        <span v-if="createdAtText">{{ createdAtText }}</span>
        <span class="blog_author">作者 {{ blog.AuthorUsername || 'unknown' }}</span>
      </div>
      <div class="blog_title">{{ blog.Title }}</div>
      <div class="blog_content">{{ blog.Content }}</div>
    </article>
  </RouterLink>
</template>

<script setup>
/*
	这个组件负责展示单篇博客的标题和摘要
*/
import { computed } from 'vue'
import { RouterLink } from 'vue-router'

const props = defineProps({
  blog: {
    type: Object,
    required: true
  }
})

// 列表和详情都改用真实博客 id
// 这样删除和权限判断不会受到列表顺序变化影响
const detailLink = computed(() => `/blog/${props.blog.ID}`)

const createdAtText = computed(() => {
  if (!props.blog.CreatedAt) {
    return ''
  }

  const date = new Date(props.blog.CreatedAt)
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

.blog_meta {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
  color: #9c6a43;
  font-size: 13px;
  font-weight: 700;
  letter-spacing: 0.04em;
}

.blog_author {
  color: #6d4d33;
}

.blog_title {
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

.blog_content {
  color: #5f6f82;
  line-height: 1.75;
  word-break: break-word;
  display: -webkit-box;
  -webkit-line-clamp: 3;
  -webkit-box-orient: vertical;
  overflow: hidden;
  text-overflow: ellipsis;
}
</style>
