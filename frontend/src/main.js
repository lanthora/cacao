import { createWebHistory, createRouter } from 'vue-router'
import { createApp } from 'vue'
import App from './App.vue'

import RegisterView from './views/RegisterView.vue'
import LoginView from './views/LoginView.vue'
import AdminUserView from './views/AdminUser.vue'
import AdminSetting from './views/AdminSetting.vue'
import LoadingView from './views/LoadingView.vue'
import RouteView from './views/RouteView.vue'
import UserView from './views/UserView.vue'
import DeviceView from './views/DeviceView.vue'
import NetworkView from './views/NetworkView.vue'
import StatisticsView from './views/StatisticsView.vue'

const routes = [
  { path: '/', component: LoadingView },
  { path: '/login', component: LoginView },
  { path: '/register', component: RegisterView },
  { path: '/statistics', component: StatisticsView },
  { path: '/network', component: NetworkView },
  { path: '/device', component: DeviceView },
  { path: '/route', component: RouteView },
  { path: '/user', component: UserView },
  { path: '/admin/user', component: AdminUserView },
  { path: '/admin/setting', component: AdminSetting },
  { path: '/:pathMatch(.*)', redirect: '/' }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

createApp(App).use(router).mount('#app')
