<template>
  <a-layout-header :style="{ background: '#fff', padding: 0 }">
    <a-page-header title="Setting" sub-title="system configuration" />
  </a-layout-header>
  <a-layout-content :style="{ margin: '24px 16px 0' }">
    <div :style="{ padding: '24px', background: '#fff' }">
      <a-form :label-col="{ style: { width: '150px' } }">
        <a-form-item label="Registration Allowed">
          <a-switch v-model:checked="openRegister" @change="setOpenRegisterConfig" />
        </a-form-item>
        <a-form-item label="Registration Interval">
          <a-input-number
            v-model:value="registerInterval"
            :controls="false"
            @change="setRegisterIntervalConfig"
          >
            <template #addonAfter> min </template>
          </a-input-number>
        </a-form-item>
      </a-form>
    </div>
  </a-layout-content>
</template>

<script setup>
import axios from 'axios'
import { ref, onMounted } from 'vue'
const openRegister = ref()
const registerInterval = ref()
var registerIntervalTimer = null

onMounted(() => {
  getOpenRegisterConfig()
  getRegisterIntervalConfig()
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
  }, 3000)
}
</script>
