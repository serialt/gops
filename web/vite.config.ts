import vue from '@vitejs/plugin-vue'
import { fileURLToPath, URL } from 'node:url'
import { defineConfig } from 'vite'
import vueDevTools from 'vite-plugin-vue-devtools'

export default defineConfig(({ mode }) => {
  const isDev = mode === 'development'

  return {
    plugins: [
      vue(),
      ...(isDev ? [vueDevTools({ launchEditor: 'code' })] : []),
    ],
    resolve: {
      alias: {
        '@': fileURLToPath(new URL('./src', import.meta.url)),
      },
    },
    server: {
      port: 5173,
      proxy: {
        '/api': {
          target: 'http://localhost:9876',
          changeOrigin: true,
          ws: true,
        },
      },
    },
  }
})
