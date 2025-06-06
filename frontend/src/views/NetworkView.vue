<template>
  <a-layout style="min-height: 100vh">
    <user-sider value="network" />
    <a-layout>
      <a-layout-header :style="{ background: '#fff', padding: 0 }">
        <a-page-header :title="$t('network.title')" :sub-title="$t('network.subtitle')" />
      </a-layout-header>
      <a-layout-content :style="{ margin: '24px 16px 0' }">
        <div :style="{ padding: '24px', background: '#fff' }">
          <a-space style="margin-bottom: 16px">
            <a-button type="primary" @click="openNetDialog(null)">
              {{ $t('network.add') }}
            </a-button>
          </a-space>
          <a-table :columns="netColumns" :dataSource="netSource" :scroll="{ x: 'max-content' }">
            <template #bodyCell="{ column, record }">
              <template v-if="column.key === 'action'">
                <a-space wrap>
                  <a-button type="primary" size="small" @click="openNetDialog(record)">
                    {{ $t('network.edit') }}
                  </a-button>
                  <a-button danger type="primary" size="small" @click="deleteNet(record)">
                    {{ $t('network.delete') }}
                  </a-button>
                </a-space>
              </template>
            </template>
          </a-table>
        </div>
        <a-modal v-model:open="netDialogOpen" :title="$t('network.modalTitle')" @ok="handleNetDialog">
          <a-form :model="netDialogState" :style="{ margin: '24px 0 0' }">
            <a-form-item>
              <a-input 
                v-model:value="netDialogState.netname" 
                :placeholder="$t('network.inputNetname')"
              />
            </a-form-item>
            <a-form-item>
              <a-input 
                v-model:value="netDialogState.password" 
                :placeholder="$t('network.inputPassword')"
              />
            </a-form-item>
            <a-form-item>
              <a-input 
                v-model:value="netDialogState.dhcp" 
                :placeholder="$t('network.inputDhcp')"
              />
            </a-form-item>
            <a-form-item>
              <a-select v-model:value="netDialogState.broadcast" :placeholder="$t('network.broadcast')">
                <a-select-option value="true">{{ $t('network.enable') }}</a-select-option>
                <a-select-option value="false">{{ $t('network.disable') }}</a-select-option>
              </a-select>
            </a-form-item>
            <a-form-item>
              <a-input-number
                style="width: 100%"
                :controls="false"
                v-model:value="netDialogState.lease"
                :placeholder="$t('network.inputLease')"
              />
            </a-form-item>
          </a-form>
        </a-modal>
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

const netColumns = computed(() => [
  {
    title: t('network.netName'),
    dataIndex: 'netname',
    key: 'netname',
    align: 'center',
    size: '150px'
  },
  {
    title: t('network.password'),
    dataIndex: 'password',
    key: 'password',
    align: 'center'
  },
  {
    title: t('network.dhcp'),
    dataIndex: 'dhcp',
    key: 'dhcp',
    align: 'center'
  },
  {
    title: t('network.broadcast'),
    dataIndex: 'broadcast',
    key: 'broadcast',
    align: 'center',
    customRender: (text) => {
      return text.value ? t('network.enable') : t('network.disable')
    }
  },
  {
    title: t('network.lease'),
    dataIndex: 'lease',
    key: 'lease',
    align: 'center'
  },
  {
    title: t('network.action'),
    key: 'action',
    align: 'center'
  }
])

const netDialogOpen = ref(false)

const netSource = ref([])

const updateNetSource = async () => {
  const response = await axios.post('/api/net/show')

  const status = response.data.status
  if (status == 0) {
    netSource.value = response.data.data.nets
  }
}

onMounted(() => {
  updateNetSource()
})

const netDialogState = ref({
  netid: 0,
  netname: '',
  password: '',
  dhcp: '',
  broadcast: false,
  lease: 0
})

const openNetDialog = (record) => {
  netDialogState.value.netid = record ? record.netid : null
  netDialogState.value.netname = record ? record.netname : null
  netDialogState.value.password = record ? record.password : null
  netDialogState.value.dhcp = record ? record.dhcp : null
  netDialogState.value.broadcast = record ? (record.broadcast ? 'true' : 'false') : null
  netDialogState.value.lease = record ? record.lease : null
  netDialogOpen.value = true
}

const handleNetDialog = () => {
  if (netDialogState.value.netid == null) {
    addNet()
  } else {
    editNet()
  }
}

const addNet = async () => {
  const response = await axios.post('/api/net/insert', {
    netname: netDialogState.value.netname,
    password: netDialogState.value.password,
    dhcp: netDialogState.value.dhcp,
    broadcast: netDialogState.value.broadcast === 'true',
    lease: netDialogState.value.lease
  })

  const status = response.data.status
  if (status == 0) {
    netDialogOpen.value = false
    updateNetSource()
  }
}

const editNet = async () => {
  const response = await axios.post('/api/net/edit', {
    netid: netDialogState.value.netid,
    netname: netDialogState.value.netname,
    password: netDialogState.value.password,
    dhcp: netDialogState.value.dhcp,
    broadcast: netDialogState.value.broadcast === 'true',
    lease: netDialogState.value.lease
  })

  const status = response.data.status
  if (status == 0) {
    netDialogOpen.value = false
    updateNetSource()
  }
}

const deleteNet = async (record) => {
  const response = await axios.post('/api/net/delete', {
    netid: record.netid
  })

  const status = response.data.status
  if (status == 0) {
    updateNetSource()
  }
}
</script>
