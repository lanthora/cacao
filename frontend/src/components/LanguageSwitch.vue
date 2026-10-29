<template>
  <a-select
    v-model:value="locale"
    class="language-switch"
    aria-label="Language / 语言"
    @change="handleLangChange"
  >
    <a-select-option value="en-US">English</a-select-option>
    <a-select-option value="zh-CN">中文</a-select-option>
  </a-select>
</template>

<script setup>
import { useI18n } from 'vue-i18n'
import { LOCAL_LANGUAGE_KEY } from '../i18n'

const { locale } = useI18n({ useScope: 'global' })

const handleLangChange = (value) => {
  localStorage.setItem(LOCAL_LANGUAGE_KEY, value)
}
</script>

<style scoped>
.language-switch {
  width: 120px;
  flex-shrink: 0;
}
</style>
