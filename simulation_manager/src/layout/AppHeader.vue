<template>
  <div class="header-bar">
    <div class="logo"> traffic simulation</div>
    <div class="spacer"></div> <!-- 中间撑开的元素 -->
    <div class="user-section">
      <el-avatar
        :src="user.isLoggedIn ? user.avatar : ''"
        size="medium"
        @click="toggleLoginDialog"
        class="avatar"
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

const user = useUserStore()
const showLogin = ref(false)

const toggleLoginDialog = () => {
  if (!user.isLoggedIn) {
    showLogin.value = true
  } else {
    // 可选弹出菜单登出等
    ElMessageBox.confirm('是否退出登录？', '提示', {
      type: 'warning',
    }).then(() => {
      user.logout()
    })
  }
} 
 

onMounted(async () => {
  const data = await checkTokenAndGetUserInfo()
  if(data == null) {
    showLogin.value = true
  } else {
    user.isLoggedIn = true
  }
  //userInfo.value = user
})


</script>


<style scoped>
.header-bar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  height: 60px;
  padding: 0 0px;
  background-color: #409eff;
  color: white;
}
.logo {
  font-size: 20px;
  margin-left: 20px;
   
}
.user-section {
  display: flex;
  align-items: center;
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
