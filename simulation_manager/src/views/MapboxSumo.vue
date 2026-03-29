<template>
  <el-card class="box-card">
    <template #header>
      <div class="card-header">
        <span>roadeditor</span>
        <el-upload
          action=""
          :http-request="handleUpload"
          :show-file-list="false"
          accept=".geojson"
        > 
        </el-upload>
      </div>
    </template>
    <div ref="mapContainer" class="map-container" />
  </el-card>
</template>

<script setup>
import { ref, onMounted, defineExpose } from 'vue';
import mapboxgl from 'mapbox-gl';
import 'mapbox-gl/dist/mapbox-gl.css';
import { ElMessage } from 'element-plus';
import { getGeojson } from '@/api/config'
const mapContainer = ref(null);
const map = ref(null);

// 必须替换为你的 Mapbox Access Token
 
const initMap = (center = [116.397, 39.908], zoom = 12, style = 'mapbox://styles/mapbox/streets-v11') => {
  map.value = new mapboxgl.Map({
    container: mapContainer.value,
    style,
    center,
    zoom,
  });

  //map.value.addControl(new mapboxgl.NavigationControl());

  let coordinates = [];

  map.value.on('load', () => {
    // 添加空的 line source
    map.value.addSource('road-line', {
      type: 'geojson',
      data: {
        type: 'Feature',
        geometry: {
          type: 'LineString',
          coordinates: []
        }
      }
    });

    // 添加线图层
    map.value.addLayer({
      id: 'road-line-layer',
      type: 'line',
      source: 'road-line',
      layout: {
        'line-join': 'round',
        'line-cap': 'round'
      },
      paint: {
        'line-color': '#ff0000',
        'line-width': 8   // ✅ 这里控制线条的“宽度”，即“道路宽度”
      }
    });

    // 添加点击事件
    map.value.on('click', (e) => {
      const lngLat = [e.lngLat.lng, e.lngLat.lat];
      coordinates.push(lngLat);

      const newGeojson = {
        type: 'Feature',
        geometry: {
          type: 'LineString',
          coordinates: coordinates
        }
      };

      map.value.getSource('road-line').setData(newGeojson);
    });
  });

};

function handleKeydown(e) {
  if (!map.value) return

  const panDistance = 100  // 每次移动距离（像素）

  switch (e.key.toLowerCase()) {
    case 'w':
      map.value.panBy([0, -panDistance]) // 向上
      break
    case 's':
      map.value.panBy([0, panDistance]) // 向下
      break
    case 'a':
      map.value.panBy([-panDistance, 0]) // 向左
      break
    case 'd':
      map.value.panBy([panDistance, 0]) // 向右
      break
  }
}

onMounted(async () => {
  initMap();
 window.addEventListener('keydown', handleKeydown)
  /*try {
    const response = await getGeojson("http://localhost:8081/SimulationConfig/j_lI-");
    loadGeoJSON(response);
  } catch (err) {
    ElMessage.error('加载远程 GeoJSON 失败');
    console.error(err);
  }*/

});

 
const loadGeoJSON = (geojson) => {
  if (!map.value) return;

  if (map.value.getSource('sumo-network')) {
    map.value.getSource('sumo-network').setData(geojson);
  } else {
    map.value.addSource('sumo-network', {
      type: 'geojson',
      data: geojson
    });

    map.value.addLayer({
      id: 'sumo-lines',
      type: 'line',
      source: 'sumo-network',
      paint: {
        'line-color': '#0074D9',
        'line-width': 1.5
      }
    });
  }
};

const handleUpload = async (options) => {
  const file = options.file;
  const reader = new FileReader();

  reader.onload = (e) => {
    try {
      const geojson = JSON.parse(e.target.result);
      loadGeoJSON(geojson);
    } catch (err) {
      ElMessage.error('无效的 GeoJSON 文件');
    }
  };

  reader.readAsText(file);
};
 
defineExpose({
  initMap,
  loadGeoJSON,
  getMapInstance: () => map.value
});
</script>

<style scoped>
.map-container {
  width: 100%;
  height: 600px;
  border: 1px solid #ccc;
  margin-top: 10px;
}
</style>
