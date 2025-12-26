<script setup lang="ts">
import { ref, onMounted } from 'vue'
import api from '../../api'
import { PlusIcon } from '@heroicons/vue/24/outline'

const tasks = ref<any[]>([])
const loading = ref(true)
const showModal = ref(false)

const newTask = ref({
  type: 'ping',
  name: '',
  target: '',
  interval: 60,
  agent_ids: [] as string[]
})

const agents = ref<any[]>([])

onMounted(async () => {
  try {
    const [tasksRes, agentsRes] = await Promise.all([
      api.get('/api/admin/tasks'),
      api.get('/api/admin/agents')
    ])
    tasks.value = tasksRes.data || []
    agents.value = agentsRes.data || []
  } finally {
    loading.value = false
  }
})

async function createTask() {
  try {
    await api.post('/api/admin/tasks', newTask.value)
    showModal.value = false
    const res = await api.get('/api/admin/tasks')
    tasks.value = res.data || []
    newTask.value = { type: 'ping', name: '', target: '', interval: 60, agent_ids: [] }
  } catch (e) {
    console.error(e)
  }
}

async function cancelTask(id: string) {
  try {
    await api.post(`/api/admin/tasks/${id}/cancel`)
    const task = tasks.value.find(t => t.id === id)
    if (task) task.status = 'canceled'
  } catch (e) {
    console.error(e)
  }
}
</script>

<template>
  <div>
    <div class="flex items-center justify-between mb-6">
      <h1 class="text-2xl font-bold text-white">Tasks</h1>
      <button @click="showModal = true" class="btn btn-primary flex items-center gap-2">
        <PlusIcon class="w-5 h-5" />
        New Task
      </button>
    </div>

    <div class="card">
      <div v-if="loading" class="py-8 text-center text-gray-400">
        Loading...
      </div>
      
      <div v-else-if="tasks.length === 0" class="py-8 text-center text-gray-400">
        No tasks created
      </div>
      
      <table v-else class="w-full">
        <thead class="bg-gray-700/50">
          <tr>
            <th class="px-4 py-3 text-left text-xs font-medium text-gray-400 uppercase">Task</th>
            <th class="px-4 py-3 text-left text-xs font-medium text-gray-400 uppercase">Type</th>
            <th class="px-4 py-3 text-left text-xs font-medium text-gray-400 uppercase">Target</th>
            <th class="px-4 py-3 text-left text-xs font-medium text-gray-400 uppercase">Interval</th>
            <th class="px-4 py-3 text-left text-xs font-medium text-gray-400 uppercase">Status</th>
            <th class="px-4 py-3 text-right text-xs font-medium text-gray-400 uppercase">Actions</th>
          </tr>
        </thead>
        <tbody class="divide-y divide-gray-700">
          <tr v-for="task in tasks" :key="task.id" class="hover:bg-gray-700/30">
            <td class="px-4 py-3">
              <p class="font-medium text-white">{{ task.name || task.id.slice(0, 8) }}</p>
            </td>
            <td class="px-4 py-3">
              <span class="px-2 py-1 bg-gray-700 rounded text-xs text-gray-300">{{ task.type }}</span>
            </td>
            <td class="px-4 py-3 text-sm text-gray-300">{{ task.target || '-' }}</td>
            <td class="px-4 py-3 text-sm text-gray-300">
              {{ task.interval ? `${task.interval}s` : 'One-time' }}
            </td>
            <td class="px-4 py-3">
              <span :class="[
                'px-2 py-1 rounded text-xs',
                task.status === 'running' ? 'bg-green-500/20 text-green-400' :
                task.status === 'pending' ? 'bg-yellow-500/20 text-yellow-400' :
                task.status === 'canceled' ? 'bg-gray-500/20 text-gray-400' :
                'bg-blue-500/20 text-blue-400'
              ]">
                {{ task.status }}
              </span>
            </td>
            <td class="px-4 py-3 text-right">
              <button 
                v-if="task.status === 'running' || task.status === 'pending'"
                @click="cancelTask(task.id)"
                class="text-red-400 hover:text-red-300 text-sm"
              >
                Cancel
              </button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- Create Task Modal -->
    <div v-if="showModal" class="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
      <div class="card w-full max-w-md mx-4">
        <h2 class="text-lg font-semibold text-white mb-4">Create Task</h2>
        
        <form @submit.prevent="createTask" class="space-y-4">
          <div>
            <label class="block text-sm text-gray-400 mb-1">Type</label>
            <select v-model="newTask.type" class="input w-full">
              <option value="ping">Ping</option>
              <option value="script">Script</option>
            </select>
          </div>
          
          <div>
            <label class="block text-sm text-gray-400 mb-1">Name</label>
            <input v-model="newTask.name" type="text" class="input w-full" placeholder="Task name" />
          </div>
          
          <div>
            <label class="block text-sm text-gray-400 mb-1">Target</label>
            <input v-model="newTask.target" type="text" class="input w-full" placeholder="e.g., google.com:80" />
          </div>
          
          <div>
            <label class="block text-sm text-gray-400 mb-1">Interval (seconds, 0 for one-time)</label>
            <input v-model.number="newTask.interval" type="number" class="input w-full" min="0" />
          </div>
          
          <div>
            <label class="block text-sm text-gray-400 mb-1">Agents</label>
            <select v-model="newTask.agent_ids" multiple class="input w-full h-32">
              <option v-for="agent in agents" :key="agent.id" :value="agent.id">
                {{ agent.custom_name || agent.hostname }}
              </option>
            </select>
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
