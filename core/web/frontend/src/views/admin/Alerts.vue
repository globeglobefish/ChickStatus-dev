<script setup lang="ts">
import { ref, onMounted } from 'vue'
import api from '../../api'
import { PlusIcon } from '@heroicons/vue/24/outline'

const rules = ref<any[]>([])
const alerts = ref<any[]>([])
const loading = ref(true)
const showModal = ref(false)

const newRule = ref({
  name: '',
  metric_type: 'cpu',
  operator: 'gt',
  threshold: 80,
  duration: 60,
  cooldown: 300,
  enabled: true,
  agent_ids: [] as string[]
})

const agents = ref<any[]>([])

onMounted(async () => {
  try {
    const [rulesRes, alertsRes, agentsRes] = await Promise.all([
      api.get('/api/admin/alerts/rules'),
      api.get('/api/admin/alerts/history?limit=50'),
      api.get('/api/admin/agents')
    ])
    rules.value = rulesRes.data || []
    alerts.value = alertsRes.data || []
    agents.value = agentsRes.data || []
  } finally {
    loading.value = false
  }
})

async function createRule() {
  try {
    await api.post('/api/admin/alerts/rules', newRule.value)
    showModal.value = false
    const res = await api.get('/api/admin/alerts/rules')
    rules.value = res.data || []
    newRule.value = { name: '', metric_type: 'cpu', operator: 'gt', threshold: 80, duration: 60, cooldown: 300, enabled: true, agent_ids: [] }
  } catch (e) {
    console.error(e)
  }
}

async function deleteRule(id: string) {
  if (!confirm('Delete this rule?')) return
  try {
    await api.delete(`/api/admin/alerts/rules/${id}`)
    rules.value = rules.value.filter(r => r.id !== id)
  } catch (e) {
    console.error(e)
  }
}
</script>

<template>
  <div>
    <div class="flex items-center justify-between mb-6">
      <h1 class="text-2xl font-bold text-white">Alerts</h1>
      <button @click="showModal = true" class="btn btn-primary flex items-center gap-2">
        <PlusIcon class="w-5 h-5" />
        New Rule
      </button>
    </div>

    <!-- Alert Rules -->
    <div class="card mb-6">
      <h2 class="text-lg font-semibold text-white mb-4">Alert Rules</h2>
      
      <div v-if="loading" class="py-4 text-center text-gray-400">Loading...</div>
      
      <div v-else-if="rules.length === 0" class="py-4 text-center text-gray-400">
        No alert rules configured
      </div>
      
      <div v-else class="space-y-3">
        <div 
          v-for="rule in rules" 
          :key="rule.id"
          class="flex items-center justify-between p-3 bg-gray-700/50 rounded-lg"
        >
          <div class="flex items-center gap-4">
            <div :class="['w-2 h-2 rounded-full', rule.enabled ? 'bg-green-500' : 'bg-gray-500']"></div>
            <div>
              <p class="font-medium text-white">{{ rule.name }}</p>
              <p class="text-xs text-gray-400">
                {{ rule.metric_type }} {{ rule.operator === 'gt' ? '>' : rule.operator === 'lt' ? '<' : '=' }} {{ rule.threshold }}%
              </p>
            </div>
          </div>
          <button @click="deleteRule(rule.id)" class="text-red-400 hover:text-red-300 text-sm">
            Delete
          </button>
        </div>
      </div>
    </div>

    <!-- Alert History -->
    <div class="card">
      <h2 class="text-lg font-semibold text-white mb-4">Alert History</h2>
      
      <div v-if="alerts.length === 0" class="py-4 text-center text-gray-400">
        No alerts
      </div>
      
      <div v-else class="space-y-2">
        <div 
          v-for="alert in alerts" 
          :key="alert.id"
          class="flex items-center gap-4 p-3 bg-gray-700/30 rounded-lg"
        >
          <div :class="['w-2 h-2 rounded-full', alert.status === 'firing' ? 'bg-red-500' : 'bg-green-500']"></div>
          <div class="flex-1 min-w-0">
            <p class="text-sm font-medium text-white">{{ alert.rule_name }}</p>
            <p class="text-xs text-gray-400">{{ alert.agent_name }} · {{ alert.metric_type }}: {{ alert.value.toFixed(1) }}%</p>
          </div>
          <div class="text-right">
            <span :class="['text-xs px-2 py-1 rounded', alert.status === 'firing' ? 'bg-red-500/20 text-red-400' : 'bg-green-500/20 text-green-400']">
              {{ alert.status }}
            </span>
            <p class="text-xs text-gray-500 mt-1">{{ new Date(alert.triggered_at).toLocaleString() }}</p>
          </div>
        </div>
      </div>
    </div>

    <!-- Create Rule Modal -->
    <div v-if="showModal" class="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
      <div class="card w-full max-w-md mx-4">
        <h2 class="text-lg font-semibold text-white mb-4">Create Alert Rule</h2>
        
        <form @submit.prevent="createRule" class="space-y-4">
          <div>
            <label class="block text-sm text-gray-400 mb-1">Name</label>
            <input v-model="newRule.name" type="text" class="input w-full" placeholder="High CPU Alert" required />
          </div>
          
          <div class="grid grid-cols-3 gap-3">
            <div>
              <label class="block text-sm text-gray-400 mb-1">Metric</label>
              <select v-model="newRule.metric_type" class="input w-full">
                <option value="cpu">CPU</option>
                <option value="memory">Memory</option>
                <option value="disk">Disk</option>
              </select>
            </div>
            <div>
              <label class="block text-sm text-gray-400 mb-1">Operator</label>
              <select v-model="newRule.operator" class="input w-full">
                <option value="gt">></option>
                <option value="lt"><</option>
                <option value="eq">=</option>
              </select>
            </div>
            <div>
              <label class="block text-sm text-gray-400 mb-1">Threshold %</label>
              <input v-model.number="newRule.threshold" type="number" class="input w-full" min="0" max="100" />
            </div>
          </div>
          
          <div class="grid grid-cols-2 gap-3">
            <div>
              <label class="block text-sm text-gray-400 mb-1">Duration (sec)</label>
              <input v-model.number="newRule.duration" type="number" class="input w-full" min="0" />
            </div>
            <div>
              <label class="block text-sm text-gray-400 mb-1">Cooldown (sec)</label>
              <input v-model.number="newRule.cooldown" type="number" class="input w-full" min="0" />
            </div>
          </div>
          
          <div class="flex gap-3 pt-2">
            <button type="button" @click="showModal = false" class="btn flex-1 bg-gray-700 hover:bg-gray-600 text-white">
              Cancel
            </button>
            <button type="submit" class="btn btn-primary flex-1">
              Create
            </button>
          </div>
        </form>
      </div>
    </div>
  </div>
</template>
