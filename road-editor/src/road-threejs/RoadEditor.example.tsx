/**
 * RoadEditor.example.tsx
 * 演示如何在 iTowns（纯 Three.js）项目中使用 useRoadLayer。
 */

import { useEffect, useRef } from "react";
import * as itowns from "itowns";
import * as THREE from "three";
import { useRoadLayer } from "@/road-threejs";
import type { Edge, Junction } from "@/road-threejs";

// ── 示例数据（替换成你的真实数据）──────────────────────────

const edges: Edge[] = [
  {
    spreadType: "right",
    lanes: [
      {
        width: 3.5,
        shape: [
          new THREE.Vector3(0,   0, 0),
          new THREE.Vector3(50,  0, 0),
          new THREE.Vector3(100, 0, 20),
        ],
      },
      {
        width: 3.5,
        shape: [
          new THREE.Vector3(0,   0, 3.5),
          new THREE.Vector3(50,  0, 3.5),
          new THREE.Vector3(100, 0, 23.5),
        ],
      },
    ],
  },
];

const junctions: Junction[] = [
  {
    shape: [
      new THREE.Vector3(100, 0, 0),
      new THREE.Vector3(120, 0, 0),
      new THREE.Vector3(120, 0, 20),
      new THREE.Vector3(100, 0, 20),
    ],
  },
];

// ── 组件 ───────────────────────────────────────────────────

export default function RoadEditor() {
  const viewerRef   = useRef<HTMLDivElement>(null);
  const viewRef     = useRef<itowns.GlobeView | null>(null);

  // 简单的纹理（实际替换成你的 roadTexture）
  const roadTexture = new THREE.TextureLoader().load("/road.png");

  // ① 用 hook 管理道路层，传入 onNeedRender 触发 iTowns 重绘
  const { addToScene } = useRoadLayer({
    edges,
    junctions,
    roadTexture,
    container: viewerRef.current,         // 用于读取视口尺寸
    onNeedRender: () => viewRef.current?.notifyChange(),
  });

  useEffect(() => {
    const container = viewerRef.current;
    if (!container) return;

    let rafId: number;

    const init = () => {
      if (container.clientWidth === 0 || container.clientHeight === 0) {
        rafId = requestAnimationFrame(init);
        return;
      }

      const placement = {
        coord: new itowns.Coordinates("EPSG:4326", 104.0125, 30.4751),
        range: 5000,
      };

      const view = new itowns.GlobeView(container, placement);
      viewRef.current = view;

      view.addLayer(
        new itowns.ColorLayer("Ortho", {
          source: new itowns.TMSSource({
            crs: "EPSG:3857",
            isInverted: true,
            format: "image/png",
            url: "https://tile.openstreetmap.org/${z}/${x}/${y}.png",
          }),
        })
      );

      // ② iTowns 初始化完成后，把道路层加入其 Three.js scene
      const scene = (view as any).scene as THREE.Scene;
      addToScene(scene);

      window.dispatchEvent(new Event("resize"));
    };

    rafId = requestAnimationFrame(init);

    return () => {
      cancelAnimationFrame(rafId);
      viewRef.current?.dispose();
    };
  // addToScene 已由 useRoadLayer 内部稳定，不会无限循环
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  return (
    <div style={{ position: "fixed", inset: 0 }}>
      <div ref={viewerRef} style={{ width: "100%", height: "100%" }} />
      <button
        onClick={() => window.history.back()}
        style={{ position: "absolute", top: 20, left: 20, zIndex: 10000 }}
      >
        退出编辑器
      </button>
    </div>
  );
}
