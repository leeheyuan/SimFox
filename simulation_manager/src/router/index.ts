import { createRouter, createWebHistory } from 'vue-router'
import Dashboard from '../pages/Dashboard.vue'
import Projects from '../pages/Projects.vue'
import Simulation from '../pages/Simulation.vue'
import Results from '../pages/Results.vue'
import Cluster from '../pages/Cluster.vue'
import Settings from '../pages/Settings.vue'

const routes = [
  { path: '/', name: 'Dashboard', component: Dashboard },
  { path: '/projects', name: 'Projects', component: Projects },
  { path: '/tasks', name: 'SimulationTasks', component: Simulation },
  { path: '/results', name: 'Results', component: Results },
  { path: '/cluster', name: 'Cluster', component: Cluster },
  { path: '/settings', name: 'Settings', component: Settings },
]

const router = createRouter({
  history: createWebHistory(),
  routes,
})

export default router
