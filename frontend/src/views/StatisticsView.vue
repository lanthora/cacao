<template>
  <a-layout style="min-height: 100vh">
    <user-sider value="statistics" />
    <a-layout>
      <a-layout-header :style="{ background: '#fff', padding: 0 }">
        <a-page-header :title="$t('statistics.title')" :sub-title="$t('statistics.subtitle')" />
      </a-layout-header>
      <a-layout-content :style="{ margin: '24px 16px 0' }">
        <div :style="{ padding: '24px', background: '#fff' }">
          <a-row>
            <a-col :xs="24" :sm="24" :md="12" :lg="6">
              <a-statistic :title="$t('statistics.columns.net')" :value="statistics.netnum" />
            </a-col>
            <a-col :xs="24" :sm="24" :md="12" :lg="6">
              <a-statistic :title="$t('statistics.columns.device')" :value="statistics.devnum" />
            </a-col>
            <a-col :xs="24" :sm="24" :md="12" :lg="6">
              <a-statistic :title="$t('statistics.columns.rx')" :value="statistics.rxsum" />
            </a-col>
            <a-col :xs="24" :sm="24" :md="12" :lg="6">
              <a-statistic :title="$t('statistics.columns.tx')" :value="statistics.txsum" />
            </a-col>
          </a-row>
        </div>
      </a-layout-content>
      <footer-view />
    </a-layout>
  </a-layout>
</template>

<script setup>
import axios from 'axios'
import { ref, onMounted } from 'vue'

const statistics = ref({
  netnum: '',
  devnum: '',
  rxsum: '',
  txsum: ''
})

onMounted(() => {
  getUserInformation()
})

const getUserInformation = async () => {
  const response = await axios.post('/api/user/statistics')
  const status = response.data.status
  if (status == 0) {
    statistics.value = response.data.data
  }
}
</script>
