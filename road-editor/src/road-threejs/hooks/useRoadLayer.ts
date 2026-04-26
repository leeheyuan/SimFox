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
import { createRoadMesh, createJunctionMesh } from "../utils/roadGeometry";
import {
  buildLaneLines,
  updateLaneLinesResolution,
  disposeLaneLines,
  type LaneLineOptions,
} from "../utils/laneLines";
import type { Edge, Junction } from "../utils/roadGeometry";

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
  const sceneRef = useRef<THREE.Scene | null>(null);
  const roadObjectRef = useRef<THREE.Object3D | null>(null);
  const junctionObjectRef = useRef<THREE.Object3D | null>(null);
  const laneLines = useRef<THREE.Object3D | null>(null);

  // ── 销毁并从 scene 移除当前所有对象 ──
  const removeFromScene = useCallback(() => {
    const scene = sceneRef.current;
    if (!scene) return;

    if (roadObjectRef.current) {
      scene.remove(roadObjectRef.current);
      roadObjectRef.current = null;
    }
    if (junctionObjectRef.current) {
      scene.remove(junctionObjectRef.current);
      junctionObjectRef.current = null;
    }

    if (laneLines.current) {
      scene.remove(laneLines.current);
      laneLines.current = null;
    }

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
        roadObjectRef.current = g;
      }

      const junctionResult = createJunctionMesh(junctions, roadTexture);
      if (junctionResult) {
          scene.add(junctionResult);
        junctionObjectRef.current = junctionResult;
      }

      const opts: LaneLineOptions = {
        ...laneLineOptions,
      };
      const line = buildLaneLines(edges, opts);
      if (line) {
        scene.add(line);
        laneLines.current = line;
      }
      onNeedRender?.();
    },
    // eslint-disable-next-line react-hooks/exhaustive-deps
    [edges, junctions, roadTexture, laneLineOptions, container, onNeedRender]
  );

  // ── 更新分辨率 ──
  const updateResolution = useCallback((width: number, height: number) => {
    //updateLaneLinesResolution(laneLines.current, width, height);
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
      const w = container?.clientWidth ?? window.innerWidth;
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
