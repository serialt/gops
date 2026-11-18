<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { getDashboardStats } from '@/api/dashboard'

const router = useRouter()
const stats = ref({ hosts_total: 0, hosts_online: 0, hosts_offline: 0, clusters_total: 0, clusters_offline: 0, nodes_total: 0, nodes_offline: 0, pods_total: 0, pods_abnormal: 0, domains_total: 0, domains_normal: 0, domains_abnormal: 0, domains_paused: 0, domain_certificates: [] as Array<{ domain: string; ssl_expiry: string }>, today_operations: 0, active_alerts: 0 })

const topCards = computed(() => [
  {
    title: '主机总数',
    value: stats.value.hosts_total,
    secondary: `在线 ${stats.value.hosts_online} · 离线 ${stats.value.hosts_offline}`,
    accent: 'blue',
    tone: 'blue',
  },
  {
    title: 'K8s 集群',
    value: stats.value.clusters_total,
    secondary: `离线 ${stats.value.clusters_offline} · 节点 ${stats.value.nodes_total} · Pod ${stats.value.pods_total}`,
    accent: 'green',
    tone: 'green',
  },
  {
    title: '今日操作',
    value: stats.value.today_operations,
    secondary: '审计记录',
    accent: 'cyan',
    tone: 'cyan',
  },
  {
    title: '活跃告警',
    value: stats.value.active_alerts,
    secondary: '待处理告警',
    accent: 'red',
    tone: 'red',
  },
])

const onlineRate = computed(() => {
  if (!stats.value.hosts_total) return 0
  return Math.min(100, Math.round((stats.value.hosts_online / stats.value.hosts_total) * 100))
})

const donutStyle = computed(() => ({
  background: `conic-gradient(#2bb673 0 ${onlineRate.value}%, #dfe5ec ${onlineRate.value}% 100%)`,
}))

const dotValue = computed(() => Math.max(stats.value.hosts_online || 0, 0))
const clustersOnline = computed(() => Math.max(stats.value.clusters_total - stats.value.clusters_offline, 0))
const clusterOnlineRate = computed(() => stats.value.clusters_total ? Math.round((clustersOnline.value / stats.value.clusters_total) * 100) : 0)
const nodesOnline = computed(() => Math.max(stats.value.nodes_total - stats.value.nodes_offline, 0))
const podsNormal = computed(() => Math.max(stats.value.pods_total - stats.value.pods_abnormal, 0))
const openHosts = () => router.push({ name: 'hosts' })
const openClusters = () => router.push({ name: 'kubernetes-clusters' })
const openDomains = () => router.push({ name: 'monitor-domains' })
const expiryDays = (value: string) => Math.ceil((new Date(value).getTime() - Date.now()) / 86400000)
const expiryText = (value: string) => { const days = expiryDays(value); return days < 0 ? `已过期 ${Math.abs(days)} 天` : days === 0 ? '今天到期' : `${days} 天后到期` }
const expiryDate = (value: string) => value ? new Date(value).toLocaleDateString('zh-CN') : '—'

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
          <button class="link-btn" type="button" @click="openHosts">查看全部</button>
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
          <div class="status-row"><span class="legend legend--green"></span> 在线 <span class="count">{{ dotValue
              }}</span></div>
          <div class="status-row"><span class="legend legend--gray"></span> 离线 <span class="count">{{
            stats.hosts_offline }}</span></div>
          <div class="status-row"><span class="legend legend--default"></span> 资产总量 <span class="count">{{ stats.hosts_total
              }}</span></div>
        </div>
      </section>

      <section class="panel panel-right">
        <div class="panel-header">
          <h3>K8s 集群资源概览</h3>
          <button class="link-btn" type="button" @click="openClusters">查看全部</button>
        </div>

        <div class="resource-summary">
          <div class="summary-box cluster-summary">
            <div class="cluster-total-line"><span>集群总数</span><strong>{{ stats.clusters_total }}</strong></div>
            <div class="cluster-status-line"><span>在线 {{ clustersOnline }}</span><i>·</i><span>离线 {{ stats.clusters_offline }}</span></div>
          </div>
          <div class="summary-box status-summary">
            <div class="cluster-total-line"><span>节点总数</span><strong>{{ stats.nodes_total }}</strong></div>
            <div class="cluster-status-line"><span>在线 {{ nodesOnline }}</span><i>·</i><span>离线 {{ stats.nodes_offline }}</span></div>
          </div>
          <div class="summary-box status-summary">
            <div class="cluster-total-line"><span>Pod 总数</span><strong>{{ stats.pods_total }}</strong></div>
            <div class="cluster-status-line"><span>正常 {{ podsNormal }}</span><i>·</i><span>异常 {{ stats.pods_abnormal }}</span></div>
          </div>
        </div>

        <div class="resource-details"><div><span>集群在线率</span><strong :class="{ 'offline-value': stats.clusters_offline > 0 }">{{ clusterOnlineRate }}%</strong></div><div><span>平均节点数</span><strong>{{ stats.clusters_total ? (stats.nodes_total / stats.clusters_total).toFixed(1) : '0' }}</strong></div><div><span>平均 Pod 数</span><strong>{{ stats.clusters_total ? (stats.pods_total / stats.clusters_total).toFixed(1) : '0' }}</strong></div></div>
      </section>

      <section class="panel domain-panel">
        <div class="panel-header">
          <h3>域名监控概览</h3>
          <button class="link-btn" type="button" @click="openDomains">查看全部</button>
        </div>
        <div class="domain-summary">
          <div class="domain-status-summary">
            <div class="cluster-total-line"><span>监控总数</span><strong>{{ stats.domains_total }}</strong></div>
            <div class="domain-status-line"><span class="normal">正常 {{ stats.domains_normal }}</span><i>·</i><span class="abnormal">异常 {{ stats.domains_abnormal }}</span><i>·</i><span class="paused">暂停 {{ stats.domains_paused }}</span></div>
          </div>
        </div>
        <div class="certificate-expiry">
          <div class="certificate-expiry-head"><strong>证书到期时间</strong><span>按到期时间正序</span></div>
          <div v-if="stats.domain_certificates.length" class="certificate-expiry-list">
            <div v-for="item in stats.domain_certificates" :key="`${item.domain}-${item.ssl_expiry}`" class="certificate-expiry-row">
              <span class="certificate-domain">{{ item.domain }}</span>
              <time>{{ expiryDate(item.ssl_expiry) }}</time>
              <strong :class="{ danger: expiryDays(item.ssl_expiry) <= 10, warning: expiryDays(item.ssl_expiry) > 10 && expiryDays(item.ssl_expiry) <= 30 }">{{ expiryText(item.ssl_expiry) }}</strong>
            </div>
          </div>
          <div v-else class="certificate-expiry-empty">暂无已采集证书到期时间的域名</div>
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
  font-size: 12px;
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

.domain-panel {
  grid-column: 1 / -1;
}

.domain-summary {
  display: block;
}

.domain-status-summary {
  display: flex;
  min-width: 0;
  min-height: 78px;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 10px 18px;
  border: 1px solid var(--line);
  border-radius: 8px;
  background: var(--bg-soft);
}

.domain-status-line {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 7px;
  margin-top: 7px;
  font-size: 11px;
}

.domain-status-line .normal {
  color: var(--success);
}

.domain-status-line .abnormal {
  color: var(--danger);
}

.domain-status-line .paused,
.domain-status-line i {
  color: var(--text-muted);
}

.domain-status-line i {
  font-style: normal;
}

.certificate-expiry {
  margin-top: 14px;
  border: 1px solid var(--line);
  border-radius: 8px;
  overflow: hidden;
}

.certificate-expiry-head,
.certificate-expiry-row {
  display: grid;
  grid-template-columns: minmax(180px, 1fr) 140px 120px;
  align-items: center;
  gap: 12px;
}

.certificate-expiry-head {
  padding: 9px 12px;
  background: var(--bg-soft);
}

.certificate-expiry-head strong {
  font-size: 12px;
}

.certificate-expiry-head span {
  grid-column: 2 / -1;
  color: var(--text-muted);
  font-size: 10px;
  text-align: right;
}

.certificate-expiry-row {
  min-height: 38px;
  padding: 0 12px;
  border-top: 1px solid var(--line);
  font-size: 11px;
}

.certificate-domain {
  overflow: hidden;
  color: var(--text);
  font-weight: 600;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.certificate-expiry-row time {
  color: var(--text-secondary);
}

.certificate-expiry-row strong {
  color: var(--success);
  text-align: right;
}

.certificate-expiry-row strong.warning {
  color: var(--warning);
}

.certificate-expiry-row strong.danger {
  color: var(--danger);
}

.certificate-expiry-empty {
  padding: 18px 12px;
  color: var(--text-muted);
  font-size: 11px;
  text-align: center;
}

.panel-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 16px;
}

.panel-header h3 {
  margin: 0;
  font-size: 15px;
  font-weight: 600;
  color: var(--text);
}

.link-btn {
  background: transparent;
  border: 0;
  color: var(--primary);
  font-size: 12px;
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
  font-size: 13px;
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
  background: rgba(255, 255, 255, 0.42);
  border: 1px solid rgba(148, 163, 184, 0.16);
  border-radius: 12px;
  padding: 18px 10px 14px;
  text-align: center;
}

.cluster-summary,
.status-summary {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
}

.cluster-total-line {
  display: flex;
  align-items: baseline;
  justify-content: center;
  gap: 10px;
  color: var(--text-secondary);
  font-size: 12px;
}

.cluster-total-line strong {
  color: var(--text);
  font-size: 34px;
  line-height: 1;
}

.cluster-status-line {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  margin-top: 7px;
  color: var(--text-secondary);
  font-size: 11px;
}

.cluster-status-line span:first-child {
  color: var(--success);
}

.cluster-status-line span:last-child {
  color: var(--danger);
}

.cluster-status-line i {
  color: var(--text-muted);
  font-style: normal;
}

.summary-value {
  font-size: 34px;
  font-weight: 800;
  color: #0f172a;
  margin-bottom: 8px;
}

.summary-label {
  font-size: 12px;
  color: #5c6978;
}

.summary-value--online {
  color: var(--success);
}

.summary-value--offline {
  color: var(--danger);
}

.resource-details {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 12px;
  margin-top: 22px;
}

.resource-details>div {
  display: flex;
  min-height: 92px;
  align-items: center;
  justify-content: space-between;
  padding: 0 20px;
  border: 1px solid var(--line);
  border-radius: 8px;
  background: var(--bg-soft);
  color: var(--text-soft);
}

.resource-details strong {
  color: var(--text);
  font-size: 24px;
}

.resource-details strong.offline-value {
  color: var(--danger);
}

.chart-area {
  position: relative;
  height: 260px;
  margin-top: 22px;
  border-radius: 14px;
  background: linear-gradient(180deg, rgba(255, 255, 255, 0.25), rgba(255, 255, 255, 0.08));
  border: 1px solid rgba(148, 163, 184, 0.12);
  overflow: hidden;
}

.chart-grid {
  position: absolute;
  inset: 0;
  background-image: linear-gradient(to top, rgba(148, 163, 184, 0.14) 1px, transparent 1px);
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

.bar--blue {
  background: linear-gradient(180deg, #8ab9ff, #2a82f5);
}

.bar--green {
  background: linear-gradient(180deg, #8fe0bb, #2bb673);
}

.legend-row {
  display: flex;
  justify-content: flex-end;
  gap: 18px;
  margin-top: 18px;
  color: #54657b;
  font-size: 13px;
}

.legend-row span {
  display: inline-flex;
  align-items: center;
  gap: 8px;
}

@media (max-width: 1200px) {
  .cards-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .panel-grid {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 640px) {
  .cards-grid {
    grid-template-columns: 1fr;
    gap: 12px;
  }

  .panel-grid {
    gap: 12px;
  }

  .mini-card {
    min-height: 96px;
    padding: 14px;
  }

  .card-value {
    font-size: 26px;
  }

  .panel {
    padding: 14px;
  }

  .donut-wrap {
    min-height: 190px;
  }

  .donut-chart {
    width: 160px;
    height: 160px;
  }

  .chart-area {
    height: 200px;
  }

  .summary-value {
    font-size: 26px;
  }

  .resource-summary {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .cluster-summary {
    grid-column: 1 / -1;
  }

  .resource-details {
    grid-template-columns: 1fr;
  }

  .certificate-expiry-head,
  .certificate-expiry-row {
    grid-template-columns: minmax(0, 1fr) 92px;
  }

  .certificate-expiry-head span,
  .certificate-expiry-row strong {
    display: none;
  }
}
</style>
