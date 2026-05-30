<template>
  <section class="page">
    <div class="page-header">
      <div>
        <h2>Results</h2>
        <p>Browse completed simulation runs and open their generated logs.</p>
      </div>
      <el-button type="primary" :loading="loading" @click="refreshResults">Refresh</el-button>
    </div>

    <el-table v-loading="loading" :data="rows" style="width: 100%">
      <el-table-column prop="taskId" label="Task ID" width="120" />
      <el-table-column prop="project" label="Project" />
      <el-table-column prop="status" label="Status" width="130" />
      <el-table-column prop="summary" label="Summary" min-width="260" />
      <el-table-column prop="generatedAt" label="Generated At" width="180" />
      <el-table-column label="Actions" width="220">
        <template #default="scope">
          <el-button size="small" @click="openDetails(scope.row.taskId)">View</el-button>
          <el-button
            size="small"
            type="primary"
            :disabled="!canOpenResult(scope.row.taskId)"
            @click="openResult(scope.row.taskId)"
          >
            Download
          </el-button>
        </template>
      </el-table-column>
    </el-table>

    <el-dialog v-model="showDetailsDialog" title="Result Details" width="600px">
      <template v-if="selectedResult">
        <div class="meta-grid">
          <div><strong>Task ID:</strong> {{ selectedResult.taskId }}</div>
          <div><strong>Project:</strong> {{ selectedResult.projectName || `Project ${selectedResult.projectId}` }}</div>
          <div><strong>Status:</strong> {{ selectedResult.status }}</div>
          <div><strong>Generated:</strong> {{ formatDate(selectedResult.generatedAt) }}</div>
        </div>
        <el-form label-position="top">
          <el-form-item label="Summary">
            <el-input :model-value="describeResult(selectedResult)" type="textarea" :rows="3" readonly />
          </el-form-item>
          <el-form-item label="Log Path">
            <el-input :model-value="selectedResult.logUrl || 'No log path recorded'" readonly />
          </el-form-item>
          <el-form-item label="Last Error">
            <el-input :model-value="selectedResult.lastError || 'No error recorded'" type="textarea" :rows="4" readonly />
          </el-form-item>
        </el-form>
      </template>
      <template #footer>
        <el-button @click="showDetailsDialog = false">Close</el-button>
        <el-button
          v-if="selectedResult && canOpenResult(selectedResult.taskId)"
          type="primary"
          @click="openResult(selectedResult.taskId)"
        >
          Open Log
        </el-button>
      </template>
    </el-dialog>
  </section>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { listResults, type ResultListItem } from '@/api/simulation'

const loading = ref(false)
const results = ref<ResultListItem[]>([])
const showDetailsDialog = ref(false)
const selectedResult = ref<ResultListItem | null>(null)

const rows = computed(() =>
  results.value.map((result) => ({
    taskId: result.taskId,
    project: result.projectName || `Project ${result.projectId}`,
    status: result.status,
    summary: describeResult(result),
    generatedAt: formatDate(result.generatedAt),
  })),
)

onMounted(async () => {
  await refreshResults()
})

async function refreshResults() {
  loading.value = true
  try {
    const data = await listResults()
    results.value = data.results ?? []
  } catch (_error) {
    ElMessage.error('Failed to load results')
  } finally {
    loading.value = false
  }
}

function openDetails(taskId: number) {
  selectedResult.value = results.value.find((item) => item.taskId === taskId) ?? null
  showDetailsDialog.value = true
}

function canOpenResult(taskId: number) {
  const result = results.value.find((item) => item.taskId === taskId)
  return Boolean(result?.logUrl && window.simfox?.openPath)
}

async function openResult(taskId: number) {
  const result = results.value.find((item) => item.taskId === taskId)
  if (!result?.logUrl || !window.simfox?.openPath) {
    return
  }

  try {
    await window.simfox.openPath(result.logUrl)
  } catch (_error) {
    ElMessage.error('Failed to open result log')
  }
}

function describeResult(result: ResultListItem) {
  if (result.status === 'succeeded') {
    return result.logUrl ? 'Simulation completed successfully. Log file is available.' : 'Simulation completed successfully.'
  }
  if (result.status === 'failed') {
    return result.lastError || 'Simulation failed.'
  }
  if (result.status === 'cancelled') {
    return result.lastError || 'Simulation was cancelled.'
  }
  return 'Result available.'
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

.meta-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 8px 16px;
  margin-bottom: 16px;
  color: #334155;
}

h2 {
  margin: 0 0 6px;
}

p {
  margin: 0;
  color: #64748b;
}
</style>
