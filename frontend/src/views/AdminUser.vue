<template>
  <a-layout style="min-height: 100vh">
    <admin-sider value="user" />
    <a-layout>
      <a-layout-header :style="{ background: '#fff', padding: 0 }">
        <a-page-header :title="$t('adminUser.title')" :sub-title="$t('adminUser.subtitle')" />
      </a-layout-header>
      <a-layout-content :style="{ margin: '24px 16px 0' }">
        <div :style="{ padding: '24px', background: '#fff' }">
          <a-form :model="userState">
            <a-form-item>
              <a-input 
                v-model:value="userState.username" 
                :placeholder="$t('adminUser.placeholder.username')"
              >
                <template #prefix><UserOutlined style="color: rgba(0, 0, 0, 0.25)" /></template>
              </a-input>
            </a-form-item>
            <a-form-item>
              <a-input
                v-model:value="userState.password"
                type="password"
                autocomplete="new-password"
                :placeholder="$t('adminUser.placeholder.password')"
              >
                <template #prefix><LockOutlined style="color: rgba(0, 0, 0, 0.25)" /></template>
              </a-input>
            </a-form-item>
            <a-form-item>
              <a-space>
                <a-button
                  type="primary"
                  @click="adminAddUser"
                  :disabled="userState.username === '' || userState.password === ''"
                >
                  {{ $t('adminUser.create') }}
                </a-button>
                <a-button
                  type="primary"
                  @click="adminUpdateUserPassword"
                  :disabled="userState.username === '' || userState.password === ''"
                >
                  {{ $t('adminUser.update') }}
                </a-button>
              </a-space>
            </a-form-item>
          </a-form>
        </div>
        <div :style="{ padding: '24px', margin: '24px 0px 0px', background: '#fff' }">
          <a-table :columns="userColumns" :dataSource="userSource" :scroll="{ x: 'max-content' }">
            <template #bodyCell="{ column, record }">
              <template v-if="column.key === 'action'">
                <a-space wrap>
                  <a-popconfirm
                    :title="$t('adminUser.confirmDelete')"
                    :ok-text="$t('adminUser.yes')"
                    :cancel-text="$t('adminUser.no')"
                    @confirm="deleteUser(record.userid)"
                  >
                    <a-button danger type="primary" size="small">
                      {{ $t('adminUser.delete') }}
                    </a-button>
                  </a-popconfirm>
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
import { message } from 'ant-design-vue'
import axios from 'axios'
import { onMounted, ref, computed } from 'vue'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()

const formatRxTx = (value) => {
  var cnt = 0
  var unit = ['B', 'KB', 'MB', 'GB', 'TB', 'EB']
  while (value > 1024) {
    cnt += 1
    value /= 1024
  }
  return value.toFixed(3) + ' ' + unit[cnt]
}

const userState = ref({
  username: '',
  password: ''
})

const adminAddUser = async () => {
  const response = await axios.post('/api/admin/addUser', {
    username: userState.value.username,
    password: userState.value.password
  })

  const status = response.data.status
  if (status == 0) {
    message.success(response.data.msg)
    userState.value.username = ''
    userState.value.password = ''
    updateUserSource()
  }
}

const adminUpdateUserPassword = async () => {
  const response = await axios.post('/api/admin/updateUserPassword', {
    username: userState.value.username,
    password: userState.value.password
  })

  const status = response.data.status
  if (status == 0) {
    message.success(response.data.msg)
    userState.value.username = ''
    userState.value.password = ''
    updateUserSource()
  }
}

const userColumns = computed(() => [
  {
    title: t('adminUser.columns.username'),
    dataIndex: 'username',
    key: 'username',
    align: 'center'
  },
  {
    title: t('adminUser.columns.role'),
    dataIndex: 'role',
    key: 'role',
    align: 'center'
  },
  {
    title: t('adminUser.columns.network'),
    dataIndex: 'netnum',
    key: 'netnum',
    align: 'center'
  },
  {
    title: t('adminUser.columns.device'),
    dataIndex: 'devnum',
    key: 'devnum',
    align: 'center'
  },
  {
    title: t('adminUser.columns.rx'),
    dataIndex: 'rxsum',
    key: 'rxsum',
    customRender: (text) => formatRxTx(text.value),
    align: 'center'
  },
  {
    title: t('adminUser.columns.tx'),
    dataIndex: 'txsum',
    key: 'txsum',
    customRender: (text) => formatRxTx(text.value),
    align: 'center'
  },
  {
    title: t('adminUser.columns.lastActiveTime'),
    dataIndex: 'lastActiveTime',
    key: 'lastActiveTime',
    align: 'center'
  },
  {
    title: t('adminUser.columns.action'),
    key: 'action',
    align: 'center'
  }
])

const userSource = ref([])

const updateUserSource = async () => {
  const response = await axios.post('/api/admin/showUsers')

  const status = response.data.status
  if (status == 0) {
    userSource.value = response.data.data.users
  }
}

onMounted(() => {
  updateUserSource()
})

const deleteUser = async (e) => {
  const response = await axios.post('/api/admin/deleteUser', {
    userid: e
  })

  const status = response.data.status
  if (status == 0) {
    updateUserSource()
  }
}
</script>
