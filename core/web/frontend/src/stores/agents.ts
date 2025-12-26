import { defineStore } from 'pinia'
import { ref } from 'vue'
import api from '../api'

export interface Agent {
  id: string
  name: string
  status: 'online' | 'offline'
  group_id?: string
  location?: {
    country: string
    country_code: string
    region?: string
  }
  metrics?: {
    cpu: number
    memory_percent: number
    disk_percent: number
  }
  traffic?: {
    used: number
    limit: number
    percent: number
  }
}

export const useAgentsStore = defineStore('agents', () => {
  const agents = ref<Agent[]>([])
  const loading = ref(false)
  const ws = ref<WebSocket | null>(null)

  async function fetchPublicAgents() {
    loading.value = true
    try {
      const response = await api.get('/api/public/agents')
      agents.value = response.data
    } finally {
      loading.value = false
    }
  }

  function connectWebSocket() {
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
    const wsUrl = `${protocol}//${window.location.host}/ws/dashboard`
    
    ws.value = new WebSocket(wsUrl)
    
    ws.value.onmessage = (event) => {
      const data = JSON.parse(event.data)
      
      if (data.type === 'init') {
        agents.value = data.data
      } else if (data.type === 'update') {
        for (const update of data.data) {
          const agent = agents.value.find(a => a.id === update.id)
          if (agent) {
            agent.status = update.status
            if (update.metrics) {
              agent.metrics = update.metrics
            }
            if (update.traffic) {
              agent.traffic = update.traffic
            }
          }
        }
      }
    }
    
    ws.value.onclose = () => {
      setTimeout(connectWebSocket, 3000)
    }
  }

  function disconnectWebSocket() {
    if (ws.value) {
      ws.value.close()
      ws.value = null
    }
  }

  return {
    agents,
    loading,
    fetchPublicAgents,
    connectWebSocket,
    disconnectWebSocket
  }
})
