<template>
  <a-layout style="min-height: 98vh">
    <a-layout-sider breakpoint="lg">
      <div class="logo">
        <a-avatar src="/favicon.ico" />
      </div>
      <a-menu v-model:selectedKeys="selectedKeys" theme="dark" mode="inline">
        <a-menu-item key="user">
          <team-outlined />
          <span class="nav-text">User</span>
        </a-menu-item>
        <a-menu-item key="setting">
          <setting-outlined />
          <span class="nav-text">Setting</span>
        </a-menu-item>
      </a-menu>
    </a-layout-sider>
    <a-layout>
      <AdminUserView v-if="showUser()"></AdminUserView>
      <AdminSettingView v-if="showSetting()"></AdminSettingView>
      <a-layout-footer style="text-align: center">Cacao © 2024</a-layout-footer>
    </a-layout>
  </a-layout>
</template>

<script setup>
import { ref } from 'vue'

const selectedKeys = ref(['user'])

const showUser = () => {
  return selectedKeys.value.includes('user')
}

const showSetting = () => {
  return selectedKeys.value.includes('setting')
}
</script>

<style scoped>
.logo {
  height: 32px;
  margin: 16px;
  text-align: center;
}
</style>
