<template>
  <a-layout-sider breakpoint="lg" collapsed-width="0">
    <logo-view />
    <a-menu v-model:selectedKeys="selectedKeys" theme="dark" mode="inline" @click="handleMenuClick">
      <a-menu-item key="setting">
        <setting-outlined />
        <span class="nav-text">Setting</span>
      </a-menu-item>
      <a-menu-item key="user">
        <team-outlined />
        <span class="nav-text">User</span>
      </a-menu-item>
      <a-menu-item key="license">
        <copyright-outlined />
        <span class="nav-text">License</span>
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
    router.push('/admin/' + item.key)
  }
}
</script>
