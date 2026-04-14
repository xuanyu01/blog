<!--
/*
	这个文件定义前端应用的根组件
*/
-->
<template>
  <div class="site-shell">
    <AppHeader />
    <main class="site-main">
      <router-view />
    </main>
  </div>
</template>

<script setup>
/*
	这个组件负责组织页面外壳并挂载公共导航
*/
import AppHeader from './components/AppHeader.vue'
</script>
