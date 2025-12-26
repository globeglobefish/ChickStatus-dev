<script setup lang="ts">
import { ref } from 'vue'
import { useRouter, RouterLink, RouterView } from 'vue-router'
import { 
  HomeIcon, 
  ServerStackIcon, 
  ClipboardDocumentListIcon,
  BellAlertIcon,
  Cog6ToothIcon,
  ArrowRightOnRectangleIcon,
  Bars3Icon
} from '@heroicons/vue/24/outline'

const router = useRouter()
const sidebarOpen = ref(false)

const navigation = [
  { name: 'Dashboard', href: '/admin', icon: HomeIcon },
  { name: 'Agents', href: '/admin/agents', icon: ServerStackIcon },
  { name: 'Tasks', href: '/admin/tasks', icon: ClipboardDocumentListIcon },
  { name: 'Alerts', href: '/admin/alerts', icon: BellAlertIcon },
  { name: 'Settings', href: '/admin/settings', icon: Cog6ToothIcon },
]

function logout() {
  localStorage.removeItem('token')
  router.push('/login')
}
</script>

<template>
  <div class="min-h-screen bg-gray-900 flex">
    <!-- Sidebar -->
    <aside 
      :class="[
        'fixed inset-y-0 left-0 z-50 w-64 bg-gray-800 border-r border-gray-700 transform transition-transform duration-200 lg:translate-x-0 lg:static',
        sidebarOpen ? 'translate-x-0' : '-translate-x-full'
      ]"
    >
      <div class="flex items-center gap-3 px-6 py-5 border-b border-gray-700">
        <div class="w-10 h-10 bg-primary-600 rounded-lg flex items-center justify-center">
          <svg class="w-6 h-6 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" />
          </svg>
        </div>
        <span class="text-lg font-bold text-white">Probe Admin</span>
      </div>

      <nav class="px-4 py-4 space-y-1">
        <RouterLink
          v-for="item in navigation"
          :key="item.name"
          :to="item.href"
          class="flex items-center gap-3 px-3 py-2.5 rounded-lg text-gray-300 hover:bg-gray-700 hover:text-white transition-colors"
          active-class="bg-primary-600/20 text-primary-400 hover:bg-primary-600/30 hover:text-primary-300"
          @click="sidebarOpen = false"
        >
          <component :is="item.icon" class="w-5 h-5" />
          {{ item.name }}
        </RouterLink>
      </nav>

      <div class="absolute bottom-0 left-0 right-0 p-4 border-t border-gray-700">
        <button 
          @click="logout"
          class="flex items-center gap-3 px-3 py-2.5 w-full rounded-lg text-gray-300 hover:bg-gray-700 hover:text-white transition-colors"
        >
          <ArrowRightOnRectangleIcon class="w-5 h-5" />
          Logout
        </button>
      </div>
    </aside>

    <!-- Overlay -->
    <div 
      v-if="sidebarOpen" 
      class="fixed inset-0 bg-black/50 z-40 lg:hidden"
      @click="sidebarOpen = false"
    ></div>

    <!-- Main Content -->
    <div class="flex-1 flex flex-col min-w-0">
      <!-- Top Bar -->
      <header class="bg-gray-800 border-b border-gray-700 px-4 py-3 flex items-center gap-4 lg:hidden">
        <button @click="sidebarOpen = true" class="text-gray-400 hover:text-white">
          <Bars3Icon class="w-6 h-6" />
        </button>
        <span class="font-semibold text-white">Probe Admin</span>
      </header>

      <!-- Page Content -->
      <main class="flex-1 p-6 overflow-auto">
        <RouterView />
      </main>
    </div>
  </div>
</template>
