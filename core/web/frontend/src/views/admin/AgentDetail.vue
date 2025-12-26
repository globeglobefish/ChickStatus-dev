<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import api from '../../api'

const route = useRoute()
const agent = ref<any>(null)
const metrics = ref<any>(null)
const traffic = ref<any>(null)
const loading = ref(true)

const formatBytes = (bytes: number) => {
  if (!bytes) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}

onMounted(async () => {
  const id = route.params.id as string
  try {
    const [agentRes, metricsRes, trafficRes] = await Promise.all([
      api.get(`/api/admin/agents/${id}`),
      api.get(`/api/admin/agents/${id}/metrics`),
      api.get(`/api/admin/agents/${id}/traffic`)
    ])
    agent.value = agentRes.data
    metrics.value = metricsRes.data
    traffic.value = trafficRes.data
  } finally {
    loading.value = false
  }
})

function getCountryFlag(code: string) {
  if (!code) return '🌐'
  const offset = 127397
  return String.fromCodePoint(...[...code.toUpperCase()].map(c => c.charCodeAt(0) + offset))
}
</script>

<template>
  <div v-if="loading" class="py-8 text-center text-gray-400">
    Loading...
  </div>

  <div v-else-if="agent">
    <!-- Header -->
    <div class="flex items-start justify-between mb-6">
      <div class="flex items-center gap-4">
        <span class="text-4xl">{{ getCountryFlag(agent.location?.country_code) }}</span>
        <div>
          <h1 class="text-2xl font-bold text-white">{{ agent.custom_name || agent.hostname }}</h1>
          <p class="text-gray-400">{{ agent.ip }} · {{ agent.location?.city || agent.location?.country }}</p>
        </div>
      </div>
      <span :class="[
        'inline-flex items-center gap-2 px-3 py-1.5 rounded-full text-sm font-medium',
        agent.status === 'online' ? 'bg-green-500/20 text-green-400' : 'bg-red-500/20 text-red-400'
      ]">
        <span :class="['w-2 h-2 rounded-full', agent.status === 'online' ? 'bg-green-400' : 'bg-red-400']"></span>
        {{ agent.status }}
      </span>
    </div>

    <!-- Metrics Grid -->
    <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4 mb-6">
      <div class="card">
        <p class="text-sm text-gray-400 mb-2">CPU Usage</p>
        <p class="text-3xl font-bold text-white">{{ metrics?.cpu?.toFixed(1) || 0 }}%</p>
        <div class="progress-bar mt-3">
          <div class="progress-fill bg-blue-500" :style="{ width: `${metrics?.cpu || 0}%` }"></div>
        </div>
      </div>

      <div class="card">
        <p class="text-sm text-gray-400 mb-2">Memory Usage</p>
        <p class="text-3xl font-bold text-white">{{ metrics?.memory?.percent?.toFixed(1) || 0 }}%</p>
        <p class="text-xs text-gray-500 mt-1">
          {{ formatBytes(metrics?.memory?.used || 0) }} / {{ formatBytes(metrics?.memory?.total || 0) }}
        </p>
        <div class="progress-bar mt-3">
          <div class="progress-fill bg-purple-500" :style="{ width: `${metrics?.memory?.percent || 0}%` }"></div>
        </div>
      </div>

      <div class="card">
        <p class="text-sm text-gray-400 mb-2">Disk Usage</p>
        <p class="text-3xl font-bold text-white">{{ metrics?.disks?.[0]?.percent?.toFixed(1) || 0 }}%</p>
        <p class="text-xs text-gray-500 mt-1">
          {{ formatBytes(metrics?.disks?.[0]?.used || 0) }} / {{ formatBytes(metrics?.disks?.[0]?.total || 0) }}
        </p>
        <div class="progress-bar mt-3">
          <div class="progress-fill bg-yellow-500" :style="{ width: `${metrics?.disks?.[0]?.percent || 0}%` }"></div>
        </div>
      </div>

      <div class="card">
        <p class="text-sm text-gray-400 mb-2">Network</p>
        <div class="flex items-baseline gap-2">
          <span class="text-green-400">↑</span>
          <span class="text-xl font-bold text-white">{{ formatBytes(metrics?.network?.bytes_sent_rate || 0) }}/s</span>
        </div>
        <div class="flex items-baseline gap-2 mt-1">
          <span class="text-blue-400">↓</span>
          <span class="text-xl font-bold text-white">{{ formatBytes(metrics?.network?.bytes_recv_rate || 0) }}/s</span>
        </div>
      </div>
    </div>

    <!-- Traffic Stats -->
    <div v-if="traffic" class="card mb-6">
      <h2 class="text-lg font-semibold text-white mb-4">Traffic Usage</h2>
      <div class="flex items-center justify-between mb-3">
        <span class="text-gray-400">Current Cycle</span>
        <span class="text-white font-medium">
          {{ formatBytes(traffic.total_bytes) }} / {{ traffic.limit ? formatBytes(traffic.limit) : 'Unlimited' }}
        </span>
      </div>
      <div class="progress-bar h-4">
        <div 
          class="progress-fill bg-gradient-to-r from-blue-500 to-purple-500" 
          :style="{ width: `${Math.min(traffic.percent || 0, 100)}%` }"
        ></div>
      </div>
      <div class="flex justify-between mt-2 text-xs text-gray-500">
        <span>{{ traffic.percent?.toFixed(1) || 0 }}% used</span>
        <span>Resets: {{ new Date(traffic.cycle_end).toLocaleDateString() }}</span>
      </div>
    </div>

    <!-- Agent Info -->
    <div class="card">
      <h2 class="text-lg font-semibold text-white mb-4">Agent Information</h2>
      <div class="grid grid-cols-2 gap-4 text-sm">
        <div>
          <p class="text-gray-400">Hostname</p>
          <p class="text-white">{{ agent.hostname }}</p>
        </div>
        <div>
          <p class="text-gray-400">IP Address</p>
          <p class="text-white">{{ agent.ip }}</p>
        </div>
        <div>
          <p class="text-gray-400">OS</p>
          <p class="text-white">{{ agent.os }} / {{ agent.arch }}</p>
        </div>
        <div>
          <p class="text-gray-400">Version</p>
          <p class="text-white">{{ agent.version }}</p>
        </div>
        <div>
          <p class="text-gray-400">Last Seen</p>
          <p class="text-white">{{ new Date(agent.last_seen_at).toLocaleString() }}</p>
        </div>
        <div>
          <p class="text-gray-400">Created</p>
          <p class="text-white">{{ new Date(agent.created_at).toLocaleString() }}</p>
        </div>
      </div>
    </div>
  </div>
</template>
