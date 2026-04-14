/*
	这个文件是 Vue 前端的启动入口
*/
import { createApp } from 'vue'
import App from './App.vue'
import router from './router'
import './styles.css'

createApp(App).use(router).mount('#app')
