<template>
  <el-container style="height: 100vh">
    <!-- 顶部 -->
    <el-header height="60px" class="custom-header" >
      <AppHeader />
    </el-header>

    <!-- 主体区域 -->
    <el-container>
      <!-- 左侧菜单 -->
      <el-aside width="200px">
        <SidebarMenu />
      </el-aside>

      <!-- 内容区 -->
      <el-main>
        <router-view />
      </el-main>
    </el-container>
  </el-container>
</template>

<script setup>
import AppHeader from './AppHeader.vue'
import SidebarMenu from './SidebarMenu.vue'
</script>
<style scoped>
.custom-header {
  padding: 0 !important;
}
</style>