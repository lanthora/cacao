<template>
  <div class="container">
    <a-form
      :model="loginState"
      :hideRequiredMark="true"
      name="login"
      class="login-form"
      @finish="onFinish"
    >
      <a-form-item name="username" :rules="[{ required: true, message: 'Input your username' }]">
        <a-input v-model:value="loginState.username">
          <template #prefix>
            <user-outlined class="site-form-item-icon" />
          </template>
        </a-input>
      </a-form-item>

      <a-form-item name="password" :rules="[{ required: true, message: 'Input your password' }]">
        <a-input-password v-model:value="loginState.password">
          <template #prefix>
            <lock-outlined class="site-form-item-icon" />
          </template>
        </a-input-password>
      </a-form-item>

      <a-form-item>
        <a-button type="primary" html-type="submit" class="login-form-button"> Log in </a-button>
        Or
        <a href="/register">register now</a>
      </a-form-item>
    </a-form>
  </div>
</template>

<script setup>
import { reactive } from 'vue'
import axios from 'axios'
import { useRouter } from 'vue-router'

const loginState = reactive({
  username: '',
  password: ''
})

const router = useRouter()

const userLogin = async (username, password) => {
  const response = await axios.post('/api/user/login', {
    username: username,
    password: password
  })

  const status = response.data.status
  if (status == 0) {
    router.push('/')
  }
}

const onFinish = (values) => {
  userLogin(values.username, values.password)
}
</script>

<style scoped>
.container {
  display: flex;
  justify-content: center;
  align-items: center;
  height: 100vh;
}

.login-form {
  max-width: 300px;
}
.login-form-button {
  width: 100%;
}
</style>
