/**
 * laneLines.ts
 * 纯 Three.js 工具函数，不依赖任何框架。
 * 负责生成车道线 Line2 对象列表。
 */

import * as THREE from "three";
import { Line2 } from "three/examples/jsm/lines/Line2";
import { LineGeometry } from "three/examples/jsm/lines/LineGeometry";
import { LineMaterial } from "three/examples/jsm/lines/LineMaterial";
import type { Edge } from "./roadGeometry";

// ─────────────────────────────────────────────
// 类型
// ─────────────────────────────────────────────

export interface LaneLineOptions {
  color?: number;
  linewidth?: number;
  dashSize?: number;
  gapSize?: number;
  yOffset?: number;
  renderOrder?: number;
  /** 视口宽度（px），LineMaterial 需要此值计算线宽 */
  resolutionWidth?: number;
  /** 视口高度（px） */
  resolutionHeight?: number;
}

const DEFAULTS: Required<LaneLineOptions> = {
  color: 0xffffff,
  linewidth: 3,
  dashSize: 2,
  gapSize: 1,
  yOffset: 0.01,
  renderOrder: 999,
  resolutionWidth: window.innerWidth,
  resolutionHeight: window.innerHeight,
};

// ─────────────────────────────────────────────
// 单条 Line2 工厂
// ─────────────────────────────────────────────

export function createLaneLine(
  positions: number[],
  dashed: boolean,
  opts: LaneLineOptions = {}
): Line2 {
  const o = { ...DEFAULTS, ...opts };

  const geometry = new LineGeometry();
  geometry.setPositions(positions);

  const material = new LineMaterial({
    color: o.color,
    linewidth: o.linewidth,
    dashed,
    dashSize: o.dashSize,
    gapSize: o.gapSize,
    depthTest: true,
    transparent: true,
    resolution: new THREE.Vector2(o.resolutionWidth, o.resolutionHeight),
  });

  // 模板测试：只在路面（stencilRef=1）区域绘制
  material.stencilWrite = true;
  material.stencilRef = 1;
  material.stencilFunc = THREE.EqualStencilFunc;
  material.stencilFail = THREE.KeepStencilOp;
  material.stencilZFail = THREE.KeepStencilOp;
  material.stencilZPass = THREE.KeepStencilOp;

  material.polygonOffset = true;
  material.polygonOffsetFactor = -1;
  material.polygonOffsetUnits = -1;

  const line = new Line2(geometry, material);
  line.renderOrder = o.renderOrder;
  line.computeLineDistances();

  return line;
}

// ─────────────────────────────────────────────
// 批量构建车道线
// ─────────────────────────────────────────────

/**
 * 规则：
 * - index === 0（最左侧 lane）→ 实线（道路外边界）
 * - index > 0               → 虚线（车道分隔线）
 * - 最后一条 lane 额外生成右侧外边界实线
 */
export function buildLaneLines(edges: Edge[], opts: LaneLineOptions = {}): Line2[] {
  const { yOffset = DEFAULTS.yOffset } = opts;
  const lines: Line2[] = [];

  for (const edge of edges) {
    if (edge.function === "internal") continue;

    edge.lanes.forEach((lane, index) => {
      if (lane.shape.length < 2) return;

      const isFirst = index === 0;
      const isLast = index === edge.lanes.length - 1;

      // 左侧边界
      const leftPos = buildSidePositions(lane.shape, lane.width, "left", yOffset);
      if (leftPos.length >= 6) {
        lines.push(createLaneLine(leftPos, !isFirst, opts));
      }

      // 最后一条 lane 的右侧外边界（实线）
      if (isLast) {
        const rightPos = buildSidePositions(lane.shape, lane.width, "right", yOffset);
        if (rightPos.length >= 6) {
          lines.push(createLaneLine(rightPos, false, opts));
        }
      }
    });
  }

  return lines;
}

// ─────────────────────────────────────────────
// 更新所有 Line2 的视口分辨率（resize 时调用）
// ─────────────────────────────────────────────

export function updateLaneLinesResolution(
  lines: Line2[],
  width: number,
  height: number
): void {
  lines.forEach((line) => {
    (line.material as LineMaterial).resolution.set(width, height);
  });
}

// ─────────────────────────────────────────────
// 销毁 Line2 列表（释放 GPU 资源）
// ─────────────────────────────────────────────

export function disposeLaneLines(lines: Line2[]): void {
  lines.forEach((line) => {
    line.geometry.dispose();
    (line.material as THREE.Material).dispose();
  });
}

// ─────────────────────────────────────────────
// 私有辅助
// ─────────────────────────────────────────────

function buildSidePositions(
  points: THREE.Vector3[],
  width: number,
  side: "left" | "right",
  yOffset: number
): number[] {
  const positions: number[] = [];
  for (let i = 0; i < points.length - 1; i++) {
    const p1 = points[i];
    const p2 = points[i + 1];
    const dir = new THREE.Vector3().subVectors(p2, p1).normalize();
    const n = new THREE.Vector3(-dir.z, 0, dir.x).multiplyScalar(width / 2);
    const sp1 = side === "left" ? p1.clone().add(n) : p1.clone().sub(n);
    const sp2 = side === "left" ? p2.clone().add(n) : p2.clone().sub(n);
    positions.push(sp1.x, sp1.y + yOffset, sp1.z, sp2.x, sp2.y + yOffset, sp2.z);
  }
  return positions;
}
