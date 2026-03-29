<template>
  <div style="padding: 20px;">
    <h2>仿真运行</h2>
    <el-alert title="提示" type="info" class="mb-4">
      选择一个项目并运行仿真。支持实时查看仿真状态。
    </el-alert>
    <el-select v-model="selectedProject" placeholder="选择项目" style="width: 300px;" class="mb-4">
      <el-option v-for="item in projects" :key="item.name" :label="item.name" :value="item.name" />
    </el-select>
    <el-button type="primary" @click="runSimulation">开始仿真</el-button>

    <el-divider />
    <div v-if="simulating">
      <p>仿真中：{{ selectedProject }}</p>
      <el-progress :percentage="progress" :status="progress === 100 ? 'success' : 'active'" />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'

const selectedProject = ref('')
const simulating = ref(false)
const progress = ref(0)

const projects = [
  { name: '交叉口仿真A' },
  { name: '公交信号优先' }
]

function runSimulation() {
  simulating.value = true
  progress.value = 0
  const timer = setInterval(() => {
    progress.value += 10
    if (progress.value >= 100) {
      clearInterval(timer)
    }
  }, 500)
}
</script>
