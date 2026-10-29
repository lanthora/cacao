<template>
  <a-config-provider
    :theme="{ token: { colorPrimary: '#3b806b', colorInfo: '#3b806b', borderRadius: 8 } }"
  >
    <div class="auth-layout">
      <page-header />
      <div class="auth-content">
        <div class="auth-panel">
          <section class="auth-card" aria-labelledby="auth-title">
            <h1 id="auth-title" class="sr-only">{{ title }}</h1>
            <a-alert
              v-if="error"
              class="auth-error"
              type="error"
              show-icon
              :message="error"
              role="alert"
            />
            <slot />
            <footer class="auth-footer"><slot name="footer" /></footer>
          </section>
          <div class="auth-utilities">
            <span>{{ $t('components.footer.copyright', { year: new Date().getFullYear() }) }}</span>
          </div>
        </div>
      </div>
    </div>
  </a-config-provider>
</template>

<script setup>
import PageHeader from './PageHeader.vue'

defineProps({
  title: { type: String, required: true },
  error: { type: String, default: '' }
})
</script>

<style scoped>
.auth-layout {
  display: flex;
  flex-direction: column;
  min-height: 100vh;
  min-height: 100dvh;
  box-sizing: border-box;
  background: #f1f6f3;
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial,
    'Noto Sans', sans-serif;
}
.auth-layout :deep(.page-header) {
  flex-shrink: 0;
  border-bottom: 1px solid #e3ebe6;
}
.auth-content {
  display: flex;
  flex: 1;
  align-items: center;
  justify-content: center;
  padding: 40px 20px;
}
.auth-panel {
  width: 400px;
  max-width: 100%;
  min-width: 0;
}
.auth-card {
  box-sizing: border-box;
  padding: 40px 32px 24px;
  background: #fff;
  border: 1px solid #dfe9e3;
  border-radius: 14px;
  box-shadow:
    0 12px 40px rgb(31 68 48 / 5%),
    0 2px 5px rgb(31 68 48 / 2%);
}
.sr-only {
  position: absolute;
  width: 1px;
  height: 1px;
  padding: 0;
  margin: -1px;
  overflow: hidden;
  clip-path: inset(50%);
  white-space: nowrap;
  border: 0;
}
.auth-error {
  margin-bottom: 24px;
  overflow-wrap: anywhere;
}
.auth-footer {
  margin-top: 20px;
  text-align: center;
  font-size: 12px;
  color: #8a9990;
  line-height: 1.8;
}
.auth-footer :deep(a) {
  margin-left: 5px;
  color: #3b806b;
  text-decoration: none;
  font-weight: 500;
}
.auth-footer :deep(a:hover) {
  color: #28634f;
}
.auth-footer :deep(a:focus-visible) {
  outline: 2px solid #3b806b;
  outline-offset: 3px;
  border-radius: 2px;
}
.auth-utilities {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 16px;
  margin-top: 18px;
  padding: 0 8px;
  font-size: 11px;
  color: #8f9e95;
}
:deep(.auth-submit) {
  width: 100%;
  height: 44px;
  font-size: 14px;
  font-weight: 500;
  box-shadow: none;
}
:deep(.ant-form-item) {
  margin-bottom: 20px;
}
:deep(.ant-form-item:last-child) {
  margin-bottom: 0;
  padding-top: 4px;
}
:deep(.ant-input-affix-wrapper),
:deep(.ant-input-lg) {
  min-height: 44px;
  border-color: #dce5df;
  border-radius: 8px;
  font-size: 14px;
}
:deep(.ant-input-affix-wrapper .ant-input) {
  min-height: 0;
  font-size: 14px;
}
:deep(.ant-input::placeholder) {
  color: #b0bcb4;
}
:deep(.ant-form-item-explain-error) {
  font-size: 12px;
}
@media (max-width: 480px) {
  .auth-content {
    padding: 32px 20px;
  }
  .auth-card {
    padding: 36px 24px 24px;
    border-radius: 12px;
  }
}
</style>
