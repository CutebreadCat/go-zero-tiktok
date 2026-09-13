<script setup>
/** 发布页：multipart 上传视频（对应 POST /videos），发布成功后跳转推荐流 */
import { ref, computed } from 'vue'
import { useRouter } from 'vue-router'
import { publishVideo } from '../api/video'

const router = useRouter()

const file = ref(null)
const previewURL = ref('')
const title = ref('')
const description = ref('')
const publishing = ref(false)
const error = ref('')
const success = ref(false)

const canPublish = computed(() => file.value && title.value.trim() && !publishing.value)

function pickFile(e) {
  const f = e.target.files?.[0]
  if (!f) return
  file.value = f
  previewURL.value = URL.createObjectURL(f)
}

async function submit() {
  publishing.value = true
  error.value = ''
  success.value = false
  try {
    await publishVideo({ file: file.value, title: title.value.trim(), description: description.value.trim() })
    success.value = true
    setTimeout(() => router.push('/'), 900)
  } catch (e) {
    error.value = e.message
  } finally {
    publishing.value = false
  }
}
</script>

<template>
  <div class="page-pad">
    <h2 class="page-title">发布视频</h2>
    <div class="publish-grid">
      <div class="card upload-area">
        <label class="drop">
          <input type="file" accept="video/mp4,video/quicktime,video/webm" hidden @change="pickFile" />
          <video v-if="previewURL" :src="previewURL" controls></video>
          <div v-else class="drop-hint">
            <div class="icon">📤</div>
            <p>点击选择视频文件</p>
            <p class="tip">支持 MP4 / MOV / WEBM</p>
          </div>
        </label>
      </div>

      <div class="card form">
        <div class="field">
          <label>标题（必填）</label>
          <input v-model="title" maxlength="60" placeholder="给视频起个标题" />
        </div>
        <div class="field">
          <label>描述</label>
          <textarea v-model="description" rows="4" maxlength="200" placeholder="补充说明（可选）"></textarea>
        </div>
        <p v-if="error" class="error">{{ error }}</p>
        <p v-if="success" class="ok">✅ 发布成功，正在跳转…</p>
        <button class="btn btn-primary submit" :disabled="!canPublish" @click="submit">
          {{ publishing ? '上传中…' : '发布' }}
        </button>
      </div>
    </div>
  </div>
</template>

<style scoped>
.publish-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 18px; }
.upload-area { padding: 12px; }
.drop { display: block; cursor: pointer; }
.drop video { width: 100%; max-height: 360px; border-radius: var(--radius); background: #000; display: block; }
.drop-hint {
  height: 320px; border: 2px dashed var(--border); border-radius: var(--radius);
  display: flex; flex-direction: column; align-items: center; justify-content: center; gap: 8px;
  color: var(--text-dim);
}
.drop-hint .icon { font-size: 40px; }
.drop-hint .tip { font-size: 12px; }
.error { color: var(--accent); font-size: 13px; margin-bottom: 10px; }
.ok { color: var(--success); font-size: 13px; margin-bottom: 10px; }
.submit { width: 100%; padding: 12px; }
@media (max-width: 768px) { .publish-grid { grid-template-columns: 1fr; } }
</style>
