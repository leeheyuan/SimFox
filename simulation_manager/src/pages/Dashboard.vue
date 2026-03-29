<template>
  <el-container style="height: 100vh;">  
    <el-container> 
      <el-main>
        <el-row :gutter="20">
          <el-col :span="6">
            <el-card>
              <div>项目总数</div>
              <h2>{{ stats.projectCount }}</h2>
            </el-card>
          </el-col>
          <el-col :span="6">
            <el-card>
              <div>仿真次数</div>
              <h2>{{ stats.simulationCount }}</h2>
            </el-card>
          </el-col>
          <el-col :span="6">
            <el-card>
              <div>地图数量</div>
              <h2>{{ stats.mapCount }}</h2>
            </el-card>
          </el-col>
        </el-row>

        <el-divider>最近打开的项目</el-divider>

        <el-table :data="recentProjects" style="width: 100%">
          <el-table-column prop="name" label="项目名" />
          <el-table-column prop="updated" label="更新时间" />
          <el-table-column prop="status" label="状态" />
          <el-table-column label="操作">
            <template #default="scope">
              <el-button size="small" type="primary" @click="handleEnterProject(scope.row)">进入</el-button>
              <el-button size="small" @click="handleSimulate(scope.row)">仿真</el-button>
            </template>
          </el-table-column>
        </el-table>
      </el-main>
    </el-container>
  </el-container>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { getOverview } from '@/api/user'
import { useRouter } from 'vue-router'

const router = useRouter()

const stats = ref({
  projectCount: 0,
  simulationCount: 0,
  mapCount: 0
})


const recentProjects = ref([])

// <script setup> 写法（Composition API）
const handleEnterProject = (row) => {
  console.log('进入项目：', row)
  // 路由跳转示例
  // router.push(`/project/${row.id}`)
  router.push('/sim-monitor')
}

const handleSimulate = (row) => {
  console.log('开始仿真：', row)
  // ElMessage.success('开始仿真...')
}

onMounted(async () => {
  const data = await getOverview()
  if(data == null) { 

  } else {
    stats.value.mapCount = data.map_count
    stats.value.projectCount = data.project_count
    stats.value.simulationCount = data.simulation_count 
    recentProjects.value = data.recent_projects
  } 
})



</script>
