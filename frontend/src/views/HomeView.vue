<!--
/*
	这个文件定义首页页面组件
*/
-->
<template>
  <section class="page-block home-page">
    <div class="home-layout">
      <div class="home-main">
        <div class="blog-list" v-if="store.blogs.length">
          <BlogCard
            v-for="blog in store.blogs"
            :key="blog.ID"
            :blog="blog"
          />
        </div>

        <div class="empty-state" v-else>
          <h3>还没有内容</h3>
          <p>当前还没有可展示的博客内容</p>
        </div>
      </div>

      <aside class="home-side">
        <div class="create-module">
          <p class="create-kicker">Blog Studio</p>
          <h3>创作</h3>
          <p class="create-text">写下新的标题和内容 发布一篇新的博客</p>

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
	这个页面负责加载并展示博客列表
*/
import { onMounted } from 'vue'
import { RouterLink } from 'vue-router'
import BlogCard from '../components/BlogCard.vue'
import { appStore as store, refreshAppState } from '../store/appStore'

// 页面挂载后加载首页数据
onMounted(async () => {
  await refreshAppState()
})
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
}

.blog-list {
  display: grid;
  gap: 20px;
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
</style>
