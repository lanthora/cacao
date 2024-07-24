<template>
  <LoadingView v-if="showLoading()" class="container"></LoadingView>
  <NormalView v-if="showNormal()"></NormalView>
  <AdminView v-if="showAdmin()"></AdminView>
</template>

<script setup>
import axios from 'axios'
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'

const router = useRouter()
const component = ref('loading')

onMounted(() => {
  userInfo()
})

const showLoading = () => {
  return !showNormal() && !showAdmin()
}

const showNormal = () => {
  return component.value == 'normal'
}

const showAdmin = () => {
  return component.value == 'admin'
}

const userInfo = async () => {
  const response = await axios.post('/api/user/info')
  const status = response.data.status
  if (status == 0) {
    component.value = response.data.data.role
  } else if (status == 2) {
    router.push('/login')
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
