<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useRouter } from 'vue-router'
import api from '../../api'
import { MagnifyingGlassIcon, EyeIcon, EyeSlashIcon } from '@heroicons/vue/24/outline'

const router = useRouter()
const agents = ref<any[]>([])
const groups = ref<any[]>([])
const loading = ref(true)
const search = ref('')
const filterGroup = ref('')
const filterStatus = ref('')

const filteredAgents = computed(() => {
  return agents.value.filter(agent => {
    if (search.value) {
      const s = search.value.toLowerCase()
      if (!agent.custom_name?.toLowerCase().includes(s) && 
          !agent.hostname?.toLowerCase().includes(s) &&
          !agent.ip?.toLowerCase().includes(s)) {
        return false
      }
    }
    if (filterGroup.value && agent.group_id !== filterGroup.value) {
      return false
    }
    if (filterStatus.value && agent.status !== filterStatus.value) {
      return false
    }
    return true
  })
})

onMounted(async () => {
  try {
    const [agentsRes, groupsRes] = await Promise.all([
      api.get('/api/admin/agents'),
      api.get('/api/admin/groups')
    ])
    agents.value = agentsRes.data || []
    groups.value = groupsRes.data || []
  } finally {
    loading.value = false
  }
})

async function toggleVisibility(agent: any) {
  try {
    await api.patch(`/api/admin/agents/${agent.id}/visibility`, {
      visible: !agent.public_visible
    })
    agent.public_visible = !agent.public_visible
  } catch (e) {
    console.error(e)
  }
}

function viewAgent(id: string) {
  router.push(`/admin/agents/${id}`)
}

function getCountryFlag(code: string) {
  if (!code) return '🌐'
  const offset = 127397
  return String.fromCodePoint(...[...code.toUpperCase()].map(c => c.charCodeAt(0) + offset))
}
</script>

<template>
  <div>
    <div class="flex items-center justify-between mb-6">
      <h1 class="text-2xl font-bold text-white">Agents</h1>
    </div>

    <!-- Filters -->
    <div class="card mb-6">
      <div class="flex flex-wrap gap-4">
        <div class="flex-1 min-w-[200px]">
          <div class="relative">
            <MagnifyingGlassIcon class="w-5 h-5 absolute left-3 top-1/2 -translate-y-1/2 text-gray-400" />
            <input 
              v-model="search"
              type="text" 
              placeholder="Search agents..." 
              class="input w-full pl-10"
            />
          </div>
        </div>
        
        <select v-model="filterGroup" class="input">
          <option value="">All Groups</option>
          <option v-for="g in groups" :key="g.id" :value="g.id">{{ g.name }}</option>
        </select>
        
        <select v-model="filterStatus" class="input">
          <option value="">All Status</option>
          <option value="online">Online</option>
          <option value="offline">Offline</option>
        </select>
      </div>
    </div>

    <!-- Agents Table -->
    <div class="card overflow-hidden">
      <div v-if="loading" class="py-8 text-center text-gray-400">
        Loading...
      </div>
      
      <div v-else-if="filteredAgents.length === 0" class="py-8 text-center text-gray-400">
        No agents found
      </div>
      
      <table v-else class="w-full">
        <thead class="bg-gray-700/50">
          <tr>
            <th class="px-4 py-3 text-left text-xs font-medium text-gray-400 uppercase">Agent</th>
            <th class="px-4 py-3 text-left text-xs font-medium text-gray-400 uppercase">IP / Location</th>
            <th class="px-4 py-3 text-left text-xs font-medium text-gray-400 uppercase">Status</th>
            <th class="px-4 py-3 text-left text-xs font-medium text-gray-400 uppercase">Group</th>
            <th class="px-4 py-3 text-left text-xs font-medium text-gray-400 uppercase">Public</th>
            <th class="px-4 py-3 text-right text-xs font-medium text-gray-400 uppercase">Actions</th>
          </tr>
        </thead>
        <tbody class="divide-y divide-gray-700">
          <tr 
            v-for="agent in filteredAgents" 
            :key="agent.id"
            class="hover:bg-gray-700/30 cursor-pointer"
            @click="viewAgent(agent.id)"
          >
            <td class="px-4 py-3">
              <div>
                <p class="font-medium text-white">{{ agent.custom_name || agent.hostname }}</p>
                <p class="text-xs text-gray-500">{{ agent.hostname }}</p>
              </div>
            </td>
            <td class="px-4 py-3">
              <div class="flex items-center gap-2">
                <span>{{ getCountryFlag(agent.location?.country_code) }}</span>
                <div>
                  <p class="text-sm text-white">{{ agent.ip }}</p>
                  <p class="text-xs text-gray-500">{{ agent.location?.city || agent.location?.country || 'Unknown' }}</p>
                </div>
              </div>
            </td>
            <td class="px-4 py-3">
              <span :class="[
                'inline-flex items-center gap-1.5 px-2 py-1 rounded-full text-xs font-medium',
                agent.status === 'online' ? 'bg-green-500/20 text-green-400' : 'bg-red-500/20 text-red-400'
              ]">
                <span :class="['w-1.5 h-1.5 rounded-full', agent.status === 'online' ? 'bg-green-400' : 'bg-red-400']"></span>
                {{ agent.status }}
              </span>
            </td>
            <td class="px-4 py-3 text-sm text-gray-300">
              {{ groups.find(g => g.id === agent.group_id)?.name || '-' }}
            </td>
            <td class="px-4 py-3">
              <button 
                @click.stop="toggleVisibility(agent)"
                :class="[
                  'p-1.5 rounded-lg transition-colors',
                  agent.public_visible ? 'text-green-400 hover:bg-green-500/20' : 'text-gray-500 hover:bg-gray-700'
                ]"
              >
                <EyeIcon v-if="agent.public_visible" class="w-5 h-5" />
                <EyeSlashIcon v-else class="w-5 h-5" />
              </button>
            </td>
            <td class="px-4 py-3 text-right">
              <button class="text-primary-400 hover:text-primary-300 text-sm">
                View
              </button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>
