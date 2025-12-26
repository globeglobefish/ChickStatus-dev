<script setup lang="ts">
import { ref, onMounted } from 'vue'
import api from '../../api'
import { ServerStackIcon, ExclamationTriangleIcon, CheckCircleIcon } from '@heroicons/vue/24/outline'

const stats = ref({
  total: 0,
  online: 0,
  offline: 0,
  alerts: 0
})

const recentAlerts = ref<any[]>([])
const loading = ref(true)

onMounted(async () => {
  try {
    const [agentsRes, alertsRes] = await Promise.all([
      api.get('/api/admin/agents'),
      api.get('/api/admin/alerts/active')
    ])
    
    const agents = agentsRes.data || []
    stats.value.total = agents.length
    stats.value.online = agents.filter((a: any) => a.status === 'online').length
    stats.value.offline = agents.filter((a: any) => a.status === 'offline').length
    
    recentAlerts.value = alertsRes.data || []
    stats.value.alerts = recentAlerts.value.length
  } finally {
    loading.value = false
  }
})
</script>

<template>
  <div>
    <h1 class="text-2xl font-bold text-white mb-6">Dashboard</h1>

    <!-- Stats Grid -->
    <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4 mb-8">
      <div class="card">
        <div class="flex items-center gap-4">
          <div class="w-12 h-12 bg-blue-500/20 rounded-lg flex items-center justify-center">
            <ServerStackIcon class="w-6 h-6 text-blue-400" />
          </div>
          <div>
            <p class="text-sm text-gray-400">Total Agents</p>
            <p class="text-2xl font-bold text-white">{{ stats.total }}</p>
          </div>
        </div>
      </div>

      <div class="card">
        <div class="flex items-center gap-4">
          <div class="w-12 h-12 bg-green-500/20 rounded-lg flex items-center justify-center">
            <CheckCircleIcon class="w-6 h-6 text-green-400" />
          </div>
          <div>
            <p class="text-sm text-gray-400">Online</p>
            <p class="text-2xl font-bold text-green-400">{{ stats.online }}</p>
          </div>
        </div>
      </div>

      <div class="card">
        <div class="flex items-center gap-4">
          <div class="w-12 h-12 bg-red-500/20 rounded-lg flex items-center justify-center">
            <ServerStackIcon class="w-6 h-6 text-red-400" />
          </div>
          <div>
            <p class="text-sm text-gray-400">Offline</p>
            <p class="text-2xl font-bold text-red-400">{{ stats.offline }}</p>
          </div>
        </div>
      </div>

      <div class="card">
        <div class="flex items-center gap-4">
          <div class="w-12 h-12 bg-yellow-500/20 rounded-lg flex items-center justify-center">
            <ExclamationTriangleIcon class="w-6 h-6 text-yellow-400" />
          </div>
          <div>
            <p class="text-sm text-gray-400">Active Alerts</p>
            <p class="text-2xl font-bold text-yellow-400">{{ stats.alerts }}</p>
          </div>
        </div>
      </div>
    </div>

    <!-- Recent Alerts -->
    <div class="card">
      <h2 class="text-lg font-semibold text-white mb-4">Active Alerts</h2>
      
      <div v-if="loading" class="py-8 text-center text-gray-400">
        Loading...
      </div>
      
      <div v-else-if="recentAlerts.length === 0" class="py-8 text-center text-gray-400">
        No active alerts
      </div>
      
      <div v-else class="space-y-3">
        <div 
          v-for="alert in recentAlerts" 
          :key="alert.id"
          class="flex items-center gap-4 p-3 bg-gray-700/50 rounded-lg"
        >
          <div class="w-2 h-2 rounded-full bg-red-500"></div>
          <div class="flex-1 min-w-0">
            <p class="text-sm font-medium text-white truncate">{{ alert.rule_name }}</p>
            <p class="text-xs text-gray-400">{{ alert.agent_name }} · {{ alert.metric_type }}</p>
          </div>
          <div class="text-right">
            <p class="text-sm font-medium text-red-400">{{ alert.value.toFixed(1) }}%</p>
            <p class="text-xs text-gray-500">threshold: {{ alert.threshold }}%</p>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
