import http from './http'

// 视频域：Feed / 热门 / 搜索 / 作者作品 / 发布

/** Feed 流（新游标协议） */
export const getFeed = ({ scene = 'recommend', cursor = '', limit = 5 } = {}) =>
  http.get('/feed-items', { params: { scene, cursor, limit } })

/** 热门视频（页码分页） */
export const getPopular = ({ pageNum = 1, pageSize = 10 } = {}) =>
  http.get('/videos/popular', { params: { page_num: pageNum, page_size: pageSize } })

/** 关键词搜索 */
export const searchVideos = (keyword, { pageNum = 1, pageSize = 20 } = {}) =>
  http.get('/videos/search', { params: { keyword, page_num: pageNum, page_size: pageSize } })

/** 作者视频列表 */
export const getUserVideos = (userId, { pageNum = 1, pageSize = 20 } = {}) =>
  http.get(`/users/${userId}/videos`, { params: { page_num: pageNum, page_size: pageSize } })

/** 发布视频（multipart 真文件上传） */
export const publishVideo = ({ file, title, description }) => {
  const form = new FormData()
  form.append('file', file)
  form.append('title', title)
  form.append('description', description || '')
  return http.post('/videos', form)
}
