/** 时间与数字的展示格式化 */
export function timeAgo(iso) {
  if (!iso) return ''
  const t = new Date(iso.replace(' ', 'T')).getTime()
  if (Number.isNaN(t)) return iso
  const diff = Date.now() - t
  const m = Math.floor(diff / 60000)
  if (m < 1) return '刚刚'
  if (m < 60) return `${m} 分钟前`
  const h = Math.floor(m / 60)
  if (h < 24) return `${h} 小时前`
  const d = Math.floor(h / 24)
  if (d < 30) return `${d} 天前`
  return iso.slice(0, 10)
}

export function compact(n) {
  if (n == null) return '0'
  if (n >= 100000000) return (n / 100000000).toFixed(1) + '亿'
  if (n >= 10000) return (n / 10000).toFixed(1) + 'w'
  return String(n)
}

/** 用户名首字头像色 */
export function avatarColor(name = '') {
  const colors = ['#ff4d6d', '#3a86ff', '#8338ec', '#fb8500', '#2a9d8f', '#e63946']
  let h = 0
  for (const c of name) h = (h * 31 + c.charCodeAt(0)) % 997
  return colors[h % colors.length]
}
