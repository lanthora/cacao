<template>
  <a-layout style="min-height: 98vh">
    <user-sider />
    <a-layout>
      <a-layout-header :style="{ background: '#fff', padding: 0 }">
        <a-page-header title="User" sub-title="user information" />
      </a-layout-header>
      <a-layout-content :style="{ margin: '24px 16px 0' }">
        <div :style="{ padding: '24px', background: '#fff' }">
          <a-form :label-col="{ style: { width: '150px' } }">
            <a-form-item label="User Name"> {{ username }} </a-form-item>
            <a-form-item label="Change Password">
              <a-form :model="passwordState" @finish="changePassword">
                <a-form-item>
                  <a-input
                    v-model:value="passwordState.old"
                    type="password"
                    placeholder="Old Password"
                  >
                  </a-input>
                </a-form-item>
                <a-form-item>
                  <a-input
                    v-model:value="passwordState.new"
                    type="password"
                    placeholder="New Password"
                  >
                  </a-input>
                </a-form-item>
                <a-form-item>
                  <a-button
                    type="primary"
                    html-type="submit"
                    :disabled="passwordState.old === '' || passwordState.new === ''"
                  >
                    Update
                  </a-button>
                </a-form-item>
              </a-form>
            </a-form-item>
            <a-form-item label="Logout">
              <a-button @click="logout"> Confirm </a-button>
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
import { ref, onMounted } from 'vue'
import { message } from 'ant-design-vue'
import { useRouter } from 'vue-router'

const router = useRouter()

const username = ref('')

onMounted(() => {
  fetchUsername()
})

const fetchUsername = async () => {
  const response = await axios.post('/api/user/info')
  const status = response.data.status
  if (status == 0) {
    username.value = response.data.data.name
  }
}

const passwordState = ref({
  old: '',
  new: ''
})

const changePassword = async () => {
  const response = await axios.post('/api/user/changePassword', {
    old: passwordState.value.old,
    new: passwordState.value.new
  })

  const status = response.data.status
  if (status == 0) {
    message.success(response.data.msg)
  }
  passwordState.value.old = ''
  passwordState.value.new = ''
}

const logout = async () => {
  const response = await axios.post('/api/user/logout')

  const status = response.data.status
  if (status == 0) {
    router.push('/login')
  }
}
</script>
