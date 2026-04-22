<!--
/*
	这个文件定义管理员页面组件
*/
-->
<template>
  <section class="page-block admin-page" v-if="ready">
    <div v-if="canEnter" class="admin-layout">
      <aside class="admin-sidebar">
        <div class="admin-sidebar-title">管理导航</div>
        <button
          type="button"
          class="admin-nav-item"
          :class="{ active: activeTab === 'users' }"
          @click="activeTab = 'users'"
        >
          用户管理
        </button>
      </aside>

      <div class="admin-content">
        <section v-if="activeTab === 'users'" class="panel user-manage-panel">
          <div class="panel-head">
            <div>
              <h2>用户管理</h2>
              <p>分页查看全部用户 并执行权限调整或删除操作</p>
            </div>
          </div>

          <div class="table-wrap" v-if="users.length">
            <table class="user-table">
              <thead>
                <tr>
                  <th>用户名</th>
                  <th>显示名称</th>
                  <th>权限</th>
                  <th>操作</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="item in users" :key="item.username">
                  <td>{{ item.username }}</td>
                  <td>{{ item.displayName || '-' }}</td>
                  <td>
                    <template v-if="isAdmin && item.username !== store.user.userName">
                      <select
                        :value="permissionDrafts[item.username] || item.permission"
                        class="field-select compact-select"
                        @change="setPermissionDraft(item.username, $event.target.value)"
                      >
                        <option value="user">普通用户</option>
                        <option value="user_admin">用户管理员</option>
                        <option v-if="item.permission === 'admin'" value="admin">系统管理员</option>
                      </select>
                    </template>
                    <template v-else>
                      {{ permissionLabel(item.permission) }}
                    </template>
                  </td>
                  <td>
                    <div class="action-row">
                      <button
                        v-if="canEditPermission(item)"
                        type="button"
                        class="secondary-btn"
                        :disabled="permissionUpdatingFor === item.username"
                        @click="handleUpdatePermission(item)"
                      >
                        {{ permissionUpdatingFor === item.username ? '保存中...' : '保存权限' }}
                      </button>

                      <button
                        v-if="canDeleteUser(item)"
                        type="button"
                        class="danger-btn"
                        :disabled="deletingFor === item.username"
                        @click="handleDeleteUser(item)"
                      >
                        {{ deletingFor === item.username ? '删除中...' : '删除用户' }}
                      </button>
                    </div>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>

          <div v-else class="empty-card admin-empty">
            <h3>暂无用户数据</h3>
            <p>当前没有可展示的用户信息</p>
          </div>

          <div class="pager">
            <button type="button" class="secondary-btn" :disabled="page <= 1 || loading" @click="goToPage(page - 1)">
              上一页
            </button>
            <span>第 {{ page }} 页 / 共 {{ totalPages }} 页</span>
            <button
              type="button"
              class="secondary-btn"
              :disabled="page >= totalPages || loading"
              @click="goToPage(page + 1)"
            >
              下一页
            </button>
          </div>

          <p v-if="message" :class="success ? 'feedback success' : 'feedback error'">
            {{ message }}
          </p>
        </section>
      </div>
    </div>

    <div v-else class="empty-card">
      <h3>无权访问管理员界面</h3>
      <p>只有用户管理员和系统管理员可以进入这里</p>
    </div>
  </section>
</template>

<script setup>
/*
	这个页面负责展示管理员功能入口和用户管理能力
*/
import { computed, onMounted, reactive, ref } from 'vue'
import { deleteManagedUser, getAdminUsers, updateUserPermission } from '../api/client'
import { appStore as store, refreshCurrentUser } from '../store/appStore'

const ready = ref(false)
const activeTab = ref('users')
const loading = ref(false)
const message = ref('')
const success = ref(false)
const deletingFor = ref('')
const permissionUpdatingFor = ref('')
const users = ref([])
const page = ref(1)
const pageSize = ref(8)
const total = ref(0)
const permissionDrafts = reactive({})

const canEnter = computed(() => {
  return store.user.permission === 'admin' || store.user.permission === 'user_admin'
})

const isAdmin = computed(() => store.user.permission === 'admin')

const totalPages = computed(() => {
  const value = Math.ceil(total.value / pageSize.value)
  return value > 0 ? value : 1
})

function permissionLabel(permission) {
  switch (permission) {
    case 'admin':
      return '系统管理员'
    case 'user_admin':
      return '用户管理员'
    case 'user':
    default:
      return '普通用户'
  }
}

function setPermissionDraft(username, permission) {
  permissionDrafts[username] = permission
}

function canEditPermission(item) {
  if (!isAdmin.value) {
    return false
  }
  if (item.username === store.user.userName) {
    return false
  }
  return item.permission !== 'admin'
}

function canDeleteUser(item) {
  if (item.username === store.user.userName) {
    return false
  }

  if (store.user.permission === 'admin') {
    return item.permission !== 'admin'
  }

  if (store.user.permission === 'user_admin') {
    return item.permission === 'user'
  }

  return false
}

async function loadUsers(targetPage = page.value) {
  loading.value = true
  try {
    const data = await getAdminUsers(targetPage, pageSize.value)
    users.value = data.items || []
    page.value = data.page || targetPage
    pageSize.value = data.pageSize || pageSize.value
    total.value = data.total || 0

    for (const item of users.value) {
      permissionDrafts[item.username] = item.permission
    }
  } finally {
    loading.value = false
  }
}

async function goToPage(targetPage) {
  message.value = ''
  await loadUsers(targetPage)
}

async function handleUpdatePermission(item) {
  if (!canEditPermission(item)) {
    return
  }

  permissionUpdatingFor.value = item.username
  message.value = ''

  try {
    await updateUserPermission({
      username: item.username,
      permission: permissionDrafts[item.username] || item.permission
    })
    success.value = true
    message.value = `已更新 ${item.username} 的权限`
    await loadUsers(page.value)
  } catch (error) {
    success.value = false
    message.value = error.message
  } finally {
    permissionUpdatingFor.value = ''
  }
}

async function handleDeleteUser(item) {
  if (!canDeleteUser(item)) {
    return
  }

  const confirmed = window.confirm(`确定删除用户 ${item.username} 吗`)
  if (!confirmed) {
    return
  }

  deletingFor.value = item.username
  message.value = ''

  try {
    await deleteManagedUser(item.username)
    success.value = true
    message.value = `已删除用户 ${item.username}`

    const remaining = users.value.length - 1
    const nextPage = remaining <= 0 && page.value > 1 ? page.value - 1 : page.value
    await loadUsers(nextPage)
  } catch (error) {
    success.value = false
    message.value = error.message
  } finally {
    deletingFor.value = ''
  }
}

onMounted(async () => {
  await refreshCurrentUser()
  ready.value = true

  if (canEnter.value) {
    await loadUsers(1)
  }
})
</script>

<style scoped>
.admin-layout {
  display: grid;
  grid-template-columns: 220px minmax(0, 1fr);
  gap: 24px;
  align-items: start;
}

.admin-sidebar {
  display: grid;
  gap: 12px;
  padding: 20px;
  border-radius: 24px;
  background: rgba(255, 255, 255, 0.88);
  box-shadow: 0 18px 40px rgba(40, 58, 80, 0.08);
  position: sticky;
  top: 16px;
}

.admin-sidebar-title {
  color: #8a5b36;
  font-size: 12px;
  font-weight: 700;
  letter-spacing: 0.08em;
  text-transform: uppercase;
}

.admin-nav-item {
  border: none;
  border-radius: 16px;
  padding: 14px 16px;
  background: #eef2f6;
  color: #203040;
  font-weight: 600;
  text-align: left;
  cursor: pointer;
}

.admin-nav-item.active {
  background: #203040;
  color: #fff;
}

.admin-content {
  min-width: 0;
}

.user-manage-panel {
  display: grid;
  gap: 20px;
  padding: 24px;
  border-radius: 24px;
  background: rgba(255, 255, 255, 0.88);
  box-shadow: 0 18px 40px rgba(40, 58, 80, 0.08);
}

.panel-head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
}

.panel-head h2,
.panel-head p {
  margin: 0;
}

.panel-head p {
  color: #5f6f82;
  line-height: 1.7;
}

.table-wrap {
  overflow-x: auto;
}

.user-table {
  width: 100%;
  border-collapse: collapse;
}

.user-table th,
.user-table td {
  padding: 14px 12px;
  border-bottom: 1px solid #e6edf3;
  text-align: left;
  vertical-align: middle;
}

.user-table th {
  color: #5f6f82;
  font-size: 13px;
  font-weight: 700;
}

.action-row {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.field-select {
  width: 100%;
  padding: 10px 12px;
  border: 1px solid #d5dee8;
  border-radius: 12px;
  background: #fff;
  font-size: 14px;
}

.compact-select {
  min-width: 140px;
}

.secondary-btn,
.danger-btn {
  border: none;
  border-radius: 12px;
  padding: 10px 14px;
  font-weight: 600;
  cursor: pointer;
}

.secondary-btn {
  background: #e8edf2;
  color: #203040;
}

.danger-btn {
  background: #a53a3a;
  color: #fff;
}

.secondary-btn:disabled,
.danger-btn:disabled {
  opacity: 0.72;
  cursor: wait;
}

.pager {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 12px;
}

.admin-empty {
  text-align: center;
}

@media (max-width: 900px) {
  .admin-layout {
    grid-template-columns: 1fr;
  }

  .admin-sidebar {
    position: static;
  }
}
</style>
