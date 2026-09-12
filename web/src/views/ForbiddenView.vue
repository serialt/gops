<script setup lang="ts">
import { useRouter } from 'vue-router'
const router = useRouter()
</script>

<template>
  <div class="forbidden">
    <el-result icon="warning" title="403" sub-title="当前账号没有访问此页面的权限">
      <template #extra><el-button type="primary" @click="router.push('/')">返回首页</el-button></template>
    </el-result>
  </div>
</template>

<style scoped>
.forbidden {
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
}
</style>
