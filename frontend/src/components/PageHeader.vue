<template>
  <a-layout-header class="page-header">
    <a-page-header
      v-if="title || subtitle"
      class="page-header-title"
      :title="title"
      :sub-title="subtitle"
    />
    <div class="page-header-actions"><language-switch /></div>
  </a-layout-header>
</template>

<script setup>
defineProps({
  title: String,
  subtitle: String
})
</script>

<style scoped>
.page-header {
  box-sizing: border-box;
  display: flex;
  align-items: center;
  justify-content: space-between;
  flex-wrap: wrap;
  gap: 8px 16px;
  height: auto;
  min-height: 64px;
  padding: 12px 24px;
  line-height: normal;
  background: #fff;
}

.page-header-actions {
  display: flex;
  align-items: center;
  margin-left: auto;
  flex-shrink: 0;
}

.page-header-title {
  flex: 1 1 240px;
  min-width: 0;
  padding: 0;
}

.page-header-title :deep(.ant-page-header-heading-left) {
  min-width: 0;
  flex-wrap: wrap;
}

.page-header-title :deep(.ant-page-header-heading-title),
.page-header-title :deep(.ant-page-header-heading-sub-title) {
  white-space: normal;
  overflow-wrap: anywhere;
}

@media (max-width: 480px) {
  .page-header {
    padding: 12px 16px;
  }
}
</style>
