import http from './http'

// 互动域：点赞 / 收藏 / 评论 / 回复 / 评论点赞

export const likeVideo = (videoId) => http.put(`/videos/${videoId}/like`)
export const unlikeVideo = (videoId) => http.delete(`/videos/${videoId}/like`)
export const favoriteVideo = (videoId) => http.put(`/videos/${videoId}/favorite`)
export const unfavoriteVideo = (videoId) => http.delete(`/videos/${videoId}/favorite`)

export const getMyLikes = ({ pageNum = 1, pageSize = 20 } = {}) =>
  http.get('/users/me/likes', { params: { page_num: pageNum, page_size: pageSize } })

export const getMyFavorites = ({ pageNum = 1, pageSize = 20 } = {}) =>
  http.get('/users/me/favorites', { params: { page_num: pageNum, page_size: pageSize } })

export const getComments = (videoId, { pageNum = 1, pageSize = 20 } = {}) =>
  http.get(`/videos/${videoId}/comments`, { params: { page_num: pageNum, page_size: pageSize } })

export const postComment = (videoId, commentText) =>
  http.post(`/videos/${videoId}/comments`, null, { params: { comment_text: commentText } })

export const postReply = (commentId, commentText) =>
  http.post(`/comments/${commentId}/replies`, null, { params: { comment_text: commentText } })

export const deleteComment = (commentId) => http.delete(`/comments/${commentId}`)

export const likeComment = (commentId) => http.put(`/comments/${commentId}/like`)
export const unlikeComment = (commentId) => http.delete(`/comments/${commentId}/like`)
