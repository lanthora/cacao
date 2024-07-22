import { createWebHistory, createRouter } from 'vue-router'
import { createApp } from 'vue'
import App from './App.vue'

import LoadingView from './LoadingView.vue'
import RegisterView from './RegisterView.vue'
import LoginView from './LoginView.vue'
import UserView from './UserView.vue'
import AdminView from './AdminView.vue'

const routes = [
  { path: '/', component: LoadingView },
  { path: '/register', component: RegisterView },
  { path: '/login', component: LoginView },
  { path: '/user', component: UserView },
  { path: '/admin', component: AdminView }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

createApp(App).use(router).mount('#app')
