<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import { ChatDotRound, Connection, DataLine, Refresh } from '@element-plus/icons-vue'
import { getIOTDashboard, type IOTClient, type IOTMessage } from '@/api/iot'

type TrendPoint = { label: string; count: number }
type TopicRank = { topic: string; count: number }

const loading = ref(false)
const refreshedAt = ref('')
const topicTotal = ref(0)
const messageTotal = ref(0)
const clientTotal = ref(0)
const onlineTotal = ref(0)
const recentMessages = ref<IOTMessage[]>([])
const recentClients = ref<IOTClient[]>([])
let refreshTimer: ReturnType<typeof setInterval> | undefined

const onlineRate = computed(() => clientTotal.value ? Math.round(onlineTotal.value / clientTotal.value * 100) : 0)
const trend = computed<TrendPoint[]>(() => {
  const now = new Date()
  now.setMinutes(0, 0, 0)
  const points = Array.from({ length: 12 }, (_, index) => {
    const time = new Date(now.getTime() - (11 - index) * 60 * 60 * 1000)
    return { time, label: `${String(time.getHours()).padStart(2, '0')}:00`, count: 0 }
  })
  for (const message of recentMessages.value) {
    const timestamp = Date.parse(message.received_at)
    const index = Math.floor((timestamp - points[0].time.getTime()) / 3_600_000)
    if (index >= 0 && index < points.length) points[index].count += 1
  }
  return points.map(({ label, count }) => ({ label, count }))
})
const trendMax = computed(() => Math.max(1, ...trend.value.map(item => item.count)))
const topicRanks = computed<TopicRank[]>(() => {
  const counts = new Map<string, number>()
  recentMessages.value.forEach(message => counts.set(message.topic, (counts.get(message.topic) || 0) + 1))
  return [...counts.entries()].map(([topic, count]) => ({ topic, count })).sort((a, b) => b.count - a.count).slice(0, 6)
})
const topicRankMax = computed(() => Math.max(1, ...topicRanks.value.map(item => item.count)))

async function load() {
  if (loading.value) return
  loading.value = true
  try {
    const response = await getIOTDashboard()
    const data = response.data.data
    topicTotal.value = data.topic_total
    messageTotal.value = data.message_total
    clientTotal.value = data.client_total
    onlineTotal.value = data.online_total
    recentMessages.value = data.recent_messages
    recentClients.value = data.recent_clients
    refreshedAt.value = new Date().toLocaleTimeString('zh-CN', { hour12: false })
  } finally {
    loading.value = false
  }
}

function formatTime(value: string) {
  return value ? value.replace('T', ' ').slice(0, 19) : '-'
}

function shortPayload(value: string) {
  return value.length > 72 ? `${value.slice(0, 72)}…` : value || '-'
}

onMounted(() => {
  void load()
  refreshTimer = setInterval(() => void load(), 5000)
})
onBeforeUnmount(() => {
  if (refreshTimer) clearInterval(refreshTimer)
})
</script>

<template>
  <section v-loading="loading" class="iot-dashboard">
    <header class="dashboard-header">
      <div>
        <h1>IOT 实时大屏</h1>
        <p>设备连接、Topic 与 MQTT 消息运行概览</p>
      </div>
      <div class="refresh-state"><span class="live-dot" />实时更新 · {{ refreshedAt || '--:--:--' }}<el-button text :icon="Refresh" @click="load">刷新</el-button></div>
    </header>

    <div class="metric-grid">
      <article class="metric-card"><el-icon><Connection /></el-icon><div><span>在线设备</span><strong>{{ onlineTotal }}</strong><small>共 {{ clientTotal }} 个客户端</small></div></article>
      <article class="metric-card"><el-icon><DataLine /></el-icon><div><span>在线率</span><strong>{{ onlineRate }}%</strong><small>{{ clientTotal - onlineTotal }} 个设备离线</small></div></article>
      <article class="metric-card"><el-icon><ChatDotRound /></el-icon><div><span>累计消息</span><strong>{{ messageTotal }}</strong><small>最近采样 {{ recentMessages.length }} 条</small></div></article>
      <article class="metric-card"><el-icon><DataLine /></el-icon><div><span>Topic 总数</span><strong>{{ topicTotal }}</strong><small>{{ topicRanks.length }} 个活跃 Topic</small></div></article>
    </div>

    <div class="dashboard-grid">
      <article class="panel trend-panel">
        <div class="panel-title"><div><h2>消息趋势</h2><p>最近 12 小时消息数量</p></div><span>总计 {{ trend.reduce((sum, item) => sum + item.count, 0) }}</span></div>
        <div class="trend-chart">
          <div v-for="item in trend" :key="item.label" class="trend-column"><span class="trend-value">{{ item.count }}</span><div class="trend-bar" :style="{ height: `${Math.max(3, item.count / trendMax * 100)}%` }" /><small>{{ item.label }}</small></div>
        </div>
      </article>

      <article class="panel topic-panel">
        <div class="panel-title"><div><h2>活跃 Topic</h2><p>最近消息分布</p></div></div>
        <div v-if="topicRanks.length" class="rank-list">
          <div v-for="(item, index) in topicRanks" :key="item.topic" class="rank-item"><b>{{ index + 1 }}</b><div><span>{{ item.topic }}</span><i :style="{ width: `${item.count / topicRankMax * 100}%` }" /></div><strong>{{ item.count }}</strong></div>
        </div>
        <div v-else class="empty-state">暂无消息数据</div>
      </article>

      <article class="panel message-panel">
        <div class="panel-title"><div><h2>最新消息</h2><p>实时 MQTT 消息流</p></div><router-link to="/iot/received-messages">查看全部</router-link></div>
        <div class="compact-table">
          <div class="table-head"><span>Topic</span><span>客户端</span><span>内容</span><span>时间</span></div>
          <div v-for="item in recentMessages.slice(0, 8)" :key="item.id" class="table-row"><strong>{{ item.topic }}</strong><span>{{ item.client_id || '-' }}</span><code :title="item.payload">{{ shortPayload(item.payload) }}</code><time>{{ formatTime(item.received_at) }}</time></div>
          <div v-if="!recentMessages.length" class="empty-state">暂无消息</div>
        </div>
      </article>

      <article class="panel client-panel">
        <div class="panel-title"><div><h2>设备状态</h2><p>最近连接客户端</p></div><router-link to="/iot/clients">查看全部</router-link></div>
        <div class="client-list">
          <div v-for="item in recentClients" :key="item.id" class="client-item"><span :class="['state-dot', item.status]" /><div><strong>{{ item.client_id }}</strong><small>{{ item.username || '未认证用户' }} · {{ item.remote_address || '-' }}</small></div><time>{{ formatTime(item.connected_at) }}</time></div>
          <div v-if="!recentClients.length" class="empty-state">暂无客户端记录</div>
        </div>
      </article>
    </div>
  </section>
</template>

<style scoped>
.iot-dashboard{height:100%;min-width:0;min-height:0;display:flex;flex-direction:column;gap:8px;padding:0 8px 8px;box-sizing:border-box;color:var(--text);font-size:11px;overflow:auto}.dashboard-header{display:flex;align-items:center;justify-content:space-between;gap:12px;padding:10px 14px;border:1px solid var(--line);border-radius:10px;background:var(--bg-panel)}.dashboard-header h1{margin:0;font-size:18px}.dashboard-header p,.panel-title p{margin:3px 0 0;color:var(--text-muted);font-size:10px}.refresh-state{display:flex;align-items:center;gap:6px;color:var(--text-muted)}.live-dot,.state-dot{width:7px;height:7px;border-radius:50%;background:var(--success);box-shadow:0 0 0 3px rgba(39,174,96,.12)}.metric-grid{display:grid;grid-template-columns:repeat(4,minmax(0,1fr));gap:8px}.metric-card{display:flex;min-width:0;align-items:center;gap:12px;padding:12px 14px;border:1px solid var(--line);border-radius:10px;background:var(--bg-panel)}.metric-card>.el-icon{width:34px;height:34px;flex:none;border-radius:8px;background:var(--primary-soft);color:var(--primary);font-size:18px}.metric-card div{display:flex;min-width:0;flex-direction:column}.metric-card span,.metric-card small{color:var(--text-muted)}.metric-card strong{margin:2px 0;font-size:22px;line-height:1.1}.dashboard-grid{display:grid;min-height:0;grid-template-columns:minmax(0,1.45fr) minmax(280px,.55fr);grid-template-rows:minmax(230px,.8fr) minmax(260px,1fr);gap:8px;flex:1}.panel{min-width:0;min-height:0;padding:12px 14px;border:1px solid var(--line);border-radius:10px;background:var(--bg-panel);overflow:hidden}.panel-title{display:flex;align-items:flex-start;justify-content:space-between;gap:10px}.panel-title h2{margin:0;font-size:13px}.panel-title>a,.panel-title>span{color:var(--primary);font-size:10px;text-decoration:none}.trend-panel{display:flex;flex-direction:column}.trend-chart{display:flex;min-height:150px;flex:1;align-items:flex-end;gap:8px;padding:24px 4px 0}.trend-column{position:relative;display:flex;height:100%;min-width:0;flex:1;align-items:center;flex-direction:column;justify-content:flex-end}.trend-bar{width:min(28px,70%);min-height:3px;border-radius:5px 5px 2px 2px;background:linear-gradient(180deg,var(--primary),var(--primary-soft))}.trend-value{position:absolute;top:-17px;color:var(--text-muted);font-size:9px}.trend-column small{margin-top:6px;color:var(--text-muted);font-size:9px;white-space:nowrap}.rank-list{display:flex;flex-direction:column;gap:12px;margin-top:16px}.rank-item{display:grid;grid-template-columns:20px minmax(0,1fr) 30px;align-items:center;gap:8px}.rank-item>b{display:grid;width:20px;height:20px;place-items:center;border-radius:5px;background:var(--primary-soft);color:var(--primary)}.rank-item>div{display:flex;min-width:0;flex-direction:column;gap:5px}.rank-item span{overflow:hidden;text-overflow:ellipsis;white-space:nowrap}.rank-item i{display:block;height:4px;border-radius:2px;background:var(--primary)}.rank-item>strong{text-align:right}.message-panel{display:flex;flex-direction:column}.compact-table{min-height:0;margin-top:10px;overflow:auto}.table-head,.table-row{display:grid;grid-template-columns:minmax(130px,1fr) minmax(90px,.65fr) minmax(180px,1.4fr) 132px;align-items:center;gap:10px;padding:8px;border-bottom:1px solid var(--line)}.table-head{color:var(--text-muted);font-weight:600}.table-row strong,.table-row span,.table-row code{min-width:0;overflow:hidden;text-overflow:ellipsis;white-space:nowrap}.table-row code{color:var(--text-soft);font:10px ui-monospace,SFMono-Regular,Menlo,monospace}.table-row time,.client-item time{color:var(--text-muted);font-size:9px;text-align:right}.client-list{display:flex;min-height:0;flex-direction:column;margin-top:8px;overflow:auto}.client-item{display:grid;grid-template-columns:8px minmax(0,1fr) 124px;align-items:center;gap:9px;padding:9px 4px;border-bottom:1px solid var(--line)}.client-item>div{display:flex;min-width:0;flex-direction:column;gap:2px}.client-item strong,.client-item small{overflow:hidden;text-overflow:ellipsis;white-space:nowrap}.client-item small{color:var(--text-muted);font-size:9px}.state-dot.disconnected{background:var(--text-muted);box-shadow:none}.empty-state{display:grid;min-height:90px;place-items:center;color:var(--text-muted)}@media(max-width:1100px){.metric-grid{grid-template-columns:repeat(2,minmax(0,1fr))}.dashboard-grid{grid-template-columns:1fr;grid-template-rows:repeat(4,minmax(240px,auto))}}@media(max-width:680px){.dashboard-header{align-items:flex-start;flex-direction:column}.metric-grid{grid-template-columns:1fr}.table-head,.table-row{grid-template-columns:minmax(120px,1fr) 100px 120px}.table-head span:nth-child(3),.table-row code{display:none}.dashboard-grid{display:flex;flex-direction:column}.panel{min-height:260px}}
</style>
