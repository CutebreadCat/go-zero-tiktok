<script setup>
/** 登录 + 注册（后端表单编码 + HttpOnly Cookie，前端只管提交与探测会话） */
import { ref, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'

const route = useRoute()
const router = useRouter()
const auth = useAuthStore()

const mode = ref('login')
const username = ref('')
const password = ref('')
const mfaCode = ref('')
const needMfa = ref(false)
const loading = ref(false)
const error = ref('')

const canSubmit = computed(() => username.value.trim() && password.value && !loading.value)

async function submit() {
  loading.value = true
  error.value = ''
  try {
    if (mode.value === 'login') {
      await auth.login(username.value.trim(), password.value, mfaCode.value)
    } else {
      await auth.register(username.value.trim(), password.value)
      await auth.login(username.value.trim(), password.value, '')
    }
    router.push(route.query.redirect || '/')
  } catch (e) {
    if (e.code === 10001 || /mfa/i.test(e.message)) needMfa.value = true
    error.value = e.message
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="login-view">
    <div class="box">
      <h1 class="logo">Go<span>Tik</span></h1>
      <p class="sub">go-zero-tiktok 前端 · 深色短视频体验</p>

      <div class="tabs">
        <button :class="{ on: mode === 'login' }" @click="mode = 'login'; error = ''">登录</button>
        <button :class="{ on: mode === 'register' }" @click="mode = 'register'; error = ''">注册</button>
      </div>

      <div class="field">
        <label>用户名</label>
        <input v-model="username" placeholder="username" @keyup.enter="submit" />
      </div>
      <div class="field">
        <label>密码</label>
        <input v-model="password" type="password" placeholder="password" @keyup.enter="submit" />
      </div>
      <div class="field" v-if="needMfa && mode === 'login'">
        <label>MFA 动态码（该账号已开启两步验证）</label>
        <input v-model="mfaCode" placeholder="6 位动态码" @keyup.enter="submit" />
      </div>

      <p v-if="error" class="error">{{ error }}</p>
      <button class="btn btn-primary submit" :disabled="!canSubmit" @click="submit">
        {{ loading ? '请稍候…' : mode === 'login' ? '登录' : '注册并登录' }}
      </button>
    </div>
  </div>
</template>

<style scoped>
.login-view {
  height: 100%; display: flex; align-items: center; justify-content: center;
  background: radial-gradient(1200px 600px at 70% -10%, rgba(254,44,85,.12), transparent), var(--bg);
}
.box { width: 380px; max-width: 92vw; }
.logo { font-size: 40px; font-weight: 800; text-align: center; }
.logo span { color: var(--accent); }
.sub { text-align: center; color: var(--text-dim); margin: 6px 0 26px; font-size: 13px; }
.tabs {
  display: flex; background: var(--bg-elevated); border: 1px solid var(--border);
  border-radius: 999px; padding: 4px; margin-bottom: 20px;
}
.tabs button { flex: 1; padding: 9px 0; border-radius: 999px; color: var(--text-dim); font-weight: 600; }
.tabs button.on { background: var(--accent); color: #fff; }
.error { color: var(--accent); font-size: 13px; margin: -6px 0 10px; }
.submit { width: 100%; padding: 12px; }
</style>
