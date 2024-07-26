<template>
  <a-layout-sider breakpoint="lg" collapsed-width="0">
    <logo-view />
    <a-menu v-model:selectedKeys="selectedKeys" theme="dark" mode="inline" @click="handleMenuClick">
      <a-menu-item key="statistics">
        <bar-chart-outlined />
        <span class="nav-text">Statistics</span>
      </a-menu-item>
      <a-menu-item key="network">
        <apartment-outlined />
        <span class="nav-text">Network</span>
      </a-menu-item>
      <a-menu-item key="device">
        <desktop-outlined />
        <span class="nav-text">Device</span>
      </a-menu-item>
      <a-menu-item key="route">
        <thunderbolt-outlined />
        <span class="nav-text">Route</span>
      </a-menu-item>
      <a-menu-item key="user">
        <user-outlined />
        <span class="nav-text">User</span>
      </a-menu-item>
      <a-menu-item key="logout">
        <logout-outlined />
        <span class="nav-text">Logout</span>
      </a-menu-item>
    </a-menu>
  </a-layout-sider>
</template>

<script setup>
import { useRouter } from 'vue-router'
import { ref } from 'vue'
import axios from 'axios'

const router = useRouter()

const props = defineProps({
  value: String
})

const selectedKeys = ref([props.value])

const handleMenuClick = async (item) => {
  if (item.key === 'logout') {
    const response = await axios.post('/api/user/logout')
    const status = response.data.status
    if (status == 0) {
      router.push('/login')
    }
  } else {
    router.push('/' + item.key)
  }
}
</script>
