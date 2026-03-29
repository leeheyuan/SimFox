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
          <h3>实时3D仿真视图</h3>
          <div id="three-container" class="three-box">
            <!-- 这里放 Three.js / Unreal Pixel Streaming -->
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
import { ref, onMounted } from "vue"
import * as echarts from "echarts"

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

.three-box {
  height: 400px;
  background: #0b1120;
}

.chart-box {
  height: 300px;
}
</style>