<script setup>
/**
 * Feed 主页：四种 scene 切换 + 全屏上下滑动 + 自动播放 + 行为埋点。
 * 调用流程：getFeed(scene, cursor) → VideoCard 渲染 → IntersectionObserver
 * 判定可见 → impression 上报 → 播放/完播上报（对应推荐飞轮的客户端侧）。
 */
import { ref, onMounted, onBeforeUnmount, nextTick, watch } from 'vue'
import { useRouter } from 'vue-router'
import { getFeed } from '../api/video'
import { likeVideo, unlikeVideo, favoriteVideo, unfavoriteVideo } from '../api/interaction'
import { follow as followUser } from '../api/communication'
import { useAuthStore } from '../stores/auth'
import { beginFeedSession, trackImpression, trackPlay, trackComplete, trackAction } from '../utils/tracker'
import VideoCard from '../components/VideoCard.vue'
import CommentPanel from '../components/CommentPanel.vue'

const auth = useAuthStore()
const router = useRouter()

const scenes = [
  { key: 'recommend', label: '推荐' },
  { key: 'timeline', label: '关注' },
  { key: 'hot', label: '热门' },
]
const scene = ref('recommend')
const items = ref([])
const cursor = ref('')
const hasMore = ref(false)
const loading = ref(false)
const error = ref('')
const activeIndex = ref(0)
const commentFor = ref(null) // 正在查看评论的 item

let session = null
const feedEl = ref(null)
let observer = null
const visibleTimers = new WeakMap()

async function loadFeed(reset = true) {
  if (loading.value) return
  loading.value = true
  error.value = ''
  try {
    if (reset) {
      cursor.value = ''
      items.value = []
      session = beginFeedSession(scene.value)
    }
    const resp = await getFeed({ scene: scene.value, cursor: cursor.value, limit: 5 })
    items.value.push(...(resp.items || []))
    cursor.value = resp.next_cursor || ''
    hasMore.value = !!resp.has_more
    await nextTick()
    observeItems()
  } catch (e) {
    error.value = e.message
  } finally {
    loading.value = false
  }
}

/** 可见性判定：可见 ≥50% 且停留 ≥800ms 记一次 impression（真实曝光口径） */
function observeItems() {
  if (!observer) {
    observer = new IntersectionObserver(
      (entries) => {
        for (const en of entries) {
          const idx = Number(en.target.dataset.index)
          const vid = items.value[idx]?.videos?.video_id
          if (!vid) continue
          if (en.intersectionRatio >= 0.5) {
            if (!visibleTimers.has(en.target)) {
              visibleTimers.set(
                en.target,
                setTimeout(() => {
                  if (session) trackImpression(vid, session)
                  activeIndex.value = idx
                  playAt(idx)
                }, 800)
              )
            }
          } else {
            clearTimeout(visibleTimers.get(en.target))
            visibleTimers.delete(en.target)
            pauseAt(idx)
          }
        }
      },
      { threshold: [0, 0.5, 1] }
    )
  }
  feedEl.value?.querySelectorAll('.feed-slot').forEach((el) => observer.observe(el))
}

function playAt(idx) {
  const el = feedEl.value?.querySelector(`[data-index="${idx}"] video`)
  el?.play().catch(() => {})
}
function pauseAt(idx) {
  const el = feedEl.value?.querySelector(`[data-index="${idx}"] video`)
  el?.pause()
}

function onPlay(item) {
  if (session) trackPlay(item.videos.video_id, session, 0)
}
function onEnded(item) {
  const v = item.videos
  if (session) trackComplete(v.video_id, session, v.duration_ms || 0, v.duration_ms || 0)
  // 自动滑到下一条
  if (activeIndex.value < items.value.length - 1) {
    scrollToIndex(activeIndex.value + 1)
  }
}

function scrollToIndex(idx) {
  const el = feedEl.value?.querySelector(`[data-index="${idx}"]`)
  el?.scrollIntoView({ behavior: 'smooth' })
}

function onScroll() {
  const el = feedEl.value
  if (!el || loading.value || !hasMore.value) return
  if (el.scrollTop + el.clientHeight >= el.scrollHeight - 200) {
    loadFeed(false)
  }
}

async function toggleLike(item) {
  if (!auth.isLoggedIn) return router.push('/login')
  try {
    if (item.liked) { await unlikeVideo(item.videos.video_id); item.liked = false; item.videos_popular.like_count-- }
    else { await likeVideo(item.videos.video_id); item.liked = true; item.videos_popular.like_count++ }
  } catch (e) { error.value = e.message }
}
async function toggleFavorite(item) {
  if (!auth.isLoggedIn) return router.push('/login')
  try {
    if (item.favorited) { await unfavoriteVideo(item.videos.video_id); item.favorited = false; item.videos_popular.favorite_count-- }
    else { await favoriteVideo(item.videos.video_id); item.favorited = true; item.videos_popular.favorite_count++ }
  } catch (e) { error.value = e.message }
}
async function onFollow(item) {
  if (!auth.isLoggedIn) return router.push('/login')
  try {
    await followUser(item.author.user_id)
    trackAction('follow', item.videos.video_id, { follow_user_id: String(item.author.user_id), action: 'follow' })
  } catch (e) { error.value = e.message }
}

watch(scene, () => loadFeed(true))
onMounted(() => loadFeed(true))
onBeforeUnmount(() => observer?.disconnect())
</script>

<template>
  <div class="feed-view">
    <div class="scene-tabs">
      <button
        v-for="s in scenes" :key="s.key"
        class="tab" :class="{ active: scene === s.key }"
        @click="scene = s.key"
      >{{ s.label }}</button>
    </div>

    <div ref="feedEl" class="feed-scroll" @scroll.passive="onScroll">
      <div v-if="error && !items.length" class="empty">
        <div class="icon">📡</div>
        <p>{{ error }}</p>
        <button class="btn btn-primary" @click="loadFeed(true)">重试</button>
        <p class="tip">如果后端不在本机，请点右上角"连接设置"切换 API 地址</p>
      </div>
      <div v-else-if="!items.length && loading" class="empty"><div class="spinner"></div></div>

      <div
        v-for="(item, i) in items" :key="item.videos?.video_id ?? i"
        class="feed-slot" :data-index="i"
      >
        <VideoCard
          :item="item"
          :active="i === activeIndex"
          @toggle-like="toggleLike"
          @toggle-favorite="toggleFavorite"
          @open-comments="commentFor = item"
          @follow="onFollow"
          @play="onPlay(item)"
          @ended="onEnded(item)"
        />
      </div>

      <div v-if="loading && items.length" class="loading-more"><div class="spinner"></div></div>
      <div v-if="items.length && !hasMore && !loading" class="feed-end">— 已经到底啦 —</div>
    </div>

    <CommentPanel
      v-if="commentFor"
      :video-id="String(commentFor.videos.video_id)"
      @close="commentFor = null"
    />
  </div>
</template>

<style scoped>
.feed-view { height: 100%; display: flex; }
.scene-tabs {
  position: absolute; top: 14px; left: 50%; transform: translateX(-50%);
  display: flex; gap: 6px; z-index: 10;
  background: rgba(23,23,29,.85); backdrop-filter: blur(8px);
  border-radius: 999px; padding: 4px;
}
.tab { padding: 7px 20px; border-radius: 999px; color: var(--text-dim); font-weight: 600; }
.tab.active { background: var(--bg-card); color: var(--text); }
.feed-scroll {
  flex: 1; overflow-y: auto; scroll-snap-type: y mandatory;
  display: flex; flex-direction: column;
}
.feed-slot {
  height: 100%; flex-shrink: 0; scroll-snap-align: start;
  padding: 10px 14px; display: flex;
}
.feed-slot > * { flex: 1; }
.loading-more { padding: 20px 0; }
.feed-end { text-align: center; color: var(--text-dim); padding: 24px 0; }
.empty { height: 100%; justify-content: center; padding: 0 24px; text-align: center; }
.tip { font-size: 12px; opacity: .7; }
@media (max-width: 768px) {
  .feed-slot { padding: 0; }
}
</style>
