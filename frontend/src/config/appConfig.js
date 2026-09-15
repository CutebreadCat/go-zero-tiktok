/**
 * 应用运行时配置 —— "随时切换后端地址"的唯一入口。
 *
 * 三层优先级（高 → 低）：
 *   1. localStorage（设置面板运行时修改，立即生效）
 *   2. 构建期环境变量 VITE_API_BASE（.env / .env.production）
 *   3. 同源模式（留空或根相对路径如 /api：请求发到当前页面域名，
 *      开发由 Vite 代理转发、生产由 Nginx 反代转发到网关）
 *
 * 所有 API 请求统一经过 api/http.js，它每次请求都会调用 getApiBase()，
 * 因此这里改完地址，下一个请求立即生效，无需刷新或重启。
 */

const STORAGE_KEY = 'gotik:api-base'

/** 读取当前生效的 API Base URL（'' 表示同源模式） */
export function getApiBase() {
  const stored = localStorage.getItem(STORAGE_KEY)
  if (stored !== null) return normalize(stored)
  const fromEnv = import.meta.env.VITE_API_BASE || ''
  return normalize(fromEnv)
}

/** 设置 API Base URL 并持久化；传 '' 表示回到同源模式 */
export function setApiBase(base) {
  const normalized = normalize(base)
  if (normalized) localStorage.setItem(STORAGE_KEY, normalized)
  else localStorage.removeItem(STORAGE_KEY)
  return normalized
}

/** 用户输入的容错：根相对路径原样保留，其余补协议，统一去尾部斜杠，空白视为同源 */
function normalize(base) {
  const raw = (base || '').trim()
  if (!raw) return ''
  // 根相对路径（如同源反代前缀 /api）不能补协议，原样规范化
  if (raw.startsWith('/')) return raw.replace(/\/+$/, '')
  let url = raw
  if (!/^https?:\/\//i.test(url)) url = 'http://' + url
  return url.replace(/\/+$/, '')
}

/** 是否为同源模式（空或根相对路径都表示请求发到当前域名） */
export function isSameOriginMode() {
  const base = getApiBase()
  return base === '' || base.startsWith('/')
}
