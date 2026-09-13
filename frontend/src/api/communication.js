import http from './http'

// 关系域 + 消息中心（communication）

export const follow = (userId) => http.put(`/users/me/following/${userId}`)
export const unfollow = (userId) => http.delete(`/users/me/following/${userId}`)

export const getFollowing = ({ pageNum = 1, pageSize = 50 } = {}) =>
  http.get('/users/me/following', { params: { page_num: pageNum, page_size: pageSize } })

export const getFollowers = ({ pageNum = 1, pageSize = 50 } = {}) =>
  http.get('/users/me/followers', { params: { page_num: pageNum, page_size: pageSize } })

export const getFriends = ({ pageNum = 1, pageSize = 50 } = {}) =>
  http.get('/users/me/friends', { params: { page_num: pageNum, page_size: pageSize } })

/** 消息列表（游标分页 + 类型筛选） */
export const getMessages = ({ type = '', cursor = '', limit = 20 } = {}) =>
  http.get('/messages', { params: { type: type || undefined, cursor, limit } })

export const getUnreadCount = () => http.get('/message-stats/unread')

/** 批量已读；messageIds 为空数组表示全部已读 */
export const markRead = (messageIds = []) => http.patch('/messages', { message_ids: messageIds })
