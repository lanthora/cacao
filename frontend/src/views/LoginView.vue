<template>
  <auth-layout :title="$t('login.login')" :error="error">
    <a-form
      :model="formState"
      :hide-required-mark="true"
      name="login"
      layout="vertical"
      size="large"
      :aria-busy="submitting"
      @finish="onFinish"
    >
      <a-form-item
        name="username"
        :rules="[{ required: true, message: $t('login.inputUsername') }]"
      >
        <a-input
          v-model:value="formState.username"
          :placeholder="$t('login.username')"
          :aria-label="$t('login.username')"
          autocomplete="username"
          :disabled="submitting"
          :autocapitalize="'none'"
          :spellcheck="false"
        />
      </a-form-item>
      <a-form-item
        name="password"
        :rules="[{ required: true, message: $t('login.inputPassword') }]"
      >
        <a-input-password
          v-model:value="formState.password"
          :placeholder="$t('login.password')"
          :aria-label="$t('login.password')"
          autocomplete="current-password"
          :disabled="submitting"
        />
      </a-form-item>
      <a-form-item>
        <a-button type="primary" html-type="submit" class="auth-submit" :loading="submitting">{{
          $t('login.login')
        }}</a-button>
      </a-form-item>
    </a-form>
    <template #footer>
      {{ $t('login.noAccount') }}
      <router-link to="/register">{{ $t('register.register') }}</router-link>
    </template>
  </auth-layout>
</template>

<script setup>
import { reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import axios from 'axios'
import AuthLayout from '../components/AuthLayout.vue'

const router = useRouter()
const { t } = useI18n()
const formState = reactive({ username: '', password: '' })
const submitting = ref(false)
const error = ref('')

const onFinish = async (values) => {
  if (submitting.value) return
  submitting.value = true
  error.value = ''
  try {
    const { data } = await axios.post('/api/user/login', {
      username: values.username,
      password: values.password
    })
    if (data.status === 0) {
      await router.push('/')
    } else {
      error.value = data.msg || t('auth.requestFailed')
    }
  } catch {
    error.value = t('auth.requestFailed')
  } finally {
    submitting.value = false
  }
}
</script>
