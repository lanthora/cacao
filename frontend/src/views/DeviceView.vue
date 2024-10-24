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
                  <a-popconfirm
                    title="Are you sure delete this device?"
                    ok-text="Yes"
                    cancel-text="No"
                    @confirm="deleteDevice(record)"
                  >
                    <a-button danger type="primary" size="small"> Delete </a-button>
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
    title: 'Network',
    dataIndex: 'netid',
    key: 'netid',
    align: 'center',
    customRender: (text) => {
      var net = getNetByID(text.value)
      return net ? net.netname : ''
    },
    sorter: (a, b) => a.netid - b.netid
  },
  {
    title: 'IP',
    dataIndex: 'ip',
    key: 'ip',
    align: 'center',
    sorter: (a, b) => compareDottedDecimal(a.ip, b.ip)
  },
  {
    title: 'Country',
    dataIndex: 'country',
    key: 'country',
    align: 'center',
    sorter: (a, b) => a.country.localeCompare(b.country)
  },
  {
    title: 'City',
    dataIndex: 'city',
    key: 'city',
    align: 'center',
    sorter: (a, b) => {
      const tmp = a.country.localeCompare(b.country)
      if (tmp == 0) {
        return a.city.localeCompare(b.city)
      }
      return tmp
    }
  },
  {
    title: 'RX',
    dataIndex: 'rx',
    key: 'rx',
    align: 'center',
    customRender: (text) => formatRxTx(text.value),
    sorter: (a, b) => a.rx - b.rx
  },
  {
    title: 'TX',
    dataIndex: 'tx',
    key: 'tx',
    align: 'center',
    customRender: (text) => formatRxTx(text.value),
    sorter: (a, b) => a.tx - b.tx
  },
  {
    title: 'Online',
    dataIndex: 'online',
    key: 'online',
    align: 'center',
    customRender: (text) => {
      return text.value ? 'true' : 'false'
    },
    sorter: (a, b) => a.online - b.online
  },
  {
    title: 'OS',
    dataIndex: 'os',
    key: 'os',
    align: 'center',
    sorter: (a, b) => a.os.localeCompare(b.os)
  },
  {
    title: 'Version',
    dataIndex: 'version',
    key: 'version',
    align: 'center',
    sorter: (a, b) => compareDottedDecimal(a.version, b.version)
  },
  {
    title: 'Last Active At',
    dataIndex: 'lastActiveTime',
    key: 'lastActiveTime',
    align: 'center',
    sorter: (a, b) => a.lastActiveTime.localeCompare(b.lastActiveTime)
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

const compareDottedDecimal = (a, b) => {
  const lista = a.split('.')
  const listb = b.split('.')

  for (let i = 0; i < Math.max(lista.length, listb.length); i++) {
    const itema = parseInt(lista[i] || '0', 10)
    const itemb = parseInt(listb[i] || '0', 10)

    if (itema > itemb) {
      return 1
    } else if (itema < itemb) {
      return -1
    }
  }

  return 0
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
