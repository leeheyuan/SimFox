<template>
  <div style="padding: 20px;">
    <h2>地图编辑</h2>
    <el-alert title="提示" type="warning" class="mb-4">
      这是地图编辑页面。你可以在此导入地图文件、标注道路、编辑交叉口信息等。
    </el-alert>

    <el-upload
      class="upload-demo"
      drag
      action="#"
      :auto-upload="true"
      :http-request="customUpload"
      :show-file-list="false"
      multiple
      style="width: 100%;"
    >
      <i class="el-icon-upload" />
      <div class="el-upload__text">将地图文件拖到此处，或<em>点击上传</em></div>
    </el-upload>

    <el-divider />
    <p>你可以通过地图编辑器进行可视化标注（功能待开发）。</p>
  </div>
</template>


<script lang="ts" setup>
import { ElMessage } from 'element-plus' 
import { uploadMap } from '@/api/user'

// Element Plus Upload 自定义上传函数类型
interface UploadRequestOptions {
  file: File
  onSuccess: (response: any) => void
  onError: (error: any) => void
  onProgress?: (event: { percent: number }) => void
}

const customUpload = async (options: UploadRequestOptions) => {

  const { file, onSuccess, onError, onProgress } = options
  const formData = new FormData()
  formData.append('netFile', file)

  const success = await uploadMap(formData, (percent) => {
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