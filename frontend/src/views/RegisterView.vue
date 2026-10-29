<template>
  <div class="container">
    <div class="register-box">
      <div class="language-toolbar">
        <language-switch />
      </div>
      <a-form
        :model="registerState"
        :hideRequiredMark="true"
        name="register"
        class="register-form"
        @finish="onFinish"
      >
        <a-form-item
          name="username"
          :rules="[{ required: true, message: $t('register.inputUsername') }]"
        >
          <a-input v-model:value="registerState.username" :placeholder="$t('register.username')">
            <template #prefix>
              <UserOutlined class="site-form-item-icon" />
            </template>
          </a-input>
        </a-form-item>

        <a-form-item
          name="password"
          :rules="[{ required: true, message: $t('register.inputPassword') }]"
        >
          <a-input
            type="password"
            autocomplete="new-password"
            v-model:value="registerState.password"
            :placeholder="$t('register.password')"
          >
            <template #prefix>
              <LockOutlined class="site-form-item-icon" />
            </template>
          </a-input>
        </a-form-item>

        <a-form-item>
          <a-button type="primary" html-type="submit" class="register-form-button">
            {{ $t('register.register') }}
          </a-button>
        </a-form-item>
      </a-form>
    </div>
  </div>
</template>

<script setup>
import { reactive } from 'vue'
import axios from 'axios'
import { useRouter } from 'vue-router'

const registerState = reactive({
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
</script>

<style scoped>
.container {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 100vh;
  box-sizing: border-box;
  padding: 24px 16px;
}

.register-box {
  width: 300px;
  max-width: 100%;
}

.language-toolbar {
  display: flex;
  justify-content: flex-end;
  margin-bottom: 24px;
}
.register-form-button {
  width: 100%;
}
</style>
