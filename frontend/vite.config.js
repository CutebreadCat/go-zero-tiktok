import { defineConfig, loadEnv } from 'vite'
import vue from '@vitejs/plugin-vue'

// 代理目标同样配置化：.env 里改 VITE_PROXY_TARGET 即可指向任意后端 IP
export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd(), '')
  const proxyTarget = env.VITE_PROXY_TARGET || 'http://127.0.0.1:8888'
  return {
    plugins: [vue()],
    server: {
      port: 5173,
      proxy: {
        // 仅当 API Base 留空（同源模式）时走代理；显式填写完整地址则直连
        '/api': {
          target: proxyTarget,
          changeOrigin: true,
          rewrite: (path) => path.replace(/^\/api/, ''),
        },
      },
    },
  }
})
