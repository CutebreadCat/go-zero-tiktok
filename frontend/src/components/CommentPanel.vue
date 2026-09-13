<script setup>
/** 评论面板：列表 + 发表 + 回复 + 评论点赞（对应 interaction 域契约） */
import { ref, onMounted } from 'vue'
import { getComments, postComment, postReply, likeComment, unlikeComment, deleteComment } from '../api/interaction'
import { useAuthStore } from '../stores/auth'
import { timeAgo, avatarColor } from '../utils/format'
import { ApiError } from '../api/http'

const props = defineProps({ videoId: { type: String, required: true } })
const auth = useAuthStore()

const comments = ref([])
const loading = ref(false)
const page = ref(1)
const hasMore = ref(false)
const text = ref('')
const replyTo = ref(null) // 正在回复的评论
const error = ref('')

async function load(reset = false) {
  if (loading.value) return
  loading.value = true
  error.value = ''
  try {
    if (reset) { page.value = 1; comments.value = [] }
    const resp = await getComments(props.videoId, { pageNum: page.value, pageSize: 20 })
    comments.value.push(...(resp.comments || []))
    const total = resp.total ?? 0
    hasMore.value = comments.value.length < total
    page.value += 1
  } catch (e) {
    error.value = e.message
  } finally {
    loading.value = false
  }
}

async function submit() {
  const content = text.value.trim()
  if (!content) return
  try {
    if (replyTo.value) {
      await postReply(replyTo.value.comment_id, content)
    } else {
      await postComment(props.videoId, content)
    }
    text.value = ''
    replyTo.value = null
    await load(true)
  } catch (e) {
    error.value = e.message
  }
}

async function toggleLike(c) {
  try {
    if (c._liked) { await unlikeComment(c.comment_id); c._liked = false; c.like_count = Math.max(0, (c.like_count || 0) - 1) }
    else { await likeComment(c.comment_id); c._liked = true; c.like_count = (c.like_count || 0) + 1 }
  } catch (e) {
    error.value = e.message
  }
}

async function remove(c) {
  try {
    await deleteComment(c.comment_id)
    comments.value = comments.value.filter((x) => x.comment_id !== c.comment_id)
  } catch (e) {
    error.value = e.message
  }
}

defineExpose({ load })
onMounted(() => load(true))
</script>

<template>
  <div class="panel">
    <div class="panel-head">
      <span>评论</span>
      <button class="btn-ghost" @click="$emit('close')">✕</button>
    </div>

    <div class="list">
      <div v-if="loading && !comments.length" class="spinner-wrap"><div class="spinner"></div></div>
      <div v-else-if="!comments.length" class="empty"><div class="icon">💬</div>还没有评论</div>
      <div v-for="c in comments" :key="c.comment_id" class="comment">
        <div class="avatar" :style="{ background: avatarColor(c.user_id ? String(c.user_id) : 'u') }">
          {{ 'U' }}
        </div>
        <div class="body">
          <div class="meta">
            <span class="uid">用户 {{ String(c.user_id).slice(-4) }}</span>
            <span class="time">{{ timeAgo(c.created_at) }}</span>
          </div>
          <div class="content">
            <span v-if="c.parent_comment_id && c.parent_comment_id !== '0'" class="reply-tag">回复</span>
            {{ c.content }}
          </div>
          <div class="ops">
            <button class="op" :class="{ on: c._liked }" @click="toggleLike(c)">
              👍 {{ c.like_count || 0 }}
            </button>
            <button class="op" @click="replyTo = replyTo?.comment_id === c.comment_id ? null : c">
              回复
            </button>
            <button
              v-if="auth.isLoggedIn && String(c.user_id) === String(auth.userId)"
              class="op danger" @click="remove(c)"
            >删除</button>
          </div>
          <div v-if="replyTo?.comment_id === c.comment_id" class="replying">
            正在回复该评论 · <button class="op" @click="replyTo = null">取消</button>
          </div>
        </div>
      </div>
      <div v-if="hasMore" class="more-wrap">
        <button class="btn" :disabled="loading" @click="load()">加载更多</button>
      </div>
    </div>

    <div class="composer" v-if="auth.isLoggedIn">
      <input
        v-model="text"
        :placeholder="replyTo ? `回复该评论…` : '说点什么…'"
        @keyup.enter="submit"
      />
      <button class="btn btn-primary" @click="submit">发送</button>
    </div>
    <div class="composer disabled" v-else>登录后参与评论</div>
    <p v-if="error" class="err">{{ error }}</p>
  </div>
</template>

<script>
export default { emits: ['close'] }
</script>

<style scoped>
.panel {
  width: 380px; max-width: 90vw; height: 100%;
  background: var(--bg-card); border-left: 1px solid var(--border);
  display: flex; flex-direction: column;
}
.panel-head {
  display: flex; justify-content: space-between; align-items: center;
  padding: 14px 16px; border-bottom: 1px solid var(--border); font-weight: 700;
}
.list { flex: 1; overflow-y: auto; padding: 12px 16px; }
.spinner-wrap { padding: 40px 0; }
.comment { display: flex; gap: 10px; margin-bottom: 16px; }
.avatar {
  width: 32px; height: 32px; border-radius: 50%; flex-shrink: 0;
  display: flex; align-items: center; justify-content: center; color: #fff; font-weight: 700;
}
.body { flex: 1; min-width: 0; }
.meta { display: flex; gap: 8px; font-size: 12px; color: var(--text-dim); margin-bottom: 2px; }
.content { font-size: 14px; white-space: pre-wrap; word-break: break-word; }
.reply-tag { color: var(--secondary); margin-right: 4px; font-size: 12px; }
.ops { display: flex; gap: 14px; margin-top: 4px; }
.op { font-size: 12px; color: var(--text-dim); padding: 0; }
.op.on { color: var(--accent); }
.op.danger:hover { color: var(--accent); }
.replying { font-size: 12px; color: var(--secondary); margin-top: 4px; }
.more-wrap { text-align: center; padding: 8px 0; }
.composer {
  display: flex; gap: 8px; padding: 12px 16px; border-top: 1px solid var(--border);
}
.composer.disabled { color: var(--text-dim); font-size: 13px; justify-content: center; }
.err { padding: 0 16px 12px; color: var(--accent); font-size: 12px; }
</style>
