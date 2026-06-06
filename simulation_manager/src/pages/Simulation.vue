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
          <strong>{{ stats.queued }}</strong>
      </el-card>
    </el-col>
    <el-col :span="6">
      <el-card shadow="never">
        <div class="label">Running</div>
          <strong>{{ stats.running }}</strong>
      </el-card>
    </el-col>
    <el-col :span="6">
      <el-card shadow="never">
        <div class="label">Succeeded</div>
          <strong>{{ stats.succeeded }}</strong>
      </el-card>
    </el-col>
    <el-col :span="6">
      <el-card shadow="never">
        <div class="label">Failed</div>
          <strong>{{ stats.failed }}</strong>
      </el-card>
    </el-col>
  </el-row>

    <el-table v-loading="loading" :data="rows" style="width: 100%">
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
        <template #default="scope">
          <el-button size="small" @click="openLogs(scope.row.id)">Logs</el-button>
          <el-button
            size="small"
            :disabled="!canCancelTask(scope.row.id)"
            @click="handleCancel(scope.row.id)"
          >
            Cancel
          </el-button>
          <el-button
            size="small"
            type="primary"
            :disabled="!canRetryTask(scope.row.id)"
            @click="handleRetry(scope.row.id)"
          >
            Retry
          </el-button>
        </template>
      </el-table-column>
    </el-table>

    <el-dialog v-model="showLogsDialog" title="Task Logs" width="560px">
      <template v-if="selectedTask">
        <div class="log-meta">
          <div><strong>Task ID:</strong> {{ selectedTask.id }}</div>
          <div><strong>Project:</strong> {{ selectedTask.projectName || `Project ${selectedTask.projectId}` }}</div>
          <div><strong>Status:</strong> {{ selectedTask.status }}</div>
          <div><strong>Updated:</strong> {{ formatDate(selectedTask.endedAt || selectedTask.startedAt || selectedTask.submittedAt || selectedTask.createdAt) }}</div>
        </div>
        <el-form label-position="top">
          <el-form-item label="Last Error">
            <el-input
              :model-value="selectedTask.lastError || 'No error recorded'"
              type="textarea"
              :rows="4"
              readonly
            />
          </el-form-item>
          <el-form-item label="Log Path">
            <el-input :model-value="selectedTask.logUrl || 'No log path recorded'" readonly />
          </el-form-item>
        </el-form>
      </template>
      <template #footer>
        <el-button @click="showLogsDialog = false">Close</el-button>
        <el-button
          v-if="canOpenLogPath"
          type="primary"
          @click="openLogPath"
        >
          Open Log
        </el-button>
      </template>
    </el-dialog>
  </section>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { cancelTask, listTasks, retryTask, type TaskListItem } from '@/api/simulation'

const loading = ref(false)
const tasks = ref<TaskListItem[]>([])
const showLogsDialog = ref(false)
const selectedTask = ref<TaskListItem | null>(null)
const canOpenLogPath = computed(() => Boolean(selectedTask.value?.logUrl && window.simfox?.openPath))
let refreshTimer: number | null = null

const rows = computed(() =>
  tasks.value.map((task) => ({
    id: task.id,
    project: task.projectName || `Project ${task.projectId}`,
    status: task.status,
    worker: task.workerId ? `worker-${task.workerId}` : '-',
    progress: task.progress,
    submittedAt: formatDate(task.submittedAt || task.createdAt),
  })),
)

const stats = computed(() => {
  const summary = {
    queued: 0,
    running: 0,
    succeeded: 0,
    failed: 0,
  }

  for (const task of tasks.value) {
    if (task.status === 'queued' || task.status === 'dispatching') {
      summary.queued += 1
    } else if (task.status === 'running') {
      summary.running += 1
    } else if (task.status === 'succeeded') {
      summary.succeeded += 1
    } else if (task.status === 'failed' || task.status === 'cancelled') {
      summary.failed += 1
    }
  }

  return summary
})

onMounted(async () => {
  await refreshTasks()
  refreshTimer = window.setInterval(() => {
    void refreshTasks(true)
  }, 3000)
})

onBeforeUnmount(() => {
  if (refreshTimer !== null) {
    window.clearInterval(refreshTimer)
    refreshTimer = null
  }
})

async function refreshTasks(isBackground = false) {
  if (!isBackground) {
    loading.value = true
  }
  try {
    const data = await listTasks()
    tasks.value = data.tasks ?? []
  } catch (_error) {
    if (!isBackground) {
      ElMessage.error('Failed to load tasks')
    }
  } finally {
    if (!isBackground) {
      loading.value = false
    }
  }
}

function canCancelTask(taskId: number) {
  const task = tasks.value.find((item) => item.id === taskId)
  if (!task) {
    return false
  }

  return task.status === 'queued' || task.status === 'dispatching' || task.status === 'running'
}

async function handleCancel(taskId: number) {
  if (!canCancelTask(taskId)) {
    return
  }

  try {
    await cancelTask(taskId)
    ElMessage.success('Task cancelled')
    await refreshTasks()
  } catch (_error) {
    ElMessage.error('Failed to cancel task')
  }
}

function canRetryTask(taskId: number) {
  const task = tasks.value.find((item) => item.id === taskId)
  if (!task) {
    return false
  }

  return task.status === 'failed' || task.status === 'cancelled' || task.status === 'succeeded'
}

async function handleRetry(taskId: number) {
  if (!canRetryTask(taskId)) {
    return
  }

  try {
    await retryTask(taskId)
    ElMessage.success('Task queued for retry')
    await refreshTasks()
  } catch (_error) {
    ElMessage.error('Failed to retry task')
  }
}

function openLogs(taskId: number) {
  selectedTask.value = tasks.value.find((task) => task.id === taskId) ?? null
  showLogsDialog.value = true
}

async function openLogPath() {
  if (!selectedTask.value?.logUrl || !window.simfox?.openPath) {
    return
  }
  await window.simfox.openPath(selectedTask.value.logUrl)
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

.log-meta {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 8px 16px;
  margin-bottom: 16px;
  color: #334155;
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
