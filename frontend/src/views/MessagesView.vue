<script setup>
/** 消息中心：类型筛选 + 游标分页 + 未读数 + 批量/单条已读（对应 communication 消息域） */
import { ref, onMounted, onBeforeUnmount } from 'vue'
import { getMessages } from '../api/communication'
import { useMessageStore } from '../stores/messages'
import { timeAgo } from '../utils/format'

const messages = useMessageStore()

const types = [
  { key: '', label: '全部' },
  { key: 'LIKE', label: '点赞' },
  { key: 'COMMENT', label: '评论' },
  { key: 'FOLLOW', label: '关注' },
  { key: 'SYSTEM', label: '系统' },
]
const type = ref('')
const items = ref([])
const cursor = ref('')
const hasMore = ref(false)
const loading = ref(false)
const error = ref('')

async function load(reset = true) {
  if (loading.value) return
  loading.value = true
  error.value = ''
  try {
    if (reset) { items.value = []; cursor.value = '' }
    const resp = await getMessages({ type: type.value, cursor: cursor.value, limit: 20 })
    items.value.push(...(resp.items || []))
    cursor.value = resp.next_cursor || ''
    hasMore.value = !!resp.has_more
  } catch (e) {
    error.value = e.message
  } finally {
    loading.value = false
  }
}

async function markOne(m) {
  if (m.is_read) return
  await messages.markRead([Number(m.message_id)])
  m.is_read = true
}

async function markAll() {
  await messages.markRead([])
  items.value.forEach((m) => { m.is_read = true })
}

const typeIcon = { LIKE: '❤️', COMMENT: '💬', FOLLOW: '👤', SYSTEM: '📣' }
const typeText = { LIKE: '赞了你的视频', COMMENT: '评论了你的视频', FOLLOW: '关注了你', SYSTEM: '系统通知' }

onMounted(() => { load(true); messages.startPolling() })
onBeforeUnmount(() => messages.stopPolling())
</script>

<template>
  <div class="page-pad msg-page">
    <div class="head">
      <h2 class="page-title">消息中心</h2>
      <div class="head-ops">
        <span class="unread tag" v-if="messages.unreadCount > 0">{{ messages.unreadCount }} 条未读</span>
        <button class="btn" @click="markAll">全部已读</button>
      </div>
    </div>

    <div class="type-tabs">
      <button
        v-for="t in types" :key="t.key"
        :class="{ on: type === t.key }"
        @click="type = t.key; load(true)"
      >{{ t.label }}</button>
    </div>

    <div v-if="error" class="empty"><div class="icon">📡</div>{{ error }}</div>
    <div v-else-if="loading && !items.length" style="padding:30px 0"><div class="spinner"></div></div>
    <div v-else-if="!items.length" class="empty"><div class="icon">✉️</div>暂无消息</div>

    <div v-else class="msg-list">
      <div
        v-for="m in items" :key="m.message_id"
        class="msg-row" :class="{ unread: !m.is_read }"
        @click="markOne(m)"
      >
        <span class="m-ico">{{ typeIcon[m.type] || '📣' }}</span>
        <div class="m-body">
          <div class="m-title">
            <strong v-if="m.sender_nickname">{{ m.sender_nickname }}</strong>
            <strong v-else>用户 {{ String(m.sender_id).slice(-6) }}</strong>
            <span class="m-action">{{ typeText[m.type] || m.type }}</span>
            <span v-if="!m.is_read" class="dot"></span>
          </div>
          <div class="m-content">{{ m.content }}</div>
        </div>
        <span class="m-time">{{ timeAgo(m.created_at) }}</span>
      </div>

      <div v-if="hasMore" class="more">
        <button class="btn" :disabled="loading" @click="load(false)">加载更多</button>
      </div>
    </div>
  </div>
</template>

<style scoped>
.msg-page { max-width: 720px; }
.head { display: flex; justify-content: space-between; align-items: center; }
.head-ops { display: flex; align-items: center; gap: 10px; }
.type-tabs { display: flex; gap: 6px; margin: 4px 0 16px; }
.type-tabs button {
  padding: 7px 16px; border-radius: 999px; font-size: 13px;
  background: var(--bg-elevated); border: 1px solid var(--border); color: var(--text-dim);
}
.type-tabs button.on { color: #fff; background: var(--accent); border-color: var(--accent); }
.msg-list { display: flex; flex-direction: column; gap: 8px; }
.msg-row {
  display: flex; align-items: center; gap: 12px;
  background: var(--bg-card); border: 1px solid var(--border);
  border-radius: var(--radius-lg); padding: 14px 16px; cursor: pointer;
}
.msg-row.unread { border-left: 3px solid var(--accent); }
.m-ico { font-size: 22px; }
.m-body { flex: 1; min-width: 0; }
.m-title { display: flex; align-items: center; gap: 8px; font-size: 14px; }
.m-action { color: var(--text-dim); font-size: 13px; }
.dot { width: 7px; height: 7px; background: var(--accent); border-radius: 50%; }
.m-content { font-size: 13px; color: var(--text-dim); white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.m-time { font-size: 12px; color: var(--text-dim); flex-shrink: 0; }
.more { text-align: center; padding: 12px 0; }
</style>
