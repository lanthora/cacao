<template>
  <div class="container">
    <LoadingView v-if="showLoading()"></LoadingView>
    <UserView v-if="showUser()"></UserView>
    <AdminView v-if="showAdmin()"></AdminView>
  </div>
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
  return !showUser() && !showAdmin()
}

const showUser = () => {
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
