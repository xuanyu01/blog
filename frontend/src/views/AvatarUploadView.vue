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
          <p class="upload-desc">支持 PNG、JPG、JPEG、GIF 图片，文件不超过 512KB，尺寸不超过 1024×1024。</p>

          <div class="avatar-preview">
            <img v-if="previewImage" :src="previewImage" class="avatar-large" :alt="displayNameForView" />
            <div v-else class="avatar-large avatar-fallback">
              {{ initials }}
            </div>
          </div>

          <label class="upload-field">
            <span>选择图片</span>
            <input ref="fileInput" type="file" accept=".png,.jpg,.jpeg,.gif,image/png,image/jpeg,image/gif" @change="handleFileChange" />
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
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { appStore as store, refreshCurrentUser, uploadAvatarAndSync } from '../store/appStore'

const MAX_AVATAR_SIZE = 512 * 1024
const ALLOWED_AVATAR_TYPES = new Set(['image/png', 'image/jpeg', 'image/gif'])

const router = useRouter()
const ready = ref(false)
const uploading = ref(false)
const selectedFile = ref(null)
const previewUrl = ref('')
const message = ref('')
const success = ref(false)
const fileInput = ref(null)

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

onBeforeUnmount(() => {
  revokePreviewUrl()
})

function handleFileChange(event) {
  const [file] = event.target.files || []
  resetSelectedFile()

  if (!file) {
    return
  }

  if (!ALLOWED_AVATAR_TYPES.has(file.type)) {
    message.value = '头像仅支持 PNG、JPG、JPEG、GIF 格式'
    return
  }

  if (file.size > MAX_AVATAR_SIZE) {
    message.value = '头像文件不能超过 512KB'
    return
  }

  selectedFile.value = file
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
    message.value = avatarErrorText(error.message)
  } finally {
    uploading.value = false
  }
}

function resetSelectedFile() {
  selectedFile.value = null
  success.value = false
  message.value = ''
  revokePreviewUrl()

  if (fileInput.value) {
    fileInput.value.value = ''
  }
}

function revokePreviewUrl() {
  if (previewUrl.value) {
    URL.revokeObjectURL(previewUrl.value)
    previewUrl.value = ''
  }
}

function avatarErrorText(rawMessage) {
  const errorMap = {
    'avatar file cannot be larger than 512 KB': '头像文件不能超过 512KB',
    'avatar image dimensions cannot exceed 1024x1024': '头像尺寸不能超过 1024×1024',
    'only png jpg jpeg gif images are allowed': '头像仅支持 PNG、JPG、JPEG、GIF 格式',
    'file type does not match the allowed image format': '文件内容与图片格式不匹配',
    'failed to read avatar image': '无法读取头像图片，请换一张图片重试',
    'avatar file is required': '请选择头像图片',
    unauthorized: '登录状态已失效，请重新登录'
  }

  return errorMap[rawMessage] || rawMessage || '头像上传失败，请稍后重试'
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
