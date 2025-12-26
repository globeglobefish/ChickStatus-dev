<script setup lang="ts">
import { onMounted, onUnmounted, computed } from 'vue'
import { useAgentsStore } from '../stores/agents'
import AgentCard from '../components/AgentCard.vue'

const store = useAgentsStore()

const onlineCount = computed(() => store.agents.filter(a => a.status === 'online').length)
const offlineCount = computed(() => store.agents.filter(a => a.status === 'offline').length)
const allOnline = computed(() => offlineCount.value === 0 && store.agents.length > 0)

onMounted(() => {
  store.fetchPublicAgents()
  store.connectWebSocket()
})

onUnmounted(() => {
  store.disconnectWebSocket()
})
</script>

<template>
  <div class="min-h-screen bg-gray-900">
    <!-- Header -->
    <header class="border-b border-gray-800 bg-gray-900/80 backdrop-blur-sm sticky top-0 z-10">
      <div class="max-w-7xl mx-auto px-4 py-4 flex items-center justify-between">
        <div class="flex items-center gap-3">
          <div class="w-10 h-10 bg-primary-600 rounded-lg flex items-center justify-center">
            <svg class="w-6 h-6 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" />
            </svg>
          </div>
          <h1 class="text-xl font-bold text-white">Probe System</h1>
        </div>
        
        <div class="flex items-center gap-4">
          <div class="flex items-center gap-2">
            <span :class="['status-dot', allOnline ? 'status-online animate-pulse-glow' : 'status-offline']"></span>
            <span class="text-sm text-gray-300">
              {{ allOnline ? 'All Systems Operational' : `${offlineCount} System(s) Down` }}
            </span>
          </div>
          
          <div class="text-sm text-gray-400">
            <span class="text-green-400">{{ onlineCount }}</span> Online · 
            <span class="text-red-400">{{ offlineCount }}</span> Offline
          </div>
        </div>
      </div>
    </header>

    <!-- Main Content -->
    <main class="max-w-7xl mx-auto px-4 py-8">
      <div v-if="store.loading" class="flex items-center justify-center py-20">
        <div class="animate-spin rounded-full h-12 w-12 border-b-2 border-primary-500"></div>
      </div>
      
      <div v-else-if="store.agents.length === 0" class="text-center py-20">
        <p class="text-gray-400">No servers available</p>
      </div>
      
      <div v-else class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-4">
        <AgentCard 
          v-for="agent in store.agents" 
          :key="agent.id" 
          :agent="agent" 
        />
      </div>
    </main>

    <!-- Footer -->
    <footer class="border-t border-gray-800 py-6 mt-auto">
      <div class="max-w-7xl mx-auto px-4 text-center text-sm text-gray-500">
        Powered by Probe System · Real-time monitoring
      </div>
    </footer>
  </div>
</template>
