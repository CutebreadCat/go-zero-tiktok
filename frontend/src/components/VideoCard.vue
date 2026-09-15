<script setup>
/**
 * Feed 单条视频卡片：播放器 + 右侧操作栏 + 底部信息。
 * 对后端契约：Item { videos, videos_popular, author, liked, favorited }
 */
import { computed } from 'vue'
import { compact, avatarColor } from '../utils/format'

const props = defineProps({
  item: { type: Object, required: true },
  active: { type: Boolean, default: false },
})
const emit = defineEmits(['toggle-like', 'toggle-favorite', 'open-comments', 'follow', 'play', 'ended', 'visible', 'hidden'])

const video = computed(() => props.item.videos || {})
const stats = computed(() => props.item.videos_popular || {})
const author = computed(() => props.item.author || {})
</script>

<template>
  <div class="video-card" :class="{ active }">
    <video
      :src="video.video_url"
      :poster="video.cover_url || undefined"
      playsinline
      loop
      muted
      @play="$emit('play', $event)"
      @ended="$emit('ended', $event)"
    ></video>

    <div class="overlay">
      <div class="info">
        <div class="author-line">
          <span class="avatar" :style="{ background: avatarColor(author.username) }">
            {{ (author.username || '?')[0]?.toUpperCase() }}
          </span>
          <span class="name">@{{ author.username || '未知用户' }}</span>
        </div>
        <div class="title">{{ video.title || '未命名视频' }}</div>
        <div v-if="video.description" class="desc">{{ video.description }}</div>
      </div>

      <div class="actions">
        <button class="act" :class="{ on: item.liked }" @click="$emit('toggle-like', item)">
          <span class="ico">❤️</span>
          <span class="num">{{ compact(stats.like_count) }}</span>
        </button>
        <button class="act" :class="{ on: item.favorited }" @click="$emit('toggle-favorite', item)">
          <span class="ico">⭐</span>
          <span class="num">{{ compact(stats.favorite_count) }}</span>
        </button>
        <button class="act" @click="$emit('open-comments', item)">
          <span class="ico">💬</span>
          <span class="num">{{ compact(stats.comment_count) }}</span>
        </button>
        <button class="act" @click="$emit('follow', item)">
          <span class="ico">➕</span>
          <span class="num">关注</span>
        </button>
      </div>
    </div>
  </div>
</template>

<style scoped>
.video-card {
  height: 100%; width: 100%; position: relative;
  background: #000; overflow: hidden;
  border-radius: var(--radius-lg);
}
.video-card video {
  width: 100%; height: 100%; object-fit: contain;
  display: block;
}
.overlay {
  position: absolute; inset: 0;
  display: flex; align-items: flex-end; justify-content: space-between;
  padding: 20px 18px; pointer-events: none;
  background: linear-gradient(transparent 55%, rgba(0,0,0,.65));
  border-radius: var(--radius-lg);
}
.info { max-width: 70%; pointer-events: auto; }
.author-line { display: flex; align-items: center; gap: 8px; margin-bottom: 6px; }
.avatar {
  width: 34px; height: 34px; border-radius: 50%; flex-shrink: 0;
  display: flex; align-items: center; justify-content: center;
  font-weight: 700; color: #fff;
}
.name { font-weight: 700; text-shadow: 0 1px 3px rgba(0,0,0,.5); }
.title { font-size: 15px; margin-bottom: 2px; text-shadow: 0 1px 3px rgba(0,0,0,.5); }
.desc { font-size: 13px; color: #d8d8de; text-shadow: 0 1px 3px rgba(0,0,0,.5); }
.actions {
  display: flex; flex-direction: column; gap: 18px;
  align-items: center; pointer-events: auto;
}
.act { display: flex; flex-direction: column; align-items: center; gap: 2px; }
.act .ico { font-size: 30px; filter: grayscale(1) brightness(2); transition: filter .15s, transform .1s; }
.act:active .ico { transform: scale(1.25); }
.act.on .ico { filter: none; }
.act .num { font-size: 12px; color: #eee; }
@media (max-width: 768px) {
  .video-card { border-radius: 0; }
  .overlay { border-radius: 0; }
}
</style>
