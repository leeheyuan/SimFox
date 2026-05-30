<template>
  <section class="page">
    <div class="page-header">
      <div>
        <h2>Projects</h2>
        <p>Manage simulation input packages before they are submitted to the cluster.</p>
      </div>
      <div class="actions">
        <el-button type="primary" :loading="importing" @click="openImportPicker">Import Project</el-button>
      </div>
    </div>

    <el-input v-model="search" class="search" placeholder="Search projects" clearable />

    <el-table v-loading="loading" :data="filteredProjects" style="width: 100%">
      <el-table-column prop="name" label="Project" />
      <el-table-column prop="status" label="Status" width="140" />
      <el-table-column prop="createdAt" label="Created At" width="180" />
      <el-table-column prop="updatedAt" label="Updated At" width="180" />
      <el-table-column label="Actions" width="260">
        <template #default="scope">
          <el-button size="small">Files</el-button>
          <el-button size="small" type="primary" @click="submitTask(scope.row)">Submit Task</el-button>
          <el-button size="small" type="danger" @click="removeProject(scope.row)">Delete</el-button>
        </template>
      </el-table-column>
    </el-table>

    <input
      ref="fileInput"
      class="hidden-input"
      type="file"
      accept=".sumocfg,.net.xml"
      @change="handleImportSelection"
    />

    <el-dialog v-model="showSubmitDialog" title="Submit Task" width="420px">
      <el-form label-position="top" @submit.prevent>
        <el-form-item label="Project">
          <el-input :model-value="selectedProject?.name ?? ''" disabled />
        </el-form-item>
        <el-form-item label="Simulation Time (seconds)" required>
          <el-input-number v-model="submitForm.simulationTime" :min="1" :max="86400" />
        </el-form-item>
        <el-form-item label="Speed" required>
          <el-input-number v-model="submitForm.speed" :min="0.1" :max="100" :step="0.1" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="closeSubmitDialog">Cancel</el-button>
        <el-button type="primary" :loading="submittingTask" @click="confirmSubmitTask">Submit</el-button>
      </template>
    </el-dialog>
  </section>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  deleteProject,
  enqueueProjectTask,
  importProjectFromConfig,
  listProjects,
  type ProjectListItem,
} from '@/api/simulation'

const fileInput = ref<HTMLInputElement | null>(null)
const search = ref('')
const loading = ref(false)
const importing = ref(false)
const submittingTask = ref(false)
const showSubmitDialog = ref(false)
const projects = ref<ProjectListItem[]>([])
const selectedProject = ref<ProjectListItem | null>(null)
const submitForm = ref({
  simulationTime: 60,
  speed: 1,
})

const filteredProjects = computed(() =>
  projects.value.filter((project) => project.name.toLowerCase().includes(search.value.toLowerCase())),
)

onMounted(async () => {
  await refreshProjects()
})

async function refreshProjects() {
  loading.value = true
  try {
    const data = await listProjects()
    projects.value = data.projects ?? []
  } catch (_error) {
    ElMessage.error('Failed to load projects')
  } finally {
    loading.value = false
  }
}

function openImportPicker() {
  if (window.simfox?.pickProjectFiles) {
    void importWithElectron()
    return
  }
  fileInput.value?.click()
}

async function importWithElectron() {
  const desktopApi = window.simfox
  if (!desktopApi) {
    ElMessage.error('Electron desktop API is unavailable')
    return
  }

  importing.value = true
  try {
    const bundle = await desktopApi.pickProjectFiles()
    if (!bundle) {
      return
    }
    if (bundle.missing.length > 0) {
      throw new Error(`Missing dependency files: ${bundle.missing.join(', ')}`)
    }

    const projectFiles = await Promise.all(
      bundle.files.map(async (entry) => {
        const bytes = await desktopApi.readFileBytes(entry.absolutePath)
        const file = new File([new Uint8Array(bytes)], entry.name, {
          type: 'application/octet-stream',
        })
        return {
          file,
          relativePath: entry.relativePath,
        }
      }),
    )

    await importProjectFromConfig({
      projectName: bundle.projectName,
      simulationTime: 60,
      speed: 1,
      isNowRun: false,
      configFile: await resolveConfigFile(projectFiles),
      projectFiles,
    })

    ElMessage.success('Project imported successfully')
    await refreshProjects()
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : 'Failed to import project')
  } finally {
    importing.value = false
  }
}

async function handleImportSelection(event: Event) {
  const input = event.target as HTMLInputElement
  const selectedFile = input.files?.[0]
  if (!selectedFile) {
    return
  }

  importing.value = true
  try {
    const prepared = await prepareImportPayload([selectedFile])
    await importProjectFromConfig(prepared)
    ElMessage.success('Project imported successfully')
    await refreshProjects()
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : 'Failed to import project')
  } finally {
    importing.value = false
    input.value = ''
  }
}

async function prepareImportPayload(files: File[]) {
  if (files.length !== 1) {
    throw new Error('Import accepts exactly one file')
  }

  const selectedFile = files[0]
  const isSumoConfig = selectedFile.name.toLowerCase().endsWith('.sumocfg')
  if (isSumoConfig && !window.simfox?.pickProjectFiles) {
    throw new Error(
      'Single .sumocfg import requires the Electron desktop app so sibling dependency files can be collected automatically',
    )
  }

  const projectFiles = new Map<string, File>()
  const filesByRelativePath = new Map<string, File>()

  files.forEach((file) => {
    const relativePath = normalizePath(relativeFilePath(file))
    projectFiles.set(relativePath, file)
    filesByRelativePath.set(relativePath, file)
  })

  const sumoConfigFiles = files.filter((file) => file.name.toLowerCase().endsWith('.sumocfg'))
  for (const configFile of sumoConfigFiles) {
    const configRelativePath = normalizePath(relativeFilePath(configFile))
    const configDirectory = dirname(configRelativePath)
    const dependencies = extractDependencies(await configFile.text())

    for (const dependency of dependencies) {
      const relativePath = normalizePath(joinPath(configDirectory, dependency))
      const dependencyFile = filesByRelativePath.get(relativePath)
      if (!dependencyFile) {
        throw new Error(`Missing dependency file: ${dependency}`)
      }
      projectFiles.set(relativePath, dependencyFile)
    }
  }

  const resolvedProjectFiles = Array.from(projectFiles.entries()).map(([relativePath, file]) => ({
    file,
    relativePath,
  }))

  const configFile = await resolveConfigFile(resolvedProjectFiles)
  return {
    projectName: stripExtension(configFile.file.name),
    simulationTime: 60,
    speed: 1,
    isNowRun: false,
    configFile,
    projectFiles: resolvedProjectFiles,
  }
}

function extractDependencies(configText: string): string[] {
  const parser = new DOMParser()
  const doc = parser.parseFromString(configText, 'application/xml')
  const parserError = doc.querySelector('parsererror')
  if (parserError) {
    throw new Error('The selected sumocfg file could not be parsed')
  }

  const values = [
    getAttributeValue(doc, 'input > net-file'),
    getAttributeValue(doc, 'input > route-files'),
    getAttributeValue(doc, 'input > additional-files'),
    getAttributeValue(doc, 'gui_only > gui-settings-file'),
  ]

  return values
    .flatMap((value) => splitPaths(value))
    .filter((value, index, array) => value !== '' && array.indexOf(value) === index)
}

function getAttributeValue(doc: Document, selector: string): string {
  return doc.querySelector(selector)?.getAttribute('value')?.trim() ?? ''
}

function splitPaths(raw: string): string[] {
  return raw
    .split(',')
    .map((item) => normalizePath(item))
    .filter(Boolean)
}

function normalizePath(value: string): string {
  return value.replace(/\\/g, '/').replace(/^\.?\//, '').trim()
}

function dirname(value: string): string {
  const normalized = normalizePath(value)
  const segments = normalized.split('/')
  segments.pop()
  return segments.join('/')
}

function joinPath(left: string, right: string): string {
  const parts = [left, right].filter(Boolean).join('/')
  return normalizePath(parts)
}

function stripExtension(value: string): string {
  return value.replace(/\.[^.]+$/, '')
}

function submitTask(project: ProjectListItem) {
  selectedProject.value = project
  submitForm.value = {
    simulationTime: 60,
    speed: 1,
  }
  showSubmitDialog.value = true
}

async function confirmSubmitTask() {
  if (!selectedProject.value) {
    return
  }

  submittingTask.value = true
  try {
    await enqueueProjectTask(selectedProject.value.id, {
      simulationTime: submitForm.value.simulationTime,
      speed: submitForm.value.speed,
    })
    ElMessage.success('Task submitted successfully')
    closeSubmitDialog()
    await refreshProjects()
  } catch (_error) {
    ElMessage.error('Failed to submit task')
  } finally {
    submittingTask.value = false
  }
}

async function removeProject(project: ProjectListItem) {
  try {
    await ElMessageBox.confirm(
      `Delete project "${project.name}"? Completed tasks and uploaded project records will be removed.`,
      'Delete Project',
      {
        type: 'warning',
        confirmButtonText: 'Delete',
        cancelButtonText: 'Cancel',
      },
    )
  } catch {
    return
  }

  try {
    await deleteProject(project.id)
    ElMessage.success('Project deleted')
    await refreshProjects()
  } catch (error: any) {
    const message = error?.response?.data?.error || 'Failed to delete project'
    ElMessage.error(message)
  }
}

function closeSubmitDialog() {
  showSubmitDialog.value = false
  selectedProject.value = null
}

function relativeFilePath(file: File): string {
  return (file as File & { webkitRelativePath?: string }).webkitRelativePath || file.name
}

async function resolveConfigFile(
  projectFiles: Array<{ file: File; relativePath: string }>
) {
  const sumoConfig = projectFiles.find((entry) => entry.file.name.toLowerCase().endsWith('.sumocfg'))
  if (sumoConfig) {
    return sumoConfig
  }

  const netFile = projectFiles.find((entry) => entry.file.name.toLowerCase().endsWith('.net.xml'))
  if (!netFile) {
    throw new Error('Select at least one .sumocfg or .net.xml file')
  }

  const generated = [
    '<sumoConfiguration>',
    '  <input>',
    `    <net-file value="${netFile.relativePath}" />`,
    '  </input>',
    '</sumoConfiguration>',
  ].join('\n')

  return {
    file: new File([generated], `${stripExtension(netFile.file.name)}.sumocfg`, {
      type: 'application/xml',
    }),
    relativePath: `${stripExtension(netFile.relativePath)}.sumocfg`,
  }
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

.actions {
  display: flex;
  gap: 8px;
}

.search {
  width: 320px;
  margin-bottom: 16px;
}

.hidden-input {
  display: none;
}

h2 {
  margin: 0 0 6px;
}

p {
  margin: 0;
  color: #64748b;
}
</style>
