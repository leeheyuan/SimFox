<template>
  <section class="page">
    <div class="page-header">
      <div>
        <h2>Task Queue</h2>
        <p>Submit, monitor, cancel, and retry SUMO jobs executed by private cloud workers.</p>
      </div>
      <el-button type="primary">Submit Task</el-button>
    </div>

    <el-row :gutter="16" class="stats">
      <el-col :span="6">
        <el-card shadow="never">
          <div class="label">Queued</div>
          <strong>2</strong>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="never">
          <div class="label">Running</div>
          <strong>1</strong>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="never">
          <div class="label">Succeeded</div>
          <strong>8</strong>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="never">
          <div class="label">Failed</div>
          <strong>0</strong>
        </el-card>
      </el-col>
    </el-row>

    <el-table :data="tasks" style="width: 100%">
      <el-table-column prop="id" label="Task ID" width="120" />
      <el-table-column prop="project" label="Project" />
      <el-table-column prop="status" label="Status" width="140" />
      <el-table-column prop="worker" label="Worker" width="160" />
      <el-table-column prop="progress" label="Progress" width="180">
        <template #default="scope">
          <el-progress :percentage="scope.row.progress" />
        </template>
      </el-table-column>
      <el-table-column prop="submittedAt" label="Submitted At" width="180" />
      <el-table-column label="Actions" width="260">
        <template #default>
          <el-button size="small">Logs</el-button>
          <el-button size="small">Cancel</el-button>
          <el-button size="small" type="primary">Retry</el-button>
        </template>
      </el-table-column>
    </el-table>
  </section>
</template>

<script setup lang="ts">
const tasks = [
  {
    id: 1001,
    project: 'Intersection baseline',
    status: 'running',
    worker: 'worker-a',
    progress: 42,
    submittedAt: '2026-05-24 09:10',
  },
  {
    id: 1002,
    project: 'Bus priority experiment',
    status: 'queued',
    worker: '-',
    progress: 0,
    submittedAt: '2026-05-24 09:12',
  },
]
</script>

<style scoped>
.page {
  padding: 20px;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
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
