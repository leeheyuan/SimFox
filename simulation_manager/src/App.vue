 

<template>
  <MainLayout />
</template>

<script setup>
import MainLayout from './layout/MainLayout.vue'
import { ref, onMounted } from 'vue'
import { checkTokenAndGetUserInfo } from '@/utils/auth'

const userInfo = ref(null)

onMounted(async () => {
  /*const user = await checkTokenAndGetUserInfo()
  userInfo.value = user*/
})
</script>

<style scoped>
html, body {
  margin: 0;
  padding: 0;
}

</style>