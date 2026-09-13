<script setup>
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import { useMessageStore } from '../stores/messages'
import { avatarColor } from '../utils/format'

const route = useRoute()
const auth = useAuthStore()
const messages = useMessageStore()

const items = computed(() => [
  { name: 'feed', label: '推荐', icon: '🏠', to: '/' },
  { name: 'search', label: '搜索', icon: '🔍', to: '/search' },
  { name: 'publish', label: '发布', icon: '➕', to: '/publish', auth: true },
  { name: 'messages', label: '消息', icon: '✉️', to: '/messages', auth: true, badge: messages.unreadCount },
  { name: 'profile', label: '我的', icon: '👤', to: '/profile', auth: true },
])
</script>

<template>
  <nav class="side-nav">
    <div class="logo">Go<span>Tik</span></div>
    <router-link
      v-for="it in items"
      :key="it.name"
      :to="it.to"
      class="nav-item"
      :class="{ active: route.name === it.name }"
    >
      <span class="icon">{{ it.icon }}</span>
      <span class="label">{{ it.label }}</span>
      <span v-if="it.badge > 0" class="badge">{{ it.badge > 99 ? '99+' : it.badge }}</span>
    </router-link>

    <div class="nav-footer">
      <template v-if="auth.isLoggedIn">
        <div class="me">
          <div class="avatar" :style="{ background: avatarColor(auth.user?.username) }">
            {{ (auth.user?.username || '?')[0].toUpperCase() }}
          </div>
          <div class="me-meta">
            <div class="name">{{ auth.user?.username }}</div>
            <button class="btn-ghost link" @click="auth.logout(); $router.push('/login')">退出</button>
          </div>
        </div>
      </template>
      <router-link v-else to="/login" class="btn btn-primary login-btn">登录 / 注册</router-link>
      <button class="btn-ghost link" @click="$emit('open-settings')">⚙️ 连接设置</button>
    </div>
  </nav>
</template>

<script>
// 追加 emit 声明（与 setup 并存写法）
export default { emits: ['open-settings'] }
</script>

<style scoped>
.side-nav {
  width: var(--nav-w); flex-shrink: 0; height: 100%;
  background: var(--bg-elevated); border-right: 1px solid var(--border);
  display: flex; flex-direction: column; padding: 18px 12px;
}
.logo { font-size: 22px; font-weight: 800; padding: 0 10px 20px; }
.logo span { color: var(--accent); }
.nav-item {
  display: flex; align-items: center; gap: 12px;
  padding: 11px 12px; border-radius: var(--radius);
  color: var(--text-dim); margin-bottom: 4px; position: relative;
}
.nav-item:hover { background: var(--bg-card); color: var(--text); }
.nav-item.active { color: var(--accent); background: var(--accent-soft); font-weight: 600; }
.nav-item .icon { font-size: 18px; }
.badge {
  position: absolute; right: 12px;
  background: var(--accent); color: #fff; font-size: 11px;
  border-radius: 999px; padding: 1px 7px;
}
.nav-footer { margin-top: auto; display: flex; flex-direction: column; gap: 10px; padding: 0 6px; }
.me { display: flex; align-items: center; gap: 10px; }
.avatar {
  width: 36px; height: 36px; border-radius: 50%;
  display: flex; align-items: center; justify-content: center;
  font-weight: 700; color: #fff; flex-shrink: 0;
}
.me-meta { display: flex; flex-direction: column; }
.me-meta .name { font-weight: 600; font-size: 13px; }
.link { padding: 0; text-align: left; font-size: 12px; }
.login-btn { width: 100%; }
@media (max-width: 768px) {
  .side-nav { flex-direction: row; width: 100%; height: 56px; padding: 0 8px; align-items: center; border-right: none; border-top: 1px solid var(--border); position: fixed; bottom: 0; z-index: 50; }
  .logo, .nav-footer { display: none; }
  .nav-item { flex-direction: column; gap: 2px; flex: 1; padding: 6px 0; font-size: 12px; }
  .app-shell { flex-direction: column-reverse; }
}
</style>
