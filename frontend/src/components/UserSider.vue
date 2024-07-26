<template>
  <a-layout-sider breakpoint="lg" collapsed-width="0">
    <logo-view />
    <a-menu v-model:selectedKeys="selectedKeys" theme="dark" mode="inline" @click="changeRoute">
      <a-menu-item key="overview">
        <bar-chart-outlined />
        <span class="nav-text">Overview</span>
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
    </a-menu>
  </a-layout-sider>
</template>

<script setup>
import { useRouter } from 'vue-router'
import { ref } from 'vue'

const props = defineProps({
  value: String
})

const selectedKeys = ref([props.value])

const router = useRouter()
const changeRoute = (item) => {
  router.push('/' + item.key)
}
</script>
