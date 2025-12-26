<script setup lang="ts">
import { ref, onMounted } from 'vue'
import api from '../../api'

const settings = ref({
  data_retention_days: 7,
  telegram_bot_token: '',
  telegram_chat_id: '',
  smtp_host: '',
  smtp_port: 587,
  smtp_username: '',
  smtp_password: '',
  smtp_from: '',
  alert_email_to: ''
})

const loading = ref(true)
const saving = ref(false)
const message = ref('')

onMounted(async () => {
  try {
    const res = await api.get('/api/admin/settings')
    settings.value = { ...settings.value, ...res.data }
  } finally {
    loading.value = false
  }
})

async function saveSettings() {
  saving.value = true
  message.value = ''
  
  try {
    await api.put('/api/admin/settings', settings.value)
    message.value = 'Settings saved successfully'
  } catch (e: any) {
    message.value = e.response?.data?.error || 'Failed to save settings'
  } finally {
    saving.value = false
  }
}
</script>

<template>
  <div>
    <h1 class="text-2xl font-bold text-white mb-6">Settings</h1>

    <div v-if="loading" class="py-8 text-center text-gray-400">
      Loading...
    </div>

    <form v-else @submit.prevent="saveSettings" class="space-y-6">
      <!-- General -->
      <div class="card">
        <h2 class="text-lg font-semibold text-white mb-4">General</h2>
        <div>
          <label class="block text-sm text-gray-400 mb-1">Data Retention (days)</label>
          <input 
            v-model.number="settings.data_retention_days" 
            type="number" 
            class="input w-full max-w-xs" 
            min="1" 
            max="365"
          />
          <p class="text-xs text-gray-500 mt-1">Metrics older than this will be automatically deleted</p>
        </div>
      </div>

      <!-- Telegram -->
      <div class="card">
        <h2 class="text-lg font-semibold text-white mb-4">Telegram Notifications</h2>
        <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
          <div>
            <label class="block text-sm text-gray-400 mb-1">Bot Token</label>
            <input 
              v-model="settings.telegram_bot_token" 
              type="password" 
              class="input w-full" 
              placeholder="123456:ABC-DEF..."
            />
          </div>
          <div>
            <label class="block text-sm text-gray-400 mb-1">Chat ID</label>
            <input 
              v-model="settings.telegram_chat_id" 
              type="text" 
              class="input w-full" 
              placeholder="-1001234567890"
            />
          </div>
        </div>
      </div>

      <!-- Email -->
      <div class="card">
        <h2 class="text-lg font-semibold text-white mb-4">Email Notifications</h2>
        <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
          <div>
            <label class="block text-sm text-gray-400 mb-1">SMTP Host</label>
            <input v-model="settings.smtp_host" type="text" class="input w-full" placeholder="smtp.example.com" />
          </div>
          <div>
            <label class="block text-sm text-gray-400 mb-1">SMTP Port</label>
            <input v-model.number="settings.smtp_port" type="number" class="input w-full" />
          </div>
          <div>
            <label class="block text-sm text-gray-400 mb-1">Username</label>
            <input v-model="settings.smtp_username" type="text" class="input w-full" />
          </div>
          <div>
            <label class="block text-sm text-gray-400 mb-1">Password</label>
            <input v-model="settings.smtp_password" type="password" class="input w-full" />
          </div>
          <div>
            <label class="block text-sm text-gray-400 mb-1">From Address</label>
            <input v-model="settings.smtp_from" type="email" class="input w-full" placeholder="alerts@example.com" />
          </div>
          <div>
            <label class="block text-sm text-gray-400 mb-1">Alert Recipient</label>
            <input v-model="settings.alert_email_to" type="email" class="input w-full" placeholder="admin@example.com" />
          </div>
        </div>
      </div>

      <!-- Save -->
      <div class="flex items-center gap-4">
        <button type="submit" class="btn btn-primary" :disabled="saving">
          {{ saving ? 'Saving...' : 'Save Settings' }}
        </button>
        <span v-if="message" :class="message.includes('success') ? 'text-green-400' : 'text-red-400'" class="text-sm">
          {{ message }}
        </span>
      </div>
    </form>
  </div>
</template>
