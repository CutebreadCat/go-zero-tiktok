<script setup>
/** 搜索页：关键词 → 视频网格（对应 GET /videos/search） */
import { ref } from 'vue'
import { searchVideos } from '../api/video'
import { compact } from '../utils/format'

const keyword = ref('')
const results = ref([])
const searching = ref(false)
const searched = ref(false)
const error = ref('')

async function doSearch() {
  const kw = keyword.value.trim()
  if (!kw) return
  searching.value = true
  error.value = ''
  try {
    const resp = await searchVideos(kw)
    results.value = resp.videos || []
    searched.value = true
  } catch (e) {
    error.value = e.message
  } finally {
    searching.value = false
  }
}
</script>

<template>
  <div class="page-pad">
    <h2 class="page-title">搜索</h2>
    <div class="search-bar">
      <input v-model="keyword" placeholder="搜索视频标题或描述…" @keyup.enter="doSearch" />
      <button class="btn btn-primary" :disabled="searching" @click="doSearch">
        {{ searching ? '搜索中…' : '搜索' }}
      </button>
    </div>

    <div v-if="error" class="empty"><div class="icon">📡</div>{{ error }}</div>
    <div v-else-if="searching" style="padding:40px 0"><div class="spinner"></div></div>
    <div v-else-if="searched && !results.length" class="empty"><div class="icon">🎬</div>没有找到相关视频</div>

    <div v-else class="grid">
      <div v-for="v in results" :key="v.video_id" class="result-card">
        <video :src="v.video_url" :poster="v.cover_url || undefined" controls preload="metadata"></video>
        <div class="rc-body">
          <div class="rc-title">{{ v.title || '未命名视频' }}</div>
          <div class="rc-desc">{{ v.description }}</div>
          <div class="rc-meta">作者 {{ v.author_id }} · {{ v.created_at?.slice(0, 10) }}</div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.search-bar { display: flex; gap: 10px; margin-bottom: 22px; }
.grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(240px, 1fr)); gap: 16px; }
.result-card { background: var(--bg-card); border: 1px solid var(--border); border-radius: var(--radius-lg); overflow: hidden; }
.result-card video { width: 100%; aspect-ratio: 16/10; object-fit: cover; background: #000; display: block; }
.rc-body { padding: 12px; }
.rc-title { font-weight: 600; margin-bottom: 4px; }
.rc-desc { font-size: 13px; color: var(--text-dim); margin-bottom: 8px; min-height: 18px; }
.rc-meta { font-size: 12px; color: var(--text-dim); }
</style>
