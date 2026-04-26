/**
 * roadGeometry.ts
 * 纯 Three.js 工具函数，不依赖任何框架。
 * 负责生成路面、路口的几何体并合并。
 */

import * as THREE from "three";
import { Coordinates } from '@itowns/geographic'
import { mergeGeometries } from "three/examples/jsm/utils/BufferGeometryUtils";
import { wgs84ToENU, wgs84ToItowns } from "./parseNetXML";

const asphaltTexture = new THREE.TextureLoader().load('/textures/沥青4.jpg');
asphaltTexture.wrapS = THREE.RepeatWrapping;
asphaltTexture.wrapT = THREE.RepeatWrapping;
asphaltTexture.colorSpace = THREE.SRGBColorSpace;
asphaltTexture.repeat.set(1, 6);

// ─────────────────────────────────────────────
// 公共类型
// ─────────────────────────────────────────────

export interface Lane {
  shape: THREE.Vector3[];
  width: number;
}

export interface Edge {
  id?: string;
  function?: string;
  spreadType: string;
  lanes: Lane[];
}

export interface Junction {
  shape: THREE.Vector3[];
}

export interface RoadMeshResult {
  mesh: THREE.Object3D;
  dispose: () => void;
}

function isFiniteGeoPoint(point: THREE.Vector3 | undefined): point is THREE.Vector3 {
  return !!point
    && Number.isFinite(point.x)
    && Number.isFinite(point.y)
    && Number.isFinite(point.z);
}

// ─────────────────────────────────────────────
// 内部：几何体构造
// ─────────────────────────────────────────────

function toShape(points: THREE.Vector3[]): THREE.Shape {
  return new THREE.Shape(points.map((p) => new THREE.Vector2(p.x, p.z)));
}

/**
 * 根据路口轮廓点生成 ShapeGeometry。
 */
export function createJunctionGeometry(pathPoints: THREE.Vector3[]): {
  geometry: THREE.BufferGeometry;
  position: THREE.Vector3;
} {
  const base = pathPoints[0].clone();
  const shifted = pathPoints.map((p) => p.clone().sub(base));
  const geometry = new THREE.ShapeGeometry(toShape(shifted));
  geometry.rotateX(Math.PI / 2);
  return { geometry, position: base };
}

/**
 * 根据折线中心点 + 宽度生成道路几何体（填补转角间隙）。
 */
export function createRoadGeometry(
  pathPoints: THREE.Vector3[],
  width: number
): {
  geometry: THREE.BufferGeometry | null;
  position: THREE.Vector3;
} {
  const base = pathPoints[0].clone();
  const pts = pathPoints.map((p) => p.clone().sub(base));
  const shapes: THREE.Shape[] = [];
  let prevLeft: THREE.Vector3 | null = null;
  let prevRight: THREE.Vector3 | null = null;

  for (let i = 0; i < pts.length - 1; i++) {
    const p1 = pts[i];
    const p2 = pts[i + 1];
    const dir = new THREE.Vector3().subVectors(p2, p1).normalize();
    const n = new THREE.Vector3(-dir.z, 0, dir.x).multiplyScalar(width / 2);

    const tl = p1.clone().add(n);
    const bl = p2.clone().add(n);
    const tr = p1.clone().sub(n);
    const br = p2.clone().sub(n);

    shapes.push(toShape([tl, bl, br, tr]));

    if (prevLeft && prevRight) {
      shapes.push(toShape([prevLeft, tl, p1]));
      shapes.push(toShape([prevRight, tr, p1]));
    }

    prevLeft = bl.clone();
    prevRight = br.clone();
  }

  if (shapes.length === 0) return { geometry: null, position: base };
  const geometry = new THREE.ShapeGeometry(shapes);
  geometry.rotateX(Math.PI / 2);
  return { geometry, position: base };
}

// ─────────────────────────────────────────────
// 批量合并几何体
// ─────────────────────────────────────────────

export function buildMergedRoadGeometry(edges: Edge[]): THREE.Group | null {

  let g = new THREE.Group();
  for (const edge of edges) {

    if (edge.function === "internal") continue;
    for (const lane of edge.lanes) {
      if (lane.shape.length < 2) continue;

      const mesh = createRoadMeshFromLonLat(lane.shape, lane.width);
      if (mesh) {
        g.add(mesh)
        g.updateMatrixWorld();
      }
    }
  }
  return g;

}

export function buildMergedJunctionGeometry(junctions: Junction[]): THREE.BufferGeometry | null {
  const geos: THREE.BufferGeometry[] = [];

  for (const junction of junctions) {
    if (junction.shape.length < 3) continue;
    const { geometry, position } = createJunctionGeometry(junction.shape);
    geometry.translate(position.x, position.y, position.z);
    geos.push(geometry);
  }

  return geos.length > 0 ? mergeGeometries(geos, false) : null;
}

export function buildMergedJunctionGroup(
  junctions: Junction[],
  material?: THREE.MeshStandardMaterial
): THREE.Group | null {
  const group = new THREE.Group();

  for (const junction of junctions) {
    if (junction.shape.length < 3) continue;
    const mesh = createJunctionMeshFromLonLat(junction.shape, material);
    if (mesh) {
      group.add(mesh);
      group.updateMatrixWorld();
    }
  }

  return group.children.length > 0 ? group : null;
}

// ─────────────────────────────────────────────
// 创建路面 Mesh（含模板缓冲，供车道线遮罩）
// ─────────────────────────────────────────────

export function createRoadMesh(
  edges: Edge[],
  roadTexture: THREE.Texture
): THREE.Group | null {

  return buildMergedRoadGeometry(edges)
}

// ─────────────────────────────────────────────
// 创建路口 Mesh
// ─────────────────────────────────────────────

export function createJunctionMesh(
  junctions: Junction[],
  _roadTexture: THREE.Texture
): THREE.Group | null {
  const material = new THREE.MeshStandardMaterial({
    map: asphaltTexture,
    color: 0xd6d6d6,
    side: THREE.DoubleSide,
    roughness: 0.95,
    metalness: 0.02,
  });

  return buildMergedJunctionGroup(junctions, material);

}


function createECEFOrientation(origin: THREE.Vector3, altitudeOffset = 0.35) {
  const altitude = (origin.z || 0) + altitudeOffset;
  const coord = new Coordinates('EPSG:4326', origin.x, origin.y, altitude)
  const position = coord.as('EPSG:4978')
  const posVec = new THREE.Vector3().copy(position)
  const up = posVec.clone().normalize()
  const coordEast = new Coordinates('EPSG:4326', origin.x + 0.0001, origin.y, altitude)
  const eastPos = coordEast.as('EPSG:4978');
  const east = new THREE.Vector3().subVectors(eastPos, position).normalize()
  const north = new THREE.Vector3().crossVectors(up, east).normalize()
  const correctedEast = new THREE.Vector3().crossVectors(north, up).normalize()
  const matrix = new THREE.Matrix4()
  matrix.makeBasis(correctedEast, north, up)
  const quaternion = new THREE.Quaternion().setFromRotationMatrix(matrix)

  return { position: posVec, quaternion }
}

function getAltitudeOffsetByLength(lengthMeters: number, baseOffset = 0.35) {
  if (lengthMeters <= 0) {
    return baseOffset;
  }

  return Math.max(baseOffset, Math.min(3, baseOffset + lengthMeters / 1200));
}

export function createJunctionMeshFromLonLat(
  pathPoints: THREE.Vector3[],
  material?: THREE.Material
): THREE.Mesh | null {
  const validPoints = pathPoints.filter(isFiniteGeoPoint);
  if (validPoints.length < 3) return null;

  const origin = validPoints[0];
  const enuPoints: THREE.Vector3[] = validPoints.map((point) =>
    wgs84ToENU(
      { lon: point.x, lat: point.y, alt: point.z || 0 },
      { lon: origin.x, lat: origin.y, alt: origin.z || 0 }
    )
  );

  const shape = new THREE.Shape(
    enuPoints.map((point) => new THREE.Vector2(point.x, point.y))
  );

  const geometry = new THREE.ShapeGeometry(shape);
  const defaultMaterial = new THREE.MeshStandardMaterial({
    map: asphaltTexture,
    color: 0xd6d6d6,
    side: THREE.DoubleSide,
    roughness: 0.95,
    metalness: 0.02,
    polygonOffset: true,
    polygonOffsetFactor: -2,
    polygonOffsetUnits: -2,
  });

  const mesh = new THREE.Mesh(geometry, material ?? defaultMaterial);
  const { position, quaternion } = createECEFOrientation(origin, 0.8);
  mesh.position.copy(position);
  mesh.quaternion.copy(quaternion);

  return mesh;
}

export function createRoadMeshFromLonLat(
  pathPoints: THREE.Vector3[],
  width: number,
  material?: THREE.Material
): THREE.Mesh | null {
  const validPoints = pathPoints.filter(isFiniteGeoPoint);
  if (validPoints.length < 2) return null;

  // 1️⃣ 选取参考中心点（避免精度问题）
  const worldPoints = validPoints.map((point) =>
    wgs84ToItowns(point.x, point.y, point.z || 0)
  );

  // 2️⃣ 转 ENU
  const totalWorldLength = worldPoints.reduce((sum, point, idx) => {
    if (idx === 0) {
      return sum;
    }
    return sum + point.distanceTo(worldPoints[idx - 1]);
  }, 0);
  const altitudeOffset = getAltitudeOffsetByLength(totalWorldLength, 0.8);
  const liftedWorldPoints = validPoints.map((point) =>
    wgs84ToItowns(point.x, point.y, (point.z || 0) + altitudeOffset)
  );
  const basePosition = liftedWorldPoints[0].clone();

  const positions: number[] = [];
  const uvs: number[] = [];
  const indices: number[] = [];

  let index = 0;
  let totalLength = 0;

  for (let i = 0; i < liftedWorldPoints.length - 1; i++) {
    const p1 = liftedWorldPoints[i];
    const p2 = liftedWorldPoints[i + 1];

    const dir = new THREE.Vector3().subVectors(p2, p1).normalize();

    // ⚠️ 注意：ENU 平面是 x=东 y=北
    const up1 = p1.clone().normalize();
    const up2 = p2.clone().normalize();
    const side1 = new THREE.Vector3().crossVectors(up1, dir).normalize().multiplyScalar(width / 2);
    const side2 = new THREE.Vector3().crossVectors(up2, dir).normalize().multiplyScalar(width / 2);

    const left1 = p1.clone().add(side1).sub(basePosition);
    const right1 = p1.clone().sub(side1).sub(basePosition);
    const left2 = p2.clone().add(side2).sub(basePosition);
    const right2 = p2.clone().sub(side2).sub(basePosition);

    // ===== 写入 position =====
    positions.push(
      left1.x, left1.y, left1.z,
      right1.x, right1.y, right1.z,
      left2.x, left2.y, left2.z,
      right2.x, right2.y, right2.z
    );

    // ===== UV（沿道路方向展开）=====
    const segmentLength = p1.distanceTo(p2);
    const v1 = totalLength;
    const v2 = totalLength + segmentLength;

    uvs.push(
      0, v1,
      1, v1,
      0, v2,
      1, v2
    );

    // ===== 两个三角形 =====
    indices.push(
      index, index + 2, index + 1,
      index + 2, index + 3, index + 1
    );

    index += 4;
    totalLength += segmentLength;
  }

  // 3️⃣ 构建 geometry
  const geometry = new THREE.BufferGeometry();
  geometry.setAttribute(
    'position',
    new THREE.Float32BufferAttribute(positions, 3)
  );
  geometry.setAttribute(
    'uv',
    new THREE.Float32BufferAttribute(uvs, 2)
  );
  geometry.setIndex(indices);
  geometry.computeVertexNormals();
  geometry.computeBoundingSphere();

  // 4️⃣ 材质
  const defaultMaterial = new THREE.MeshStandardMaterial({
    map: asphaltTexture,
    color: 0xcfcfcf,
    side: THREE.DoubleSide,
    wireframe: false,
    roughness: 0.92,
    metalness: 0.02,
    polygonOffset: true,
    polygonOffsetFactor: -2,
    polygonOffsetUnits: -2
  });

  const mesh = new THREE.Mesh(geometry, material ?? defaultMaterial);
  mesh.position.copy(basePosition)

  return mesh//createFlatPlaneAtCoord(origin.x,  origin.y,10000,100);
}
