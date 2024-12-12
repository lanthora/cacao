<template>
  <a-layout style="min-height: 100vh">
    <user-sider value="user" />
    <a-layout>
      <a-layout-header :style="{ background: '#fff', padding: 0 }">
        <a-page-header :title="$t('user.title')" :sub-title="$t('user.subtitle')" />
      </a-layout-header>
      <a-layout-content :style="{ margin: '24px 16px 0' }">
        <div :style="{ padding: '24px', background: '#fff' }">
          <a-form :label-col="{ style: { width: '150px' } }">
            <a-form-item :label="$t('user.username')"> {{ username }} </a-form-item>
            <a-form-item :label="$t('user.changePassword')">
              <a-form :model="passwordState" @finish="changePassword">
                <a-form-item>
                  <a-input
                    v-model:value="passwordState.old"
                    type="password"
                    autocomplete="new-password"
                    :placeholder="$t('user.oldPassword')"
                  >
                  </a-input>
                </a-form-item>
                <a-form-item>
                  <a-input
                    v-model:value="passwordState.new"
                    type="password"
                    autocomplete="new-password"
                    :placeholder="$t('user.newPassword')"
                  >
                  </a-input>
                </a-form-item>
                <a-form-item>
                  <a-button
                    type="primary"
                    html-type="submit"
                    :disabled="passwordState.old === '' || passwordState.new === ''"
                  >
                    {{ $t('user.update') }}
                  </a-button>
                </a-form-item>
              </a-form>
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
    if (response.data.data.role === 'normal') {
      username.value = response.data.data.name
    } else {
      router.push('/')
    }
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
</script>
