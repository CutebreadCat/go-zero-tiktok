<script setup>
import { ref, provide } from 'vue'
import SideNav from './components/SideNav.vue'
import TopBar from './components/TopBar.vue'
import ApiSettingsModal from './components/ApiSettingsModal.vue'

// 全局唯一的设置弹窗入口：任何组件 inject 后都能打开
const showSettings = ref(false)
provide('openApiSettings', () => { showSettings.value = true })
</script>

<template>
  <div class="app-shell">
    <SideNav />
    <div class="app-main">
      <TopBar @open-settings="showSettings = true" />
      <div class="app-content">
        <router-view />
      </div>
    </div>
    <ApiSettingsModal v-if="showSettings" @close="showSettings = false" />
  </div>
</template>
