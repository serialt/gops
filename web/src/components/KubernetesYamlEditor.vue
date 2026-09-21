<script setup lang="ts">
import { onBeforeUnmount, onMounted, ref, watch } from 'vue'
import * as monaco from 'monaco-editor'
import EditorWorker from 'monaco-editor/editor/editor.worker?worker'

;(self as typeof self & { MonacoEnvironment?: unknown }).MonacoEnvironment = {
  getWorker: () => new EditorWorker(),
}

const props = withDefaults(defineProps<{ modelValue: string; minHeight?: string }>(), { minHeight: '440px' })
const emit = defineEmits<{ 'update:modelValue': [value: string] }>()
const host = ref<HTMLDivElement>()
let editor: monaco.editor.IStandaloneCodeEditor | undefined
let model: monaco.editor.ITextModel | undefined
let contentSubscription: monaco.IDisposable | undefined

onMounted(() => {
  if (!host.value) return
  model = monaco.editor.createModel(props.modelValue, 'yaml')
  editor = monaco.editor.create(host.value, {
    model,
    automaticLayout: true,
    minimap: { enabled: true },
    lineNumbers: 'on',
    glyphMargin: false,
    folding: true,
    guides: { indentation: true, highlightActiveIndentation: true, bracketPairs: true },
    renderLineHighlight: 'all',
    renderWhitespace: 'selection',
    bracketPairColorization: { enabled: true },
    tabSize: 2,
    insertSpaces: true,
    detectIndentation: false,
    wordWrap: 'off',
    scrollBeyondLastLine: false,
    padding: { top: 10, bottom: 10 },
    fontFamily: 'ui-monospace, SFMono-Regular, Menlo, Consolas, monospace',
    fontSize: 12,
    lineHeight: 20,
    contextmenu: true,
    smoothScrolling: true,
  })
  contentSubscription = model.onDidChangeContent(() => emit('update:modelValue', model?.getValue() ?? ''))
})

watch(() => props.modelValue, (value) => {
  if (!model || value === model.getValue()) return
  model.setValue(value)
})

onBeforeUnmount(() => {
  contentSubscription?.dispose()
  editor?.dispose()
  model?.dispose()
})
</script>

<template>
  <div ref="host" class="kubernetes-yaml-editor" :style="{ minHeight }" />
</template>

<style scoped>
.kubernetes-yaml-editor{width:100%;min-width:0;overflow:hidden;border:1px solid var(--line);border-radius:8px;background:#fff}.kubernetes-yaml-editor:focus-within{border-color:var(--primary);box-shadow:0 0 0 1px color-mix(in srgb,var(--primary) 15%,transparent)}
</style>
