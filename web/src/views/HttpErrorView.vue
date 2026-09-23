<script setup lang="ts">
import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import gopsLogoUrl from '@/assets/gops.svg'
import { useSystemConfigStore } from '@/stores/system-config'

type ErrorCode = 401 | 403 | 404 | 429

const props = defineProps<{ code: ErrorCode }>()
const route = useRoute()
const router = useRouter()
const systemConfig = useSystemConfigStore()

const content = computed(() => ({
  401: {
    eyebrow: 'AUTHENTICATION REQUIRED',
    title: '登录状态已失效',
    description: '当前会话已过期或尚未登录，请重新登录后继续访问。',
    action: '重新登录',
  },
  403: {
    eyebrow: 'ACCESS DENIED',
    title: '暂无访问权限',
    description: '当前账号没有访问此页面的权限，如有需要请联系系统管理员。',
    action: '返回首页',
  },
  404: {
    eyebrow: 'PAGE NOT FOUND',
    title: '页面不存在',
    description: '您访问的页面可能已被移动、删除，或输入了错误的地址。',
    action: '返回首页',
  },
  429: {
    eyebrow: 'TOO MANY REQUESTS',
    title: '请求过于频繁',
    description: '系统正在处理较多请求，请稍候片刻后再试。',
    action: '重新尝试',
  },
}[props.code]))

function primaryAction() {
  if (props.code === 401) {
    const redirect = typeof route.query.redirect === 'string' ? route.query.redirect : '/'
    router.replace({ path: '/login', query: { redirect } })
    return
  }
  if (props.code === 429) {
    const redirect = typeof route.query.redirect === 'string' ? route.query.redirect : ''
    if (redirect && redirect !== route.fullPath) router.replace(redirect)
    else router.go(0)
    return
  }
  router.replace('/')
}

function goBack() {
  if (window.history.length > 1) router.back()
  else router.replace('/')
}
</script>

<template>
  <main class="error-page" :class="`status-${code}`">
    <div class="ambient ambient-one"></div>
    <div class="ambient ambient-two"></div>

    <section class="error-card">
      <div class="illustration" aria-hidden="true">
        <span class="orbit orbit-one"></span>
        <span class="orbit orbit-two"></span>
        <div class="status-window">
          <div class="window-bar"><i></i><i></i><i></i></div>
          <div class="status-code">{{ code }}</div>
          <div class="status-lines"><i></i><i></i></div>
        </div>
        <span class="status-badge">!</span>
      </div>

      <div class="content">
        <div class="brand"><img :src="gopsLogoUrl" :alt="systemConfig.systemName" /><strong>{{ systemConfig.systemName }}</strong></div>
        <p class="eyebrow">{{ content.eyebrow }}</p>
        <h1>{{ content.title }}</h1>
        <p class="description">{{ content.description }}</p>
        <div class="actions">
          <button class="primary" type="button" @click="primaryAction">{{ content.action }}</button>
          <button v-if="code !== 401" class="secondary" type="button" @click="goBack">返回上一页</button>
        </div>
        <p class="hint">错误代码：HTTP {{ code }}</p>
      </div>
    </section>
  </main>
</template>

<style scoped>
.error-page {
  --accent: #665be9;
  --accent-soft: #eef0ff;
  position: relative;
  display: grid;
  min-height: 100vh;
  overflow: hidden;
  place-items: center;
  padding: 32px;
  background: #f7f8fc;
  color: #182235;
}

.status-401 { --accent: #5275ee; --accent-soft: #edf2ff; }
.status-403 { --accent: #7658e8; --accent-soft: #f1edff; }
.status-404 { --accent: #4d8edb; --accent-soft: #eaf4ff; }
.status-429 { --accent: #e49a35; --accent-soft: #fff4df; }

.ambient {
  position: absolute;
  border-radius: 50%;
  background: var(--accent-soft);
  filter: blur(2px);
}

.ambient-one { top: -260px; left: -180px; width: 560px; height: 560px; }
.ambient-two { right: -130px; bottom: -230px; width: 480px; height: 480px; opacity: .68; }

.error-card {
  position: relative;
  z-index: 1;
  display: grid;
  width: min(900px, 100%);
  grid-template-columns: minmax(320px, 1fr) minmax(300px, .88fr);
  overflow: hidden;
  border: 1px solid #e6e9f1;
  border-radius: 22px;
  background: rgba(255, 255, 255, .96);
  box-shadow: 0 24px 70px rgba(45, 58, 99, .12);
}

.illustration {
  position: relative;
  display: grid;
  min-height: 430px;
  overflow: hidden;
  place-items: center;
  background: linear-gradient(145deg, var(--accent-soft), #fbfcff 72%);
}

.orbit {
  position: absolute;
  border: 1px solid color-mix(in srgb, var(--accent) 24%, transparent);
  border-radius: 50%;
}

.orbit-one { width: 310px; height: 310px; }
.orbit-two { width: 410px; height: 410px; opacity: .55; }

.status-window {
  position: relative;
  z-index: 2;
  width: 250px;
  padding: 18px 22px 26px;
  border: 1px solid rgba(255, 255, 255, .85);
  border-radius: 18px;
  background: #fff;
  box-shadow: 0 24px 48px color-mix(in srgb, var(--accent) 18%, transparent);
  transform: rotate(-2deg);
}

.window-bar { display: flex; gap: 6px; padding-bottom: 18px; border-bottom: 1px solid #edf0f5; }
.window-bar i { width: 7px; height: 7px; border-radius: 50%; background: #d9deea; }
.status-code { margin: 24px 0 18px; color: var(--accent); font-size: 72px; font-weight: 800; line-height: 1; letter-spacing: -5px; }
.status-lines { display: grid; gap: 9px; }
.status-lines i { width: 80%; height: 7px; border-radius: 4px; background: #e9ecf3; }
.status-lines i:last-child { width: 56%; }

.status-badge {
  position: absolute;
  z-index: 3;
  right: calc(50% - 148px);
  bottom: 88px;
  display: grid;
  width: 44px;
  height: 44px;
  place-items: center;
  border: 5px solid var(--accent-soft);
  border-radius: 50%;
  background: var(--accent);
  color: #fff;
  font-size: 20px;
  font-weight: 800;
}

.content { display: flex; flex-direction: column; justify-content: center; padding: 52px; }
.brand { display: flex; align-items: center; gap: 10px; margin-bottom: 54px; color: #667085; font-size: 11px; }
.brand img { width: 38px; height: 38px; border-radius: 10px; object-fit: contain; box-shadow: 0 8px 18px rgba(76, 83, 204, .2); }
.brand strong { color: #475467; font-weight: 600; }
.eyebrow { margin: 0 0 10px; color: var(--accent); font-size: 10px; font-weight: 800; letter-spacing: 1.5px; }
h1 { margin: 0; font-size: 30px; line-height: 1.25; }
.description { min-height: 48px; margin: 14px 0 28px; color: #7b8495; font-size: 12px; line-height: 1.9; }
.actions { display: flex; gap: 10px; }
button { min-width: 104px; height: 40px; padding: 0 18px; border-radius: 8px; font: inherit; font-size: 12px; font-weight: 600; cursor: pointer; }
.primary { border: 1px solid var(--accent); background: var(--accent); color: #fff; box-shadow: 0 8px 18px color-mix(in srgb, var(--accent) 22%, transparent); }
.secondary { border: 1px solid #dfe3eb; background: #fff; color: #5f6878; }
button:hover { filter: brightness(.98); }
.hint { margin: 24px 0 0; color: #b0b7c4; font-size: 10px; }

@media (max-width: 760px) {
  .error-page { padding: 18px; }
  .error-card { grid-template-columns: 1fr; }
  .illustration { min-height: 250px; }
  .orbit-one { width: 220px; height: 220px; }
  .orbit-two { width: 290px; height: 290px; }
  .status-window { width: 205px; padding: 14px 18px 20px; }
  .status-code { margin: 18px 0 14px; font-size: 54px; }
  .status-badge { right: calc(50% - 122px); bottom: 35px; }
  .content { padding: 34px 30px 38px; }
  .brand { margin-bottom: 32px; }
}

@media (max-width: 420px) {
  .illustration { display: none; }
  .content { min-height: calc(100vh - 36px); padding: 28px 24px; }
  .actions { flex-direction: column; }
  button { width: 100%; }
}
</style>
