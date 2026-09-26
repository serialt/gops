<script setup lang="ts">
import { onBeforeUnmount, onMounted, ref, watch } from 'vue'
import * as monaco from 'monaco-editor'
import EditorWorker from 'monaco-editor/editor/editor.worker?worker'

  ; (self as typeof self & { MonacoEnvironment?: unknown }).MonacoEnvironment = {
    getWorker: () => new EditorWorker(),
  }

const props = withDefaults(defineProps<{ modelValue: string; minHeight?: string }>(), { minHeight: '500px' })
const emit = defineEmits<{ 'update:modelValue': [value: string] }>()
const host = ref<HTMLDivElement>()
let editor: monaco.editor.IStandaloneCodeEditor | undefined
let model: monaco.editor.ITextModel | undefined
let contentSubscription: monaco.IDisposable | undefined

onMounted(() => {
  if (!host.value) return
  model = monaco.editor.createModel(props.modelValue, 'yaml')
  monaco.editor.defineTheme('kubernetes-yaml-light', {
    base: 'vs',
    inherit: true,
    rules: [
      { token: 'type', foreground: '6F42C1', fontStyle: 'bold' },
      { token: 'keyword', foreground: '6F42C1', fontStyle: 'bold' },
      { token: 'string', foreground: '238636' },
      { token: 'number', foreground: '0969DA' },
      { token: 'number.float', foreground: '0969DA' },
      { token: 'comment', foreground: '7D8590', fontStyle: 'italic' },
      { token: 'operators', foreground: 'CF5700' },
      { token: 'tag', foreground: 'CF222E' },
      { token: 'namespace', foreground: '8250DF' },
    ],
    colors: {
      'editor.background': '#FFFFFF',
      'editor.foreground': '#24292F',
      'editorLineNumber.foreground': '#9AA0AA',
      'editorLineNumber.activeForeground': '#635BFF',
      'editor.lineHighlightBackground': '#F7F7FF',
      'editorIndentGuide.background1': '#E6E8EE',
      'editorIndentGuide.activeBackground1': '#A9A4FF',
      'editor.selectionBackground': '#DCD9FF',
    },
  })
  editor = monaco.editor.create(host.value, {
    model,
    theme: 'kubernetes-yaml-light',
    automaticLayout: true,
    minimap: { enabled: true, maxColumn: 30 },
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
    scrollbar: { verticalScrollbarSize: 4, horizontalScrollbarSize: 4, useShadows: false },
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
.kubernetes-yaml-editor {
  width: 100%;
  min-width: 0;
  overflow: hidden;
  border: 1px solid var(--line);
  border-radius: 8px;
  background: #fff
}

.kubernetes-yaml-editor:focus-within {
  border-color: var(--primary);
  box-shadow: 0 0 0 1px color-mix(in srgb, var(--primary) 15%, transparent)
}
</style>
