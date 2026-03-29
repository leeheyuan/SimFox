<template>
  <el-menu
    default-active="$route.path"
    class="el-menu-vertical"
    :router="true"
    background-color="#2d3a4b"
    text-color="#fff"
    active-text-color="#ffd04b"
  >
    <el-menu-item index="/">
      <el-icon><House /></el-icon>
      <span>首页</span>
    </el-menu-item>
    <el-menu-item index="/Projects">
      <el-icon><Document /></el-icon>
      <span>我的项目</span>
    </el-menu-item>
     <el-menu-item index="/Simulation">
      <el-icon><Cpu /></el-icon>
      <span>仿真运行</span>
    </el-menu-item>
    <!-- <el-menu-item index="/map-editor">
      <el-icon><MapLocation /></el-icon>
      <span>地图编辑</span>
    </el-menu-item> -->   
  </el-menu>
</template>

<script setup>
import { useRoute } from 'vue-router' 
import { House,Document,Cpu, MapLocation, Setting } from '@element-plus/icons-vue'
const $route = useRoute()
</script>

<style scoped>
.el-menu-vertical {
  height: 100%;
  border-right: none;
}
</style>

 <!--<el-menu-item index="Dashboard">首页</el-menu-item>
      <el-menu-item index="Projects">我的项目</el-menu-item>
      <el-menu-item index="Simulation">仿真运行</el-menu-item>
      <el-menu-item index="MapEditor">地图编辑</el-menu-item>
      <el-menu-item index="/settings">系统设置</el-menu-item>-->
