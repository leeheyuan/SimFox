<template>
  <div class="map-editor-page">
    <div class="toolbar">
      <div>
        <h2>地图编辑</h2>
        <p class="toolbar-hint">点击“开始框选”后，在地图上拖拽一个矩形区域，系统会自动抓取 OSM 并转换为 net.xml。</p>
      </div>
      <div class="toolbar-actions">
        <el-button type="primary" :disabled="selecting" @click="enableSelection">
          {{ selecting ? '框选中...' : '开始框选' }}
        </el-button>
        <el-button type="success" :disabled="!selectedBounds || importing" @click="handleImport">
          {{ importing ? '导入中...' : '抓取 OSM 并转换' }}
        </el-button>
        <el-button :disabled="!selectedBounds && !networkLoaded" @click="resetSelection">
          清空选择
        </el-button>
      </div>
    </div>

    <el-alert type="info" :closable="false" class="mb-4">
      当前流程：框选区域 -> 后端拉取 OSM -> 服务器执行 netconvert -> 前端直接回显 GeoJSON。
    </el-alert>

    <div class="map-shell">
      <div ref="mapContainer" class="map-container" />
      <div v-if="dragBoxStyle" class="selection-box" :style="dragBoxStyle" />
    </div>

    <div class="result-panel">
      <div v-if="selectedBounds" class="bounds-card">
        <strong>框选经纬度</strong>
        <div>南西角：{{ selectedBounds.south.toFixed(6) }}, {{ selectedBounds.west.toFixed(6) }}</div>
        <div>东北角：{{ selectedBounds.north.toFixed(6) }}, {{ selectedBounds.east.toFixed(6) }}</div>
      </div>
      <div v-if="netFilePath" class="bounds-card">
        <strong>生成结果</strong>
        <div>net.xml: {{ netFilePath }}</div>
      </div>
    </div>
  </div>
</template>

<script lang="ts" setup>
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import { ElMessage } from 'element-plus'
import mapboxgl from 'mapbox-gl'
import 'mapbox-gl/dist/mapbox-gl.css'
import { importOsmBounds } from '@/api/user'

type Bounds = {
  south: number
  west: number
  north: number
  east: number
}

type ScreenPoint = {
  x: number
  y: number
}

const mapContainer = ref<HTMLElement | null>(null)
const mapInstance = ref<mapboxgl.Map | null>(null)
const selecting = ref(false)
const importing = ref(false)
const networkLoaded = ref(false)
const selectedBounds = ref<Bounds | null>(null)
const netFilePath = ref('')

const dragStart = ref<ScreenPoint | null>(null)
const dragCurrent = ref<ScreenPoint | null>(null)

const mapStyle = {
  version: 8 as const,
  sources: {
    osm: {
      type: 'raster' as const,
      tiles: ['https://tile.openstreetmap.org/{z}/{x}/{y}.png'],
      tileSize: 256,
      attribution: '© OpenStreetMap contributors',
    },
  },
  layers: [
    {
      id: 'osm',
      type: 'raster' as const,
      source: 'osm',
    },
  ],
}

const dragBoxStyle = computed(() => {
  if (!dragStart.value || !dragCurrent.value) {
    return null
  }

  const left = Math.min(dragStart.value.x, dragCurrent.value.x)
  const top = Math.min(dragStart.value.y, dragCurrent.value.y)
  const width = Math.abs(dragCurrent.value.x - dragStart.value.x)
  const height = Math.abs(dragCurrent.value.y - dragStart.value.y)

  return {
    left: `${left}px`,
    top: `${top}px`,
    width: `${width}px`,
    height: `${height}px`,
  }
})

function enableSelection() {
  selecting.value = true
  ElMessage.info('请在地图上按住鼠标左键拖拽框选区域')
}

function resetSelection() {
  selectedBounds.value = null
  netFilePath.value = ''
  networkLoaded.value = false
  dragStart.value = null
  dragCurrent.value = null

  const map = mapInstance.value
  if (!map) return

  if (map.getSource('selection-area')) {
    ;(map.getSource('selection-area') as mapboxgl.GeoJSONSource).setData(emptyPolygon())
  }
  if (map.getSource('sumo-network')) {
    ;(map.getSource('sumo-network') as mapboxgl.GeoJSONSource).setData(emptyFeatureCollection())
  }
}

async function handleImport() {
  if (!selectedBounds.value) {
    ElMessage.warning('请先框选一个区域')
    return
  }

  importing.value = true
  try {
    const result = await importOsmBounds(selectedBounds.value)
    netFilePath.value = result.netFile || ''
    loadGeoJSON(result.geojson)
    networkLoaded.value = true
    ElMessage.success('OSM 已抓取并完成 net.xml 转换')
  } catch (error: any) {
    ElMessage.error(error?.response?.data?.error || '导入失败')
  } finally {
    importing.value = false
  }
}

function emptyPolygon() {
  return {
    type: 'FeatureCollection' as const,
    features: [],
  }
}

function emptyFeatureCollection() {
  return {
    type: 'FeatureCollection' as const,
    features: [],
  }
}

function buildSelectionFeature(bounds: Bounds) {
  return {
    type: 'FeatureCollection' as const,
    features: [
      {
        type: 'Feature' as const,
        geometry: {
          type: 'Polygon' as const,
          coordinates: [[
            [bounds.west, bounds.south],
            [bounds.east, bounds.south],
            [bounds.east, bounds.north],
            [bounds.west, bounds.north],
            [bounds.west, bounds.south],
          ]],
        },
        properties: {},
      },
    ],
  }
}

function loadGeoJSON(geojson: any) {
  const map = mapInstance.value
  if (!map) return

  if (map.getSource('sumo-network')) {
    ;(map.getSource('sumo-network') as mapboxgl.GeoJSONSource).setData(geojson)
    return
  }

  map.addSource('sumo-network', {
    type: 'geojson',
    data: geojson,
  })

  map.addLayer({
    id: 'sumo-network-line',
    type: 'line',
    source: 'sumo-network',
    paint: {
      'line-color': '#38bdf8',
      'line-width': 2,
    },
  })
}

function attachSelectionHandlers(map: mapboxgl.Map) {
  map.on('mousedown', (event) => {
    if (!selecting.value || event.originalEvent.button !== 0) {
      return
    }

    map.dragPan.disable()
    dragStart.value = { x: event.point.x, y: event.point.y }
    dragCurrent.value = { x: event.point.x, y: event.point.y }
  })

  map.on('mousemove', (event) => {
    if (!dragStart.value) {
      return
    }

    dragCurrent.value = { x: event.point.x, y: event.point.y }
  })

  map.on('mouseup', (event) => {
    if (!dragStart.value || !dragCurrent.value) {
      return
    }

    const startLngLat = map.unproject([dragStart.value.x, dragStart.value.y])
    const endLngLat = map.unproject([event.point.x, event.point.y])
    const bounds = {
      south: Math.min(startLngLat.lat, endLngLat.lat),
      west: Math.min(startLngLat.lng, endLngLat.lng),
      north: Math.max(startLngLat.lat, endLngLat.lat),
      east: Math.max(startLngLat.lng, endLngLat.lng),
    }

    selectedBounds.value = bounds
    selecting.value = false
    dragStart.value = null
    dragCurrent.value = null
    map.dragPan.enable()

    if (map.getSource('selection-area')) {
      ;(map.getSource('selection-area') as mapboxgl.GeoJSONSource).setData(buildSelectionFeature(bounds))
    }

    ElMessage.success('区域已选中，可以开始抓取 OSM')
  })
}

onMounted(() => {
  if (!mapContainer.value) {
    return
  }

  const map = new mapboxgl.Map({
    container: mapContainer.value,
    style: mapStyle,
    center: [104.0665, 30.5365],
    zoom: 13,
  })

  mapInstance.value = map

  map.on('load', () => {
    map.addSource('selection-area', {
      type: 'geojson',
      data: emptyPolygon(),
    })

    map.addLayer({
      id: 'selection-area-fill',
      type: 'fill',
      source: 'selection-area',
      paint: {
        'fill-color': '#22c55e',
        'fill-opacity': 0.18,
      },
    })

    map.addLayer({
      id: 'selection-area-outline',
      type: 'line',
      source: 'selection-area',
      paint: {
        'line-color': '#22c55e',
        'line-width': 2,
      },
    })
  })

  attachSelectionHandlers(map)
})

onBeforeUnmount(() => {
  mapInstance.value?.remove()
})
</script>

<style scoped>
.map-editor-page {
  padding: 20px;
}

.toolbar {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 24px;
  margin-bottom: 16px;
}

.toolbar-hint {
  margin: 8px 0 0;
  color: #64748b;
}

.toolbar-actions {
  display: flex;
  gap: 12px;
  flex-wrap: wrap;
}

.map-shell {
  position: relative;
  border: 1px solid #dbeafe;
  border-radius: 12px;
  overflow: hidden;
}

.map-container {
  width: 100%;
  height: 640px;
}

.selection-box {
  position: absolute;
  border: 2px solid #22c55e;
  background: rgba(34, 197, 94, 0.18);
  pointer-events: none;
}

.result-panel {
  margin-top: 16px;
  display: flex;
  gap: 16px;
  flex-wrap: wrap;
}

.bounds-card {
  padding: 12px 16px;
  border-radius: 10px;
  background: #f8fafc;
  border: 1px solid #e2e8f0;
  color: #0f172a;
  line-height: 1.8;
}
</style>
