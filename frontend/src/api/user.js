import http from './http'

// 账户域：注册/登录/刷新（表单编码）、当前用户、头像、MFA
export const register = (username, password) =>
  http.post('/users', null, { params: { username, password } })

export const login = (username, password, mfaCode = '') =>
  http.post('/sessions', null, { params: { username, password, mfa_code: mfaCode } })

export const fetchMe = () => http.get('/users/me')

export const updatePhoto = (file) => {
  const form = new FormData()
  form.append('file', file)
  return http.put('/users/me/photo', form)
}

export const getMfaQR = () => http.get('/users/me/mfa/qr')

export const bindMfa = (mfaSecret, mfaCode) =>
  http.post('/users/me/mfa/bind', null, { params: { mfa_secret: mfaSecret, mfa_code: mfaCode } })
