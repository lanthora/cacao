<template>
  <div class="login-container">
    <div class="login-box">
      <div class="login-content">
        <div class="logo">
          <img src="/favicon.ico" alt="Logo" />
          <h2 class="title">{{ $t('login.login') }}</h2>
        </div>
        <a-form
          :model="loginState"
          :hideRequiredMark="true"
          class="login-form"
          @finish="onFinish"
        >
          <a-form-item 
            name="username" 
            :rules="[{ required: true, message: $t('login.inputUsername') }]"
          >
            <a-input
              v-model:value="loginState.username"
              :placeholder="$t('login.username')"
              size="large"
            >
              <template #prefix>
                <user-outlined class="site-form-item-icon" />
              </template>
            </a-input>
          </a-form-item>

          <a-form-item
            name="password"
            :rules="[{ required: true, message: $t('login.inputPassword') }]"
          >
            <a-input-password
              v-model:value="loginState.password"
              :placeholder="$t('login.password')"
              size="large"
            >
              <template #prefix>
                <lock-outlined class="site-form-item-icon" />
              </template>
            </a-input-password>
          </a-form-item>

          <a-form-item>
            <a-button
              type="primary"
              html-type="submit"
              size="large"
              class="login-button"
            >
              {{ $t('login.login') }}
            </a-button>
          </a-form-item>

          <div class="register-link">
            {{ $t('login.register') }}
            <router-link to="/register" class="link">
              {{ $t('login.registerNow') }}
            </router-link>
          </div>
        </a-form>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import axios from 'axios'
import { UserOutlined, LockOutlined } from '@ant-design/icons-vue'

const router = useRouter()

const loginState = ref({
  username: '',
  password: ''
})

const onFinish = async (values) => {
  const response = await axios.post('/api/user/login', {
    username: values.username,
    password: values.password
  })

  const status = response.data.status
  if (status == 0) {
    router.push('/')
  }
}
</script>

<style scoped>
.login-container {
  width: 100vw;
  height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background-color: #f5f5f5;
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial,
    'Noto Sans', sans-serif, 'Apple Color Emoji', 'Segoe UI Emoji', 'Segoe UI Symbol',
    'Noto Color Emoji';
}

.login-box {
  width: 360px;
  background: white;
  border-radius: 8px;
  padding: 40px;
  box-shadow: 0 2px 10px rgba(0, 0, 0, 0.08);
}

.logo {
  text-align: center;
  margin-bottom: 30px;
}

.logo img {
  width: 48px;
  height: 48px;
  margin-bottom: 16px;
}

.title {
  color: #1a1a1a;
  font-size: 24px;
  font-weight: 500;
  margin: 0;
  letter-spacing: -0.5px;
}

.site-form-item-icon {
  color: rgba(0, 0, 0, 0.25);
}

.login-button {
  width: 100%;
}

.register-link {
  margin-top: 16px;
  text-align: center;
  font-size: 14px;
  color: rgba(0, 0, 0, 0.45);
}

.link {
  color: #1890ff;
  text-decoration: none;
  margin-left: 4px;
  transition: color 0.3s;
}

.link:hover {
  color: #40a9ff;
}

:deep(.ant-input-affix-wrapper) {
  border-radius: 6px;
}

:deep(.ant-form-item) {
  margin-bottom: 24px;
}

:deep(.ant-form-item-explain-error) {
  font-size: 13px;
}

@media (max-width: 480px) {
  .login-box {
    width: 90%;
    padding: 30px 20px;
  }
}
</style>
