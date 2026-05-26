<template>
  <div class="header-bar">
    <div class="logo">SimFox</div>
    <div class="subtitle">Private cloud simulation management</div>
    <div class="spacer"></div>
    <div class="user-section">
      <el-avatar
        :src="user.isLoggedIn ? user.avatar : ''"
        size="default"
        class="avatar"
        @click="toggleLoginDialog"
      >
        <el-icon><User /></el-icon>
      </el-avatar>
      <span v-if="user.isLoggedIn" class="username">{{ user.name }}</span>
    </div>
    <LoginDialog v-model="showLogin" />
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { checkTokenAndGetUserInfo } from '@/utils/auth'
import { useUserStore } from '@/stores/user'
import LoginDialog from '@/components/LoginDialog.vue'
import { User } from '@element-plus/icons-vue'
import { ElMessageBox } from 'element-plus'

const user = useUserStore()
const showLogin = ref(false)

const toggleLoginDialog = () => {
  if (!user.isLoggedIn) {
    showLogin.value = true
    return
  }

  ElMessageBox.confirm('Sign out of SimFox?', 'Account', {
    type: 'warning',
  }).then(() => {
    user.logout()
  })
}

onMounted(async () => {
  const data = await checkTokenAndGetUserInfo()
  if (data == null) {
    showLogin.value = true
  } else {
    user.isLoggedIn = true
  }
})
</script>

<style scoped>
.header-bar {
  display: flex;
  align-items: center;
  height: 60px;
  padding: 0;
  background-color: #0f172a;
  color: white;
}

.logo {
  font-size: 20px;
  font-weight: 700;
  margin-left: 20px;
}

.subtitle {
  margin-left: 16px;
  color: #cbd5e1;
  font-size: 13px;
}

.user-section {
  display: flex;
  align-items: center;
  padding-right: 20px;
}

.avatar {
  cursor: pointer;
}

.username {
  margin-left: 10px;
}

.spacer {
  flex: 1;
}
</style>
