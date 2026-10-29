<template>
  <a-layout style="min-height: 100vh">
    <user-sider value="device" />
    <a-layout>
      <page-header :title="$t('device.title')" :subtitle="$t('device.subtitle')" />
      <a-layout-content :style="{ margin: '24px 16px 0' }">
        <div :style="{ padding: '24px', background: '#fff' }">
          <a-table
            row-key="devid"
            :columns="deviceColumns"
            :dataSource="deviceSource"
            :scroll="{ x: 'max-content' }"
          >
            <template #bodyCell="{ column, record }">
              <template v-if="column.key === 'action'">
                <a-space wrap>
                  <a-button type="primary" size="small" @click="openDeviceDialog(record)">
                    {{ $t('device.edit') }}
                  </a-button>
                  <a-popconfirm
                    :title="$t('device.confirmDelete')"
                    :ok-text="$t('device.yes')"
                    :cancel-text="$t('device.no')"
                    @confirm="deleteDevice(record)"
                  >
                    <a-button danger type="primary" size="small">
                      {{ $t('device.delete') }}
                    </a-button>
                  </a-popconfirm>
                </a-space>
              </template>
            </template>
          </a-table>
        </div>
        <a-modal
          v-model:open="deviceDialogOpen"
          :title="$t('device.editTitle')"
          :ok-text="$t('device.save')"
          :cancel-text="$t('device.cancel')"
          :confirm-loading="deviceSaving"
          :closable="!deviceSaving"
          :mask-closable="!deviceSaving"
          :keyboard="!deviceSaving"
          :cancel-button-props="{ disabled: deviceSaving }"
          @ok="saveDevice"
        >
          <a-alert type="info" show-icon :message="$t('device.editHint')" />
          <a-form layout="vertical" style="margin-top: 20px">
            <a-form-item :label="$t('device.columns.hostname')">
              <a-input :value="deviceDialog.hostname" disabled />
            </a-form-item>
            <a-form-item :label="$t('device.columns.network')">
              <a-input :value="getNetByID(deviceDialog.netid)?.netname || ''" disabled />
            </a-form-item>
            <a-form-item :label="$t('device.addressRange')">
              <a-input :value="getNetByID(deviceDialog.netid)?.dhcp || ''" disabled />
            </a-form-item>
            <a-form-item
              :label="$t('device.columns.ip')"
              :validate-status="deviceError ? 'error' : ''"
              :help="deviceError"
            >
              <a-input v-model:value="deviceDialog.ip" :disabled="deviceSaving" />
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

const deviceColumns = computed(() => [
  {
    title: t('device.columns.hostname'),
    dataIndex: 'hostname',
    key: 'hostname',
    align: 'center'
  },
  {
    title: t('device.columns.network'),
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
    title: t('device.columns.ip'),
    dataIndex: 'ip',
    key: 'ip',
    align: 'center',
    sorter: (a, b) => compareDottedDecimal(a.ip, b.ip)
  },
  {
    title: t('device.columns.country'),
    dataIndex: 'country',
    key: 'country',
    align: 'center',
    sorter: (a, b) => a.country.localeCompare(b.country)
  },
  {
    title: t('device.columns.region'),
    dataIndex: 'region',
    key: 'region',
    align: 'center',
    sorter: (a, b) => {
      const tmp = a.country.localeCompare(b.country)
      if (tmp == 0) {
        return a.region.localeCompare(b.region)
      }
      return tmp
    }
  },
  {
    title: t('device.columns.rx'),
    dataIndex: 'rx',
    key: 'rx',
    align: 'center',
    customRender: (text) => formatRxTx(text.value),
    sorter: (a, b) => a.rx - b.rx
  },
  {
    title: t('device.columns.tx'),
    dataIndex: 'tx',
    key: 'tx',
    align: 'center',
    customRender: (text) => formatRxTx(text.value),
    sorter: (a, b) => a.tx - b.tx
  },
  {
    title: t('device.columns.online'),
    dataIndex: 'online',
    key: 'online',
    align: 'center',
    customRender: (text) => {
      return text.value ? t('device.status.online') : t('device.status.offline')
    },
    sorter: (a, b) => a.online - b.online
  },
  {
    title: t('device.columns.os'),
    dataIndex: 'os',
    key: 'os',
    align: 'center',
    sorter: (a, b) => a.os.localeCompare(b.os)
  },
  {
    title: t('device.columns.version'),
    dataIndex: 'version',
    key: 'version',
    align: 'center',
    sorter: (a, b) => compareDottedDecimal(a.version, b.version)
  },
  {
    title: t('device.columns.lastActiveTime'),
    dataIndex: 'lastActiveTime',
    key: 'lastActiveTime',
    align: 'center',
    sorter: (a, b) => a.lastActiveTime.localeCompare(b.lastActiveTime)
  },
  {
    title: t('device.columns.action'),
    key: 'action',
    align: 'center'
  }
])

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
const deviceDialogOpen = ref(false)
const deviceSaving = ref(false)
const deviceError = ref('')
const deviceDialog = ref({ devid: 0, netid: 0, hostname: '', ip: '' })

const openDeviceDialog = (record) => {
  deviceDialog.value = { ...record }
  deviceError.value = ''
  deviceDialogOpen.value = true
}

const saveDevice = async () => {
  if (deviceSaving.value) return
  const ip = deviceDialog.value.ip.trim()
  const octets = ip.split('.')
  if (
    octets.length !== 4 ||
    !octets.every((octet) => /^(0|[1-9]\d{0,2})$/.test(octet) && Number(octet) <= 255)
  ) {
    deviceError.value = t('device.invalidIP')
    return
  }
  deviceSaving.value = true
  deviceError.value = ''
  try {
    const response = await axios.post('/api/device/update', {
      devid: deviceDialog.value.devid,
      ip
    })
    if (response.data.status !== 0) {
      deviceError.value = t(
        {
          'invalid ip address': 'device.invalidIP',
          'device ip already in use': 'device.ipConflict',
          'device not exists': 'device.notExists'
        }[response.data.msg] || 'device.updateFailed'
      )
      return
    }
    deviceDialogOpen.value = false
    await updateDeviceSource()
  } catch {
    deviceError.value = t('auth.requestFailed')
  } finally {
    deviceSaving.value = false
  }
}

const updateDeviceSource = async () => {
  const response = await axios.post('/api/device/show')

  const status = response.data.status
  if (status == 0) {
    deviceSource.value = response.data.data.devices
  }
}

const netMap = ref(new Map())

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
