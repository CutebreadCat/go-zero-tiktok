import { createApp } from 'vue'
import { createPinia } from 'pinia'
import App from './App.vue'
import { router } from './router'
import { setupAuthExpiry } from './stores/auth'
import './styles/main.css'

const app = createApp(App)
app.use(createPinia())
app.use(router)
setupAuthExpiry()
app.mount('#app')
