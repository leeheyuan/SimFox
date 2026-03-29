// stores/user.js
import { defineStore } from 'pinia'

export const useUserStore = defineStore('user', {
  state: () => ({
    isLoggedIn: false,
    name: '',
    avatar: '',
  }),
  actions: {
    login(name, avatar) {
      this.isLoggedIn = true
      this.name = name
      this.avatar = avatar
    },
    logout() {
      this.isLoggedIn = false
      this.name = ''
      this.avatar = ''
    },
  },
})
