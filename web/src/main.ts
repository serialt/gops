import ElementPlus from 'element-plus'
import 'element-plus/dist/index.css'
import { createPinia } from 'pinia'
import TDesign from 'tdesign-vue-next'
import 'tdesign-vue-next/es/style/index.css'
import { createApp } from 'vue'
import App from './App.vue'
import { permissionDirective } from './directives/permission'
import router from './router'
import './styles/global.css'

const app = createApp(App)
app.use(createPinia())
app.use(router)
app.use(ElementPlus)
app.use(TDesign)
app.directive('permission', permissionDirective)
app.mount('#app')
