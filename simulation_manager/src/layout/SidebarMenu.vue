<template>
  <el-menu
    :default-active="$route.path"
    class="el-menu-vertical"
    :router="true"
    background-color="#1f2937"
    text-color="#e5e7eb"
    active-text-color="#7dd3fc"
  >
    <el-menu-item index="/">
      <el-icon><House /></el-icon>
      <span>Overview</span>
    </el-menu-item>
    <el-menu-item index="/projects">
      <el-icon><Document /></el-icon>
      <span>Projects</span>
    </el-menu-item>
    <el-menu-item index="/tasks">
      <el-icon><Cpu /></el-icon>
      <span>Tasks</span>
    </el-menu-item>
    <el-menu-item index="/results">
      <el-icon><DataAnalysis /></el-icon>
      <span>Results</span>
    </el-menu-item>
    <el-menu-item index="/cluster">
      <el-icon><Connection /></el-icon>
      <span>Cluster</span>
    </el-menu-item>
    <el-menu-item index="/settings">
      <el-icon><Setting /></el-icon>
      <span>Settings</span>
    </el-menu-item>
  </el-menu>
</template>

<script setup>
import { useRoute } from 'vue-router'
import { Connection, Cpu, DataAnalysis, Document, House, Setting } from '@element-plus/icons-vue'

const $route = useRoute()
</script>

<style scoped>
.el-menu-vertical {
  height: 100%;
  border-right: none;
}
</style>
