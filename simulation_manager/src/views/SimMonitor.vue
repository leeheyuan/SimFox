<template>
  <div class="sim-container">
    <!-- 顶部KPI -->
    <el-row :gutter="20" class="kpi-row">
      <el-col :span="4" v-for="item in kpiList" :key="item.label">
        <el-card shadow="hover" class="kpi-card">
          <div class="kpi-value">{{ item.value }}</div>
          <div class="kpi-label">{{ item.label }}</div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 主体 -->
    <el-row :gutter="20" class="main-row">
      <!-- 左侧 -->
      <el-col :span="6">
        <el-card class="panel">
          <h3>拥堵路段 TOP10</h3>
          <el-table :data="topEdges" size="small" height="300">
            <el-table-column prop="id" label="路段" />
            <el-table-column prop="speed" label="速度" />
            <el-table-column prop="queue" label="排队长度" />
          </el-table>
        </el-card>
      </el-col>

      <!-- 中间 3D 视图 -->
      <el-col :span="12">
        <el-card class="panel">
          <div class="panel-header">
            <h3>实时3D仿真视图</h3>
            <a
              class="editor-link"
              :href="roadEditorUrl"
              target="_blank"
              rel="noopener noreferrer"
            >
              新窗口打开
            </a>
          </div>
          <div id="three-container" class="three-box">
            <iframe
              class="road-editor-frame"
              :src="roadEditorUrl"
              title="Road Editor"
            />
            <div v-if="!roadEditorUrl" class="frame-fallback">
              未配置 road-editor 地址
            </div>
          </div>
        </el-card>
      </el-col>

      <!-- 右侧 -->
      <el-col :span="6">
        <el-card class="panel">
          <h3>路口排队 TOP10</h3>
          <el-table :data="topJunctions" size="small" height="300">
            <el-table-column prop="id" label="路口" />
            <el-table-column prop="delay" label="延误" />
            <el-table-column prop="queue" label="排队" />
          </el-table>
        </el-card>
      </el-col>
    </el-row>

    <!-- 底部趋势 -->
    <el-row>
      <el-col :span="24">
        <el-card class="panel">
          <h3>全局趋势</h3>
          <div ref="chartRef" class="chart-box"></div>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup>
import { computed, ref, onMounted } from "vue"
import * as echarts from "echarts"

const roadEditorBaseUrl =
  import.meta.env.VITE_ROAD_EDITOR_URL || "http://localhost:5174/editor"

const roadEditorUrl = computed(() => {
  const token = localStorage.getItem("token")
  if (!roadEditorBaseUrl) return ""

  try {
    const url = new URL(roadEditorBaseUrl)
    if (token) {
      url.searchParams.set("token", token)
    }
    return url.toString()
  } catch {
    return roadEditorBaseUrl
  }
})

const kpiList = ref([
  { label: "仿真时间", value: "0 s" },
  { label: "车辆总数", value: 0 },
  { label: "平均车速", value: 0 },
  { label: "平均延误", value: 0 },
  { label: "拥堵指数", value: 0 },
])

const topEdges = ref([])
const topJunctions = ref([])

const chartRef = ref(null)
let chartInstance = null

// 初始化图表
function initChart() {
  chartInstance = echarts.init(chartRef.value)
  chartInstance.setOption({
    tooltip: { trigger: "axis" },
    legend: { data: ["平均车速"] },
    xAxis: { type: "category", data: [] },
    yAxis: { type: "value" },
    series: [
      {
        name: "平均车速",
        type: "line",
        data: [],
      },
    ],
  })
}

// 更新趋势图
function updateChart(time, speed) {
  const option = chartInstance.getOption()
  option.xAxis[0].data.push(time)
  option.series[0].data.push(speed)

  if (option.xAxis[0].data.length > 50) {
    option.xAxis[0].data.shift()
    option.series[0].data.shift()
  }

  chartInstance.setOption(option)
}

// WebSocket连接
function initWebSocket() {
  const ws = new WebSocket("ws://localhost:8080/ws")

  ws.onmessage = (event) => {
    const data = JSON.parse(event.data)

    // 更新KPI
    kpiList.value[0].value = data.sim_time + " s"
    kpiList.value[1].value = data.vehicle_count
    kpiList.value[2].value = data.avg_speed.toFixed(1)
    kpiList.value[3].value = data.avg_delay.toFixed(1)
    kpiList.value[4].value = data.congestion_index

    // 更新表格
    topEdges.value = data.top_edges || []
    topJunctions.value = data.top_junctions || []

    // 更新图表
    updateChart(data.sim_time, data.avg_speed)
  }
}

onMounted(() => {
  initChart()
  initWebSocket()
})
</script>


<style scoped>
.sim-container {
  background: #0f172a;
  padding: 20px;
  color: #fff;
}

.kpi-row {
  margin-bottom: 20px;
}

.kpi-card {
  text-align: center;
  background: #1e293b;
  color: #00f5ff;
}

.kpi-value {
  font-size: 26px;
  font-weight: bold;
}

.kpi-label {
  font-size: 14px;
  color: #94a3b8;
}

.panel {
  background: #1e293b;
  color: #fff;
}

.panel-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.editor-link {
  color: #38bdf8;
  font-size: 13px;
  text-decoration: none;
}

.editor-link:hover {
  color: #7dd3fc;
}

.three-box {
  position: relative;
  height: 400px;
  background: #0b1120;
  overflow: hidden;
  border-radius: 10px;
}

.road-editor-frame {
  width: 100%;
  height: 100%;
  border: 0;
  display: block;
  background: #020617;
}

.frame-fallback {
  position: absolute;
  inset: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #94a3b8;
  font-size: 14px;
}

.chart-box {
  height: 300px;
}
</style>
