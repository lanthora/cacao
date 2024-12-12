<template>
  <a-layout style="min-height: 100vh">
    <admin-sider value="setting" />
    <a-layout>
      <a-layout-header :style="{ background: '#fff', padding: 0 }">
        <a-page-header :title="$t('adminSetting.title')" :sub-title="$t('adminSetting.subtitle')" />
      </a-layout-header>
      <a-layout-content :style="{ margin: '24px 16px 0' }">
        <div :style="{ padding: '24px', background: '#fff' }">
          <a-form :label-col="{ style: { width: '200px' } }">
            <a-form-item :label="$t('adminSetting.register.allowed')">
              <a-switch v-model:checked="openRegister" @change="setOpenRegisterConfig" />
            </a-form-item>
            <a-form-item :label="$t('adminSetting.register.interval')">
              <a-input-number
                v-model:value="registerInterval"
                :controls="false"
                @change="setRegisterIntervalConfig"
              >
                <template #addonAfter> {{ $t('adminSetting.register.intervalUnit') }} </template>
              </a-input-number>
            </a-form-item>
            <a-divider />
            <a-form-item :label="$t('adminSetting.userClean.auto')">
              <a-switch v-model:checked="autoCleanUser" @change="setAutoCleanUserConfig" />
            </a-form-item>
            <a-form-item :label="$t('adminSetting.userClean.threshold')">
              <a-input-number
                v-model:value="inactiveUserThreshold"
                :controls="false"
                @change="setInactiveUserThresholdConfig"
              >
                <template #addonAfter> {{ $t('adminSetting.userClean.thresholdUnit') }} </template>
              </a-input-number>
            </a-form-item>
            <a-form-item :label="$t('adminSetting.userClean.manual')">
              <a-button danger type="primary" size="small" @click="cleanInactiveUser">
                {{ $t('adminSetting.userClean.clean') }}
              </a-button>
            </a-form-item>
          </a-form>
        </div>
      </a-layout-content>
      <footer-view />
    </a-layout>
  </a-layout>
</template>

<script setup>
import axios from 'axios'
import { onMounted, ref } from 'vue'
import { message } from 'ant-design-vue'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()

const openRegister = ref()
const registerInterval = ref()
var registerIntervalTimer = null

const autoCleanUser = ref()
const inactiveUserThreshold = ref()
var inactiveUserThresholdTimer = null

onMounted(() => {
  getOpenRegisterConfig()
  getRegisterIntervalConfig()
  getAutoCleanUserConfig()
  getInactiveUserThresholdConfig()
})

const getOpenRegisterConfig = async () => {
  const response = await axios.post('/api/admin/getOpenRegisterConfig')
  const status = response.data.status
  if (status == 0) {
    openRegister.value = response.data.data.openreg
  }
}

const setOpenRegisterConfig = async () => {
  const response = await axios.post('/api/admin/setOpenRegisterConfig', {
    openreg: openRegister.value
  })
  const status = response.data.status
  if (status != 0) {
    openRegister.value = !openRegister.value
  }
}

const getRegisterIntervalConfig = async () => {
  const response = await axios.post('/api/admin/getRegisterIntervalConfig')
  const status = response.data.status
  if (status == 0) {
    registerInterval.value = response.data.data.reginterval
  }
}

const setRegisterIntervalConfig = async () => {
  if (registerIntervalTimer) {
    clearTimeout(registerIntervalTimer)
  }
  registerIntervalTimer = setTimeout(() => {
    axios.post('/api/admin/setRegisterIntervalConfig', {
      reginterval: registerInterval.value
    })
  }, 1000)
}

const getAutoCleanUserConfig = async () => {
  const response = await axios.post('/api/admin/getAutoCleanUserConfig')
  const status = response.data.status
  if (status == 0) {
    autoCleanUser.value = response.data.data.autoCleanUser
  }
}

const setAutoCleanUserConfig = async () => {
  const response = await axios.post('/api/admin/setAutoCleanUserConfig', {
    autoCleanUser: autoCleanUser.value
  })
  const status = response.data.status
  if (status != 0) {
    autoCleanUser.value = !autoCleanUser.value
  }
}

const getInactiveUserThresholdConfig = async () => {
  const response = await axios.post('/api/admin/getInactiveUserThresholdConfig')
  const status = response.data.status
  if (status == 0) {
    inactiveUserThreshold.value = response.data.data.inactiveUserThreshold
  }
}

const setInactiveUserThresholdConfig = async () => {
  if (inactiveUserThresholdTimer) {
    clearTimeout(inactiveUserThresholdTimer)
  }
  inactiveUserThresholdTimer = setTimeout(() => {
    axios.post('/api/admin/setInactiveUserThresholdConfig', {
      inactiveUserThreshold: inactiveUserThreshold.value
    })
  }, 1000)
}

const cleanInactiveUser = async () => {
  const response = await axios.post('/api/admin/cleanInactiveUser')
  const status = response.data.status
  if (status == 0) {
    message.success(t('adminSetting.userClean.success'))
  }
}
</script>
