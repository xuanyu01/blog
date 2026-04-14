<!--
/*
	这个文件定义博客列表中的单条博客组件
*/
-->
<template>
  <article class="blog">
    <div class="blog_title">{{ blog.Title }}</div>
    <div class="blog_content">{{ blog.Content }}</div>
  </article>
</template>

<script setup>
/*
	这个组件负责展示单篇博客的标题和内容
*/
defineProps({
  blog: {
    type: Object,
    required: true
  }
})
</script>
