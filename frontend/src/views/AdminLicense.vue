<template>
  <a-layout style="min-height: 98vh">
    <admin-sider value="license" />
    <a-layout>
      <a-layout-header :style="{ background: '#fff', padding: 0 }">
        <a-page-header title="License" sub-title="license information" />
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
                    Renew
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
import { ref } from 'vue'

const licenseColumns = [
  {
    title: 'Name',
    dataIndex: 'name',
    key: 'name',
    align: 'center'
  },
  {
    title: 'Description',
    dataIndex: 'desc',
    key: 'desc',
    align: 'center'
  },
  {
    title: 'Expire',
    dataIndex: 'expire',
    key: 'expire',
    align: 'center'
  },
  {
    title: 'Action',
    key: 'action',
    align: 'center'
  }
]

const licenseSource = ref([])
</script>
