<template>
  <div class="container">
    <a-spin size="large" />
  </div>
</template>

<script setup>
import axios from 'axios'
import { onMounted } from 'vue'
import { useRouter } from 'vue-router'

const router = useRouter()

onMounted(() => {
  userInfo()
})

const userInfo = async () => {
  const response = await axios.post('/api/user/info')
  const status = response.data.status
  if (status == 0) {
    const role = response.data.data.role
    if (role === 'admin') {
      router.push('/admin/user')
    } else {
      router.push('/overview')
    }
  }
}
</script>

<style scoped>
.container {
  display: flex;
  justify-content: center;
  align-items: center;
  height: 100vh;
}
</style>
