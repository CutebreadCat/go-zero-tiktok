import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '../stores/auth'

const routes = [
  { path: '/', name: 'feed', component: () => import('../views/FeedView.vue') },
  { path: '/search', name: 'search', component: () => import('../views/SearchView.vue') },
  { path: '/publish', name: 'publish', component: () => import('../views/PublishView.vue'), meta: { auth: true } },
  { path: '/messages', name: 'messages', component: () => import('../views/MessagesView.vue'), meta: { auth: true } },
  { path: '/profile', name: 'profile', component: () => import('../views/ProfileView.vue'), meta: { auth: true } },
  { path: '/users/:id/videos', name: 'user-videos', component: () => import('../views/ProfileView.vue') },
  { path: '/login', name: 'login', component: () => import('../views/LoginView.vue') },
  { path: '/:pathMatch(.*)*', redirect: '/' },
]

export const router = createRouter({
  history: createWebHistory(),
  routes,
})

// 路由守卫：meta.auth 的页面必须已登录
router.beforeEach(async (to) => {
  const auth = useAuthStore()
  await auth.bootstrap()
  if (to.meta.auth && !auth.isLoggedIn) {
    return { name: 'login', query: { redirect: to.fullPath } }
  }
  if (to.name === 'login' && auth.isLoggedIn) {
    return { name: 'feed' }
  }
})
