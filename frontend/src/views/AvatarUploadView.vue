<!--
/*
	这个文件定义头像上传页面组件。
*/
-->
<template>
  <section class="page-block user-page" v-if="ready">
    <div class="container" v-if="user.isLogin">
      <div class="upload-layout">
        <div class="upload-card">
          <h2>上传头像</h2>
          <p class="upload-desc">支持 png、jpg、jpeg、gif 图片，上传后会保存到 frontend/img 目录。</p>

          <div class="avatar-preview">
            <img v-if="previewImage" :src="previewImage" class="avatar-large" :alt="displayNameForView" />
            <div v-else class="avatar-large avatar-fallback">
              {{ initials }}
            </div>
          </div>

          <label class="upload-field">
            <span>选择图片</span>
            <input type="file" accept=".png,.jpg,.jpeg,.gif,image/png,image/jpeg,image/gif" @change="handleFileChange" />
          </label>

          <p class="upload-name" v-if="selectedFile">{{ selectedFile.name }}</p>

          <div class="upload-actions">
            <button type="button" class="secondary-btn" @click="router.push(`/user/${store.user.id}/edit`)">返回资料编辑</button>
            <button type="button" class="primary-btn" :disabled="uploading || !selectedFile" @click="handleUpload">
              {{ uploading ? '上传中...' : '上传头像' }}
            </button>
          </div>

          <p v-if="message" :class="success ? 'feedback success' : 'feedback error'">
            {{ message }}
          </p>
        </div>
      </div>
    </div>

    <div class="empty-card" v-else>
      <h3>你还没有登录</h3>
      <p>请先登录后再上传头像。</p>
    </div>
  </section>
</template>

<script setup>
/*
	这个页面负责上传并更新当前用户头像。
*/
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { appStore as store, refreshCurrentUser, uploadAvatarAndSync } from '../store/appStore'

const router = useRouter()
const ready = ref(false)
const uploading = ref(false)
const selectedFile = ref(null)
const previewUrl = ref('')
const message = ref('')
const success = ref(false)

const user = computed(() => store.user)
const displayNameForView = computed(() => user.value.displayName || user.value.userName || '用户')
const initials = computed(() => displayNameForView.value.slice(0, 1).toUpperCase())
const previewImage = computed(() => previewUrl.value || (user.value.imageRoute ? `/img/${user.value.imageRoute}` : ''))

onMounted(async () => {
  const currentUser = await refreshCurrentUser()
  ready.value = true

  if (!currentUser.isLogin) {
    setTimeout(() => router.push('/login'), 1200)
  }
})

function handleFileChange(event) {
  const [file] = event.target.files || []
  selectedFile.value = file || null
  message.value = ''

  if (!file) {
    previewUrl.value = ''
    return
  }

  previewUrl.value = URL.createObjectURL(file)
}

async function handleUpload() {
  if (!selectedFile.value) {
    return
  }

  uploading.value = true
  message.value = ''

  try {
    await uploadAvatarAndSync(selectedFile.value)
    success.value = true
    message.value = '头像上传成功，1 秒后返回资料编辑页'

    setTimeout(() => {
      router.push(`/user/${store.user.id}/edit`)
    }, 1000)
  } catch (error) {
    success.value = false
    message.value = error.message
  } finally {
    uploading.value = false
  }
}
</script>

<style scoped>
.upload-layout {
  display: flex;
  justify-content: center;
}

.upload-card {
  width: min(100%, 520px);
  display: grid;
  gap: 16px;
  padding: 28px;
  border-radius: 24px;
  background: rgba(255, 255, 255, 0.88);
  box-shadow: 0 18px 40px rgba(40, 58, 80, 0.08);
}

.upload-card h2 {
  margin: 0;
}

.upload-desc {
  margin: 0;
  color: #5f6f82;
}

.avatar-preview {
  display: flex;
  justify-content: center;
}

.upload-field {
  display: grid;
  gap: 8px;
}

.upload-field span {
  font-weight: 600;
  color: #203040;
}

.upload-name {
  margin: 0;
  color: #5f6f82;
}

.upload-actions {
  display: flex;
  gap: 12px;
  justify-content: flex-end;
}

.primary-btn,
.secondary-btn {
  border: none;
  border-radius: 14px;
  padding: 12px 16px;
  font-weight: 600;
  cursor: pointer;
}

.primary-btn {
  background: #203040;
  color: #fff;
}

.secondary-btn {
  background: #e8edf2;
  color: #203040;
}

.primary-btn:disabled {
  cursor: wait;
  opacity: 0.72;
}
</style>


