<template>
  <header class="top-bar">
    <div class="crumb">{{ title }}</div>
    <button class="btn-ghost net" :title="apiBase || '同源模式'" @click="$emit('open-settings')">
      <span class="dot" :class="{ ok: online }"></span>
      {{ online ? '已连接' : '未连接' }}
    </button>
  </header>
</template>

<script>
export default { emits: ['open-settings'] }
</script>

<script setup>
import { computed, onMounted, ref } from 'vue'
import { useRoute } from 'vue-router'
import { getApiBase } from '../config/appConfig'

const route = useRoute()
const apiBase = getApiBase()
const online = ref(false)
const title = computed(() => {
  const map = { feed: '推荐流', search: '搜索', publish: '发布视频', messages: '消息中心', profile: '个人主页', login: '登录' }
  return map[route.name] || 'GoTik'
})
onMounted(async () => {
  // 轻量连通性探测
  try {
    const base = apiBase
    const resp = await fetch((base || '') + '/videos/popular?page_num=1&page_size=1', { credentials: 'include' })
    online.value = resp.ok || resp.status === 401
  } catch { online.value = false }
})
</script>

<style scoped>
.top-bar {
  height: 52px; flex-shrink: 0; display: flex; align-items: center;
  justify-content: space-between; padding: 0 24px;
  border-bottom: 1px solid var(--border); background: var(--bg-elevated);
}
.crumb { font-weight: 600; }
.net { display: flex; align-items: center; gap: 7px; font-size: 13px; color: var(--text-dim); }
.dot { width: 8px; height: 8px; border-radius: 50%; background: var(--accent); }
.dot.ok { background: var(--success); }
</style>
