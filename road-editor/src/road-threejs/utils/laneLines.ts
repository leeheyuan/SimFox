import * as THREE from "three";
import { Coordinates } from "@itowns/geographic";
import { createRoadMeshFromLonLat, type Edge } from "./roadGeometry";
import { wgs84ToENU, wgs84ToItowns } from "./parseNetXML";
export interface LaneLineOptions {
  lineWidth?: number;
  elevation?: number;
  renderOrder?: number;
  dashLengthMeters?: number;
  dashGapMeters?: number;
}

const DEFAULTS: Required<LaneLineOptions> = {
  lineWidth: 1.2,
  elevation: 6,
  renderOrder: 4000,
  dashLengthMeters: 4,
  dashGapMeters: 6,
};

function isFiniteGeoPoint(point: THREE.Vector3 | undefined): point is THREE.Vector3 {
  return !!point
    && Number.isFinite(point.x)
    && Number.isFinite(point.y)
    && Number.isFinite(point.z);
}
 
function worldToWgs84(point: THREE.Vector3): THREE.Vector3 {
  const coord = new Coordinates("EPSG:4978", point.x, point.y, point.z);
  const geo = coord.as("EPSG:4326");
  return new THREE.Vector3(geo.x, geo.y, geo.z);
}

function createLaneMaterial(): THREE.MeshBasicMaterial {
  return new THREE.MeshBasicMaterial({
    color: 0xff0000,
    side: THREE.DoubleSide,
    // FIX 1: Enable depthTest so the mesh is correctly occluded/visible in the scene.
    // depthTest:false + depthWrite:false together can cause lines to be invisible
    // depending on the iTowns render order and globe geometry.
    depthTest: true,
    depthWrite: false,
    transparent: true,
    opacity: 0.55,
    toneMapped: false,
  });
}

function buildBoundaryPoints(
  points: THREE.Vector3[],
  width: number,
  side: "left" | "right",
  elevation: number,
): THREE.Vector3[] {
  const validPoints = points.filter(isFiniteGeoPoint);
  if (validPoints.length < 2) {
    return [];
  }

  const worldPoints = validPoints.map((point) => wgs84ToItowns(point.x, point.y, (point.z || 0) + elevation));
  const boundary: THREE.Vector3[] = [];

  for (let i = 0; i < validPoints.length; i++) {
    const currentWorld = worldPoints[i];
    const prevWorld = i > 0 ? worldPoints[i - 1] : null;
    const nextWorld = i < worldPoints.length - 1 ? worldPoints[i + 1] : null;

    let tangent: THREE.Vector3 | null = null;
    if (prevWorld && nextWorld) {
      tangent = nextWorld.clone().sub(prevWorld);
    } else if (nextWorld) {
      tangent = nextWorld.clone().sub(currentWorld);
    } else if (prevWorld) {
      tangent = currentWorld.clone().sub(prevWorld);
    }

    if (!tangent || tangent.lengthSq() < 1e-8) {
      continue;
    }

    tangent.normalize();
    const up = currentWorld.clone().normalize();
    const offset = new THREE.Vector3()
      .crossVectors(up, tangent)
      .normalize()
      .multiplyScalar(width / 2);

    const boundaryWorld = side === "left"
      ? currentWorld.clone().add(offset)
      : currentWorld.clone().sub(offset);

    boundary.push(worldToWgs84(boundaryWorld));
  } 
  return boundary;
}

function createSolidLaneStripMesh(
  pathPoints: THREE.Vector3[],
  opts: Required<LaneLineOptions>,
): THREE.Mesh | null {
  const mesh = createRoadMeshFromLonLat(pathPoints, opts.lineWidth, createLaneMaterial());
  if (!mesh) {
    return null;
  }

  mesh.renderOrder = opts.renderOrder;
  mesh.frustumCulled = false;
  return mesh;
}

function createDashedLaneMesh(
  pathPoints: THREE.Vector3[],
  opts: Required<LaneLineOptions>,
): THREE.Group | null {
  const validPoints = pathPoints.filter(isFiniteGeoPoint);
  if (validPoints.length < 2) {
    return null;
  }

  
  const worldPoints = validPoints.map((p) => wgs84ToItowns(p.x, p.y, p.z || 0));
  const group = new THREE.Group();
  const cycleLength = opts.dashLengthMeters + opts.dashGapMeters;
  let cycleOffset = 0;

  for (let i = 0; i < worldPoints.length - 1; i++) {
    const p1World = worldPoints[i];
    const p2World = worldPoints[i + 1];
    // WGS84 counterparts for lerp→WGS84 conversion
    const p1Geo = validPoints[i];
    const p2Geo = validPoints[i + 1];

    const segmentLength = p1World.distanceTo(p2World);

    if (!Number.isFinite(segmentLength) || segmentLength <= 0.001) {
      continue;
    }
 
    let cursor = cycleOffset > 0 ? -cycleOffset : 0;

    while (cursor < segmentLength) {
      const dashStart = Math.max(cursor, 0);
      const dashEnd = Math.min(cursor + opts.dashLengthMeters, segmentLength);

      if (dashEnd > dashStart + 0.05) {
        const startRatio = dashStart / segmentLength;
        const endRatio = dashEnd / segmentLength;

        // Lerp in WGS84 space — accurate enough for short dash segments
        // and avoids the double-coordinate-conversion that made lines vanish.
        const dashStartGeo = p1Geo.clone().lerp(p2Geo, startRatio);
        const dashEndGeo = p1Geo.clone().lerp(p2Geo, endRatio);
        const dashPath = [dashStartGeo, dashEndGeo];

        const dashMesh = createSolidLaneStripMesh(dashPath, opts);
        if (dashMesh) {
          group.add(dashMesh);
        }
      }

      cursor += cycleLength;
    }

    // How far into the next cycle are we at the end of this segment?
    const consumed = segmentLength + (cycleOffset > 0 ? cycleOffset : 0);
    cycleOffset = cycleLength - (consumed % cycleLength);
    if (cycleOffset >= cycleLength) {
      cycleOffset = 0;
    }
  }

  return group.children.length > 0 ? group : null;
}

function disposeObjectTree(object: THREE.Object3D) {
  object.traverse((child) => {
    if (!(child instanceof THREE.Mesh)) {
      return;
    }

    child.geometry.dispose();
    const materials = Array.isArray(child.material) ? child.material : [child.material];
    materials.forEach((material) => material.dispose());
  });
}

export function buildLaneLines(edges: Edge[], opts: LaneLineOptions = {}): THREE.Group | null {
  const settings = { ...DEFAULTS, ...opts };
  const group = new THREE.Group();

  for (const edge of edges) {
    if (edge.function === "internal") {
      continue;
    }

    const beforeCount = group.children.length;

    edge.lanes.forEach((lane, index) => {
      if (lane.shape.length < 2) {
        return;
      }

      const isFirst = index === 0;
      const isLast = index === edge.lanes.length - 1;
      const leftBoundary = buildBoundaryPoints(lane.shape, lane.width, "left", settings.elevation);
      const leftLine = isFirst
        ? createSolidLaneStripMesh(leftBoundary, settings)
        : createDashedLaneMesh(leftBoundary, settings);

      if (leftLine) {
        group.add(leftLine);
      }

      if (isLast) {
        const rightBoundary = buildBoundaryPoints(lane.shape, lane.width, "right", settings.elevation);
        const rightLine = createSolidLaneStripMesh(rightBoundary, settings);
        if (rightLine) {
          group.add(rightLine);
        }
      }
    });

    if (group.children.length === beforeCount) {
      edge.lanes.forEach((lane) => {
        if (lane.shape.length < 2) {
          return;
        }

        // FIX 4: Fallback path must include elevation, same as the main path.
        // Previously point.z was used raw (no elevation added), so fallback
        // lines appeared at ground level and were buried under the globe mesh.
        const fallbackPath = lane.shape.map((point) =>
          new THREE.Vector3(point.x, point.y, (point.z || 0) + settings.elevation)
        );
        const fallbackLine = createSolidLaneStripMesh(
          fallbackPath,
          {
            ...settings,
            lineWidth: Math.max(settings.lineWidth, lane.width * 0.7),
          },
        );
        if (fallbackLine) {
          group.add(fallbackLine);
        }
      });
    }
  }

  group.userData.laneLineCount = group.children.length;
  return group.children.length > 0 ? group : null;
}

export function updateLaneLinesResolution(
  _lines: THREE.Object3D[],
  _width: number,
  _height: number,
): void {
}

export function disposeLaneLines(lines: THREE.Object3D[]): void {
  lines.forEach((line) => disposeObjectTree(line));
}
