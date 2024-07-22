<template>
  <div class="container">
    <a-form
      :model="formState"
      :hideRequiredMark="true"
      name="register"
      class="register-form"
      @finish="onFinish"
    >
      <a-form-item name="username" :rules="[{ required: true, message: 'Input your username' }]">
        <a-input v-model:value="formState.username">
          <template #prefix>
            <UserOutlined class="site-form-item-icon" />
          </template>
        </a-input>
      </a-form-item>

      <a-form-item name="password" :rules="[{ required: true, message: 'Input your password' }]">
        <a-input-password v-model:value="formState.password">
          <template #prefix>
            <LockOutlined class="site-form-item-icon" />
          </template>
        </a-input-password>
      </a-form-item>

      <a-form-item>
        <a-button
          :disabled="disabled"
          type="primary"
          html-type="submit"
          class="register-form-button"
        >
          Register
        </a-button>
      </a-form-item>
    </a-form>
  </div>
</template>

<script setup>
import { reactive, computed } from 'vue'
import axios from 'axios'
import { useRouter } from 'vue-router'

const formState = reactive({
  username: '',
  password: ''
})

const router = useRouter()

const userregister = async (username, password) => {
  const response = await axios.post('/api/user/register', {
    username: username,
    password: password
  })

  const status = response.data.status
  if (status == 0) {
    router.push('/')
  }
}

const onFinish = (values) => {
  userregister(values.username, values.password)
}

const disabled = computed(() => {
  return !(formState.username && formState.password)
})
</script>

<style scoped>
.container {
  display: flex;
  justify-content: center;
  align-items: center;
  height: 100vh;
}

.register-form {
  max-width: 300px;
}
.register-form-button {
  width: 100%;
}
</style>
