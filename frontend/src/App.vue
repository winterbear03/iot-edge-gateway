<template>
  <div class="container">
    <h1>工业边缘网关</h1>
    <div class="status-card">
      <h2>系统状态</h2>
      <p>状态：<span :class="statusClass">{{ statusText }}</span></p>
      <button @click="checkHealth">刷新状态</button>
    </div>
    <div class="device-card">
      <h2>设备列表</h2>
      <ul>
        <li v-for="device in devices" :key="device.id">
          {{ device.id }} ({{ device.protocol }}, 地址: {{ device.address }})
        </li>
      </ul>
    </div>
    <div class="info">
      <p>网关版本：v0.1.0 | 运行时间：刚刚</p>
    </div>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'

const health = ref(null)
const devices = ref([])

const statusText = computed(() => {
  if (health.value === null) return '检测中...'
  return health.value.status === 'UP' ? '在线' : '离线'
})

const statusClass = computed(() => {
  if (health.value === null) return 'loading'
  return health.value.status === 'UP' ? 'online' : 'offline'
})

async function checkHealth() {
  try {
    const resp = await fetch('/api/health')
    health.value = await resp.json()
  } catch {
    health.value = { status: 'DOWN' }
  }
}

function loadDevices() {
  // 模拟设备列表，后续可替换为真实 API 调用
  devices.value = [
    { id: 'temp_sensor_01', protocol: 'Modbus TCP', address: '192.168.1.101' },
    { id: 'flow_meter_02', protocol: 'Modbus RTU', address: '/dev/ttyS0' },
  ]
}

// 初始化
checkHealth()
loadDevices()
// 每30秒自动刷新状态
setInterval(checkHealth, 30000)
</script>

<style scoped>
.container {
  max-width: 800px;
  margin: 0 auto;
}
h1 {
  color: #1a1a2e;
  border-bottom: 2px solid #16213e;
  padding-bottom: 10px;
}
.status-card, .device-card {
  background: white;
  padding: 20px;
  margin: 15px 0;
  border-radius: 8px;
  box-shadow: 0 2px 4px rgba(0,0,0,0.1);
}
button {
  background: #0f3460;
  color: white;
  border: none;
  padding: 8px 16px;
  border-radius: 4px;
  cursor: pointer;
}
button:hover { background: #16213e; }
.online { color: green; font-weight: bold; }
.offline { color: red; font-weight: bold; }
.loading { color: gray; }
</style>
