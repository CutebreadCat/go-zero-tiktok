import { defineStore } from 'pinia'
import * as userApi from '../api/user'
import { onAuthExpiredListen } from '../api/http'

/**
 * 认证与当前用户状态。
 * 鉴权细节：后端登录后发 HttpOnly Cookie（access + refresh），
 * token 对 JS 不可见（防 XSS 窃取），前端只维护"当前用户是谁"这一事实。
 */
export const useAuthStore = defineStore('auth', {
  state: () => ({
    user: null,       // UserBaseinfo | null
    checked: false,   // 是否完成过一次会话探测
    loading: false,
  }),
  getters: {
    isLoggedIn: (s) => !!s.user,
    userId: (s) => s.user?.user_id ?? null,
  },
  actions: {
    /** 会话探测：有 Cookie 就能换到用户信息；401 由 http 层先尝试刷新 */
    async bootstrap() {
      if (this.checked) return
      this.loading = true
      try {
        const resp = await userApi.fetchMe()
        this.user = resp.user
      } catch {
        this.user = null
      } finally {
        this.checked = true
        this.loading = false
      }
    },
    async login(username, password, mfaCode) {
      await userApi.login(username, password, mfaCode)
      await this.bootstrap(true)
      const resp = await userApi.fetchMe()
      this.user = resp.user
    },
    async register(username, password) {
      const resp = await userApi.register(username, password)
      return resp.user_id
    },
    /** 后端暂无登出接口：前端清状态（Cookie 到期自然失效） */
    logout() {
      this.user = null
    },
  },
})

// 登录过期 → 清空本地状态（路由守卫负责跳转）
export function setupAuthExpiry() {
  const auth = useAuthStore()
  onAuthExpiredListen(() => { auth.user = null; auth.checked = true })
}
