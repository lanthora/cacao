<template>
  <a-config-provider :theme="{ token: { colorPrimary: '#147d70', borderRadius: 8 } }">
    <a-layout class="settings-layout">
      <admin-sider value="setting" />
      <a-layout class="settings-main">
        <page-header :title="$t('adminSetting.title')" :subtitle="$t('adminSetting.subtitle')" />
        <a-layout-content class="settings-content">
          <div class="settings-container">
            <header class="settings-intro">
              <div>
                <div class="eyebrow">{{ $t('adminSetting.design.workspace') }}</div>
                <h1>{{ $t('adminSetting.design.heading') }}</h1>
                <p>{{ $t('adminSetting.design.description') }}</p>
              </div>
              <div class="intro-symbol" aria-hidden="true"><setting-outlined /></div>
            </header>

            <section class="preferences-grid" :aria-label="$t('adminSetting.design.preferences')">
              <article class="preference-card">
                <div class="card-heading">
                  <span class="section-icon"><user-add-outlined /></span>
                  <div>
                    <h2>{{ $t('adminSetting.register.title') }}</h2>
                    <p>{{ $t('adminSetting.design.registerDescription') }}</p>
                  </div>
                </div>
                <div class="setting-row">
                  <label for="open-register">{{ $t('adminSetting.register.allowed') }}</label>
                  <a-switch
                    id="open-register"
                    v-model:checked="openRegister"
                    :disabled="openRegister === undefined"
                    @change="setOpenRegisterConfig"
                  />
                </div>
                <div class="setting-row">
                  <label for="register-interval">{{ $t('adminSetting.register.interval') }}</label>
                  <a-input-number
                    id="register-interval"
                    v-model:value="registerInterval"
                    :min="0"
                    :precision="0"
                    :controls="false"
                    :disabled="registerInterval === undefined"
                    @change="setRegisterIntervalConfig"
                  >
                    <template #addonAfter>{{ $t('adminSetting.register.intervalUnit') }}</template>
                  </a-input-number>
                </div>
                <div class="card-note">
                  <check-circle-outlined />{{ $t('adminSetting.design.autoSaved') }}
                </div>
              </article>

              <article class="preference-card">
                <div class="card-heading">
                  <span class="section-icon"><clear-outlined /></span>
                  <div>
                    <h2>{{ $t('adminSetting.userClean.title') }}</h2>
                    <p>{{ $t('adminSetting.design.cleanDescription') }}</p>
                  </div>
                </div>
                <div class="setting-row">
                  <label for="auto-clean">{{ $t('adminSetting.userClean.auto') }}</label>
                  <a-switch
                    id="auto-clean"
                    v-model:checked="autoCleanUser"
                    :disabled="autoCleanUser === undefined"
                    @change="setAutoCleanUserConfig"
                  />
                </div>
                <div class="setting-row">
                  <label for="inactive-threshold">{{
                    $t('adminSetting.userClean.threshold')
                  }}</label>
                  <a-input-number
                    id="inactive-threshold"
                    v-model:value="inactiveUserThreshold"
                    :min="1"
                    :precision="0"
                    :controls="false"
                    :disabled="inactiveUserThreshold === undefined"
                    @change="setInactiveUserThresholdConfig"
                  >
                    <template #addonAfter>{{
                      $t('adminSetting.userClean.thresholdUnit')
                    }}</template>
                  </a-input-number>
                </div>
                <div class="cleanup-footer">
                  <span>{{ $t('adminSetting.userClean.manual') }}</span>
                  <a-popconfirm
                    :title="$t('adminSetting.design.cleanConfirm')"
                    :ok-text="$t('adminSetting.userClean.clean')"
                    :cancel-text="$t('adminSetting.design.cancel')"
                    @confirm="cleanInactiveUser"
                  >
                    <a-button danger type="text" size="small" :loading="cleaningUsers"
                      ><delete-outlined />{{ $t('adminSetting.userClean.clean') }}</a-button
                    >
                  </a-popconfirm>
                </div>
              </article>
            </section>

            <section class="database-section" aria-labelledby="database-heading">
              <div class="section-heading">
                <div>
                  <h2 id="database-heading">{{ $t('adminSetting.design.databaseHeading') }}</h2>
                  <p>{{ $t('adminSetting.design.databaseDescription') }}</p>
                </div>
                <span class="schedule-note"
                  ><clock-circle-outlined />{{ $t('adminSetting.design.hourly') }}</span
                >
              </div>
              <a-alert
                v-if="geoDBLoadError"
                type="error"
                show-icon
                :message="$t('adminSetting.design.loadFailed')"
                class="load-error"
              >
                <template #action
                  ><a-button size="small" @click="getGeoDBConfig">{{
                    $t('adminSetting.design.retry')
                  }}</a-button></template
                >
              </a-alert>
              <div v-if="geoDBLoading" class="database-grid" aria-busy="true">
                <div v-for="item in 2" :key="item" class="database-card">
                  <a-skeleton active :paragraph="{ rows: 6 }" />
                </div>
              </div>
              <div v-else class="database-grid">
                <article v-for="config in geoDBConfigs" :key="config.type" class="database-card">
                  <header class="database-card-header">
                    <div class="database-identity">
                      <span class="database-icon"><database-outlined /></span>
                      <div>
                        <h3>{{ config.type.toUpperCase() }}</h3>
                        <span>{{ $t(`adminSetting.design.${config.type}Description`) }}</span>
                      </div>
                    </div>
                    <span class="status-badge" :class="geoDBStatus(config)" role="status"
                      ><span class="status-dot"></span
                      >{{ $t(`adminSetting.design.${geoDBStatus(config)}`) }}</span
                    >
                  </header>
                  <div class="database-fields">
                    <div class="field-group">
                      <label :for="`${config.type}-url`">{{ $t('adminSetting.geoDB.url') }}</label>
                      <a-input
                        :id="`${config.type}-url`"
                        v-model:value="config.downloadURL"
                        :disabled="config.saving || config.updating"
                        :placeholder="`https://example.com/city.${config.type}`"
                        autocomplete="off"
                        spellcheck="false"
                      >
                        <template #prefix><link-outlined class="input-icon" /></template>
                      </a-input>
                      <p class="field-help">{{ $t('adminSetting.design.rawFile') }}</p>
                    </div>
                    <div class="interval-setting">
                      <div>
                        <label :for="`${config.type}-interval`">{{
                          $t('adminSetting.geoDB.interval')
                        }}</label>
                        <p class="field-help">{{ $t('adminSetting.design.disableHelp') }}</p>
                      </div>
                      <a-input-number
                        :id="`${config.type}-interval`"
                        v-model:value="config.updateIntervalHours"
                        :min="0"
                        :max="87600"
                        :precision="0"
                        :disabled="config.saving || config.updating"
                      >
                        <template #addonAfter>{{ $t('adminSetting.geoDB.hours') }}</template>
                      </a-input-number>
                    </div>
                    <div class="last-updated">
                      <clock-circle-outlined />
                      <div>
                        <span>{{ $t('adminSetting.geoDB.lastSuccess') }}</span
                        ><time v-if="config.lastSuccessAt" :datetime="config.lastSuccessAt">{{
                          new Date(config.lastSuccessAt).toLocaleString()
                        }}</time
                        ><strong v-else>{{ $t('adminSetting.geoDB.never') }}</strong>
                      </div>
                    </div>
                    <div v-if="config.error" class="database-error" role="alert">
                      <exclamation-circle-outlined /><span>{{ config.error }}</span>
                    </div>
                  </div>
                  <footer class="database-actions">
                    <span
                      class="save-hint"
                      :class="{ pending: geoDBDirty(config) }"
                      aria-live="polite"
                      ><edit-outlined v-if="geoDBDirty(config)" /><check-outlined v-else />{{
                        $t(
                          geoDBDirty(config)
                            ? 'adminSetting.design.unsaved'
                            : 'adminSetting.design.savedState'
                        )
                      }}</span
                    >
                    <div class="action-buttons">
                      <a-button
                        :loading="config.saving"
                        :disabled="config.updating || !geoDBDirty(config)"
                        @click="saveGeoDBConfig(config)"
                        >{{ $t('adminSetting.geoDB.save') }}</a-button
                      >
                      <a-tooltip
                        :title="geoDBDirty(config) ? $t('adminSetting.geoDB.savedHelp') : ''"
                      >
                        <span
                          ><a-button
                            type="primary"
                            :loading="config.updating"
                            :disabled="config.saving || !config.savedURL || geoDBDirty(config)"
                            @click="updateGeoDB(config)"
                            ><template #icon><sync-outlined /></template
                            >{{ $t('adminSetting.geoDB.update') }}</a-button
                          ></span
                        >
                      </a-tooltip>
                    </div>
                  </footer>
                </article>
              </div>
              <aside class="lookup-note">
                <info-circle-outlined />
                <div>
                  <div class="lookup-priority">
                    <strong>{{ $t('adminSetting.design.lookupOrder') }}</strong
                    ><span
                      >IPDB <span class="priority-arrow">→</span> MMDB
                      <span class="priority-arrow">→</span> ipinfo</span
                    >
                  </div>
                  <p>{{ $t('adminSetting.design.lookupHelp') }}</p>
                </div>
              </aside>
            </section>
          </div>
        </a-layout-content>
        <footer-view />
      </a-layout>
    </a-layout>
  </a-config-provider>
</template>

<script setup>
import axios from 'axios'
import { onMounted, onBeforeUnmount, ref } from 'vue'
import { message } from 'ant-design-vue'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()

onBeforeUnmount(() => {
  clearTimeout(registerIntervalTimer)
  clearTimeout(inactiveUserThresholdTimer)
})

const geoDBConfigs = ref([])
const geoDBLoading = ref(true)
const geoDBLoadError = ref(false)
const cleaningUsers = ref(false)
const geoDBStatus = (config) => {
  if (config.updating) return 'updating'
  if (config.error) return 'updateError'
  return config.savedURL && config.savedInterval > 0 ? 'enabled' : 'disabled'
}

const geoDBDirty = (config) =>
  config.downloadURL !== config.savedURL || config.updateIntervalHours !== config.savedInterval

const getGeoDBConfig = async () => {
  geoDBLoading.value = true
  geoDBLoadError.value = false
  try {
    const { data } = await axios.post('/api/admin/getGeoDBConfig')
    if (data.status !== 0) throw new Error(data.msg)
    geoDBConfigs.value = data.data.configs.map((config) => ({
      ...config,
      savedURL: config.downloadURL,
      savedInterval: config.updateIntervalHours,
      saving: false,
      updating: false,
      error: ''
    }))
  } catch (error) {
    geoDBLoadError.value = true
    message.error(error.message || t('adminSetting.geoDB.failed'))
  } finally {
    geoDBLoading.value = false
  }
}

const saveGeoDBConfig = async (config) => {
  if (config.updateIntervalHours === null || !Number.isInteger(config.updateIntervalHours)) {
    message.error(t('adminSetting.geoDB.invalidInterval'))
    return
  }
  config.error = ''
  config.saving = true
  try {
    const { data } = await axios.post('/api/admin/setGeoDBConfig', {
      type: config.type,
      downloadURL: config.downloadURL.trim(),
      updateIntervalHours: config.updateIntervalHours
    })
    if (data.status !== 0) throw new Error(data.msg)
    config.downloadURL = config.downloadURL.trim()
    config.savedURL = config.downloadURL
    config.savedInterval = config.updateIntervalHours
    message.success(t('adminSetting.geoDB.saved'))
  } catch (error) {
    config.error = error.message || t('adminSetting.geoDB.failed')
    message.error(config.error)
  } finally {
    config.saving = false
  }
}

const updateGeoDB = async (config) => {
  config.error = ''
  config.updating = true
  try {
    const { data } = await axios.post(
      '/api/admin/updateGeoDB',
      { type: config.type },
      { timeout: 310000 }
    )
    if (data.status !== 0) throw new Error(data.msg)
    config.lastSuccessAt = data.data.lastSuccessAt
    message.success(t('adminSetting.geoDB.updated'))
  } catch (error) {
    config.error = error.message || t('adminSetting.geoDB.failed')
    message.error(config.error)
  } finally {
    config.updating = false
  }
}

const openRegister = ref()
const registerInterval = ref()
var registerIntervalTimer = null

const autoCleanUser = ref()
const inactiveUserThreshold = ref()
var inactiveUserThresholdTimer = null

onMounted(() => {
  getGeoDBConfig()
  getOpenRegisterConfig()
  getRegisterIntervalConfig()
  getAutoCleanUserConfig()
  getInactiveUserThresholdConfig()
})

const getOpenRegisterConfig = async () => {
  const response = await axios.post('/api/admin/getOpenRegisterConfig')
  const status = response.data.status
  if (status == 0) {
    openRegister.value = response.data.data.openreg
  }
}

const setOpenRegisterConfig = async () => {
  const response = await axios.post('/api/admin/setOpenRegisterConfig', {
    openreg: openRegister.value
  })
  const status = response.data.status
  if (status != 0) {
    openRegister.value = !openRegister.value
  }
}

const getRegisterIntervalConfig = async () => {
  const response = await axios.post('/api/admin/getRegisterIntervalConfig')
  const status = response.data.status
  if (status == 0) {
    registerInterval.value = response.data.data.reginterval
  }
}

const setRegisterIntervalConfig = async () => {
  if (registerIntervalTimer) {
    clearTimeout(registerIntervalTimer)
  }
  registerIntervalTimer = setTimeout(() => {
    axios.post('/api/admin/setRegisterIntervalConfig', {
      reginterval: registerInterval.value
    })
  }, 1000)
}

const getAutoCleanUserConfig = async () => {
  const response = await axios.post('/api/admin/getAutoCleanUserConfig')
  const status = response.data.status
  if (status == 0) {
    autoCleanUser.value = response.data.data.autoCleanUser
  }
}

const setAutoCleanUserConfig = async () => {
  const response = await axios.post('/api/admin/setAutoCleanUserConfig', {
    autoCleanUser: autoCleanUser.value
  })
  const status = response.data.status
  if (status != 0) {
    autoCleanUser.value = !autoCleanUser.value
  }
}

const getInactiveUserThresholdConfig = async () => {
  const response = await axios.post('/api/admin/getInactiveUserThresholdConfig')
  const status = response.data.status
  if (status == 0) {
    inactiveUserThreshold.value = response.data.data.inactiveUserThreshold
  }
}

const setInactiveUserThresholdConfig = async () => {
  if (inactiveUserThresholdTimer) {
    clearTimeout(inactiveUserThresholdTimer)
  }
  inactiveUserThresholdTimer = setTimeout(() => {
    axios.post('/api/admin/setInactiveUserThresholdConfig', {
      inactiveUserThreshold: inactiveUserThreshold.value
    })
  }, 1000)
}

const cleanInactiveUser = async () => {
  cleaningUsers.value = true
  try {
    const response = await axios.post('/api/admin/cleanInactiveUser')
    if (response.data.status == 0) message.success(t('adminSetting.userClean.success'))
  } catch (error) {
    message.error(error.message)
  } finally {
    cleaningUsers.value = false
  }
}
</script>

<style scoped>
.settings-layout {
  min-height: 100vh;
}
.settings-main {
  min-width: 0;
  background: #f5f7f8;
}
.settings-main :deep(.page-header) {
  border-bottom: 1px solid #e8edef;
}
.settings-content {
  padding: 36px 40px 16px;
}
.settings-container {
  max-width: 1180px;
  margin: 0 auto;
  color: #23343c;
}
.settings-intro {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 28px;
  gap: 24px;
}
.eyebrow {
  color: #147d70;
  font-size: 11px;
  font-weight: 650;
  letter-spacing: 1.8px;
  margin-bottom: 9px;
}
h1,
h2,
h3,
p {
  margin: 0;
}
h1 {
  font-size: 28px;
  line-height: 1.4;
  font-weight: 650;
  letter-spacing: -0.7px;
}
.settings-intro p {
  margin-top: 8px;
  color: #78858d;
  font-size: 13px;
  line-height: 1.8;
}
.intro-symbol {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 62px;
  height: 62px;
  border: 1px solid #dde9e6;
  border-radius: 18px;
  color: #528a80;
  background: #edf4f2;
  font-size: 28px;
  flex-shrink: 0;
}
.preferences-grid,
.database-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 22px;
}
.preference-card,
.database-card {
  background: #fff;
  border: 1px solid #e1e7ea;
  border-radius: 12px;
  overflow: hidden;
  box-shadow: 0 2px 3px #23343c03;
}
.preference-card {
  padding: 22px 24px 0;
}
.card-heading {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 20px;
}
.section-icon {
  display: grid;
  place-items: center;
  width: 38px;
  height: 38px;
  background: #f1f5f5;
  color: #5a747c;
  border-radius: 10px;
  font-size: 18px;
  flex-shrink: 0;
}
h2 {
  font-size: 16px;
  line-height: 1.5;
  font-weight: 600;
}
.card-heading p {
  font-size: 12px;
  color: #849099;
  margin-top: 3px;
  line-height: 1.6;
}
.setting-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 20px;
  min-height: 51px;
  font-size: 13px;
}
.setting-row + .setting-row {
  border-top: 1px solid #f0f3f4;
}
.setting-row :deep(.ant-input-number-group-wrapper) {
  width: 138px;
  flex-shrink: 0;
}
.card-note,
.cleanup-footer {
  display: flex;
  align-items: center;
  min-height: 45px;
  border-top: 1px solid #edf1f3;
  margin: 12px -24px 0;
  padding: 0 24px;
  background: #fcfdfd;
  color: #8a969e;
  font-size: 12px;
  gap: 7px;
}
.card-note :deep(.anticon) {
  color: #719c90;
}
.cleanup-footer {
  justify-content: space-between;
}
.cleanup-footer :deep(.ant-btn) {
  font-size: 12px;
}
.database-section {
  margin-top: 34px;
}
.section-heading {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 20px;
  margin-bottom: 18px;
}
.section-heading h2 {
  font-size: 18px;
  letter-spacing: -0.3px;
}
.section-heading p {
  font-size: 12px;
  color: #849099;
  margin-top: 5px;
  line-height: 1.7;
}
.schedule-note {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  flex-shrink: 0;
  color: #7d8a92;
  font-size: 12px;
}
.database-card {
  display: flex;
  flex-direction: column;
  min-width: 0;
}
.database-card:has(.ant-skeleton) {
  padding: 24px;
}
.database-card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 12px;
  padding: 22px 24px;
  border-bottom: 1px solid #edf1f3;
}
.database-identity {
  display: flex;
  align-items: center;
  gap: 11px;
}
.database-icon {
  width: 40px;
  height: 44px;
  display: grid;
  place-items: center;
  background: #edf5f3;
  border: 1px solid #e0eeea;
  border-radius: 9px;
  color: #338a7d;
  font-size: 21px;
  flex-shrink: 0;
}
h3 {
  font-size: 17px;
  font-weight: 650;
  letter-spacing: 0.7px;
}
.database-identity div > span {
  display: block;
  margin-top: 3px;
  font-size: 11px;
  color: #8a969e;
}
.status-badge {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  background: #f1f4f5;
  color: #78868f;
  padding: 4px 9px;
  border: 1px solid #e9edef;
  border-radius: 20px;
  font-size: 11px;
  white-space: nowrap;
}
.status-dot {
  width: 5px;
  height: 5px;
  border-radius: 50%;
  background: currentColor;
}
.status-badge.enabled {
  background: #edf7f1;
  color: #438769;
  border-color: #dceee2;
}
.status-badge.updating {
  background: #ecf4ff;
  color: #427bb7;
  border-color: #dbeafb;
}
.status-badge.updateError {
  background: #fff1ef;
  color: #b95649;
  border-color: #f8dfda;
}
.database-fields {
  padding: 22px 24px;
  flex: 1;
}
.field-group label,
.interval-setting label {
  display: block;
  margin-bottom: 9px;
  font-size: 12px;
  font-weight: 550;
}
.field-group :deep(.ant-input-affix-wrapper) {
  padding: 8px 11px;
}
.field-group :deep(.ant-input) {
  font-size: 12px;
}
.input-icon {
  color: #9aabb2;
  margin-right: 5px;
}
.field-help {
  font-size: 11px;
  line-height: 1.7;
  color: #8a969e;
  margin-top: 7px;
}
.interval-setting {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 18px;
  margin-top: 24px;
}
.interval-setting label {
  margin-bottom: 0;
}
.interval-setting .field-help {
  margin-top: 5px;
}
.interval-setting :deep(.ant-input-number-group-wrapper) {
  width: 138px;
  flex-shrink: 0;
}
.last-updated {
  display: flex;
  align-items: center;
  gap: 11px;
  border: 1px solid #edf1f3;
  background: #f8fafb;
  border-radius: 8px;
  margin-top: 23px;
  padding: 12px 14px;
}
.last-updated > .anticon {
  font-size: 17px;
  color: #93a4ae;
}
.last-updated div > span {
  display: block;
  font-size: 11px;
  color: #83919a;
  margin-bottom: 4px;
}
.last-updated strong,
.last-updated time {
  display: block;
  color: #536771;
  font-size: 12px;
  font-weight: 500;
  font-variant-numeric: tabular-nums;
}
.database-actions {
  display: flex;
  align-items: center;
  justify-content: space-between;
  flex-wrap: wrap;
  gap: 12px;
  border-top: 1px solid #edf1f3;
  padding: 15px 24px;
}
.save-hint {
  display: flex;
  align-items: center;
  gap: 5px;
  font-size: 11px;
  color: #8a969e;
}
.save-hint.pending {
  color: #b18439;
}
.action-buttons {
  display: flex;
  gap: 8px;
  margin-left: auto;
}
.action-buttons :deep(.ant-btn) {
  font-size: 12px;
  box-shadow: none;
}
.database-error {
  display: flex;
  gap: 8px;
  padding: 10px 12px;
  color: #b95649;
  background: #fff5f3;
  border-radius: 6px;
  font-size: 12px;
  line-height: 1.6;
  margin-top: 12px;
  overflow-wrap: anywhere;
}
.database-error > .anticon {
  margin-top: 3px;
}
.lookup-note {
  display: flex;
  align-items: flex-start;
  gap: 10px;
  margin-top: 18px;
  padding: 16px 18px;
  background: #ecf2f3;
  border-radius: 9px;
  color: #71848c;
}
.lookup-note > .anticon {
  margin-top: 3px;
  font-size: 15px;
  color: #7b969c;
}
.lookup-priority {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 10px 16px;
  font-size: 11px;
}
.lookup-priority strong {
  font-weight: 550;
}
.lookup-priority > span {
  font-size: 11px;
  font-weight: 550;
  color: #526e76;
}
.priority-arrow {
  margin: 0 10px;
  color: #92a6ac;
}
.lookup-note p {
  margin-top: 7px;
  font-size: 11px;
  line-height: 1.8;
}
.load-error {
  margin-bottom: 18px;
}
.settings-main :deep(.ant-layout-footer) {
  background: transparent;
  font-size: 11px;
  color: #a4afb5;
  padding: 24px;
}
:deep(.ant-input-number),
:deep(.ant-input-affix-wrapper),
:deep(.ant-btn-default) {
  border-color: #dce4e8;
}
@media (min-width: 1600px) {
  .settings-content {
    padding-top: 44px;
  }
}
@media (max-width: 1120px) {
  .settings-content {
    padding: 28px 24px 12px;
  }
  .preferences-grid,
  .database-grid {
    gap: 16px;
  }
  .database-card-header,
  .database-fields {
    padding: 20px;
  }
  .database-actions {
    padding: 15px 20px;
  }
  .save-hint {
    width: 100%;
  }
  .action-buttons {
    margin-left: 0;
  }
}
@media (max-width: 760px) {
  .preferences-grid,
  .database-grid {
    grid-template-columns: 1fr;
  }
  .settings-content {
    padding: 24px 18px 8px;
  }
  .settings-intro {
    margin-bottom: 24px;
  }
  h1 {
    font-size: 25px;
  }
  .intro-symbol {
    width: 50px;
    height: 50px;
    font-size: 24px;
    border-radius: 14px;
  }
  .section-heading {
    align-items: flex-start;
    flex-direction: column;
    gap: 10px;
  }
  .save-hint {
    width: auto;
  }
  .action-buttons {
    margin-left: auto;
  }
}
@media (max-width: 380px) {
  .intro-symbol {
    display: none;
  }
  .preference-card {
    padding-left: 18px;
    padding-right: 18px;
  }
  .card-note,
  .cleanup-footer {
    margin-left: -18px;
    margin-right: -18px;
    padding-left: 18px;
    padding-right: 18px;
  }
  .interval-setting {
    gap: 10px;
  }
  .interval-setting :deep(.ant-input-number-group-wrapper) {
    width: 120px;
  }
  .save-hint {
    width: 100%;
  }
  .action-buttons {
    margin-left: 0;
  }
}
@media (max-width: 991px) {
  .settings-content {
    padding-top: 44px;
  }
}
</style>
