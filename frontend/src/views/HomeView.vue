<!--
/*
	这个文件定义首页页面组件
*/
-->
<template>
  <section class="page-block home-page">
    <div class="main">
      <div class="blog-list" v-if="store.blogs.length">
        <BlogCard v-for="blog in store.blogs" :key="`${blog.Title}-${blog.Content}`" :blog="blog" />
      </div>

      <div class="empty-state" v-else>
        <h3>还没有内容</h3>
        <p>当前还没有可展示的博客内容</p>
      </div>
    </div>
  </section>
</template>

<script setup>
/*
	这个页面负责加载并展示博客列表
*/
import { onMounted } from 'vue'
import BlogCard from '../components/BlogCard.vue'
import { appStore as store, refreshAppState } from '../store/appStore'

// 页面挂载后加载首页数据
onMounted(async () => {
  await refreshAppState()
})
</script>
