<template>
  <section class="page">
    <div class="page-header">
      <div>
        <h2>Cluster Nodes</h2>
        <p>Track worker agents that execute SUMO jobs in the private cloud.</p>
      </div>
      <el-button type="primary" @click="showRegisterDialog = true">Register Worker</el-button>
    </div>

    <el-row :gutter="16" class="stats">
      <el-col :span="6">
        <el-card shadow="never">
          <div class="label">Online Workers</div>
          <strong>{{ stats.onlineWorkers }}</strong>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="never">
          <div class="label">Running Tasks</div>
          <strong>{{ stats.runningTasks }}</strong>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="never">
          <div class="label">Queued Tasks</div>
          <strong>{{ stats.queuedTasks }}</strong>
        </el-card>
      </el-col>
    </el-row>

    <el-table v-loading="loading" :data="rows" style="width: 100%">
      <el-table-column prop="name" label="Worker" />
      <el-table-column prop="status" label="Status" width="130" />
      <el-table-column prop="address" label="Address" />
      <el-table-column prop="capacity" label="Capacity" width="140" />
      <el-table-column prop="lastHeartbeat" label="Last Heartbeat" width="180" />
    </el-table>

    <el-dialog v-model="showRegisterDialog" title="Register Worker" width="480px">
      <el-form label-position="top" @submit.prevent>
        <el-form-item label="Worker Name" required>
          <el-input v-model="registerForm.name" placeholder="worker-a" />
        </el-form-item>
        <el-form-item label="Address">
          <el-input v-model="registerForm.address" placeholder="10.0.0.12" />
        </el-form-item>
        <el-form-item label="Queue Name">
          <el-input v-model="registerForm.queueName" placeholder="default" />
        </el-form-item>
        <el-form-item label="Max Concurrency">
          <el-input-number v-model="registerForm.maxConcurrency" :min="1" :max="64" />
        </el-form-item>
        <el-form-item label="Labels JSON">
          <el-input
            v-model="registerForm.labelsJson"
            type="textarea"
            :rows="4"
            placeholder='{"region":"cn-east","runtime":"sumo"}'
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="resetRegisterDialog">Cancel</el-button>
        <el-button type="primary" :loading="registering" @click="handleRegisterWorker">Register</el-button>
      </template>
    </el-dialog>
  </section>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { ElMessage } from 'element-plus'
import {
  listTasks,
  listWorkers,
  registerWorker,
  type TaskListItem,
  type WorkerListItem,
} from '@/api/simulation'

const loading = ref(false)
const registering = ref(false)
const showRegisterDialog = ref(false)
const workers = ref<WorkerListItem[]>([])
const tasks = ref<TaskListItem[]>([])
const registerForm = ref({
  name: '',
  address: '',
  queueName: 'default',
  labelsJson: '',
  maxConcurrency: 1,
})

const rows = computed(() =>
  workers.value.map((worker) => ({
    name: worker.name,
    status: worker.status,
    address: worker.address || '-',
    capacity: `${worker.runningTasks} / ${worker.maxConcurrency}`,
    lastHeartbeat: formatDate(worker.lastHeartbeat || worker.updatedAt),
  })),
)

const stats = computed(() => {
  const queuedTasks = tasks.value.filter((task) => task.status === 'queued' || task.status === 'dispatching').length
  const runningTasks = workers.value.reduce((sum, worker) => sum + worker.runningTasks, 0)
  const onlineWorkers = workers.value.filter((worker) =>
    ['online', 'busy', 'draining'].includes(worker.status),
  ).length

  return {
    onlineWorkers,
    runningTasks,
    queuedTasks,
  }
})

onMounted(async () => {
  await refreshCluster()
})

async function refreshCluster() {
  loading.value = true
  try {
    const [workerData, taskData] = await Promise.all([listWorkers(), listTasks()])
    workers.value = workerData.workers ?? []
    tasks.value = taskData.tasks ?? []
  } catch (_error) {
    ElMessage.error('Failed to load cluster data')
  } finally {
    loading.value = false
  }
}

async function handleRegisterWorker() {
  if (!registerForm.value.name.trim()) {
    ElMessage.error('Worker name is required')
    return
  }

  if (registerForm.value.labelsJson.trim()) {
    try {
      JSON.parse(registerForm.value.labelsJson)
    } catch {
      ElMessage.error('Labels JSON is invalid')
      return
    }
  }

  registering.value = true
  try {
    await registerWorker({
      name: registerForm.value.name.trim(),
      address: registerForm.value.address.trim(),
      queueName: registerForm.value.queueName.trim() || 'default',
      labelsJson: registerForm.value.labelsJson.trim(),
      maxConcurrency: registerForm.value.maxConcurrency,
    })
    ElMessage.success('Worker registered successfully')
    resetRegisterDialog()
    await refreshCluster()
  } catch (_error) {
    ElMessage.error('Failed to register worker')
  } finally {
    registering.value = false
  }
}

function resetRegisterDialog() {
  showRegisterDialog.value = false
  registerForm.value = {
    name: '',
    address: '',
    queueName: 'default',
    labelsJson: '',
    maxConcurrency: 1,
  }
}

function formatDate(value?: string | null) {
  if (!value) {
    return '-'
  }

  const date = new Date(value)
  if (Number.isNaN(date.getTime())) {
    return value
  }

  const year = date.getFullYear()
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')
  const hours = String(date.getHours()).padStart(2, '0')
  const minutes = String(date.getMinutes()).padStart(2, '0')
  return `${year}-${month}-${day} ${hours}:${minutes}`
}
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
