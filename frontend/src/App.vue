<template>
  <main>
    <RouterView />
  </main>
</template>

<script setup>
import axios from 'axios'
import { message } from 'ant-design-vue'
import { useRouter } from 'vue-router'

const router = useRouter()

axios.interceptors.response.use(
  (response) => {
    if (response.data.status != 0) {
      message.warning(response.data.msg)
    }
    if (response.data.status == 2) {
      router.push('/login')
    } else if (response.data.status == 11) {
      router.push('/')
    }
    return response
  },
  (error) => {
    return Promise.reject(error)
  }
)
</script>
