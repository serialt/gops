<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { getDashboardStats } from '@/api'

const stats = ref({ users: 0, enabled_users: 0, roles: 0, policies: 0 })

const topCards = computed(() => [
  {
    title: '主机总数',
    value: stats.value.users,
    secondary: '在线状态',
    accent: 'blue',
    tone: 'blue',
  },
  {
    title: 'K8s 集群',
    value: '0',
    secondary: '集群资源',
    accent: 'green',
    tone: 'green',
  },
  {
    title: '今日操作',
    value: '0',
    secondary: '审计记录',
    accent: 'cyan',
    tone: 'cyan',
  },
  {
    title: '活跃告警',
    value: '0',
    secondary: '告警事件',
    accent: 'red',
    tone: 'red',
  },
])

const onlineRate = computed(() => {
  if (!stats.value.users) return 0
  return Math.min(100, Math.round((stats.value.enabled_users / stats.value.users) * 100))
})

const donutStyle = computed(() => ({
  background: `conic-gradient(#2bb673 0 ${onlineRate.value}%, #dfe5ec ${onlineRate.value}% 100%)`,
}))

const dotValue = computed(() => Math.max(stats.value.enabled_users || 0, 0))

onMounted(async () => {
  const { data } = await getDashboardStats()
  stats.value = data.data
})
</script>

<template>
  <div class="page-container dashboard-shell">
    <div class="cards-grid">
      <div v-for="item in topCards" :key="item.title" class="mini-card" :class="`tone-${item.tone}`">
        <div class="card-left">
          <div class="card-title">{{ item.title }}</div>
          <div class="card-value">{{ item.value }}</div>
        </div>
        <div class="card-icon" aria-hidden="true"></div>
        <div class="card-footer">
          <span>{{ item.secondary }}</span>
        </div>
      </div>
    </div>

    <div class="panel-grid">
      <section class="panel panel-left">
        <div class="panel-header">
          <h3>主机状态分布</h3>
          <button class="link-btn">查看全部</button>
        </div>

        <div class="donut-wrap">
          <div class="donut-chart" :style="donutStyle">
            <div class="donut-inner">
              <span class="donut-percent">{{ onlineRate }}%</span>
              <span class="donut-label">在线率</span>
            </div>
          </div>
        </div>

        <div class="status-list">
          <div class="status-row"><span class="legend legend--green"></span> 在线 <span class="count">{{ dotValue }}</span></div>
          <div class="status-row"><span class="legend legend--gray"></span> 离线 <span class="count">{{ Math.max((stats.users || 0) - dotValue, 0) }}</span></div>
          <div class="status-row"><span class="legend legend--default"></span> 资产总量 <span class="count">{{ stats.users }}</span></div>
        </div>
      </section>

      <section class="panel panel-right">
        <div class="panel-header">
          <h3>K8s 集群资源概览</h3>
          <button class="link-btn">查看全部</button>
        </div>

        <div class="resource-summary">
          <div class="summary-box">
            <div class="summary-value">0</div>
            <div class="summary-label">节点总数</div>
          </div>
          <div class="summary-box">
            <div class="summary-value">0</div>
            <div class="summary-label">Pod 总数</div>
          </div>
          <div class="summary-box">
            <div class="summary-value">0</div>
            <div class="summary-label">集群数量</div>
          </div>
        </div>

        <div class="chart-area">
          <div class="chart-grid"></div>
          <div class="chart-bars">
            <span class="bar bar--blue" style="height: 20%"></span>
            <span class="bar bar--green" style="height: 60%"></span>
            <span class="bar bar--blue" style="height: 35%"></span>
            <span class="bar bar--green" style="height: 80%"></span>
            <span class="bar bar--blue" style="height: 55%"></span>
            <span class="bar bar--green" style="height: 46%"></span>
            <span class="bar bar--blue" style="height: 30%"></span>
          </div>
        </div>

        <div class="legend-row">
          <span><i class="legend legend--green"></i>节点</span>
          <span><i class="legend legend--blue"></i>Pod</span>
        </div>
      </section>
    </div>
  </div>
</template>

<style scoped>
.dashboard-shell {
  background: var(--bg);
  min-height: 100%;
}

.cards-grid {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 16px;
  margin-bottom: 20px;
}

.mini-card {
  position: relative;
  display: flex;
  align-items: center;
  justify-content: space-between;
  background: var(--bg-card);
  border: 1px solid var(--line);
  border-radius: 8px;
  padding: 16px;
  min-height: 110px;
  box-shadow: var(--shadow);
}

.card-left {
  flex: 1;
  display: flex;
  flex-direction: column;
  justify-content: center;
  gap: 6px;
}

.card-title {
  color: var(--text-soft);
  font-size: 13px;
  font-weight: 500;
}

.card-value {
  font-size: 32px;
  font-weight: 700;
  color: var(--text);
  line-height: 1.1;
}

.card-icon {
  width: 40px;
  height: 40px;
  border-radius: 8px;
  position: relative;
  margin-left: 16px;
  box-shadow: none;
}

.tone-blue .card-icon {
  background: #e3f2fd;
}

.tone-green .card-icon {
  background: #e8f5e9;
}

.tone-cyan .card-icon {
  background: #e0f7fa;
}

.tone-red .card-icon {
  background: #ffebee;
}

.card-icon::before {
  content: "";
  position: absolute;
  top: 8px;
  left: 8px;
  width: 24px;
  height: 24px;
  border-radius: 4px;
  background: transparent;
}

.card-footer {
  position: absolute;
  left: 16px;
  right: 16px;
  bottom: 12px;
  display: flex;
  justify-content: space-between;
  align-items: center;
  color: var(--text-muted);
  font-size: 11px;
}

.panel-grid {
  display: grid;
  grid-template-columns: 1fr 1.2fr;
  gap: 16px;
}

.panel {
  background: var(--bg-card);
  border: 1px solid var(--line);
  border-radius: 8px;
  padding: 16px;
  box-shadow: var(--shadow);
}

.panel-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 16px;
}

.panel-header h3 {
  margin: 0;
  font-size: 16px;
  font-weight: 600;
  color: var(--text);
}

.link-btn {
  background: transparent;
  border: 0;
  color: var(--primary);
  font-size: 13px;
  cursor: pointer;
  padding: 0;
  font-weight: 500;
}

.donut-wrap {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 240px;
}

.donut-chart {
  position: relative;
  width: 200px;
  height: 200px;
  border-radius: 50%;
  display: grid;
  place-items: center;
}

.donut-chart::before {
  content: "";
  position: absolute;
  inset: 20px;
  background: var(--bg);
  border-radius: 50%;
  box-shadow: none;
}

.donut-inner {
  position: relative;
  z-index: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  width: 140px;
  height: 140px;
  border-radius: 50%;
  background: transparent;
}

.donut-percent {
  font-size: 32px;
  font-weight: 700;
  color: var(--text);
}

.donut-label {
  font-size: 13px;
  color: var(--text-soft);
  margin-top: 4px;
}

.status-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
  padding: 0;
}

.status-row {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 8px 12px;
  border-radius: 6px;
  background: var(--bg-soft);
  color: var(--text);
  font-size: 14px;
}

.status-row .count {
  margin-left: auto;
  color: var(--text);
  font-weight: 600;
}

.legend {
  display: inline-block;
  width: 10px;
  height: 10px;
  border-radius: 50%;
}

.legend--green {
  background: #2bb673;
}

.legend--gray {
  background: #d5dbe4;
}

.legend--blue {
  background: #3a8dff;
}

.legend--default {
  background: transparent;
  border: 2px solid #d5dbe4;
}

.resource-summary {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 12px;
  margin-top: 8px;
}

.summary-box {
  background: rgba(255,255,255,0.42);
  border: 1px solid rgba(148,163,184,0.16);
  border-radius: 12px;
  padding: 18px 10px 14px;
  text-align: center;
}

.summary-value {
  font-size: 34px;
  font-weight: 800;
  color: #0f172a;
  margin-bottom: 8px;
}

.summary-label {
  font-size: 13px;
  color: #5c6978;
}

.chart-area {
  position: relative;
  height: 260px;
  margin-top: 22px;
  border-radius: 14px;
  background: linear-gradient(180deg, rgba(255,255,255,0.25), rgba(255,255,255,0.08));
  border: 1px solid rgba(148,163,184,0.12);
  overflow: hidden;
}

.chart-grid {
  position: absolute;
  inset: 0;
  background-image: linear-gradient(to top, rgba(148,163,184,0.14) 1px, transparent 1px);
  background-size: 100% 20%;
}

.chart-bars {
  position: absolute;
  left: 18px;
  right: 18px;
  bottom: 18px;
  height: 190px;
  display: flex;
  align-items: flex-end;
  justify-content: space-between;
  gap: 14px;
}

.bar {
  width: 100%;
  max-width: 32px;
  border-radius: 10px 10px 0 0;
  display: inline-block;
}

.bar--blue { background: linear-gradient(180deg, #8ab9ff, #2a82f5); }
.bar--green { background: linear-gradient(180deg, #8fe0bb, #2bb673); }

.legend-row {
  display: flex;
  justify-content: flex-end;
  gap: 18px;
  margin-top: 18px;
  color: #54657b;
  font-size: 14px;
}

.legend-row span {
  display: inline-flex;
  align-items: center;
  gap: 8px;
}

@media (max-width: 1200px) {
  .cards-grid { grid-template-columns: repeat(2, minmax(0, 1fr)); }
  .panel-grid { grid-template-columns: 1fr; }
}
</style>
