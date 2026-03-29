/**
 * useRoadLayer.ts
 * React Hook：管理道路层（路面 + 路口 + 车道线）的生命周期。
 * 纯 Three.js，不依赖 @react-three/fiber。
 *
 * 用法：
 *   const { addToScene, removeFromScene } = useRoadLayer({ edges, junctions, roadTexture });
 *
 *   // 在拿到 scene 之后调用：
 *   addToScene(scene);
 *
 *   // 组件卸载时会自动调用 removeFromScene，也可手动调用。
 */

import { useEffect, useRef, useCallback } from "react";
import * as THREE from "three";
import { Line2 } from "three/examples/jsm/lines/Line2";
import { LineMaterial } from "three/examples/jsm/lines/LineMaterial";

import { createRoadMesh, createJunctionMesh ,createRoadMeshFromLonLat} from "../utils/roadGeometry";
import {
  buildLaneLines,
  updateLaneLinesResolution,
  disposeLaneLines,
  type LaneLineOptions,
} from "../utils/laneLines";
import type { Edge, Junction, RoadMeshResult } from "../utils/roadGeometry";

// ─────────────────────────────────────────────
// 类型
// ─────────────────────────────────────────────

export interface UseRoadLayerOptions {
  edges: Edge[];
  junctions?: Junction[];
  roadTexture: THREE.Texture;
  laneLineOptions?: LaneLineOptions;
  /** 传入后 hook 自动把图层加入 scene，无需手动调用 addToScene */
  scene?: THREE.Scene;
  /** 传入后 hook 自动监听 resize 并更新车道线分辨率 */
  container?: HTMLElement | null;
  /** iTowns / 自定义渲染器：每次数据变化后触发重绘 */
  onNeedRender?: () => void;
}

export interface UseRoadLayerReturn {
  /** 手动把图层加入指定 scene */
  addToScene: (scene: THREE.Scene) => void;
  /** 手动从 scene 移除并销毁图层 */
  removeFromScene: () => void;
  /** 手动更新车道线视口分辨率 */
  updateResolution: (width: number, height: number) => void;
}

// ─────────────────────────────────────────────
// Hook 实现
// ─────────────────────────────────────────────

export function useRoadLayer({
  edges,
  junctions = [],
  roadTexture,
  laneLineOptions,
  scene: externalScene,
  container,
  onNeedRender,
}: UseRoadLayerOptions): UseRoadLayerReturn {

  // 持有当前已加入 scene 的对象引用
  const sceneRef         = useRef<THREE.Scene | null>(null);
  const roadResultRef    = useRef<RoadMeshResult | null>(null);
  const junctionResultRef= useRef<RoadMeshResult | null>(null);
  const laneLines        = useRef<Line2[]>([]);

  // ── 销毁并从 scene 移除当前所有对象 ──
  const removeFromScene = useCallback(() => {
    const scene = sceneRef.current;
    if (!scene) return;

    if (roadResultRef.current) {
      scene.remove(roadResultRef.current.mesh);
      roadResultRef.current.dispose();
      roadResultRef.current = null;
    }
    if (junctionResultRef.current) {
      scene.remove(junctionResultRef.current.mesh);
      junctionResultRef.current.dispose();
      junctionResultRef.current = null;
    }
    laneLines.current.forEach((l) => scene.remove(l));
    disposeLaneLines(laneLines.current);
    laneLines.current = [];

    sceneRef.current = null;
  }, []);

  // ── 构建并加入 scene ──
  const addToScene = useCallback(
    (scene: THREE.Scene) => {
      // 先清理旧的
      removeFromScene();

      sceneRef.current = scene;

      // 路面
      const g = createRoadMesh(edges, roadTexture);
      if (g) {
        scene.add(g);
        //roadResultRef.current = roadResult;
      }
 
     /* const junctionResult = createJunctionMesh(junctions, roadTexture);
      if (junctionResult) {
        scene.add(junctionResult.mesh);
        junctionResultRef.current = junctionResult;
      }
 
      const opts: LaneLineOptions = {
        resolutionWidth:  container?.clientWidth  ?? window.innerWidth,
        resolutionHeight: container?.clientHeight ?? window.innerHeight,
        ...laneLineOptions,
      };
      const lines = buildLaneLines(edges, opts);
      lines.forEach((l) => scene.add(l));
      laneLines.current = lines;*/

      onNeedRender?.();
    },
    // eslint-disable-next-line react-hooks/exhaustive-deps
    [edges, junctions, roadTexture, laneLineOptions, container, onNeedRender]
  );

  // ── 更新分辨率 ──
  const updateResolution = useCallback((width: number, height: number) => {
    updateLaneLinesResolution(laneLines.current, width, height);
  }, []);

  // ── 如果外部传入了 scene，自动管理挂载 ──
  useEffect(() => {
    if (!externalScene) return;
    addToScene(externalScene);
    return () => removeFromScene();
  }, [externalScene, addToScene, removeFromScene]);

  // ── 自动监听 resize（需要 container 或 externalScene 存在时才注册）──
  useEffect(() => {
    if (!container && !externalScene) return;

    const handleResize = () => {
      const w = container?.clientWidth  ?? window.innerWidth;
      const h = container?.clientHeight ?? window.innerHeight;
      updateResolution(w, h);
      onNeedRender?.();
    };

    window.addEventListener("resize", handleResize);
    return () => window.removeEventListener("resize", handleResize);
  }, [container, externalScene, updateResolution, onNeedRender]);

  // ── 组件卸载时兜底清理 ──
  useEffect(() => {
    return () => removeFromScene();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  return { addToScene, removeFromScene, updateResolution };
}
