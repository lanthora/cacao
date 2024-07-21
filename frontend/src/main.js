import './assets/main.css'

import { createApp } from 'vue'
import { createPinia } from 'pinia'

import App from './App.vue'
import router from './router'

//全局ElementPlus
import ElementPlus from 'element-plus'
import 'element-plus/dist/index.css'

const app = createApp(App)

//使用Pinia状态管理
app.use(createPinia())

//启用vue的路由功能
app.use(router)

//使用ElementPlus
app.use(ElementPlus)
app.mount('#app')
