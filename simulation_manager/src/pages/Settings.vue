<template>
  <section class="page">
    <h2>Settings</h2>
    <p class="intro">Configure private cloud endpoints and local desktop integration.</p>

    <el-form label-width="180px" class="settings-form">
      <el-form-item label="Platform API">
        <el-input v-model="settings.apiEndpoint" placeholder="https://simfox.example.com/api" />
      </el-form-item>
      <el-form-item label="Result download folder">
        <el-input v-model="settings.downloadFolder" placeholder="Select from Electron client" />
      </el-form-item>
      <el-form-item label="Desktop client">
        <el-switch v-model="settings.desktopEnabled" />
      </el-form-item>
      <el-form-item>
        <el-button type="primary">Save</el-button>
      </el-form-item>
    </el-form>
  </section>
</template>

<script setup lang="ts">
import { reactive } from 'vue'

const settings = reactive({
  apiEndpoint: '',
  downloadFolder: '',
  desktopEnabled: false,
})
</script>

<style scoped>
.page {
  padding: 20px;
  max-width: 860px;
}

h2 {
  margin: 0 0 6px;
}

.intro {
  margin: 0 0 24px;
  color: #64748b;
}

.settings-form {
  max-width: 720px;
}
</style>
