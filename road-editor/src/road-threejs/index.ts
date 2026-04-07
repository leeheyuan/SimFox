/**
 * index.ts — road-threejs 模块统一出口
 *
 * import { useRoadLayer, createRoadMesh, buildLaneLines } from "@/road-threejs"
 */

// Hook（React + 纯 Three.js）
export { useRoadLayer }            from "./hooks/useRoadLayer";
export type { UseRoadLayerOptions, UseRoadLayerReturn } from "./hooks/useRoadLayer";

// 路面 / 路口几何体工具
export {
  createRoadGeometry,
  createJunctionGeometry,
  buildMergedRoadGeometry,
  buildMergedJunctionGeometry,
  createRoadMesh,
  createJunctionMesh,
  createJunctionMeshFromLonLat,
} from "./utils/roadGeometry";
export type { Edge, Lane, Junction, RoadMeshResult } from "./utils/roadGeometry";

// SUMO net.xml 解析
export { parseNetXML }            from "./utils/parseNetXML";
export type { ParseNetXMLResult } from "./utils/parseNetXML";

// 车道线工具
export {
  createLaneLine,
  buildLaneLines,
  updateLaneLinesResolution,
  disposeLaneLines,
} from "./utils/laneLines";
export type { LaneLineOptions } from "./utils/laneLines";


