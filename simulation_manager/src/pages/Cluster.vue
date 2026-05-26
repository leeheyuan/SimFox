<template>
  <section class="page">
    <div class="page-header">
      <div>
        <h2>Cluster Nodes</h2>
        <p>Track worker agents that execute SUMO jobs in the private cloud.</p>
      </div>
      <el-button type="primary">Register Worker</el-button>
    </div>

    <el-row :gutter="16" class="stats">
      <el-col :span="6">
        <el-card shadow="never">
          <div class="label">Online Workers</div>
          <strong>0</strong>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="never">
          <div class="label">Running Tasks</div>
          <strong>0</strong>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="never">
          <div class="label">Queued Tasks</div>
          <strong>0</strong>
        </el-card>
      </el-col>
    </el-row>

    <el-table :data="workers" style="width: 100%">
      <el-table-column prop="name" label="Worker" />
      <el-table-column prop="status" label="Status" width="130" />
      <el-table-column prop="address" label="Address" />
      <el-table-column prop="capacity" label="Capacity" width="140" />
      <el-table-column prop="lastHeartbeat" label="Last Heartbeat" width="180" />
    </el-table>
  </section>
</template>

<script setup lang="ts">
const workers = [
  {
    name: 'worker-a',
    status: 'offline',
    address: '10.0.0.12',
    capacity: '0 / 4',
    lastHeartbeat: '-',
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
