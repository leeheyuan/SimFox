<template>
  <section class="page">
    <div class="page-header">
      <div>
        <h2>Overview</h2>
        <p>Private cloud simulation workload and cluster health.</p>
      </div>
    </div>

    <el-row :gutter="16" class="stats">
      <el-col :span="6">
        <el-card shadow="never">
          <div class="label">Projects</div>
          <strong>{{ stats.projectCount }}</strong>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="never">
          <div class="label">Tasks</div>
          <strong>{{ stats.simulationCount }}</strong>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="never">
          <div class="label">Input Files</div>
          <strong>{{ stats.mapCount }}</strong>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="never">
          <div class="label">Workers</div>
          <strong>0</strong>
        </el-card>
      </el-col>
    </el-row>

    <el-table :data="recentProjects" style="width: 100%">
      <el-table-column prop="name" label="Project" />
      <el-table-column prop="updated" label="Updated At" />
      <el-table-column prop="status" label="Status" width="140" />
      <el-table-column label="Actions" width="220">
        <template #default="scope">
          <el-button size="small" type="primary" @click="openTasks(scope.row)">Tasks</el-button>
          <el-button size="small" @click="openResults">Results</el-button>
        </template>
      </el-table-column>
    </el-table>
  </section>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { getOverview, type OverviewProjectItem, type OverviewResponse } from '@/api/user'
import { useRouter } from 'vue-router'

const router = useRouter()

const stats = ref({
  projectCount: 0,
  simulationCount: 0,
  mapCount: 0,
})

const recentProjects = ref<OverviewProjectItem[]>([])

const openTasks = (_row: OverviewProjectItem) => {
  router.push('/tasks')
}

const openResults = () => {
  router.push('/results')
}

onMounted(async () => {
  const data: OverviewResponse | null = await getOverview().catch(() => null)
  if (data == null) {
    return
  }

  stats.value.mapCount = data.map_count
  stats.value.projectCount = data.project_count
  stats.value.simulationCount = data.simulation_count
  recentProjects.value = data.recent_projects
})
</script>

<style scoped>
.page {
  padding: 20px;
}

.page-header {
  margin-bottom: 16px;
}

.stats {
  margin-bottom: 16px;
}

.label {
  color: #64748b;
  margin-bottom: 8px;
}

strong {
  font-size: 26px;
}

h2 {
  margin: 0 0 6px;
}

p {
  margin: 0;
  color: #64748b;
}
</style>
