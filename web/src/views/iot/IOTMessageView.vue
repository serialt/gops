<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, reactive, ref, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { Check, CircleClose, Close, Edit, Plus, Promotion } from '@element-plus/icons-vue'
import { listIOTMessages, listIOTTopics, publishIOTMessage, type IOTMessage, type IOTTopic } from '@/api/iot'

type MessageTab = 'all' | 'received' | 'sent'
type TopicSubscription = { topic: string; enabled: boolean; qos: number; alias: string }
type DisplayMessage = IOTMessage & { direction: 'received' | 'sent'; localOnly?: boolean; sortAt?: number }
type SentMarker = { id: number; topic: string; payload: string; qos: number; retained: boolean; sentAt: string }

const sentMarkerStorageKey = 'iot-message-sent-markers'

const topic = ref('')
const payload = ref('')
const qos = ref(0)
const retain = ref(false)
const sending = ref(false)
const messages = ref<DisplayMessage[]>([])
const sentMarkers = ref<SentMarker[]>([])
const topics = ref<IOTTopic[]>([])
const subscriptions = ref<TopicSubscription[]>([])
const subscriptionTopic = ref('#')
const selectedTopic = ref('')
const contextMenu = reactive({ visible: false, x: 0, y: 0, subscription: null as TopicSubscription | null })
const subscriptionEditDialog = ref(false)
const editingSubscription = ref<TopicSubscription | null>(null)
const subscriptionForm = reactive({ topic: '', qos: 0, alias: '' })
const messageTab = ref<MessageTab>('all')
const displayFormat = ref<'JSON' | 'Plaintext'>('JSON')
const payloadFormat = ref<'JSON' | 'Plaintext'>('Plaintext')
const sidebarWidth = ref(330)
const publisherHeight = ref(230)
const messagePanel = ref<HTMLElement | null>(null)
let refreshTimer: ReturnType<typeof setInterval> | undefined
let dragStartX = 0
let dragStartWidth = 330
let dragStartY = 0
let dragStartPublisherHeight = 230

const filteredMessages = computed(() => {
	const subscribed = messages.value.filter((message) => subscriptions.value.some((subscription) => subscription.enabled && topicMatches(subscription.topic, message.topic)))
	if (messageTab.value === 'received') return subscribed.filter((message) => message.direction === 'received')
	if (messageTab.value === 'sent') return subscribed.filter((message) => message.direction === 'sent')
	return subscribed
})

async function load() {
	const [messageResponse, topicResponse] = await Promise.all([
		listIOTMessages({ page: 1, limit: 100 }),
		listIOTTopics({ page: 1, limit: 1000 }),
	])
	const cutoff = Date.now() - 7 * 24 * 60 * 60 * 1000
	sentMarkers.value = sentMarkers.value.filter((marker) => Date.parse(marker.sentAt) >= cutoff).slice(-200)
	const storedMessages = [...messageResponse.data.data.list].reverse().map((message): DisplayMessage => {
		const receivedAt = Date.parse(message.received_at)
		if (message.client_id !== 'inline') return { ...message, direction: 'received', sortAt: receivedAt }
		const marker = sentMarkers.value.find((item) => item.topic === message.topic
			&& item.payload === message.payload
			&& item.qos === message.qos
			&& Math.abs(Date.parse(item.sentAt) - receivedAt) < 60_000)
		return { ...message, direction: 'received', sortAt: marker ? Date.parse(marker.sentAt) + 1 : receivedAt }
	})
	const localMessages: DisplayMessage[] = sentMarkers.value.map((marker) => ({
		id: marker.id,
		received_at: marker.sentAt,
		client_id: '',
		username: '',
		topic: marker.topic,
		payload: marker.payload,
		qos: marker.qos,
		retained: marker.retained,
		duplicate: false,
		direction: 'sent',
		localOnly: true,
		sortAt: Date.parse(marker.sentAt),
	}))
	messages.value = [...storedMessages, ...localMessages].sort((left, right) => (left.sortAt ?? Date.parse(left.received_at)) - (right.sortAt ?? Date.parse(right.received_at)))
	topics.value = topicResponse.data.data.list
}

function topicMatches(filter: string, topicName: string) {
	const filterParts = filter.split('/')
	const topicParts = topicName.split('/')
	for (let index = 0; index < filterParts.length; index += 1) {
		if (filterParts[index] === '#') return true
		if (index >= topicParts.length) return false
		if (filterParts[index] !== '+' && filterParts[index] !== topicParts[index]) return false
	}
	return filterParts.length === topicParts.length
}

function addSubscription(value = subscriptionTopic.value) {
	const filter = value.trim()
	if (!filter) {
		ElMessage.warning('请输入 Topic')
		return
	}
	if (!subscriptions.value.some((subscription) => subscription.topic === filter)) subscriptions.value.push({ topic: filter, enabled: true, qos: 0, alias: '' })
	subscriptionTopic.value = '#'
	selectedTopic.value = ''
}

function addSelectedTopic(value: string) {
	if (value) addSubscription(value)
}

function removeSubscription(subscription: TopicSubscription) {
	subscriptions.value = subscriptions.value.filter((item) => item !== subscription)
}

function openTopicContextMenu(event: MouseEvent, subscription: TopicSubscription) {
	event.preventDefault()
	contextMenu.subscription = subscription
	contextMenu.x = Math.min(event.clientX, window.innerWidth - 180)
	contextMenu.y = Math.min(event.clientY, window.innerHeight - 120)
	contextMenu.visible = true
	window.addEventListener('click', closeTopicContextMenu)
}

function closeTopicContextMenu() {
	contextMenu.visible = false
	contextMenu.subscription = null
	window.removeEventListener('click', closeTopicContextMenu)
}

function toggleSubscription() {
	if (contextMenu.subscription) contextMenu.subscription.enabled = !contextMenu.subscription.enabled
	closeTopicContextMenu()
}

function editSubscription() {
	const subscription = contextMenu.subscription
	if (!subscription) return
	editingSubscription.value = subscription
	Object.assign(subscriptionForm, { topic: subscription.topic, qos: subscription.qos, alias: subscription.alias })
	closeTopicContextMenu()
	subscriptionEditDialog.value = true
}

function saveSubscriptionEdit() {
	const subscription = editingSubscription.value
	const nextTopic = subscriptionForm.topic.trim()
	if (!subscription || !nextTopic) {
		ElMessage.warning('Topic 过滤器不能为空')
		return
	}
	if (subscriptions.value.some((item) => item !== subscription && item.topic === nextTopic)) {
		ElMessage.warning('该 Topic 已添加')
		return
	}
	subscription.topic = nextTopic
	subscription.qos = subscriptionForm.qos
	subscription.alias = subscriptionForm.alias.trim()
	subscriptionEditDialog.value = false
	editingSubscription.value = null
}

function formatMessage(message: IOTMessage) {
	if (displayFormat.value !== 'JSON') return message.payload
	try {
		return JSON.stringify(JSON.parse(message.payload), null, 2)
	} catch {
		return message.payload
	}
}

function formatTime(value: string) {
	return value.replace('T', ' ').slice(0, 19)
}

async function publish() {
	const topicName = topic.value.trim()
	if (!topicName) {
		ElMessage.warning('请输入发布 Topic')
		return
	}
	if (payloadFormat.value === 'JSON') {
		try {
			JSON.parse(payload.value)
		} catch {
			ElMessage.warning('请输入有效的 JSON 消息')
			return
		}
	}
	sending.value = true
	try {
		await publishIOTMessage({ topic: topicName, payload: payload.value, qos: qos.value, retain: retain.value })
		sentMarkers.value.push({ id: -Date.now(), topic: topicName, payload: payload.value, qos: qos.value, retained: retain.value, sentAt: new Date().toISOString() })
		ElMessage.success('消息已发送')
		await load()
	} finally {
		sending.value = false
	}
}

function startResize(event: PointerEvent) {
	dragStartX = event.clientX
	dragStartWidth = sidebarWidth.value
	window.addEventListener('pointermove', resizeSidebar)
	window.addEventListener('pointerup', stopResize)
}

function resizeSidebar(event: PointerEvent) {
	sidebarWidth.value = Math.max(240, Math.min(480, dragStartWidth + event.clientX - dragStartX))
}

function stopResize() {
	window.removeEventListener('pointermove', resizeSidebar)
	window.removeEventListener('pointerup', stopResize)
}

function startPublisherResize(event: PointerEvent) {
	dragStartY = event.clientY
	dragStartPublisherHeight = publisherHeight.value
	window.addEventListener('pointermove', resizePublisher)
	window.addEventListener('pointerup', stopPublisherResize)
}

function resizePublisher(event: PointerEvent) {
	const panelHeight = messagePanel.value?.clientHeight ?? window.innerHeight
	const maximum = Math.max(170, panelHeight - 150)
	publisherHeight.value = Math.max(170, Math.min(maximum, dragStartPublisherHeight + dragStartY - event.clientY))
}

function stopPublisherResize() {
	window.removeEventListener('pointermove', resizePublisher)
	window.removeEventListener('pointerup', stopPublisherResize)
}

onMounted(async () => {
	try {
		const savedSubscriptions = JSON.parse(localStorage.getItem('iot-message-subscriptions') || '[]')
		if (Array.isArray(savedSubscriptions)) {
				subscriptions.value = savedSubscriptions.flatMap((item): TopicSubscription[] => {
					if (typeof item === 'string') return item ? [{ topic: item, enabled: true, qos: 0, alias: '' }] : []
					if (item && typeof item.topic === 'string' && item.topic) return [{
						topic: item.topic,
						enabled: item.enabled !== false,
						qos: [0, 1, 2].includes(item.qos) ? item.qos : 0,
						alias: typeof item.alias === 'string' ? item.alias : '',
					}]
				return []
			})
		}
		const savedSentMarkers = JSON.parse(localStorage.getItem(sentMarkerStorageKey) || '[]')
		if (Array.isArray(savedSentMarkers)) {
			sentMarkers.value = savedSentMarkers.filter((item): item is SentMarker => item
				&& typeof item.id === 'number'
				&& typeof item.topic === 'string'
				&& typeof item.payload === 'string'
				&& typeof item.qos === 'number'
				&& typeof item.sentAt === 'string')
		}
	} catch {
		subscriptions.value = []
	}
	await load()
	refreshTimer = setInterval(() => void load(), 3000)
})

	watch(subscriptions, (value) => localStorage.setItem('iot-message-subscriptions', JSON.stringify(value)), { deep: true })
	watch(sentMarkers, (value) => localStorage.setItem(sentMarkerStorageKey, JSON.stringify(value)), { deep: true })
onBeforeUnmount(() => {
	if (refreshTimer) clearInterval(refreshTimer)
	stopResize()
	stopPublisherResize()
})
</script>

<template>
	<section class="mqtt-page">
		<main class="message-workspace" :style="{ gridTemplateColumns: `${sidebarWidth}px 6px minmax(0, 1fr)` }">
			<aside class="subscription-panel">
				<div class="topic-actions">
					<el-button class="add-topic-button" :icon="Plus" @click="addSubscription()">添加 Topic</el-button>
					<el-select v-model="selectedTopic" placeholder="选择 Topic" @change="addSelectedTopic">
						<el-option v-for="item in topics" :key="item.id" :label="item.topic" :value="item.topic" />
					</el-select>
				</div>
				<el-input v-model="subscriptionTopic" placeholder="Topic 过滤器，例如 #" @keyup.enter="addSubscription()" />
				<div class="subscriptions">
					<div v-if="!subscriptions.length" class="subscription-empty">暂无订阅</div>
					<div
						v-for="subscription in subscriptions"
						:key="subscription.topic"
						:class="['subscription-item', { 'is-disabled': !subscription.enabled }]"
						@contextmenu.prevent="openTopicContextMenu($event, subscription)"
					>
						<span class="subscription-status" aria-hidden="true" />
						<div class="subscription-copy">
							<span class="subscription-name">{{ subscription.alias || subscription.topic }}</span>
							<small v-if="subscription.alias">{{ subscription.topic }}</small>
						</div>
						<el-tag size="small" effect="plain">QoS {{ subscription.qos }}</el-tag>
						<el-button text :icon="Close" aria-label="移除订阅" title="移除订阅" @click="removeSubscription(subscription)" />
					</div>
				</div>
			</aside>

			<div class="splitter" role="separator" aria-orientation="vertical" @pointerdown="startResize" />

			<section ref="messagePanel" class="message-panel">
				<div class="feed-toolbar">
					<el-select v-model="displayFormat" class="format-select" aria-label="消息显示格式">
						<el-option label="JSON" value="JSON" />
						<el-option label="Plaintext" value="Plaintext" />
					</el-select>
					<nav class="message-tabs" aria-label="消息筛选">
						<button :class="{ active: messageTab === 'all' }" @click="messageTab = 'all'">全部</button>
						<button :class="{ active: messageTab === 'received' }" @click="messageTab = 'received'">已接收</button>
						<button :class="{ active: messageTab === 'sent' }" @click="messageTab = 'sent'">已发送</button>
					</nav>
				</div>

				<section class="message-feed" aria-live="polite">
					<div v-if="!filteredMessages.length" class="feed-empty">
						{{ subscriptions.length ? '当前筛选下暂无消息' : '当前筛选下暂无消息' }}
					</div>
					<article v-for="message in filteredMessages" :key="message.id" :class="['message-item', { 'is-sent': message.direction === 'sent' }]">
						<div class="message-meta">
							<span class="topic-label">Topic:</span>
							<strong>{{ message.topic }}</strong>
							<span>QoS {{ message.qos }}</span>
							<span>{{ message.direction === 'received' ? '已接收' : '已发送' }}</span>
							<time>{{ formatTime(message.received_at) }}</time>
						</div>
						<pre>{{ formatMessage(message) }}</pre>
					</article>
				</section>

				<div class="horizontal-splitter" role="separator" aria-orientation="horizontal" @pointerdown="startPublisherResize" />

				<section class="publisher" :style="{ height: `${publisherHeight}px` }">
					<div class="publish-options">
						<el-select v-model="payloadFormat" aria-label="消息内容格式">
							<el-option label="Plaintext" value="Plaintext" />
							<el-option label="JSON" value="JSON" />
						</el-select>
						<el-select v-model="qos" aria-label="QoS">
							<el-option v-for="value in [0, 1, 2]" :key="value" :label="`QoS ${value}`" :value="value" />
						</el-select>
						<el-checkbox v-model="retain">Retain</el-checkbox>
					</div>
					<el-input v-model="topic" class="publish-topic" placeholder="发布 Topic，例如 devices/1/set" />
					<div class="payload-row">
						<el-input
							v-model="payload"
							type="textarea"
							:rows="4"
							resize="vertical"
							:placeholder="payloadFormat === 'JSON' ? '输入 JSON 消息内容' : '输入消息内容'"
							@keydown.meta.enter.prevent="publish"
							@keydown.ctrl.enter.prevent="publish"
						/>
						<el-button
							class="send-button"
							type="primary"
							circle
							:icon="Promotion"
							:loading="sending"
							aria-label="发送消息"
							title="发送消息"
							@click="publish"
						/>
					</div>
				</section>
			</section>
		</main>
		<el-dialog v-model="subscriptionEditDialog" title="编辑订阅" width="420px" destroy-on-close>
			<el-form label-position="top">
				<el-form-item label="Topic 过滤器">
					<el-input v-model="subscriptionForm.topic" placeholder="例如 devices/+/status 或 #" />
				</el-form-item>
				<el-form-item label="QoS">
					<el-select v-model="subscriptionForm.qos">
						<el-option v-for="value in [0, 1, 2]" :key="value" :label="`QoS ${value}`" :value="value" />
					</el-select>
				</el-form-item>
				<el-form-item label="别名">
					<el-input v-model="subscriptionForm.alias" placeholder="可选，用于在列表中显示" maxlength="64" />
				</el-form-item>
			</el-form>
			<template #footer>
				<el-button @click="subscriptionEditDialog = false">取消</el-button>
				<el-button type="primary" @click="saveSubscriptionEdit">保存</el-button>
			</template>
		</el-dialog>
	</section>
	<Teleport to="body">
		<div
			v-if="contextMenu.visible && contextMenu.subscription"
			class="topic-context-menu"
			:style="{ left: `${contextMenu.x}px`, top: `${contextMenu.y}px` }"
			role="menu"
			@click.stop
		>
			<button type="button" role="menuitem" @click="editSubscription">
				<el-icon><Edit /></el-icon><span>编辑</span>
			</button>
			<button type="button" role="menuitem" @click="toggleSubscription">
				<el-icon><component :is="contextMenu.subscription.enabled ? CircleClose : Check" /></el-icon>
				<span>{{ contextMenu.subscription.enabled ? '禁用' : '启用' }}</span>
			</button>
		</div>
	</Teleport>
</template>

<style scoped>
.mqtt-page {
	--mqtt-bg: var(--bg-soft);
	--mqtt-panel: var(--bg-panel);
	--mqtt-input: #fff;
	--mqtt-line: var(--line);
	--mqtt-text: var(--text);
	--mqtt-muted: var(--text-soft);
	--mqtt-accent: var(--primary);
	height: 100%;
	min-width: 0;
	overflow: hidden;
	border: 1px solid var(--mqtt-line);
	border-radius: 18px;
	background: var(--mqtt-bg);
	color: var(--mqtt-text);
	font-size: 11px;
}
.message-workspace {
	display: grid;
	width: 100%;
	height: 100%;
	min-width: 0;
	min-height: 0;
}
.subscription-panel {
	display: flex;
	min-width: 0;
	min-height: 0;
	flex-direction: column;
	gap: 12px;
	padding: 22px;
	overflow: auto;
	background: var(--mqtt-panel);
}
.topic-actions {
	display: grid;
	grid-template-columns: minmax(0, 1fr) minmax(0, 1fr);
	gap: 8px;
}
.topic-actions :deep(.el-button),
.topic-actions :deep(.el-select) {
	width: 100%;
	min-width: 0;
}
.topic-actions :deep(.el-button) {
	height: 46px;
	margin: 0;
	border: 0;
	border-radius: 12px;
	background: var(--mqtt-accent);
	color: #fff;
	font-weight: 650;
}
.topic-actions :deep(.el-button:hover) { background: var(--primary-strong); }
.topic-actions :deep(.el-button .el-icon) { margin-right: 4px; }
.topic-actions :deep(.el-select__wrapper) {
	min-height: 46px;
	border-radius: 12px;
	background: #f7f8fc;
	box-shadow: inset 0 0 0 1px var(--mqtt-line);
}
.topic-actions :deep(.el-select__placeholder),
.topic-actions :deep(.el-select__selected-item) { color: var(--mqtt-text); }
.subscription-panel > :deep(.el-input) { flex: none; }
.subscription-panel :deep(.el-input__wrapper),
.publish-topic :deep(.el-input__wrapper),
.publish-options :deep(.el-select__wrapper) {
	min-height: 52px;
	border-radius: 12px;
	background: var(--mqtt-input);
	box-shadow: inset 0 0 0 1px var(--mqtt-line);
}
.mqtt-page :deep(.el-input__inner),
.mqtt-page :deep(.el-textarea__inner),
.mqtt-page :deep(.el-select__selected-item) { color: var(--mqtt-text); }
.subscriptions { display: flex; flex-direction: column; gap: 6px; }
.subscription-empty { padding: 38px 8px; color: var(--mqtt-muted); text-align: center; }
.subscription-item {
	display: flex;
	min-height: 38px;
	align-items: center;
	justify-content: space-between;
	gap: 8px;
	padding: 0 8px 0 12px;
	border-radius: 8px;
	color: var(--mqtt-text);
}
.subscription-status {
	width: 7px;
	height: 7px;
	flex: none;
	border-radius: 50%;
	background: var(--success);
}
.subscription-item.is-disabled .subscription-status { background: var(--text-muted); opacity: 1; }
.subscription-copy { display: flex; min-width: 0; flex: 1; flex-direction: column; gap: 2px; }
.subscription-name { overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.subscription-copy small { overflow: hidden; color: var(--mqtt-muted); font-size: 10px; text-overflow: ellipsis; white-space: nowrap; }
.subscription-item :deep(.el-tag) { flex: none; }
.subscription-item:hover { background: #f3f5f9; }
.subscription-item.is-disabled { color: var(--text-muted); background: #f7f8fb; }
.subscription-item.is-disabled span { opacity: .65; }
.subscription-item span { min-width: 0; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.subscription-item :deep(.el-button) { color: var(--mqtt-muted); }
.topic-context-menu {
	position: fixed;
	z-index: 3000;
	min-width: 144px;
	padding: 4px;
	border: 1px solid var(--line);
	border-radius: 6px;
	background: #fff;
	box-shadow: 0 6px 18px rgba(24, 34, 53, .16);
}
.topic-context-menu button {
	display: flex;
	width: 100%;
	min-height: 32px;
	align-items: center;
	gap: 8px;
	padding: 0 8px;
	border: 0;
	border-radius: 4px;
	background: transparent;
	color: var(--text);
	font: inherit;
	text-align: left;
	cursor: pointer;
}
.topic-context-menu button:hover { background: var(--bg-soft); color: var(--primary); }
.splitter {
	position: relative;
	cursor: col-resize;
	background: #e6eaf2;
	touch-action: none;
}
.splitter::after {
	position: absolute;
	top: 16px;
	bottom: 16px;
	left: 2px;
	width: 3px;
	border-radius: 2px;
	background: #aebbd0;
	content: '';
}
.splitter:hover::after { background: var(--mqtt-accent); }
.message-panel {
	--mqtt-line: #c2cddd;
	display: flex;
	min-width: 0;
	min-height: 0;
	flex-direction: column;
	border: 1px solid #aebdd2;
	border-right-width: 2px;
	border-radius: 12px;
	overflow: hidden;
}
.feed-toolbar {
	display: flex;
	height: 60px;
	flex: none;
	align-items: center;
	justify-content: space-between;
	gap: 12px;
	padding: 0 20px;
	border-bottom: 1px solid var(--mqtt-line);
}
.format-select { width: 130px; }
.feed-toolbar :deep(.el-select__wrapper),
.publish-options :deep(.el-select__wrapper) {
	min-height: 46px;
	border-radius: 12px;
	background: var(--mqtt-input);
	box-shadow: inset 0 0 0 1px var(--mqtt-line);
}
.message-tabs { display: flex; height: 100%; align-items: stretch; gap: 20px; }
.message-tabs button {
	position: relative;
	padding: 0 2px;
	border: 0;
	background: transparent;
	color: var(--mqtt-muted);
	font: inherit;
	cursor: pointer;
}
.message-tabs button.active { color: var(--mqtt-accent); }
.message-tabs button.active::after {
	position: absolute;
	right: 0;
	bottom: 10px;
	left: 0;
	height: 2px;
	background: var(--mqtt-accent);
	content: '';
}
.message-feed { min-height: 0; flex: 1; overflow: auto; padding: 18px 20px; }
.feed-empty { padding-top: 64px; color: var(--mqtt-muted); text-align: center; font-size: 12px; }
.message-item {
	width: fit-content;
	max-width: 85%;
	margin: 0 auto 10px 0;
	padding: 12px 14px;
	border: 1px solid var(--mqtt-line);
	border-radius: 10px;
	background: var(--mqtt-input);
}
.message-item.is-sent { margin-right: 0; margin-left: auto; border-color: #d8def6; background: var(--primary-soft); }
.message-meta { display: flex; align-items: center; gap: 10px; color: var(--mqtt-muted); font-size: 10px; }
.message-meta .topic-label { margin-right: -6px; color: var(--mqtt-muted); }
.message-meta strong { color: var(--mqtt-text); font-size: 11px; font-weight: 600; }
.message-meta time { margin-left: auto; }
.message-item pre { margin: 10px 0 0; color: var(--mqtt-text); font: 11px/1.55 ui-monospace, SFMono-Regular, Menlo, monospace; white-space: pre-wrap; word-break: break-word; }
.horizontal-splitter {
	position: relative;
	height: 7px;
	flex: none;
	cursor: row-resize;
	background: #e6eaf2;
	touch-action: none;
}
.horizontal-splitter::after {
	position: absolute;
	top: 2px;
	right: 20px;
	left: 20px;
	height: 3px;
	border-radius: 2px;
	background: #aebbd0;
	content: '';
}
.horizontal-splitter:hover::after { background: var(--mqtt-accent); }
.publisher {
	display: flex;
	min-height: 170px;
	flex: none;
	flex-direction: column;
	gap: 6px;
	padding: 8px 20px 10px;
	background: #f8f9fc;
}
.publish-options { display: flex; min-height: 38px; align-items: center; gap: 8px; }
.publish-options .el-select { width: 126px; }
.publish-options .el-select:nth-child(2) { width: 106px; }
.publish-options :deep(.el-checkbox) { margin-left: 2px; }
.publish-options :deep(.el-checkbox__label) { color: var(--mqtt-muted); }
.publish-options :deep(.el-select__wrapper) { min-height: 38px; }
.publish-topic :deep(.el-input__wrapper) { min-height: 40px; }
.payload-row { display: flex; min-height: 82px; flex: 1; align-items: flex-end; gap: 8px; }
.payload-row :deep(.el-textarea) { height: 100%; min-width: 0; flex: 1; }
.payload-row :deep(.el-textarea__inner) {
	height: 100% !important;
	min-height: 82px !important;
	border: 1px solid #c2cddd;
	border-radius: 14px;
	background: var(--mqtt-input);
	box-shadow: none;
	line-height: 1.5;
}
.payload-row :deep(.el-textarea__inner:focus) { border-color: var(--mqtt-accent); box-shadow: 0 0 0 2px var(--primary-soft); }
.send-button { width: 46px; height: 46px; flex: none; border: 0; background: var(--mqtt-accent); color: #fff; font-size: 16px; }
.send-button:hover { background: var(--primary-strong); }
.mqtt-page :deep(.el-input__wrapper.is-focus),
.mqtt-page :deep(.el-select__wrapper.is-focused) { box-shadow: inset 0 0 0 1px var(--mqtt-accent); }
@media (max-width: 760px) {
	.mqtt-page { overflow: auto; border-radius: 12px; }
	.message-workspace { grid-template-columns: minmax(0, 1fr) !important; grid-template-rows: 190px minmax(0, 1fr); }
	.subscription-panel { padding: 12px; }
	.splitter { display: none; }
	.message-panel { min-height: 520px; }
	.feed-toolbar { padding: 0 12px; }
	.message-tabs { gap: 12px; }
	.message-feed { padding: 12px; }
	.publisher { padding: 12px; }
	.publish-options { gap: 6px; }
	.publish-options .el-select { width: 112px; }
	.publish-options .el-select:nth-child(2) { width: 94px; }
}
</style>
