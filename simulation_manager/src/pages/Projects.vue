<template>
  <div style="padding: 20px;">
    <!-- 顶部操作栏 -->
    <el-row justify="space-between" align="middle" class="mb-4">
      <el-col :span="8">
        <el-input v-model="search" placeholder="搜索项目" prefix-icon="el-icon-search" />
      </el-col>
      <el-col :span="8" style="text-align: right;">
        <el-button type="primary" @click="openCreateDialog">新建项目</el-button>
        <el-button @click="importProject">导入项目</el-button>
      </el-col>
    </el-row>

    <!-- 项目列表表格 -->
    <el-table :data="filteredProjects" style="width: 100%">
      <el-table-column prop="name" label="项目名称" />
      <el-table-column prop="status" label="状态" />
      <el-table-column prop="createdAt" label="创建时间" />
      <el-table-column prop="updatedAt" label="更新时间" />
      <el-table-column label="操作" width="240">
        <template #default="scope">
          <el-button type="primary" size="small">编辑</el-button>
          <el-button size="small">仿真</el-button>
          <el-button size="small" type="danger">删除</el-button>
        </template>
      </el-table-column>
    </el-table>
    <GenerateProject v-model="showGenerateProject" />
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import GenerateProject from '@/components/GenerateProject.vue'
const showGenerateProject = ref(false)
const search = ref('')
const projects = ref([
  { name: '交叉口仿真A', status: '已保存', createdAt: '2025-05-01', updatedAt: '2025-05-09' },
  { name: '公交信号优先', status: '运行中', createdAt: '2025-04-22', updatedAt: '2025-05-08' },
])

const filteredProjects = computed(() =>
  projects.value.filter(p => p.name.includes(search.value))
)

function openCreateDialog() {
  showGenerateProject.value = true
  console.log('新建项目弹窗')
}

function importProject() {
  console.log('导入项目')
}
</script>
