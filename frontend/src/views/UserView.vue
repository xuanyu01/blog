<!--
/*
	这个文件定义用户资料编辑页面组件。
*/
-->
<template>
  <section class="page-block user-page" v-if="ready">
    <div class="container" v-if="user.isLogin">
      <div class="user-dashboard">
        <div class="user-hero">
          <RouterLink to="/user/avatar" class="user-avatar avatar-entry" title="点击更换头像">
            <img v-if="previewImage" :src="previewImage" class="avatar-large" :alt="displayNameForView" />
            <div v-else class="avatar-large avatar-fallback">
              {{ initials }}
            </div>
            <span class="avatar-tip">点击更换头像</span>
          </RouterLink>

          <div class="user-meta">
            <h1>{{ displayNameForView }}</h1>
            <p>账号 {{ user.userName }}</p>
            <p class="user-permission">权限 {{ permissionText }}</p>
          </div>
        </div>

        <div v-if="!requiresPasswordChange" class="quick-actions">
          <RouterLink :to="profilePath" class="action-card action-card-primary">
            <strong>返回个人主页</strong>
            <span>查看自己发布、点赞和收藏过的博客</span>
          </RouterLink>

          <RouterLink to="/user/avatar" class="action-card">
            <strong>修改头像</strong>
            <span>支持上传 png、jpg、jpeg、gif 图片</span>
          </RouterLink>
        </div>

        <div v-if="requiresPasswordChange" class="force-password-panel">
          <h3>首次登录请修改密码</h3>
          <p>当前管理员账号仍在使用初始化密码。为了保证后台安全，请先修改密码，完成后才能继续使用其它功能。</p>
        </div>

        <div class="user-grid">
          <form v-if="!requiresPasswordChange" class="panel" @submit.prevent="handleProfileSubmit">
            <h3>资料信息</h3>

            <label class="field">
              <span>账号</span>
              <input :value="user.userName" type="text" readonly />
            </label>

            <label class="field">
              <span>显示名称</span>
              <input
                v-model.trim="profileForm.displayName"
                type="text"
                placeholder="请输入新的显示名称"
              />
            </label>

            <button type="submit" :disabled="profileSaving">
              {{ profileSaving ? '保存中...' : '保存资料' }}
            </button>

            <p v-if="profileMessage" :class="profileSuccess ? 'feedback success' : 'feedback error'">
              {{ profileMessage }}
            </p>
          </form>

          <form class="panel" @submit.prevent="handlePasswordSubmit">
            <h3>{{ requiresPasswordChange ? '首次登录修改密码' : '修改密码' }}</h3>

            <label class="field">
              <span>当前密码</span>
              <input
                v-model="passwordForm.currentPassword"
                type="password"
                placeholder="请输入当前密码"
              />
            </label>

            <label class="field">
              <span>新密码</span>
              <input
                v-model="passwordForm.newPassword"
                type="password"
                placeholder="请输入新密码"
              />
            </label>

            <button type="submit" :disabled="passwordSaving">
              {{ passwordSaving ? '保存中...' : (requiresPasswordChange ? '完成修改' : '修改密码') }}
            </button>

            <p v-if="passwordMessage" :class="passwordSuccess ? 'feedback success' : 'feedback error'">
              {{ passwordMessage }}
            </p>
          </form>
        </div>
      </div>
    </div>

    <div class="empty-card" v-else>
      <h3>你还没有登录</h3>
      <p>请先登录后再访问用户资料编辑页</p>
    </div>
  </section>
</template>

<script setup>
/*
	这个页面负责展示和修改当前登录用户的基本信息。
*/
import { computed, onMounted, reactive, ref } from 'vue'
import { RouterLink, useRoute, useRouter } from 'vue-router'
import {
  appStore as store,
  changeUserPassword,
  refreshCurrentUser,
  saveUserProfile
} from '../store/appStore'

const route = useRoute()
const router = useRouter()
const ready = ref(false)
const profileSaving = ref(false)
const profileMessage = ref('')
const profileSuccess = ref(false)
const passwordSaving = ref(false)
const passwordMessage = ref('')
const passwordSuccess = ref(false)

const profileForm = reactive({
  displayName: ''
})

const passwordForm = reactive({
  currentPassword: '',
  newPassword: ''
})

const user = computed(() => store.user)
const requiresPasswordChange = computed(() => Boolean(user.value.mustChangePassword))
const displayNameForView = computed(() => user.value.displayName || user.value.userName || '用户')
const permissionText = computed(() => {
  switch (user.value.permission) {
    case 'admin':
      return '系统管理员'
    case 'user_admin':
      return '用户管理员'
    default:
      return '普通用户'
  }
})
const initials = computed(() => displayNameForView.value.slice(0, 1).toUpperCase())
const previewImage = computed(() => (user.value.imageRoute ? `/img/${user.value.imageRoute}` : ''))
const profilePath = computed(() => `/user/${user.value.id}`)

function syncProfileForm(currentUser) {
  profileForm.displayName = currentUser.displayName || ''
}

onMounted(async () => {
  const currentUser = await refreshCurrentUser()
  ready.value = true

  if (!currentUser.isLogin) {
    setTimeout(() => router.push('/login'), 1200)
    return
  }

  const routeID = Number.parseInt(route.params.id, 10)
  if (routeID !== currentUser.id) {
    router.replace(`/user/${currentUser.id}/edit`)
    return
  }

  syncProfileForm(currentUser)
})

async function handleProfileSubmit() {
  profileSaving.value = true
  profileMessage.value = ''

  try {
    const updatedUser = await saveUserProfile({
      displayName: profileForm.displayName,
      imageRoute: user.value.imageRoute
    })

    syncProfileForm(updatedUser)
    profileSuccess.value = true
    profileMessage.value = '资料更新成功'
  } catch (error) {
    profileSuccess.value = false
    profileMessage.value = error.message
  } finally {
    profileSaving.value = false
  }
}

async function handlePasswordSubmit() {
  passwordSaving.value = true
  passwordMessage.value = ''

  try {
    await changeUserPassword({
      currentPassword: passwordForm.currentPassword,
      newPassword: passwordForm.newPassword
    })

    passwordForm.currentPassword = ''
    passwordForm.newPassword = ''
    const updatedUser = await refreshCurrentUser()
    passwordSuccess.value = true
    passwordMessage.value = requiresPasswordChange.value ? '密码修改成功，正在进入个人主页' : '密码修改成功'
    if (!updatedUser.mustChangePassword) {
      setTimeout(() => router.replace(`/user/${updatedUser.id}`), 800)
    }
  } catch (error) {
    passwordSuccess.value = false
    passwordMessage.value = error.message
  } finally {
    passwordSaving.value = false
  }
}
</script>


<style scoped>
.user-dashboard {
  display: grid;
  gap: 24px;
}

.user-hero {
  display: flex;
  gap: 20px;
  align-items: center;
  padding: 24px;
  border-radius: 24px;
  background: rgba(255, 255, 255, 0.82);
  box-shadow: 0 18px 40px rgba(40, 58, 80, 0.08);
}

.avatar-entry {
  position: relative;
  text-decoration: none;
}

.avatar-tip {
  position: absolute;
  left: 50%;
  bottom: -10px;
  transform: translateX(-50%);
  padding: 6px 10px;
  border-radius: 999px;
  background: rgba(32, 48, 64, 0.9);
  color: #fff;
  font-size: 12px;
  white-space: nowrap;
}

.quick-actions {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
  gap: 16px;
}

.action-card {
  display: grid;
  gap: 8px;
  padding: 20px;
  border-radius: 20px;
  background: rgba(255, 255, 255, 0.88);
  box-shadow: 0 18px 40px rgba(40, 58, 80, 0.08);
  color: #203040;
}

.action-card strong {
  font-size: 18px;
}

.action-card span {
  color: #5f6f82;
}

.action-card-primary {
  background: linear-gradient(135deg, #203040, #35506a);
  color: #fff;
}

.action-card-primary span {
  color: rgba(255, 255, 255, 0.8);
}

.user-meta h1 {
  margin: 0 0 8px;
}

.user-meta p {
  margin: 0;
  color: #5f6f82;
}

.user-permission {
  font-size: 13px;
  color: #8a5b36;
}

.user-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(260px, 1fr));
  gap: 20px;
}

.panel {
  display: grid;
  gap: 14px;
  padding: 24px;
  border-radius: 24px;
  background: rgba(255, 255, 255, 0.88);
  box-shadow: 0 18px 40px rgba(40, 58, 80, 0.08);
}

.panel h3 {
  margin: 0;
}

.field {
  display: grid;
  gap: 8px;
}

.field span {
  font-weight: 600;
  color: #203040;
}

.field input {
  width: 100%;
  padding: 12px 14px;
  border: 1px solid #d5dee8;
  border-radius: 14px;
  font-size: 14px;
}

.field input[readonly] {
  background: #f4f7fa;
  color: #6a7c8f;
}

.panel button {
  border: none;
  border-radius: 14px;
  padding: 12px 16px;
  background: #203040;
  color: #fff;
  font-weight: 600;
  cursor: pointer;
}

.panel button:disabled {
  cursor: wait;
  opacity: 0.72;
}
.force-password-panel {
  display: grid;
  gap: 8px;
  padding: 20px 24px;
  border: 1px solid rgba(165, 58, 58, 0.18);
  border-radius: 20px;
  background: rgba(255, 245, 238, 0.92);
  color: #7b3e24;
}

.force-password-panel h3,
.force-password-panel p {
  margin: 0;
}

</style>