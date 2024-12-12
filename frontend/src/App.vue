<template>
  <main>
    <div class="lang-switch">
      <a-select
        v-model:value="currentLang"
        style="width: 120px"
        @change="handleLangChange"
      >
        <a-select-option value="en-US">English</a-select-option>
        <a-select-option value="zh-CN">中文</a-select-option>
      </a-select>
    </div>
    <RouterView />
  </main>
</template>

<script setup>
import axios from 'axios'
import { message } from 'ant-design-vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { ref } from 'vue'
import { LOCAL_LANGUAGE_KEY } from './i18n'

const router = useRouter()
const { locale } = useI18n()
const currentLang = ref(locale.value)

const handleLangChange = (value) => {
  locale.value = value
  currentLang.value = value
  localStorage.setItem(LOCAL_LANGUAGE_KEY, value)
}

axios.interceptors.response.use(
  (response) => {
    if (response.data.status != 0) {
      message.warning(response.data.msg)
    }
    if (response.data.status == 2) {
      router.push('/login')
    } else if (response.data.status == 11) {
      router.push('/')
    }
    return response
  },
  (error) => {
    return Promise.reject(error)
  }
)
</script>

<style scoped>
.lang-switch {
  position: fixed;
  top: 20px;
  right: 20px;
  z-index: 1000;
}
</style>
