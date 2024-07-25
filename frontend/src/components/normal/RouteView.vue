<template>
  <a-layout-header :style="{ background: '#fff', padding: 0 }">
    <a-page-header title="Route" sub-title="multiple local area network networking" />
  </a-layout-header>
  <a-layout-content :style="{ margin: '24px 16px 0' }">
    <div :style="{ padding: '24px', background: '#fff' }">
      <a-space style="margin-bottom: 16px">
        <a-button type="primary" @click="openRouteDialog(null)"> Add </a-button>
      </a-space>
      <a-table :columns="routeColumns" :dataSource="routeSource">
        <template #bodyCell="{ column, record }">
          <template v-if="column.key === 'action'">
            <a-space wrap>
              <a-button type="primary" size="small" @click="openRouteDialog(record)">
                Edit
              </a-button>
              <a-button danger type="primary" size="small" @click="deleteRoute(record)">
                Delete
              </a-button>
            </a-space>
          </template>
        </template>
      </a-table>
    </div>
    <a-modal v-model:open="routeDialogOpen" title="Route" @ok="handleRouteDialog">
      <a-form :model="routeDialogState" :style="{ margin: '24px 0 0' }">
        <a-form-item>
          <a-input-number
            style="width: 100%"
            :controls="false"
            v-model:value="routeDialogState.netid"
            placeholder="Network"
          >
          </a-input-number>
        </a-form-item>
        <a-form-item>
          <a-input v-model:value="routeDialogState.devaddr" placeholder="Device Address"> </a-input>
        </a-form-item>
        <a-form-item>
          <a-input v-model:value="routeDialogState.devmask" placeholder="Device Mask"> </a-input>
        </a-form-item>
        <a-form-item>
          <a-input v-model:value="routeDialogState.dstaddr" placeholder="Destination Address">
          </a-input>
        </a-form-item>
        <a-form-item>
          <a-input v-model:value="routeDialogState.dstmask" placeholder="Destination Mask">
          </a-input>
        </a-form-item>
        <a-form-item>
          <a-input v-model:value="routeDialogState.nexthop" placeholder="Next Hop"> </a-input>
        </a-form-item>
        <a-form-item>
          <a-input-number
            style="width: 100%"
            :controls="false"
            v-model:value="routeDialogState.priority"
            placeholder="Priority"
          >
          </a-input-number>
        </a-form-item>
      </a-form>
    </a-modal>
  </a-layout-content>
</template>

<script setup>
import axios from 'axios'
import { onMounted, ref } from 'vue'

const routeColumns = [
  {
    title: 'Device Address',
    dataIndex: 'devaddr',
    key: 'devaddr',
    align: 'center'
  },
  {
    title: 'Device Mask',
    dataIndex: 'devmask',
    key: 'devmask',
    align: 'center'
  },
  {
    title: 'Destination Address',
    dataIndex: 'dstaddr',
    key: 'dstaddr',
    align: 'center'
  },
  {
    title: 'Destination Mask',
    dataIndex: 'dstmask',
    key: 'dstmask',
    align: 'center'
  },
  {
    title: 'Next Hop',
    dataIndex: 'nexthop',
    key: 'nexthop',
    align: 'center'
  },
  {
    title: 'Priority',
    dataIndex: 'priority',
    key: 'priority',
    align: 'center'
  },
  {
    title: 'Action',
    key: 'action',
    align: 'center'
  }
]

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

const openRouteDialog = (record) => {
  routeDialogState.value.routeid = record ? record.routeid : null
  routeDialogState.value.netid = record ? record.netid : null
  routeDialogState.value.devaddr = record ? record.devaddr : null
  routeDialogState.value.devmask = record ? record.devmask : null
  routeDialogState.value.dstaddr = record ? record.dstaddr : null
  routeDialogState.value.dstmask = record ? record.dstmask : null
  routeDialogState.value.nexthop = record ? record.nexthop : null
  routeDialogState.value.priority = record ? record.priority : null
  routeDialogOpen.value = true
}

const handleRouteDialog = () => {
  if (routeDialogState.value.routeid == null) {
    addRoute()
  } else {
    editRoute()
  }
}

const addRoute = async () => {
  const response = await axios.post('/api/route/insert', routeDialogState.value)

  const status = response.data.status
  if (status == 0) {
    routeDialogOpen.value = false
    updateRouteSource()
  }
}

const editRoute = async () => {
  const response = await axios.post('/api/route/edit', routeDialogState.value)

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
</script>
