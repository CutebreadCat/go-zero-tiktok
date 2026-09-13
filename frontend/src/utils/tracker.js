/**
 * 行为埋点批量上报器 —— 对应后端 tracking-events 链路。
 *
 * 设计：组件只调用 track(event)，事件先进队列；每 5 秒或攒满 10 条
 * 批量 POST 一次（对齐 recommend-phase2 文档的批量上报建议）。
 * 失败静默丢弃 + 计数（埋点永远不能影响主流程）。
 */
import { sendTrackingEvents } from '../api/tracking'

let queue = []
let timer = null
let dropped = 0

function newRequestId() {
  return 'r-' + Date.now().toString(36) + '-' + Math.random().toString(36).slice(2, 8)
}

/** 一次 Feed 加载会话，用于串联该批内容的曝光/播放 */
export function beginFeedSession(scene) {
  return { requestId: newRequestId(), scene, seq: 0 }
}

function flush() {
  if (!queue.length) return
  const batch = queue.splice(0, queue.length)
  sendTrackingEvents(batch).catch(() => { dropped += batch.length })
}

export function track(event, session = null) {
  const e = { timestamp: Date.now(), ...event }
  if (session) {
    e.request_id = session.requestId
    e.scene = session.scene
    if (event.event_type === 'impression') e.position = ++session.seq
  }
  queue.push(e)
  if (queue.length >= 10) flush()
  else {
    clearTimeout(timer)
    timer = setTimeout(flush, 5000)
  }
}

export function trackImpression(videoId, session) {
  track({ event_type: 'impression', video_id: String(videoId) }, session)
}
export function trackPlay(videoId, session, durationMs = 0) {
  track({ event_type: 'play', video_id: String(videoId), duration_ms: durationMs }, session)
}
export function trackComplete(videoId, session, watchMs = 0, durationMs = 0) {
  track({ event_type: 'complete', video_id: String(videoId), watch_ms: watchMs, duration_ms: durationMs }, session)
}
export function trackAction(eventType, videoId, extra = {}) {
  track({ event_type: eventType, video_id: String(videoId), ...extra })
}

// 页面关闭前尽力发送
if (typeof window !== 'undefined') {
  window.addEventListener('pagehide', flush)
}
