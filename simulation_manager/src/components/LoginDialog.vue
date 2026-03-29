<template>
  <el-dialog v-model="visible" title="登录" width="400px">
    <el-form @submit.prevent>
      <el-form-item label="用户名">
        <el-input v-model="username" placeholder="请输入用户名" />
        <el-input v-model="password" placeholder="请输入密码" />
      </el-form-item>
    </el-form>
    <template #footer>
      <el-button @click="visible = false">取消</el-button>
      <el-button type="primary" @click="handleLogin">登录</el-button>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useUserStore } from '@/stores/user'
import { login } from '@/api/user'
import { ElMessage } from 'element-plus';
const visible = defineModel<boolean>() // v-model:visible
const username = ref('')
const password = ref('') 

const handleLogin = async () => { 
  if (username.value.trim()) {
    const success = await login(username.value.trim(), password.value.trim())
    if (success) {
      visible.value = false
      ElMessage.success('登录成功');
    } else {
      ElMessage.error('登录失败，请检查用户名或密码');
    }
  }
}

</script>
