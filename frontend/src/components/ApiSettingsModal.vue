<script setup>
/**
 * API 连接设置面板 —— "随时切换后端 IP"的入口。
 * 修改保存后下一个请求立即生效（http 层每次动态读取）。
 */
import { ref, onMounted } from 'vue'
import { getApiBase, setApiBase, isSameOriginMode } from '../config/appConfig'
import { useAuthStore } from '../stores/auth'

const emit = defineEmits(['close'])
const auth = useAuthStore()

const input = ref('')
const testing = ref(false)
const result = ref(null) // { ok, msg }

onMounted(() => { input.value = getApiBase() })

async function save() {
  setApiBase(input.value)
  auth.checked = false // 强制下次路由重新探测会话（不同后端可能登录态不同）
  await auth.bootstrap()
  emit('close')
}

async function test() {
  testing.value = true
  result.value = null
  const saved = getApiBase()
  setApiBase(input.value) // 临时应用用于测试
  try {
    const resp = await fetch((getApiBase() || '') + '/videos/popular?page_num=1&page_size=1', { credentials: 'include' })
    result.value = resp.ok
      ? { ok: true, msg: '连接成功，后端在线' }
      : { ok: false, msg: `后端可达但返回 HTTP ${resp.status}（可能未启动网关路由）` }
  } catch {
    result.value = { ok: false, msg: '无法连接：检查地址、端口与后端进程' }
  } finally {
    setApiBase(saved)
    testing.value = false
  }
}
</script>

<template>
  <div class="modal-mask" @click.self="emit('close')">
    <div class="modal">
      <h3>连接设置</h3>
      <div class="field">
        <label>API 地址（后端网关）</label>
        <input
          v-model="input"
          placeholder="留空 = 同源模式；或填 http://192.168.1.5:8888"
          @keyup.enter="save"
        />
        <p class="hint">
          当前模式：{{ isSameOriginMode() ? '同源（请求发到前端所在域名）' : '自定义地址' }}。
          登录态凭据存于 Cookie，建议前端与后端同域或经代理部署。
        </p>
      </div>
      <p v-if="result" class="result" :class="{ ok: result.ok }">{{ result.msg }}</p>
      <div class="actions">
        <button class="btn" :disabled="testing" @click="test">{{ testing ? '测试中…' : '测试连接' }}</button>
        <button class="btn btn-primary" :disabled="testing" @click="save">保存</button>
      </div>
    </div>
  </div>
</template>

<style scoped>
.hint { font-size: 12px; color: var(--text-dim); margin-top: 8px; line-height: 1.6; }
.result { margin-top: 10px; font-size: 13px; color: var(--accent); }
.result.ok { color: var(--success); }
</style>
