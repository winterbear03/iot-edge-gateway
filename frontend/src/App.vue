<template>
  <div class="layout">
    <!-- ====== Sidebar ====== -->
    <aside class="sidebar">
      <div class="sidebar-brand">
        <div class="brand-icon">
          <svg width="28" height="28" viewBox="0 0 28 28" fill="none">
            <rect x="2" y="2" width="24" height="24" rx="4" stroke="var(--accent)" stroke-width="1.5" />
            <circle cx="14" cy="14" r="5" stroke="var(--accent)" stroke-width="1.5" />
            <path d="M14 9v10M9 14h10" stroke="var(--accent)" stroke-width="1" />
          </svg>
        </div>
        <div class="brand-text">
          <span class="brand-title">IoT Gateway</span>
          <span class="brand-sub">边缘网关管理</span>
        </div>
      </div>

      <nav class="sidebar-nav">
        <a v-for="item in navItems" :key="item.id" :class="['nav-item', { active: activeNav === item.id }]"
          @click="activeNav = item.id">
          <span class="nav-icon" v-html="item.icon"></span>
          <span class="nav-label">{{ item.label }}</span>
          <span v-if="item.badge" class="nav-badge">{{ item.badge }}</span>
        </a>
      </nav>

      <div class="sidebar-footer">
        <div class="version-info">
          <span class="ver-label">版本</span>
          <span class="ver-value">v0.1.0</span>
        </div>
      </div>
    </aside>

    <!-- ====== Main ====== -->
    <main class="main bg-grid">
      <!-- Header -->
      <header class="topbar">
        <div class="topbar-left">
          <span class="topbar-title">{{ navItems.find(n => n.id === activeNav)?.label || '仪表盘' }}</span>
        </div>
        <div class="topbar-right">
          <div class="status-chip" :class="healthStatus">
            <span class="status-dot"></span>
            {{ healthStatus === 'online' ? '系统在线' : healthStatus === 'checking' ? '检测中' : '系统离线' }}
          </div>
          <div class="topbar-time">{{ currentTime }}</div>
        </div>
      </header>

      <!-- Dashboard content -->
      <div class="content" v-show="activeNav === 'dashboard'">
        <!-- Row 1: Health + Metrics -->
        <div class="row cols-4">
          <!-- System Health -->
          <div class="card card-health">
            <div class="card-head">
              <span class="card-title">系统健康</span>
              <span class="card-badge" :class="healthStatus">{{ healthStatus === 'online' ? '正常' : healthStatus ===
                'checking' ? '检测' : '异常' }}</span>
            </div>
            <div class="health-grid">
              <div class="health-item">
                <span class="health-label">网关服务</span>
                <span class="health-val ok">运行中</span>
              </div>
              <div class="health-item">
                <span class="health-label">数据采集</span>
                <span class="health-val ok">正常</span>
              </div>
              <div class="health-item">
                <span class="health-label">MQTT 连接</span>
                <span class="health-val ok">已连接</span>
              </div>
              <div class="health-item">
                <span class="health-label">API 端点</span>
                <span class="health-val" :class="healthStatus">{{ healthStatus === 'online' ? '可达' : '不可达' }}</span>
              </div>
              <div class="health-item">
                <span class="health-label">运行时间</span>
                <span class="health-val ok">{{ uptime }}</span>
              </div>
              <div class="health-item">
                <span class="health-label">设备总数</span>
                <span class="health-val ok">{{ devices.length }}</span>
              </div>
            </div>
          </div>

          <!-- CPU -->
          <div class="card card-metric">
            <div class="card-head">
              <span class="card-title">CPU 利用率</span>
              <span class="card-mono">{{ cpu }}%</span>
            </div>
            <div class="gauge">
              <svg viewBox="0 0 120 120" class="gauge-svg">
                <circle cx="60" cy="60" r="50" fill="none" stroke="var(--border-default)" stroke-width="8" />
                <circle cx="60" cy="60" r="50" fill="none" :stroke="cpuColor" stroke-width="8" stroke-linecap="round"
                  :stroke-dasharray="cpu * 3.14 + ' 314'" transform="rotate(-90 60 60)" class="gauge-arc" />
              </svg>
              <div class="gauge-center">
                <span class="gauge-val">{{ cpu }}%</span>
                <span class="gauge-unit">CPU</span>
              </div>
            </div>
          </div>

          <!-- Memory -->
          <div class="card card-metric">
            <div class="card-head">
              <span class="card-title">内存占用</span>
              <span class="card-mono">{{ memUsed }} / {{ memTotal }} GB</span>
            </div>
            <div class="gauge">
              <svg viewBox="0 0 120 120" class="gauge-svg">
                <circle cx="60" cy="60" r="50" fill="none" stroke="var(--border-default)" stroke-width="8" />
                <circle cx="60" cy="60" r="50" fill="none" :stroke="memColor" stroke-width="8" stroke-linecap="round"
                  :stroke-dasharray="memPercent * 3.14 + ' 314'" transform="rotate(-90 60 60)" class="gauge-arc" />
              </svg>
              <div class="gauge-center">
                <span class="gauge-val">{{ memPercent }}%</span>
                <span class="gauge-unit">RAM</span>
              </div>
            </div>
          </div>

          <!-- Network -->
          <div class="card card-metric">
            <div class="card-head">
              <span class="card-title">网络吞吐</span>
            </div>
            <div class="net-stats">
              <div class="net-row">
                <span class="net-dir">↓ 下行</span>
                <span class="net-val">{{ netDown }} Mbps</span>
              </div>
              <div class="net-bar-track">
                <div class="net-bar down" :style="{ width: netDownPercent + '%' }"></div>
              </div>
              <div class="net-row">
                <span class="net-dir">↑ 上行</span>
                <span class="net-val">{{ netUp }} Mbps</span>
              </div>
              <div class="net-bar-track">
                <div class="net-bar up" :style="{ width: netUpPercent + '%' }"></div>
              </div>
            </div>
          </div>
        </div>

        <!-- Row 2: Devices + Protocol traffic -->
        <div class="row cols-2">
          <!-- Device list -->
          <div class="card">
            <div class="card-head">
              <span class="card-title">设备列表</span>
              <span class="card-mono">{{ devices.length }} 台</span>
            </div>
            <div class="table-wrap">
              <table class="data-table">
                <thead>
                  <tr>
                    <th>设备 ID</th>
                    <th>协议</th>
                    <th>地址</th>
                    <th>状态</th>
                    <th>数据点</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="d in devices" :key="d.id">
                    <td class="cell-id">{{ d.id }}</td>
                    <td><span class="tag-protocol" :class="d.protocolClass">{{ d.protocol }}</span></td>
                    <td class="cell-mono">{{ d.address }}</td>
                    <td><span class="tag-status" :class="d.online ? 'online' : 'offline'">{{ d.online ? '在线' : '离线'
                        }}</span></td>
                    <td class="cell-mono">{{ d.dataPoints }}</td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>

          <!-- Protocol traffic -->
          <div class="card">
            <div class="card-head">
              <span class="card-title">协议数据流</span>
              <span class="card-mono">今日</span>
            </div>
            <div class="proto-list">
              <div v-for="p in protocolStats" :key="p.name" class="proto-item">
                <div class="proto-head">
                  <span class="proto-name">{{ p.name }}</span>
                  <span class="proto-msg">{{ p.messages }} 条消息</span>
                </div>
                <div class="proto-bar-track">
                  <div class="proto-bar" :style="{ width: p.percent + '%', background: p.color }"></div>
                </div>
                <div class="proto-foot">
                  <span class="proto-bps">{{ p.bps }} msg/s</span>
                  <span class="proto-err" v-if="p.errors">错误: {{ p.errors }}</span>
                </div>
              </div>
            </div>
          </div>
        </div>

        <!-- Row 3: Event log + Quick actions -->
        <div class="row cols-2">
          <!-- Event log -->
          <div class="card">
            <div class="card-head">
              <span class="card-title">事件日志</span>
              <span class="card-mono">{{ events.length }} 条</span>
            </div>
            <div class="event-list">
              <div v-for="evt in events" :key="evt.id" class="event-item">
                <span class="event-dot" :class="evt.level"></span>
                <span class="event-time">{{ evt.time }}</span>
                <span class="event-msg">{{ evt.msg }}</span>
              </div>
            </div>
          </div>

          <!-- Quick actions -->
          <div class="card">
            <div class="card-head">
              <span class="card-title">快捷操作</span>
            </div>
            <div class="actions-grid">
              <button class="action-btn" @click="checkHealth">
                <svg width="20" height="20" viewBox="0 0 20 20" fill="none"><path d="M10 4v2m0 8v2M4 10h2m8 0h2"
                    stroke="currentColor" stroke-width="1.5" stroke-linecap="round" /></svg>
                刷新状态
              </button>
              <button class="action-btn" @click="loadDevices">
                <svg width="20" height="20" viewBox="0 0 20 20" fill="none"><rect x="3" y="3" width="14" height="14" rx="2"
                    stroke="currentColor" stroke-width="1.5" /><path d="M7 7h6M7 10h6M7 13h4" stroke="currentColor"
                    stroke-width="1" stroke-linecap="round" /></svg>
                扫描设备
              </button>
              <button class="action-btn warning">
                <svg width="20" height="20" viewBox="0 0 20 20" fill="none"><circle cx="10" cy="10" r="7"
                    stroke="currentColor" stroke-width="1.5" /><path d="M10 6v5M10 13v1" stroke="currentColor"
                    stroke-width="1.5" stroke-linecap="round" /></svg>
                重启网关
              </button>
              <button class="action-btn">
                <svg width="20" height="20" viewBox="0 0 20 20" fill="none"><circle cx="10" cy="10" r="7"
                    stroke="currentColor" stroke-width="1.5" /><path d="M10 7v5M8 9l2-2 2 2" stroke="currentColor"
                    stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round" /></svg>
                导出配置
              </button>
            </div>
          </div>
        </div>
      </div>

      <!-- Devices page -->
      <div class="content" v-show="activeNav === 'devices'">
        <div class="card">
          <div class="card-head"><span class="card-title">设备管理</span></div>
          <div class="table-wrap">
            <table class="data-table">
              <thead>
                <tr>
                  <th>设备 ID</th>
                  <th>名称</th>
                  <th>协议</th>
                  <th>地址</th>
                  <th>状态</th>
                  <th>数据点</th>
                  <th>最后通信</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="d in devices" :key="d.id">
                  <td class="cell-id">{{ d.id }}</td>
                  <td>{{ d.name }}</td>
                  <td><span class="tag-protocol" :class="d.protocolClass">{{ d.protocol }}</span></td>
                  <td class="cell-mono">{{ d.address }}</td>
                  <td><span class="tag-status" :class="d.online ? 'online' : 'offline'">{{ d.online ? '在线' : '离线'
                      }}</span></td>
                  <td class="cell-mono">{{ d.dataPoints }}</td>
                  <td class="cell-mono">{{ d.lastSeen }}</td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>
      </div>

      <!-- Logs page -->
      <div class="content" v-show="activeNav === 'logs'">
        <div class="card">
          <div class="card-head"><span class="card-title">系统日志</span></div>
          <div class="event-list">
            <div v-for="evt in events" :key="evt.id" class="event-item">
              <span class="event-dot" :class="evt.level"></span>
              <span class="event-time">{{ evt.time }}</span>
              <span class="event-msg">{{ evt.msg }}</span>
            </div>
          </div>
        </div>
      </div>

      <!-- Settings page -->
      <div class="content" v-show="activeNav === 'settings'">
        <div class="card">
          <div class="card-head"><span class="card-title">系统设置</span></div>
          <div class="settings-list">
            <div class="setting-row">
              <div>
                <div class="setting-label">数据采集间隔</div>
                <div class="setting-desc">设备轮询时间间隔</div>
              </div>
              <span class="cell-mono">{{ config.pollInterval }}ms</span>
            </div>
            <div class="setting-row">
              <div>
                <div class="setting-label">数据上报地址</div>
                <div class="setting-desc">MQTT Broker 地址</div>
              </div>
              <span class="cell-mono">{{ config.mqttBroker }}</span>
            </div>
            <div class="setting-row">
              <div>
                <div class="setting-label">日志级别</div>
                <div class="setting-desc">当前日志输出级别</div>
              </div>
              <span class="tag-protocol modbus">{{ config.logLevel }}</span>
            </div>
            <div class="setting-row">
              <div>
                <div class="setting-label">数据保留天数</div>
                <div class="setting-desc">历史数据存储时长</div>
              </div>
              <span class="cell-mono">{{ config.retentionDays }} 天</span>
            </div>
          </div>
        </div>
      </div>
    </main>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted } from 'vue'

// ---- Navigation ----
const activeNav = ref('dashboard')
const navItems = [
  { id: 'dashboard', label: '仪表盘', icon: '<svg width="18" height="18" viewBox="0 0 20 20" fill="none"><rect x="2" y="2" width="7" height="7" rx="1.5" stroke="currentColor" stroke-width="1.5"/><rect x="11" y="2" width="7" height="7" rx="1.5" stroke="currentColor" stroke-width="1.5"/><rect x="2" y="11" width="7" height="7" rx="1.5" stroke="currentColor" stroke-width="1.5"/><rect x="11" y="11" width="7" height="7" rx="1.5" stroke="currentColor" stroke-width="1.5"/></svg>' },
  { id: 'devices', label: '设备管理', icon: '<svg width="18" height="18" viewBox="0 0 20 20" fill="none"><rect x="3" y="3" width="14" height="10" rx="2" stroke="currentColor" stroke-width="1.5"/><circle cx="10" cy="8" r="2" stroke="currentColor" stroke-width="1.5"/><path d="M6 17h8M10 13v4" stroke="currentColor" stroke-width="1.5" stroke-linecap="round"/></svg>' },
  { id: 'logs', label: '系统日志', icon: '<svg width="18" height="18" viewBox="0 0 20 20" fill="none"><rect x="3" y="3" width="14" height="14" rx="2" stroke="currentColor" stroke-width="1.5"/><path d="M7 7h6M7 10h6M7 13h4" stroke="currentColor" stroke-width="1.5" stroke-linecap="round"/></svg>', badge: '3' },
  { id: 'settings', label: '系统设置', icon: '<svg width="18" height="18" viewBox="0 0 20 20" fill="none"><circle cx="10" cy="10" r="3" stroke="currentColor" stroke-width="1.5"/><path d="M10 2v2m0 12v2M2 10h2m12 0h2M4.93 4.93l1.41 1.41m7.32 7.32l1.41 1.41M4.93 15.07l1.41-1.41m7.32-7.32l1.41-1.41" stroke="currentColor" stroke-width="1.5" stroke-linecap="round"/></svg>' },
]

// ---- Clock ----
const currentTime = ref('')
let clockTimer = null
function updateClock() {
  const now = new Date()
  currentTime.value = now.toLocaleString('zh-CN', { hour: '2-digit', minute: '2-digit', second: '2-digit', year: 'numeric', month: '2-digit', day: '2-digit' })
}

// ---- Health ----
const health = ref(null)
const healthStatus = computed(() => {
  if (health.value === null) return 'checking'
  return health.value.status === 'UP' ? 'online' : 'offline'
})
const uptime = ref('--')
async function checkHealth() {
  try {
    const resp = await fetch('/api/health')
    health.value = await resp.json()
    if (health.value?.uptime) {
      uptime.value = health.value.uptime
    }
  } catch {
    health.value = { status: 'DOWN' }
  }
}

// ---- Simulated metrics ----
const cpu = ref(0)
const memPercent = ref(0)
const memUsed = ref(0)
const memTotal = ref(16)
const netDown = ref(0)
const netUp = ref(0)

const cpuColor = computed(() => cpu.value > 80 ? 'var(--accent-red)' : cpu.value > 50 ? 'var(--accent-amber)' : 'var(--accent)')
const memColor = computed(() => memPercent.value > 85 ? 'var(--accent-red)' : memPercent.value > 60 ? 'var(--accent-amber)' : 'var(--accent-green)')
const netDownPercent = computed(() => Math.min((netDown.value / 100) * 100, 100))
const netUpPercent = computed(() => Math.min((netUp.value / 50) * 100, 100))

function updateMetrics() {
  cpu.value = Math.floor(20 + Math.random() * 45)
  memPercent.value = Math.floor(35 + Math.random() * 30)
  memUsed.value = (memTotal.value * memPercent.value / 100).toFixed(1)
  netDown.value = (Math.random() * 60 + 10).toFixed(1)
  netUp.value = (Math.random() * 30 + 2).toFixed(1)
}

// ---- Devices ----
const devices = ref([])
function loadDevices() {
  devices.value = [
    { id: 'temp_sensor_01', name: '温度传感器 A', protocol: 'Modbus TCP', protocolClass: 'modbus', address: '192.168.1.101', online: true, dataPoints: 12, lastSeen: '3s 前' },
    { id: 'flow_meter_02', name: '流量计 B', protocol: 'Modbus RTU', protocolClass: 'rtu', address: '/dev/ttyS0', online: true, dataPoints: 8, lastSeen: '5s 前' },
    { id: 'press_sensor_03', name: '压力传感器 C', protocol: 'Modbus TCP', protocolClass: 'modbus', address: '192.168.1.102', online: true, dataPoints: 6, lastSeen: '12s 前' },
    { id: 'valve_ctrl_04', name: '阀门控制器 D', protocol: 'Modbus RTU', protocolClass: 'rtu', address: '/dev/ttyS1', online: false, dataPoints: 4, lastSeen: '2m 前' },
    { id: 'energy_meter_05', name: '电能表 E', protocol: 'DL/T645', protocolClass: 'dlt645', address: '192.168.1.201', online: true, dataPoints: 16, lastSeen: '1s 前' },
    { id: 'env_monitor_06', name: '环境监测仪 F', protocol: 'Modbus TCP', protocolClass: 'modbus', address: '192.168.1.103', online: true, dataPoints: 20, lastSeen: '8s 前' },
  ]
}

// ---- Protocol stats ----
const protocolStats = ref([
  { name: 'Modbus TCP', messages: 12480, bps: 23.5, errors: 2, percent: 65, color: 'var(--accent)' },
  { name: 'Modbus RTU', messages: 5420, bps: 10.2, errors: 0, percent: 28, color: 'var(--accent-green)' },
  { name: 'DL/T645', messages: 1340, bps: 2.5, errors: 1, percent: 7, color: 'var(--accent-amber)' },
])

// ---- Events ----
const events = ref([
  { id: 1, level: 'info', time: '14:32:05', msg: '设备 temp_sensor_01 数据采集成功' },
  { id: 2, level: 'warn', time: '14:31:48', msg: '设备 valve_ctrl_04 通信超时，正在重试' },
  { id: 3, level: 'info', time: '14:31:30', msg: 'MQTT 客户端已连接到 broker' },
  { id: 4, level: 'info', time: '14:31:15', msg: 'Modbus RTU 总线初始化完成' },
  { id: 5, level: 'error', time: '14:30:02', msg: '设备 valve_ctrl_04 连续 3 次无响应' },
  { id: 6, level: 'info', time: '14:29:00', msg: '网关服务启动完成，已加载 6 台设备' },
])

// ---- Config ----
const config = ref({
  pollInterval: 1000,
  mqttBroker: 'tcp://broker.local:1883',
  logLevel: 'INFO',
  retentionDays: 30,
})

// ---- Lifecycle ----
onMounted(() => {
  checkHealth()
  loadDevices()
  updateMetrics()
  updateClock()
  clockTimer = setInterval(updateClock, 1000)
  setInterval(checkHealth, 30000)
  setInterval(updateMetrics, 5000)
})

onUnmounted(() => {
  clearInterval(clockTimer)
})
</script>

<style scoped>
/* ====== Layout ====== */
.layout {
  display: flex;
  height: 100vh;
  overflow: hidden;
}

/* ====== Sidebar ====== */
.sidebar {
  width: var(--sidebar-w);
  background: var(--bg-base);
  border-right: 1px solid var(--border-subtle);
  display: flex;
  flex-direction: column;
  flex-shrink: 0;
}

.sidebar-brand {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 16px 18px;
  border-bottom: 1px solid var(--border-subtle);
}

.brand-title {
  display: block;
  font-size: 15px;
  font-weight: 600;
  color: var(--text-primary);
  letter-spacing: 0.5px;
}

.brand-sub {
  font-size: 11px;
  color: var(--text-dim);
  letter-spacing: 1px;
}

.sidebar-nav {
  flex: 1;
  padding: 12px 10px;
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.nav-item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 10px 12px;
  border-radius: 6px;
  color: var(--text-secondary);
  text-decoration: none;
  cursor: pointer;
  transition: all 0.15s;
  font-size: 13px;
  user-select: none;
}

.nav-item:hover {
  background: var(--bg-surface);
  color: var(--text-primary);
}

.nav-item.active {
  background: var(--accent-glow);
  color: var(--accent);
  box-shadow: inset 0 0 0 1px rgba(0, 212, 255, 0.25);
}

.nav-icon {
  display: flex;
  align-items: center;
  width: 18px;
  height: 18px;
  flex-shrink: 0;
}

.nav-badge {
  margin-left: auto;
  font-size: 10px;
  font-weight: 600;
  background: var(--accent-red);
  color: #fff;
  padding: 2px 6px;
  border-radius: 8px;
  min-width: 18px;
  text-align: center;
}

.sidebar-footer {
  padding: 12px 18px;
  border-top: 1px solid var(--border-subtle);
}

.version-info {
  display: flex;
  justify-content: space-between;
  font-size: 11px;
  color: var(--text-dim);
}

/* ====== Main ====== */
.main {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

/* ====== Topbar ====== */
.topbar {
  height: var(--header-h);
  background: var(--bg-header);
  border-bottom: 1px solid var(--border-subtle);
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 20px;
  flex-shrink: 0;
}

.topbar-title {
  font-size: 16px;
  font-weight: 600;
  color: var(--text-primary);
}

.topbar-right {
  display: flex;
  align-items: center;
  gap: 20px;
}

.topbar-time {
  font-family: var(--font-mono);
  font-size: 13px;
  color: var(--text-dim);
}

.status-chip {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 12px;
  font-weight: 500;
  padding: 4px 12px;
  border-radius: 12px;
  background: var(--bg-surface);
  border: 1px solid var(--border-default);
}

.status-chip.online {
  color: var(--accent-green);
  border-color: rgba(0, 200, 83, 0.3);
}

.status-chip.offline {
  color: var(--accent-red);
  border-color: rgba(255, 61, 79, 0.3);
}

.status-chip.checking {
  color: var(--accent-amber);
  border-color: rgba(240, 165, 0, 0.3);
}

.status-dot {
  width: 7px;
  height: 7px;
  border-radius: 50%;
  background: currentColor;
  box-shadow: 0 0 6px currentColor;
}

/* ====== Content ====== */
.content {
  flex: 1;
  overflow-y: auto;
  padding: 20px;
}

/* ====== Row/Col system ====== */
.row {
  display: grid;
  gap: 16px;
  margin-bottom: 16px;
}

.row.cols-4 {
  grid-template-columns: repeat(4, 1fr);
}

.row.cols-2 {
  grid-template-columns: repeat(2, 1fr);
}

@media (max-width: 1400px) {
  .row.cols-4 {
    grid-template-columns: repeat(2, 1fr);
  }
}

@media (max-width: 900px) {
  .row.cols-4,
  .row.cols-2 {
    grid-template-columns: 1fr;
  }
}

/* ====== Card ====== */
.card {
  background: var(--bg-surface);
  border: 1px solid var(--border-subtle);
  border-radius: 8px;
  padding: 18px;
  transition: border-color 0.2s;
}

.card:hover {
  border-color: var(--border-default);
}

.card-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 14px;
}

.card-title {
  font-size: 13px;
  font-weight: 600;
  color: var(--text-primary);
  letter-spacing: 0.5px;
  text-transform: uppercase;
}

.card-mono {
  font-family: var(--font-mono);
  font-size: 12px;
  color: var(--text-dim);
}

.card-badge {
  font-size: 11px;
  font-weight: 600;
  padding: 2px 8px;
  border-radius: 4px;
}

.card-badge.online {
  color: var(--accent-green);
  background: rgba(0, 200, 83, 0.12);
}

.card-badge.offline {
  color: var(--accent-red);
  background: rgba(255, 61, 79, 0.12);
}

.card-badge.checking {
  color: var(--accent-amber);
  background: rgba(240, 165, 0, 0.12);
}

/* ====== Health card ====== */
.health-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 10px;
}

.health-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px 10px;
  background: var(--bg-elevated);
  border-radius: 6px;
}

.health-label {
  font-size: 12px;
  color: var(--text-secondary);
}

.health-val {
  font-family: var(--font-mono);
  font-size: 12px;
  font-weight: 500;
}

.health-val.ok {
  color: var(--accent-green);
}

.health-val.online {
  color: var(--accent-green);
}

.health-val.offline {
  color: var(--accent-red);
}

/* ====== Gauge ====== */
.gauge {
  position: relative;
  width: 120px;
  height: 120px;
  margin: 0 auto;
}

.gauge-svg {
  width: 100%;
  height: 100%;
}

.gauge-arc {
  transition: stroke-dasharray 0.6s ease, stroke 0.6s ease;
}

.gauge-center {
  position: absolute;
  inset: 0;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
}

.gauge-val {
  font-family: var(--font-mono);
  font-size: 26px;
  font-weight: 600;
  color: var(--text-primary);
  line-height: 1;
}

.gauge-unit {
  font-size: 11px;
  color: var(--text-dim);
  margin-top: 2px;
  letter-spacing: 1px;
}

/* ====== Network stats ====== */
.net-stats {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.net-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.net-dir {
  font-size: 12px;
  color: var(--text-secondary);
}

.net-val {
  font-family: var(--font-mono);
  font-size: 13px;
  color: var(--text-primary);
  font-weight: 500;
}

.net-bar-track {
  height: 4px;
  background: var(--bg-deep);
  border-radius: 2px;
  overflow: hidden;
}

.net-bar {
  height: 100%;
  border-radius: 2px;
  transition: width 0.6s ease;
}

.net-bar.down {
  background: var(--accent);
  box-shadow: 0 0 6px var(--accent-glow);
}

.net-bar.up {
  background: var(--accent-green);
  box-shadow: 0 0 6px rgba(0, 200, 83, 0.3);
}

/* ====== Data table ====== */
.table-wrap {
  overflow-x: auto;
}

.data-table {
  width: 100%;
  border-collapse: collapse;
}

.data-table th {
  text-align: left;
  font-size: 11px;
  font-weight: 600;
  color: var(--text-dim);
  text-transform: uppercase;
  letter-spacing: 0.5px;
  padding: 8px 10px;
  border-bottom: 1px solid var(--border-subtle);
}

.data-table td {
  padding: 10px 10px;
  font-size: 13px;
  border-bottom: 1px solid rgba(26, 35, 50, 0.6);
  color: var(--text-secondary);
}

.data-table tbody tr:hover td {
  background: rgba(255, 255, 255, 0.02);
}

.cell-id {
  font-family: var(--font-mono);
  color: var(--accent);
  font-size: 12px;
}

.cell-mono {
  font-family: var(--font-mono);
  font-size: 12px;
}

/* Tags */
.tag-protocol {
  font-size: 11px;
  font-weight: 600;
  padding: 3px 8px;
  border-radius: 4px;
}

.tag-protocol.modbus {
  color: var(--accent);
  background: rgba(0, 212, 255, 0.1);
  border: 1px solid rgba(0, 212, 255, 0.2);
}

.tag-protocol.rtu {
  color: var(--accent-green);
  background: rgba(0, 200, 83, 0.1);
  border: 1px solid rgba(0, 200, 83, 0.2);
}

.tag-protocol.dlt645 {
  color: var(--accent-amber);
  background: rgba(240, 165, 0, 0.1);
  border: 1px solid rgba(240, 165, 0, 0.2);
}

.tag-status {
  font-size: 11px;
  font-weight: 600;
  padding: 3px 8px;
  border-radius: 4px;
}

.tag-status.online {
  color: var(--accent-green);
  background: rgba(0, 200, 83, 0.1);
}

.tag-status.offline {
  color: var(--accent-red);
  background: rgba(255, 61, 79, 0.1);
}

/* ====== Protocol traffic ====== */
.proto-list {
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.proto-head {
  display: flex;
  justify-content: space-between;
  margin-bottom: 4px;
}

.proto-name {
  font-size: 13px;
  font-weight: 500;
  color: var(--text-primary);
}

.proto-msg {
  font-family: var(--font-mono);
  font-size: 12px;
  color: var(--text-dim);
}

.proto-bar-track {
  height: 6px;
  background: var(--bg-deep);
  border-radius: 3px;
  overflow: hidden;
  margin-bottom: 4px;
}

.proto-bar {
  height: 100%;
  border-radius: 3px;
  transition: width 0.8s ease;
}

.proto-foot {
  display: flex;
  justify-content: space-between;
  font-size: 11px;
}

.proto-bps {
  color: var(--text-dim);
}

.proto-err {
  color: var(--accent-red);
}

/* ====== Events ====== */
.event-list {
  max-height: 200px;
  overflow-y: auto;
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.event-item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 7px 8px;
  border-radius: 4px;
  font-size: 12px;
}

.event-item:hover {
  background: var(--bg-elevated);
}

.event-dot {
  width: 7px;
  height: 7px;
  border-radius: 50%;
  flex-shrink: 0;
}

.event-dot.info {
  background: var(--accent);
  box-shadow: 0 0 4px var(--accent-glow);
}

.event-dot.warn {
  background: var(--accent-amber);
  box-shadow: 0 0 4px rgba(240, 165, 0, 0.4);
}

.event-dot.error {
  background: var(--accent-red);
  box-shadow: 0 0 4px rgba(255, 61, 79, 0.4);
}

.event-time {
  font-family: var(--font-mono);
  font-size: 11px;
  color: var(--text-dim);
  flex-shrink: 0;
  width: 60px;
}

.event-msg {
  color: var(--text-secondary);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

/* ====== Action buttons ====== */
.actions-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 10px;
}

.action-btn {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 14px 16px;
  background: var(--bg-elevated);
  border: 1px solid var(--border-default);
  border-radius: 6px;
  color: var(--text-secondary);
  font-size: 13px;
  cursor: pointer;
  transition: all 0.15s;
}

.action-btn:hover {
  border-color: var(--accent);
  color: var(--accent);
  background: var(--accent-glow);
}

.action-btn.warning:hover {
  border-color: var(--accent-red);
  color: var(--accent-red);
  background: rgba(255, 61, 79, 0.1);
}

/* ====== Settings ====== */
.settings-list {
  display: flex;
  flex-direction: column;
}

.setting-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 14px 0;
  border-bottom: 1px solid var(--border-subtle);
}

.setting-row:last-child {
  border-bottom: none;
}

.setting-label {
  font-size: 13px;
  color: var(--text-primary);
  font-weight: 500;
}

.setting-desc {
  font-size: 11px;
  color: var(--text-dim);
  margin-top: 2px;
}
</style>
