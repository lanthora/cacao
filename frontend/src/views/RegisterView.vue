<template>
  <auth-layout :title="$t('register.register')" :error="error">
    <a-form
      :model="formState"
      :hide-required-mark="true"
      name="register"
      layout="vertical"
      size="large"
      :aria-busy="submitting"
      @finish="onFinish"
    >
      <a-form-item
        name="username"
        :rules="[{ required: true, message: $t('register.inputUsername') }]"
      >
        <a-input
          v-model:value="formState.username"
          :placeholder="$t('register.username')"
          :aria-label="$t('register.username')"
          autocomplete="username"
          :disabled="submitting"
          :autocapitalize="'none'"
          :spellcheck="false"
        />
      </a-form-item>
      <a-form-item
        name="password"
        :rules="[{ required: true, message: $t('register.inputPassword') }]"
      >
        <a-input-password
          v-model:value="formState.password"
          :placeholder="$t('register.password')"
          :aria-label="$t('register.password')"
          autocomplete="new-password"
          :disabled="submitting"
        />
      </a-form-item>
      <a-form-item>
        <a-button type="primary" html-type="submit" class="auth-submit" :loading="submitting">{{
          $t('register.register')
        }}</a-button>
      </a-form-item>
    </a-form>
    <template #footer>
      {{ $t('register.haveAccount') }}
      <router-link to="/login">{{ $t('login.login') }}</router-link>
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
    const { data } = await axios.post('/api/user/register', {
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
