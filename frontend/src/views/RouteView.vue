<template>
  <a-layout style="min-height: 100vh">
    <user-sider value="route" />
    <a-layout>
      <a-layout-header :style="{ background: '#fff', padding: 0 }">
        <a-page-header :title="$t('route.title')" :sub-title="$t('route.subtitle')" />
      </a-layout-header>
      <a-layout-content :style="{ margin: '24px 16px 0' }">
        <div :style="{ padding: '24px', background: '#fff' }">
          <a-space style="margin-bottom: 16px">
            <a-button type="primary" @click="openRouteDialog">
              {{ $t('route.add') }}
            </a-button>
          </a-space>
          <a-table :columns="routeColumns" :dataSource="routeSource" :scroll="{ x: 'max-content' }">
            <template #bodyCell="{ column, record }">
              <template v-if="column.key === 'action'">
                <a-space wrap>
                  <a-button danger type="primary" size="small" @click="deleteRoute(record)">
                    {{ $t('route.delete') }}
                  </a-button>
                </a-space>
              </template>
            </template>
          </a-table>
        </div>
        <a-modal v-model:open="routeDialogOpen" :title="$t('route.modalTitle')" @ok="addRoute">
          <a-form :model="routeDialogState" :style="{ margin: '24px 0 0' }">
            <a-form-item>
              <a-select
                ref="select"
                v-model:value="routeDialogState.netid"
                :placeholder="$t('route.placeholder.network')"
                :options="netOptions"
              >
              </a-select>
            </a-form-item>
            <a-form-item>
              <a-input 
                v-model:value="routeDialogState.devaddr" 
                :placeholder="$t('route.placeholder.devAddr')"
              />
            </a-form-item>
            <a-form-item>
              <a-input 
                v-model:value="routeDialogState.devmask" 
                :placeholder="$t('route.placeholder.devMask')"
              />
            </a-form-item>
            <a-form-item>
              <a-input 
                v-model:value="routeDialogState.dstaddr" 
                :placeholder="$t('route.placeholder.dstAddr')"
              />
            </a-form-item>
            <a-form-item>
              <a-input 
                v-model:value="routeDialogState.dstmask" 
                :placeholder="$t('route.placeholder.dstMask')"
              />
            </a-form-item>
            <a-form-item>
              <a-input 
                v-model:value="routeDialogState.nexthop" 
                :placeholder="$t('route.placeholder.nextHop')"
              />
            </a-form-item>
            <a-form-item>
              <a-input-number
                style="width: 100%"
                :controls="false"
                v-model:value="routeDialogState.priority"
                :placeholder="$t('route.placeholder.priority')"
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
import { onMounted, onBeforeMount, ref, computed } from 'vue'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()

const routeColumns = computed(() => [
  {
    title: t('route.columns.network'),
    dataIndex: 'netid',
    key: 'netid',
    align: 'center',
    customRender: (text) => {
      var net = getNetByID(text.value)
      return net ? net.netname : ''
    }
  },
  {
    title: t('route.columns.devAddr'),
    dataIndex: 'devaddr',
    key: 'devaddr',
    align: 'center'
  },
  {
    title: t('route.columns.devMask'),
    dataIndex: 'devmask',
    key: 'devmask',
    align: 'center'
  },
  {
    title: t('route.columns.dstAddr'),
    dataIndex: 'dstaddr',
    key: 'dstaddr',
    align: 'center'
  },
  {
    title: t('route.columns.dstMask'),
    dataIndex: 'dstmask',
    key: 'dstmask',
    align: 'center'
  },
  {
    title: t('route.columns.nextHop'),
    dataIndex: 'nexthop',
    key: 'nexthop',
    align: 'center'
  },
  {
    title: t('route.columns.priority'),
    dataIndex: 'priority',
    key: 'priority',
    align: 'center'
  },
  {
    title: t('route.columns.action'),
    key: 'action',
    align: 'center'
  }
])

const routeDialogOpen = ref(false)

const routeSource = ref([])

const updateRouteSource = async () => {
  const response = await axios.post('/api/route/show')

  const status = response.data.status
  if (status == 0) {
    routeSource.value = response.data.data.routes
  }
}

onMounted(() => {
  updateRouteSource()
})

const routeDialogState = ref({
  routeid: 0,
  netid: 0,
  devaddr: '',
  devmask: '',
  dstaddr: '',
  dstmask: '',
  nexthop: '',
  priority: 0
})

const openRouteDialog = () => {
  routeDialogState.value.routeid = null
  routeDialogState.value.netid = null
  routeDialogState.value.devaddr = null
  routeDialogState.value.devmask = null
  routeDialogState.value.dstaddr = null
  routeDialogState.value.dstmask = null
  routeDialogState.value.nexthop = null
  routeDialogState.value.priority = null
  routeDialogOpen.value = true
}

const addRoute = async () => {
  const response = await axios.post('/api/route/insert', routeDialogState.value)

  const status = response.data.status
  if (status == 0) {
    routeDialogOpen.value = false
    updateRouteSource()
  }
}

const deleteRoute = async (record) => {
  const response = await axios.post('/api/route/delete', {
    routeid: record.routeid
  })

  const status = response.data.status
  if (status == 0) {
    updateRouteSource()
  }
}

const netMap = ref()
const netOptions = ref([])

const getNetByID = (netid) => {
  return netMap.value.get(netid)
}

const updateNetMap = async () => {
  const response = await axios.post('/api/net/show')

  const status = response.data.status
  if (status == 0) {
    const nets = response.data.data.nets
    netMap.value = new Map(
      nets.map(function (object) {
        return [object.netid, object]
      })
    )
    netOptions.value = nets.map(function (object) {
      return { label: object.netname, value: object.netid }
    })
  }
}
onBeforeMount(() => {
  updateNetMap()
})
</script>
