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
    <div :style="{ padding: '24px', background: '#fff' }">content</div>
  </a-layout-content>
</template>

<script setup>
import axios from 'axios'
import { reactive } from 'vue'

const addUserState = reactive({
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
    // TODO: refresh user list
  }
}

const handleFinish = () => {
  adminAddUser(addUserState.username, addUserState.password)
}
</script>
