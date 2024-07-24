<template>
  <a-layout style="min-height: 98vh">
    <a-layout-sider breakpoint="lg" collapsed-width="0">
      <div class="logo">
        <a-avatar src="/favicon.ico" />
      </div>
      <a-menu v-model:selectedKeys="selectedKeys" theme="dark" mode="inline">
        <a-menu-item key="user">
          <user-outlined />
          <span class="nav-text">User</span>
        </a-menu-item>
        <a-menu-item key="device">
          <desktop-outlined />
          <span class="nav-text">Device</span>
        </a-menu-item>
        <a-menu-item key="network">
          <apartment-outlined />
          <span class="nav-text">Network</span>
        </a-menu-item>
        <a-menu-item key="sdwan">
          <thunderbolt-outlined />
          <span class="nav-text">SD-WAN</span>
        </a-menu-item>
      </a-menu>
    </a-layout-sider>
    <a-layout>
      <UserInfoView v-if="showUserInfo()"></UserInfoView>
      <DeviceView v-if="showDevice()"></DeviceView>
      <NetworkView v-if="showNetwork()"></NetworkView>
      <SdwanView v-if="showSdwan()"></SdwanView>
      <a-layout-footer style="text-align: center">Cacao © 2024</a-layout-footer>
    </a-layout>
  </a-layout>
</template>

<script setup>
import { ref } from 'vue'
const selectedKeys = ref(['user'])

const showUserInfo = () => {
  return selectedKeys.value.includes('user')
}
const showDevice = () => {
  return selectedKeys.value.includes('device')
}
const showNetwork = () => {
  return selectedKeys.value.includes('network')
}
const showSdwan = () => {
  return selectedKeys.value.includes('sdwan')
}
</script>

<style scoped>
.logo {
  height: 32px;
  margin: 16px;
  text-align: center;
}
</style>
