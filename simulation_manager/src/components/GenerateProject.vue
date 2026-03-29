<template>
  <el-dialog title="创建仿真项目" v-model="dialogVisible" width="600px">
    <!-- 项目表单设置 -->
    <el-form :model="form" label-width="120px">
      <el-form-item label="项目名称">
        <el-input v-model="form.projectName" placeholder="请输入项目名称" />
      </el-form-item>

      <el-form-item label="仿真步长（秒）">
        <el-input-number v-model="form.stepLength" :min="0.1" :step="0.1" />
      </el-form-item>

      <el-form-item label="仿真总时长（秒）">
        <el-input-number v-model="form.totalTime" :min="1" />
      </el-form-item>

      <el-form-item label="仿真缩放速度">
        <el-input-number v-model="form.speedScale" :min="0.1" :step="0.1" />
      </el-form-item>

      <!-- 路网文件上传 -->
      <el-form-item label="net-file">
        <el-upload
          drag
          action="" 
          :auto-upload="false"
          :on-change="handleNetFileChange"
          @remove="handleNetFileRemove"
          :show-file-list="true"
        >
          <i class="el-icon-upload"></i>
          <div class="el-upload__text">拖拽或点击上传路路网文件</div>
        </el-upload> 
      </el-form-item> 
      <el-form-item label="routeFiles">
        <el-upload
          drag
          action=""
          multiple 
          :auto-upload="false"
          :on-change="handleRouteFileChange"
          @remove="handleRouteFileRemove"
          :show-file-list="true"
        >
          <i class="el-icon-upload"></i>
          <div class="el-upload__text">拖拽或点击上传路线需求文件</div>
        </el-upload> 
      </el-form-item>
      <el-form-item label="additionalFiles">
        <el-upload
          drag
          action=""
          multiple
          :auto-upload="false"
          @remove="handleAddiTionalRemove"
          :on-change="handleAddiTionalChange"
          :show-file-list="true"
        >
          <i class="el-icon-upload"></i>
          <div class="el-upload__text">拖拽或点击上传路线需求文件</div>
        </el-upload> 
      </el-form-item>
    </el-form>

    <!-- 底部操作 -->
    <template #footer>
      <el-button @click="dialogVisible = false">取消</el-button>
      <el-button type="primary" @click="handleConfirm">创建项目</el-button>
    </template>
 
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'

import { ElMessage, type UploadFile } from 'element-plus'  
 
 
const dialogVisible = defineModel<boolean>()
 

const form = ref({
  projectName: '',
  stepLength: 0.5,
  totalTime: 60,
  speedScale: 1.0,
  roadnet: ''
})
 

// 单文件 netFile
const netFile = ref<File | null>(null)
const handleNetFileChange = (file: UploadFile) => {
  netFile.value = file.raw as File
}
const handleNetFileRemove = () => {
  netFile.value = null
}

// 多文件 routeFiles
const routeFiles = ref<File[]>([])
const handleRouteFileChange = (_: UploadFile, uploadFiles: UploadFile[]) => {
  routeFiles.value = uploadFiles.map(f => f.raw as File)
}
const handleRouteFileRemove = (_: UploadFile, fileList: UploadFile[]) => {
  routeFiles.value = fileList.map(f => f.raw as File)
}

// 多文件 additionalFiles
const additionalFiles = ref<File[]>([])
const handleAddiTionalChange = (_: UploadFile, uploadFiles: UploadFile[]) => {
  additionalFiles.value = uploadFiles.map(f => f.raw as File)
}
const handleAddiTionalRemove = (_: UploadFile, fileList: UploadFile[]) => {
  additionalFiles.value = fileList.map(f => f.raw as File)
}
 
 
import { generate } from '@/api/user'

/*
import { useRouter } from 'vue-router'
const router = useRouter()
const openMapEdit = () => { 
  router.push('/map-Sumo')  // 跳转到 MapboxSumo.vue
}*/

const handleConfirm = async () => {
  if (!form.value.projectName || !netFile.value || routeFiles.value.length === 0) {
    ElMessage.warning('请填写完整信息并上传必要文件')
    return
  }

  const formData = new FormData()
  formData.append('name', form.value.projectName)
  formData.append('simulationTime', form.value.totalTime.toString())
  formData.append('isNowRun', 'true') // 可根据实际设置 true/false

  // 单文件 netFile
  formData.append('netFile', netFile.value)

  // 多文件 routeFiles
  routeFiles.value.forEach(file => {
    formData.append('routeFiles', file)
  })

  // 多文件 additionalFiles
  additionalFiles.value.forEach(file => {
    formData.append('additionalFiles', file)
  })

  

  const success = await generate(formData, (percent) => {
    if (onProgress) onProgress({ percent })
  })

  if (success) {
    ElMessage.success('上传成功')
    onSuccess({})
  } else {
    ElMessage.error('上传失败')
    onError(new Error('上传失败'))
  }
}
 
</script>
