<template>
  <main>
    <RouterView />
  </main>
</template>

<script setup>
import axios from 'axios'
import { message } from 'ant-design-vue'

axios.interceptors.response.use(
  (response) => {
    if (response.data.status != 0) {
      message.warning(response.data.msg)
    }
    return response
  },
  (error) => {
    return Promise.reject(error)
  }
)
</script>
