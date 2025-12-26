import { createRouter, createWebHistory } from 'vue-router'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/',
      name: 'public',
      component: () => import('../views/PublicDashboard.vue')
    },
    {
      path: '/admin',
      name: 'admin',
      component: () => import('../views/AdminLayout.vue'),
      children: [
        {
          path: '',
          name: 'admin-dashboard',
          component: () => import('../views/admin/Dashboard.vue')
        },
        {
          path: 'agents',
          name: 'admin-agents',
          component: () => import('../views/admin/Agents.vue')
        },
        {
          path: 'agents/:id',
          name: 'admin-agent-detail',
          component: () => import('../views/admin/AgentDetail.vue')
        },
        {
          path: 'tasks',
          name: 'admin-tasks',
          component: () => import('../views/admin/Tasks.vue')
        },
        {
          path: 'alerts',
          name: 'admin-alerts',
          component: () => import('../views/admin/Alerts.vue')
        },
        {
          path: 'settings',
          name: 'admin-settings',
          component: () => import('../views/admin/Settings.vue')
        }
      ]
    },
    {
      path: '/login',
      name: 'login',
      component: () => import('../views/Login.vue')
    }
  ]
})

router.beforeEach((to, from, next) => {
  if (to.path.startsWith('/admin')) {
    const token = localStorage.getItem('token')
    if (!token && to.name !== 'login') {
      next('/login')
      return
    }
  }
  next()
})

export default router
