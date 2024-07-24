<template>
  <a-layout-header :style="{ background: '#fff', padding: 0 }">
    <a-page-header title="User" sub-title="user management" />
  </a-layout-header>
  <a-layout-content :style="{ margin: '24px 16px 0' }">
    <div
      :style="{
        padding: '24px',
        background: '#fff',
        margin: '0px 0px 24px 0px'
      }"
    >
      <a-form layout="inline" :model="addUserState" @finish="handleFinish">
        <a-form-item>
          <a-input v-model:value="addUserState.username" placeholder="Username">
            <template #prefix><UserOutlined style="color: rgba(0, 0, 0, 0.25)" /></template>
          </a-input>
        </a-form-item>
        <a-form-item>
          <a-input v-model:value="addUserState.password" type="password" placeholder="Password">
            <template #prefix><LockOutlined style="color: rgba(0, 0, 0, 0.25)" /></template>
          </a-input>
        </a-form-item>
        <a-form-item>
          <a-button
            type="primary"
            html-type="submit"
            :disabled="addUserState.username === '' || addUserState.password === ''"
          >
            Create
          </a-button>
        </a-form-item>
      </a-form>
    </div>
    <div :style="{ padding: '24px', background: '#fff' }">
      <a-table :columns="userColumns" :dataSource="userSource">
        <template #bodyCell="{ column, record }">
          <template v-if="column.key === 'action'">
            <a-space wrap>
              <a-button danger size="small" @click="deleteUser(record.userid)"> Delete </a-button>
            </a-space>
          </template>
        </template>
      </a-table>
    </div>
  </a-layout-content>
</template>

<script setup>
import { message } from 'ant-design-vue'
import axios from 'axios'
import { onMounted, ref } from 'vue'

const addUserState = ref({
  username: '',
  password: ''
})

const adminAddUser = async (username, password) => {
  const response = await axios.post('/api/admin/addUser', {
    username: username,
    password: password
  })

  const status = response.data.status
  if (status == 0) {
    message.success(response.data.msg)
    addUserState.value.username = ''
    addUserState.value.password = ''
    updateUserSource()
  }
}

const handleFinish = () => {
  adminAddUser(addUserState.value.username, addUserState.value.password)
}

const userColumns = [
  {
    title: 'Username',
    dataIndex: 'username',
    key: 'username'
  },
  {
    title: 'Role',
    dataIndex: 'role',
    key: 'role'
  },
  {
    title: 'Network',
    dataIndex: 'netnum',
    key: 'netnum'
  },
  {
    title: 'Device',
    dataIndex: 'devnum',
    key: 'devnum'
  },
  {
    title: 'RX',
    dataIndex: 'rxsum',
    key: 'rxsum'
  },
  {
    title: 'TX',
    dataIndex: 'txsum',
    key: 'txsum'
  },
  {
    title: 'Register Time',
    dataIndex: 'regtime',
    key: 'regtime'
  },
  {
    title: 'Action',
    key: 'action'
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
