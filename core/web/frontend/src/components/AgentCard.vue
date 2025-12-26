<script setup lang="ts">
import { computed } from 'vue'
import type { Agent } from '../stores/agents'

const props = defineProps<{
  agent: Agent
}>()

const countryFlag = computed(() => {
  if (!props.agent.location?.country_code) return '🌐'
  const code = props.agent.location.country_code.toUpperCase()
  // Convert country code to flag emoji
  const offset = 127397
  return String.fromCodePoint(...[...code].map(c => c.charCodeAt(0) + offset))
})

const formatBytes = (bytes: number) => {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + ' ' + sizes[i]
}

const cpuColor = computed(() => {
  const cpu = props.agent.metrics?.cpu || 0
  if (cpu < 50) return 'bg-green-500'
  if (cpu < 80) return 'bg-yellow-500'
  return 'bg-red-500'
})

const memColor = computed(() => {
  const mem = props.agent.metrics?.memory_percent || 0
  if (mem < 60) return 'bg-green-500'
  if (mem < 85) return 'bg-yellow-500'
  return 'bg-red-500'
})

const trafficColor = computed(() => {
  const percent = props.agent.traffic?.percent || 0
  if (percent < 70) return 'bg-blue-500'
  if (percent < 90) return 'bg-yellow-500'
  return 'bg-red-500'
})
</script>

<template>
  <div class="card group">
    <!-- Header -->
    <div class="flex items-start justify-between mb-4">
      <div class="flex items-center gap-3">
        <span class="text-2xl">{{ countryFlag }}</span>
        <div>
          <h3 class="font-semibold text-white group-hover:text-primary-400 transition-colors">
            {{ agent.name }}
          </h3>
          <p class="text-xs text-gray-500">
            {{ agent.location?.region || agent.location?.country || 'Unknown' }}
          </p>
        </div>
      </div>
      <span :class="['status-dot', agent.status === 'online' ? 'status-online' : 'status-offline']"></span>
    </div>

    <!-- Metrics -->
    <div v-if="agent.status === 'online' && agent.metrics" class="space-y-3">
      <!-- CPU -->
      <div>
        <div class="flex justify-between text-xs mb-1">
          <span class="text-gray-400">CPU</span>
          <span class="text-white font-medium">{{ agent.metrics.cpu.toFixed(1) }}%</span>
        </div>
        <div class="progress-bar">
          <div :class="['progress-fill', cpuColor]" :style="{ width: `${agent.metrics.cpu}%` }"></div>
        </div>
      </div>

      <!-- Memory -->
      <div>
        <div class="flex justify-between text-xs mb-1">
          <span class="text-gray-400">Memory</span>
          <span class="text-white font-medium">{{ agent.metrics.memory_percent.toFixed(1) }}%</span>
        </div>
        <div class="progress-bar">
          <div :class="['progress-fill', memColor]" :style="{ width: `${agent.metrics.memory_percent}%` }"></div>
        </div>
      </div>

      <!-- Traffic -->
      <div v-if="agent.traffic && agent.traffic.limit > 0">
        <div class="flex justify-between text-xs mb-1">
          <span class="text-gray-400">Traffic</span>
          <span class="text-white font-medium">
            {{ formatBytes(agent.traffic.used) }} / {{ formatBytes(agent.traffic.limit) }}
          </span>
        </div>
        <div class="progress-bar">
          <div :class="['progress-fill', trafficColor]" :style="{ width: `${Math.min(agent.traffic.percent, 100)}%` }"></div>
        </div>
      </div>
    </div>

    <!-- Offline State -->
    <div v-else class="py-4 text-center">
      <p class="text-gray-500 text-sm">Offline</p>
    </div>
  </div>
</template>
