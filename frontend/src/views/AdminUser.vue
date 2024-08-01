<template>
  <a-layout style="min-height: 98vh">
    <admin-sider value="user" />
    <a-layout>
      <a-layout-header :style="{ background: '#fff', padding: 0 }">
        <a-page-header title="User" sub-title="user management" />
      </a-layout-header>
      <a-layout-content :style="{ margin: '24px 16px 0' }">
        <div :style="{ padding: '24px', background: '#fff' }">
          <a-form :model="userState">
            <a-form-item>
              <a-input v-model:value="userState.username" placeholder="Username">
                <template #prefix><UserOutlined style="color: rgba(0, 0, 0, 0.25)" /></template>
              </a-input>
            </a-form-item>
            <a-form-item>
              <a-input
                v-model:value="userState.password"
                type="password"
                autocomplete="new-password"
                placeholder="Password"
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
                  Create
                </a-button>
                <a-button
                  type="primary"
                  @click="adminUpdateUserPassword"
                  :disabled="userState.username === '' || userState.password === ''"
                >
                  Update
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
                  <a-button danger type="primary" size="small" @click="deleteUser(record.userid)">
                    Delete
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
import { message } from 'ant-design-vue'
import axios from 'axios'
import { onMounted, ref } from 'vue'

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

const userColumns = [
  {
    title: 'Username',
    dataIndex: 'username',
    key: 'username',
    align: 'center'
  },
  {
    title: 'Role',
    dataIndex: 'role',
    key: 'role',
    align: 'center'
  },
  {
    title: 'Network',
    dataIndex: 'netnum',
    key: 'netnum',
    align: 'center'
  },
  {
    title: 'Device',
    dataIndex: 'devnum',
    key: 'devnum',
    align: 'center'
  },
  {
    title: 'RX',
    dataIndex: 'rxsum',
    key: 'rxsum',
    align: 'center'
  },
  {
    title: 'TX',
    dataIndex: 'txsum',
    key: 'txsum',
    align: 'center'
  },
  {
    title: 'Last Active At',
    dataIndex: 'lastActiveTime',
    key: 'lastActiveTime',
    align: 'center'
  },
  {
    title: 'Action',
    key: 'action',
    align: 'center'
  }
]

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
