<script setup>
/**
 * 个人主页：作品 / 点赞 / 收藏 三个 Tab + MFA 绑定区。
 * 支持路由 /users/:id/videos（他人作品）与 /profile（自己，含点赞收藏）。
 */
import { ref, computed, onMounted, watch } from 'vue'
import { useRoute } from 'vue-router'
import { getUserVideos } from '../api/video'
import { getMyLikes, getMyFavorites } from '../api/interaction'
import { getFollowing, getFollowers, getFriends, follow as followUser, unfollow as unfollowUser } from '../api/communication'
import { getMfaQR, bindMfa } from '../api/user'
import { useAuthStore } from '../stores/auth'
import { compact } from '../utils/format'

const route = useRoute()
const auth = useAuthStore()

const viewedUserId = computed(() => String(route.params.id || auth.userId || ''))
const isSelf = computed(() => auth.isLoggedIn && viewedUserId.value === String(auth.userId))

const tab = ref('videos')
const videos = ref([])
const listTotal = ref(0)
const loading = ref(false)
const error = ref('')

// 关注 / 粉丝 / 好友
const relationTab = ref('')
const relationList = ref([])
const relationCount = ref(0)

// MFA
const mfaQR = ref(null)
const mfaSecretInput = ref('')
const mfaCodeInput = ref('')
const mfaMsg = ref('')

async function loadVideos() {
  loading.value = true
  error.value = ''
  try {
    if (isSelf.value && tab.value === 'likes') {
      const r = await getMyLikes()
      videos.value = r.video_list || []
      listTotal.value = r.like_count ?? videos.value.length
    } else if (isSelf.value && tab.value === 'favorites') {
      const r = await getMyFavorites()
      videos.value = r.video_list || []
      listTotal.value = r.favorite_count ?? videos.value.length
    } else {
      const r = await getUserVideos(viewedUserId.value)
      videos.value = r.videos || []
      listTotal.value = r.total ?? videos.value.length
    }
  } catch (e) {
    error.value = e.message
  } finally {
    loading.value = false
  }
}

async function loadRelation(kind) {
  relationTab.value = kind
  error.value = ''
  try {
    const q = { pageNum: 1, pageSize: 50 }
    if (kind === 'following') { const r = await getFollowing(q); relationList.value = r.subscriber_list || []; relationCount.value = r.subscriber_count ?? 0 }
    else if (kind === 'followers') { const r = await getFollowers(q); relationList.value = r.fans_list || []; relationCount.value = r.fans_count ?? 0 }
    else { const r = await getFriends(q); relationList.value = r.friend_list || []; relationCount.value = r.friend_count ?? 0 }
  } catch (e) { error.value = e.message }
}

async function toggleFollow(uid) {
  try {
    if (uid._following) { await unfollowUser(uid.user_id); uid._following = false }
    else { await followUser(uid.user_id); uid._following = true }
  } catch (e) { error.value = e.message }
}

async function loadMfaQR() {
  try {
    mfaQR.value = await getMfaQR()
    mfaSecretInput.value = mfaQR.value.mfa_secret || ''
  } catch (e) { mfaMsg.value = e.message }
}
async function doBindMfa() {
  try {
    await bindMfa(mfaSecretInput.value, mfaCodeInput.value)
    mfaMsg.value = '✅ 绑定成功'
  } catch (e) { mfaMsg.value = e.message }
}

watch([tab, viewedUserId], loadVideos, { immediate: false })
onMounted(() => { loadVideos(); if (isSelf.value) loadMfaQR() })
</script>

<template>
  <div class="page-pad">
    <div class="profile-head card">
      <div class="p-info">
        <h2>{{ isSelf ? auth.user?.username : `用户 ${viewedUserId}` }}</h2>
        <p class="p-sub">作品 {{ compact(listTotal) }} · 用户 ID {{ viewedUserId }}</p>
      </div>
      <div class="p-relations" v-if="isSelf">
        <button class="rel" @click="loadRelation('following')">关注</button>
        <button class="rel" @click="loadRelation('followers')">粉丝</button>
        <button class="rel" @click="loadRelation('friends')">好友</button>
      </div>
    </div>

    <div class="tabs" v-if="isSelf">
      <button :class="{ on: tab === 'videos' }" @click="tab = 'videos'">作品</button>
      <button :class="{ on: tab === 'likes' }" @click="tab = 'likes'">点赞</button>
      <button :class="{ on: tab === 'favorites' }" @click="tab = 'favorites'">收藏</button>
    </div>

    <p v-if="error" class="err">{{ error }}</p>
    <div v-if="loading" style="padding:30px 0"><div class="spinner"></div></div>
    <div v-else-if="relationTab" class="card rel-panel">
      <div class="rel-head">
        <strong>{{ { following: '我的关注', followers: '我的粉丝', friends: '互关好友' }[relationTab] }}（{{ relationCount }}）</strong>
        <button class="btn-ghost" @click="relationTab = ''">✕</button>
      </div>
      <div v-if="!relationList.length" class="empty"><div class="icon">🫥</div>暂无数据</div>
      <div v-for="u in relationList" :key="u.user_id" class="rel-row">
        <span>用户 {{ String(u.user_id).slice(-6) }}</span>
        <button class="btn" @click="toggleFollow(u)">{{ u._following ? '已关注' : '关注' }}</button>
      </div>
    </div>
    <div v-else-if="!videos.length" class="empty"><div class="icon">🎬</div>还没有视频</div>
    <div v-else class="grid">
      <div v-for="v in videos" :key="v.video_id" class="result-card">
        <video :src="v.video_url" :poster="v.cover_url || undefined" controls preload="metadata"></video>
        <div class="rc-title">{{ v.title || '未命名视频' }}</div>
      </div>
    </div>

    <div v-if="isSelf" class="card mfa">
      <h3>MFA 两步验证</h3>
      <p class="mfa-tip">用 TOTP 应用（如 Google Authenticator）扫码绑定；手动输入密钥也可。</p>
      <div class="mfa-row" v-if="mfaQR?.qr_code_url">
        <img :src="mfaQR.qr_code_url" alt="MFA QR" class="qr" />
      </div>
      <div class="mfa-row">
        <input v-model="mfaSecretInput" placeholder="MFA Secret" />
        <input v-model="mfaCodeInput" placeholder="6 位动态码" style="max-width:140px" />
        <button class="btn btn-primary" @click="doBindMfa">绑定</button>
      </div>
      <p v-if="mfaMsg" class="mfa-msg">{{ mfaMsg }}</p>
    </div>
  </div>
</template>

<style scoped>
.profile-head { display: flex; justify-content: space-between; align-items: center; margin-bottom: 18px; }
.p-sub { color: var(--text-dim); font-size: 13px; margin-top: 4px; }
.p-relations { display: flex; gap: 10px; }
.rel { padding: 8px 16px; background: var(--bg-elevated); border: 1px solid var(--border); border-radius: var(--radius); }
.tabs { display: flex; gap: 6px; margin-bottom: 16px; }
.tabs button { padding: 8px 20px; border-radius: 999px; color: var(--text-dim); background: var(--bg-elevated); border: 1px solid var(--border); }
.tabs button.on { color: #fff; background: var(--accent); border-color: var(--accent); }
.grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(220px, 1fr)); gap: 14px; }
.result-card { background: var(--bg-card); border: 1px solid var(--border); border-radius: var(--radius-lg); padding: 8px; }
.result-card video { width: 100%; aspect-ratio: 16/10; object-fit: cover; border-radius: var(--radius); background: #000; display: block; }
.rc-title { font-size: 13px; margin-top: 8px; }
.err { color: var(--accent); margin-bottom: 12px; }
.rel-panel { margin-bottom: 16px; }
.rel-head { display: flex; justify-content: space-between; margin-bottom: 10px; }
.rel-row { display: flex; justify-content: space-between; align-items: center; padding: 8px 0; border-bottom: 1px solid var(--border); }
.rel-row:last-child { border-bottom: none; }
.mfa { margin-top: 20px; }
.mfa-tip { font-size: 12px; color: var(--text-dim); margin-bottom: 10px; }
.mfa-row { display: flex; gap: 10px; margin-bottom: 10px; align-items: center; }
.qr { width: 120px; height: 120px; background: #fff; padding: 6px; border-radius: 8px; }
.mfa-msg { font-size: 13px; color: var(--success); }
</style>
