import { createRouter, createWebHistory } from 'vue-router'
import Dashboard from '../pages/Dashboard.vue'
import Projects from '../pages/Projects.vue'
import Simulation from '../pages/Simulation.vue'
import MapEditor from '../pages/MapEditor.vue'
import MapboxSumo from '../views/MapboxSumo.vue' 
import SimMonitor from '../views/SimMonitor.vue'

const routes = [
  { path: '/', name: 'Dashboard', component: Dashboard },
  { path: '/projects', name: 'Projects', component: Projects },
  { path: '/simulation', name: 'Simulation', component: Simulation },
  { path: '/map-editor', name: 'MapEditor', component: MapEditor },
  { path: '/map-Sumo', name: 'MapboxSumo', component: MapboxSumo },
  { path: '/sim-monitor', name: 'SimMonitor', component: SimMonitor },
]

const router = createRouter({
  history: createWebHistory(),
  routes,
})

export default router
