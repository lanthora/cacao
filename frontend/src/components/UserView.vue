<template>
  <a-layout>
    <a-layout-sider
      breakpoint="lg"
      collapsed-width="0"
      @collapse="onCollapse"
      @breakpoint="onBreakpoint"
      :style="{ height: '98vh' }"
    >
      <div class="logo">
        <a-avatar src="/favicon.ico" />
      </div>
      <a-menu v-model:selectedKeys="selectedKeys" theme="dark" mode="inline">
        <a-menu-item key="user">
          <user-outlined />
          <span class="nav-text">User</span>
        </a-menu-item>
        <a-menu-item key="network">
          <apartment-outlined />
          <span class="nav-text">Network</span>
        </a-menu-item>
        <a-menu-item key="route">
          <radar-chart-outlined />
          <span class="nav-text">Route</span>
        </a-menu-item>
        <a-menu-item key="device">
          <desktop-outlined />
          <span class="nav-text">Device</span>
        </a-menu-item>
      </a-menu>
    </a-layout-sider>
    <a-layout>
      <a-layout-header :style="{ background: '#fff', padding: 0 }">
        <a-page-header
          style="border: 1px solid rgb(235, 237, 240)"
          title="Title"
          sub-title="Sub Title"
        />
      </a-layout-header>
      <a-layout-content :style="{ margin: '24px 16px 0' }">
        <div :style="{ padding: '24px', background: '#fff', minHeight: '360px' }">content</div>
      </a-layout-content>
      <a-layout-footer style="text-align: center"> Cacao ©2024 Created by Canets </a-layout-footer>
    </a-layout>
  </a-layout>
</template>

<script setup>
import { ref } from 'vue'
const onCollapse = (collapsed, type) => {
  console.log(collapsed, type)
}
const onBreakpoint = (broken) => {
  console.log(broken)
}
const selectedKeys = ref(['user'])
</script>

<style scoped>
.logo {
  height: 32px;
  margin: 16px;
  text-align: center;
}

.site-layout-sub-header-background {
  background: #fff;
}

.site-layout-background {
  background: #fff;
}

[data-theme='dark'] .site-layout-sub-header-background {
  background: #141414;
}
</style>
