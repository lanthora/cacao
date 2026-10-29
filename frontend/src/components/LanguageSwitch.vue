<template>
  <a-dropdown v-model:open="open" :trigger="['click']" placement="bottomRight">
    <a-button
      type="text"
      shape="circle"
      class="language-switch"
      aria-label="Language / 语言"
      title="Language / 语言"
      aria-haspopup="menu"
      :aria-expanded="open"
    >
      <template #icon><global-outlined /></template>
    </a-button>
    <template #overlay>
      <a-menu :selected-keys="[locale]" aria-label="Language / 语言" @click="handleLangChange">
        <a-menu-item key="zh-CN">中文</a-menu-item>
        <a-menu-item key="en-US">English</a-menu-item>
      </a-menu>
    </template>
  </a-dropdown>
</template>

<script setup>
import { ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { LOCAL_LANGUAGE_KEY } from '../i18n'

const { locale } = useI18n({ useScope: 'global' })
const open = ref(false)

const handleLangChange = ({ key }) => {
  locale.value = key
  open.value = false
  localStorage.setItem(LOCAL_LANGUAGE_KEY, key)
}
</script>

<style scoped>
.language-switch {
  width: 36px;
  height: 36px;
  flex-shrink: 0;
  color: #64766c;
  font-size: 19px;
}
.language-switch:hover,
.language-switch:focus-visible {
  color: #3b806b;
  background: #edf4f0;
}
.language-switch:focus-visible {
  outline: 2px solid #3b806b;
  outline-offset: 2px;
}
</style>
