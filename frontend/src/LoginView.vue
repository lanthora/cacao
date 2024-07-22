<template>
  <a-form :model="formState" name="login" autocomplete="off" @finish="onFinish">
    <a-form-item
      label="Username"
      name="username"
      :rules="[{ required: true, message: 'Please input your username!' }]"
    >
      <a-input v-model:value="formState.username" />
    </a-form-item>

    <a-form-item
      label="Password"
      name="password"
      :rules="[{ required: true, message: 'Please input your password!' }]"
    >
      <a-input-password v-model:value="formState.password" />
    </a-form-item>

    <a-form-item>
      <a-button type="primary" html-type="submit">Submit</a-button>
    </a-form-item>
  </a-form>
</template>

<script setup>
import { reactive } from 'vue'
import axios from 'axios'
import { useRouter } from 'vue-router'

const formState = reactive({
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
    router.back()
  }
}

const onFinish = (values) => {
  userLogin(values.username, values.password)
}
</script>
