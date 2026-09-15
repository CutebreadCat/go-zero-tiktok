import http from './http'

// 埋点与 QoS：对应后端"行为埋点 + 播放质量"两条异步链路。
// 调用方是 utils/tracker.js（批量打包），不直接从组件调用。

export const sendTrackingEvents = (events) => http.post('/tracking-events', { events })

export const reportQoS = (payload) => http.post('/playback-qos-reports', payload)
