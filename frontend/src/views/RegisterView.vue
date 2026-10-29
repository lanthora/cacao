<template>
  <auth-layout :title="$t('register.register')" :error="error">
    <a-form
      ref="formRef"
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
      <a-form-item
        name="confirmPassword"
        :rules="[
          { required: true, message: $t('register.inputConfirmPassword') },
          { validator: validateConfirmPassword }
        ]"
      >
        <a-input-password
          v-model:value="formState.confirmPassword"
          :placeholder="$t('register.confirmPassword')"
          :aria-label="$t('register.confirmPassword')"
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
import { reactive, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import axios from 'axios'
import AuthLayout from '../components/AuthLayout.vue'

const router = useRouter()
const { t } = useI18n()
const formRef = ref()
const formState = reactive({ username: '', password: '', confirmPassword: '' })
const submitting = ref(false)
const error = ref('')

const validateConfirmPassword = async (_rule, value) => {
  if (value && value !== formState.password) {
    throw new Error(t('register.passwordMismatch'))
  }
}

watch(
  () => formState.password,
  () => {
    if (formState.confirmPassword) {
      formRef.value?.validateFields(['confirmPassword']).catch(() => {})
    }
  }
)

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
