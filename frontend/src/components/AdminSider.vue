<template>
  <a-layout-sider breakpoint="lg" collapsed-width="0">
    <logo-view />
    <a-menu v-model:selectedKeys="selectedKeys" theme="dark" mode="inline" @click="changeRoute">
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
  router.push('/admin/' + item.key)
}
</script>
