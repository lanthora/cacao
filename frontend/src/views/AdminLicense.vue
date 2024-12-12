<template>
  <a-layout style="min-height: 100vh">
    <admin-sider value="license" />
    <a-layout>
      <a-layout-header :style="{ background: '#fff', padding: 0 }">
        <a-page-header :title="$t('adminLicense.title')" :sub-title="$t('adminLicense.subtitle')" />
      </a-layout-header>
      <a-layout-content :style="{ margin: '24px 16px 0' }">
        <div :style="{ padding: '24px', margin: '24px 0px 0px', background: '#fff' }">
          <a-table
            :columns="licenseColumns"
            :dataSource="licenseSource"
            :scroll="{ x: 'max-content' }"
          >
            <template #bodyCell="{ column, record }">
              <template v-if="column.key === 'action'">
                <a-space wrap>
                  <a-button danger type="primary" size="small" @click="renew(record.licenseid)">
                    {{ $t('adminLicense.renew') }}
                  </a-button>
                </a-space>
              </template>
            </template>
          </a-table>
        </div>
      </a-layout-content>
      <footer-view />
    </a-layout>
  </a-layout>
</template>

<script setup>
import axios from 'axios'
import { onMounted, ref, computed } from 'vue'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()

const licenseColumns = computed(() => [
  {
    title: t('adminLicense.columns.licenseId'),
    dataIndex: 'licenseid',
    key: 'licenseid',
    align: 'center'
  },
  {
    title: t('adminLicense.columns.description'),
    dataIndex: 'desc',
    key: 'desc',
    align: 'center'
  },
  {
    title: t('adminLicense.columns.expire'),
    dataIndex: 'expire',
    key: 'expire',
    align: 'center'
  },
  {
    title: t('adminLicense.columns.action'),
    key: 'action',
    align: 'center'
  }
])

const licenseSource = ref([])

const updateLicenseSource = async () => {
  const response = await axios.post('/api/admin/showLicenses')

  const status = response.data.status
  if (status == 0) {
    licenseSource.value = response.data.data.licenses
  }
}

onMounted(() => {
  updateLicenseSource()
})

const renew = async (licenseid) => {
  const response = await axios.post('/api/admin/renewLicense', {
    licenseid: licenseid
  })

  const status = response.data.status
  if (status == 0) {
    updateLicenseSource()
  }
}
</script>
