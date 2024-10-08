<template>
  <a-layout style="min-height: 98vh">
    <user-sider value="device" />
    <a-layout>
      <a-layout-header :style="{ background: '#fff', padding: 0 }">
        <a-page-header title="Device" sub-title="view and manage devices" />
      </a-layout-header>
      <a-layout-content :style="{ margin: '24px 16px 0' }">
        <div :style="{ padding: '24px', background: '#fff' }">
          <a-table
            :columns="deviceColumns"
            :dataSource="deviceSource"
            :scroll="{ x: 'max-content' }"
          >
            <template #bodyCell="{ column, record }">
              <template v-if="column.key === 'action'">
                <a-space wrap>
                  <a-button danger type="primary" size="small" @click="deleteDevice(record)">
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
import axios from 'axios'
import { onMounted, onBeforeMount, ref } from 'vue'

const deviceColumns = [
  {
    title: 'Host Name',
    dataIndex: 'hostname',
    key: 'hostname',
    align: 'center'
  },
  {
    title: 'IP',
    dataIndex: 'ip',
    key: 'ip',
    align: 'center'
  },
  {
    title: 'Network',
    dataIndex: 'netid',
    key: 'netid',
    align: 'center',
    customRender: (text) => {
      var net = getNetByID(text.value)
      return net ? net.netname : ''
    }
  },
  {
    title: 'RX',
    dataIndex: 'rx',
    key: 'rx',
    align: 'center',
    customRender: (text) => {
      return formatRxTx(text.value)
    }
  },
  {
    title: 'TX',
    dataIndex: 'tx',
    key: 'tx',
    align: 'center',
    customRender: (text) => {
      return formatRxTx(text.value)
    }
  },
  {
    title: 'Online',
    dataIndex: 'online',
    key: 'online',
    align: 'center',
    customRender: (text) => {
      return text.value ? 'true' : 'false'
    }
  },
  {
    title: 'OS',
    dataIndex: 'os',
    key: 'os',
    align: 'center'
  },
  {
    title: 'Version',
    dataIndex: 'version',
    key: 'version',
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

const formatRxTx = (value) => {
  var cnt = 0
  var unit = ['B', 'KB', 'MB', 'GB', 'TB', 'EB']
  while (value > 1024) {
    cnt += 1
    value /= 1024
  }
  return value.toFixed(3) + ' ' + unit[cnt]
}

const deviceSource = ref([])

const updateDeviceSource = async () => {
  const response = await axios.post('/api/device/show')

  const status = response.data.status
  if (status == 0) {
    deviceSource.value = response.data.data.devices
  }
}

const netMap = ref()

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
  }
}

const deleteDevice = async (record) => {
  const response = await axios.post('/api/device/delete', {
    devid: record.devid
  })

  const status = response.data.status
  if (status == 0) {
    updateDeviceSource()
  }
}

onBeforeMount(() => {
  updateNetMap()
})

onMounted(() => {
  updateDeviceSource()
})
</script>
