import { defineStore } from 'pinia'
import { getUnreadCount, markRead } from '../api/communication'

/** 消息中心状态：未读数轮询 + 已读操作后的即时刷新 */
export const useMessageStore = defineStore('messages', {
  state: () => ({
    unreadCount: 0,
    timer: null,
  }),
  actions: {
    async refreshUnread() {
      try {
        const resp = await getUnreadCount()
        this.unreadCount = resp.unread_count ?? 0
      } catch { /* 未登录或网络异常时静默 */ }
    },
    startPolling() {
      this.stopPolling()
      this.refreshUnread()
      this.timer = setInterval(() => this.refreshUnread(), 30000)
    },
    stopPolling() {
      if (this.timer) clearInterval(this.timer)
      this.timer = null
    },
    async markRead(messageIds) {
      const resp = await markRead(messageIds)
      await this.refreshUnread()
      return resp.updated_count
    },
  },
})
