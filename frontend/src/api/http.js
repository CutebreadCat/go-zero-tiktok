/**
 * HTTP 层：全项目唯一的 axios 实例。
 *
 * 职责边界（页面/组件永远不直接 import axios）：
 *   - 动态拼接 API Base（配置化入口，见 config/appConfig.js）
 *   - 携带 Cookie 凭据（后端是 HttpOnly Cookie 鉴权，token 不进 JS）
 *   - 统一解析 { base: { status_code, status_msg } } 业务信封
 *   - 401 时自动尝试一次刷新令牌并重放原请求
 */
import axios from 'axios'
import { getApiBase } from '../config/appConfig'

const http = axios.create({
  timeout: 15000,
  withCredentials: true, // 关键：跨端口/跨 IP 也要带上 Cookie
})

// 每次请求动态取 Base，设置面板切换后立刻生效
http.interceptors.request.use((config) => {
  const base = getApiBase()
  if (base && !/^https?:\/\//i.test(config.url)) {
    config.url = base + config.url
  }
  return config
})

// 业务信封解析：go-zero 网关统一返回 { base: { status_code, status_msg } }
http.interceptors.response.use(
  (resp) => {
    const body = resp.data
    if (body && body.base && body.base.status_code !== 0) {
      return Promise.reject(new ApiError(body.base.status_code, body.base.status_msg))
    }
    return body
  },
  async (error) => {
    const { response, config } = error
    if (!response) {
      return Promise.reject(new ApiError(-1, '网络不可达：请检查 API 地址与后端是否在线'))
    }
    // 401 → 尝试刷新一次再重放（刷新接口本身失败则放行给调用方）
    if (response.status === 401 && !config._retried && !config.url.includes('/sessions/refresh')) {
      config._retried = true
      const refreshed = await tryRefresh()
      if (refreshed) return http(config)
      onAuthExpired.forEach((fn) => fn())
    }
    const msg = response.data?.base?.status_msg || `请求失败（HTTP ${response.status}）`
    return Promise.reject(new ApiError(response.data?.base?.status_code ?? response.status, msg))
  }
)

export class ApiError extends Error {
  constructor(code, message) {
    super(message)
    this.code = code
  }
}

// 刷新与登录过期回调：由 auth store 注册，避免 store 与 http 循环依赖
const onAuthExpired = new Set()
export function onAuthExpiredListen(fn) { onAuthExpired.add(fn) }

let refreshing = null
function tryRefresh() {
  if (!refreshing) {
    refreshing = axios
      .post(joinBase('/sessions/refresh'), null, { withCredentials: true })
      .then(() => true)
      .catch(() => false)
      .finally(() => { refreshing = null })
  }
  return refreshing
}

function joinBase(path) {
  const base = getApiBase()
  return base ? base + path : path
}

export default http
