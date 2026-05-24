
import React, { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import maplibregl, { type GeoJSONSource, type LngLatBoundsLike } from 'maplibre-gl';
import 'maplibre-gl/dist/maplibre-gl.css';
import { parseNetXML, type Connection } from '@/road-threejs/utils/parseNetXML';
import type { Edge, Junction } from '@/road-threejs';

interface RoadEditorWithImportProps {
  xmlUrl?: string;
}

type Bounds = {
  south: number;
  west: number;
  north: number;
  east: number;
};

type BoundsImportResponse = {
  mapName: string;
  osmFile: string;
  netFile: string;
  netXml: string;
};

type ProjectListItem = {
  id: number;
  name: string;
  status: string;
  createdAt: string;
  updatedAt: string;
};

type TaskListItem = {
  id: number;
  status: string;
  durationSeconds: number;
  speed: number;
  monitorPort: number;
  traCIPort: number;
  lastError: string;
  createdAt: string;
  startedAt?: string | null;
  endedAt?: string | null;
};

type ScreenPoint = {
  x: number;
  y: number;
};

type SelectionBoxStyle = {
  left: number;
  top: number;
  width: number;
  height: number;
};

type FeatureCollection = GeoJSON.FeatureCollection<GeoJSON.Geometry>;

type NetworkCollections = {
  roads: FeatureCollection;
  centerLines: FeatureCollection;
  laneLines: FeatureCollection;
  junctions: FeatureCollection;
  turnPaths: FeatureCollection;
  turnMarkers: FeatureCollection;
  stopLines: FeatureCollection;
  laneDirectionMarkers: FeatureCollection;
  laneDirectionShapes: FeatureCollection;
};

const OVERPASS_SERVERS = [
  'https://overpass.kumi.systems/api/interpreter',
  'https://overpass.openstreetmap.fr/api/interpreter',
  'https://overpass-api.de/api/interpreter',
];

const ROAD_SOURCE_ID = 'road-network-source';
const ROAD_CENTER_SOURCE_ID = 'road-network-center-source';
const LANE_LINE_SOURCE_ID = 'road-lane-line-source';
const JUNCTION_SOURCE_ID = 'road-junction-source';
const TURN_PATH_SOURCE_ID = 'road-turn-path-source';
const TURN_MARKER_SOURCE_ID = 'road-turn-marker-source';
const STOP_LINE_SOURCE_ID = 'road-stop-line-source';
const LANE_DIRECTION_SOURCE_ID = 'road-lane-direction-source';
const LANE_DIRECTION_SHAPE_SOURCE_ID = 'road-lane-direction-shape-source';
const SELECTION_SOURCE_ID = 'road-selection-source';

const emptyCollection = (): FeatureCollection => ({ type: 'FeatureCollection', features: [] });

async function loadXMLFile(filePath: string): Promise<string> {
  const response = await fetch(filePath);
  if (!response.ok) {
    throw new Error(`下载失败: ${response.status} ${response.statusText}`);
  }
  return response.text();
}

function getAuthorizationHeader(): string | undefined {
  const token = localStorage.getItem('token');
  if (!token) {
    return undefined;
  }
  return token.startsWith('Bearer ') ? token : `Bearer ${token}`;
}

function getJsonHeaders(): HeadersInit {
  const authorization = getAuthorizationHeader();
  return {
    'Content-Type': 'application/json',
    ...(authorization ? { Authorization: authorization } : {}),
  };
}

function buildOverpassQuery(bounds: Bounds): string {
  return `[out:xml];
(
  way["highway"](${bounds.south},${bounds.west},${bounds.north},${bounds.east});
);
(._;>;);
out body;
`;
}

function buildSelectionBox(start: ScreenPoint, end: ScreenPoint): SelectionBoxStyle {
  return {
    left: Math.min(start.x, end.x),
    top: Math.min(start.y, end.y),
    width: Math.abs(end.x - start.x),
    height: Math.abs(end.y - start.y),
  };
}

function metersPerDegreeLon(latitude: number): number {
  return 111320 * Math.cos((latitude * Math.PI) / 180);
}

function metersPerDegreeLat(): number {
  return 110540;
}

function buildOffsetBoundary(
  points: Array<[number, number]>,
  widthMeters: number,
  side: 'left' | 'right',
): Array<[number, number]> {
  if (points.length < 2) {
    return [];
  }

  const result: Array<[number, number]> = [];

  for (let i = 0; i < points.length; i += 1) {
    const current = points[i];
    const prev = i > 0 ? points[i - 1] : null;
    const next = i < points.length - 1 ? points[i + 1] : null;
    const latitude = current[1];
    const lonScale = metersPerDegreeLon(latitude);
    const latScale = metersPerDegreeLat();

    let dx = 0;
    let dy = 0;

    if (prev && next) {
      dx = (next[0] - prev[0]) * lonScale;
      dy = (next[1] - prev[1]) * latScale;
    } else if (next) {
      dx = (next[0] - current[0]) * lonScale;
      dy = (next[1] - current[1]) * latScale;
    } else if (prev) {
      dx = (current[0] - prev[0]) * lonScale;
      dy = (current[1] - prev[1]) * latScale;
    }

    const length = Math.hypot(dx, dy);
    if (length < 1e-6) {
      result.push(current);
      continue;
    }

    const normalX = (-dy / length) * (widthMeters / 2);
    const normalY = (dx / length) * (widthMeters / 2);
    const offsetX = side === 'left' ? normalX : -normalX;
    const offsetY = side === 'left' ? normalY : -normalY;

    result.push([
      current[0] + offsetX / lonScale,
      current[1] + offsetY / latScale,
    ]);
  }

  return result;
}

function bboxPolygon(bounds: Bounds): FeatureCollection {
  return {
    type: 'FeatureCollection',
    features: [{
      type: 'Feature',
      properties: {},
      geometry: {
        type: 'Polygon',
        coordinates: [[
          [bounds.west, bounds.south],
          [bounds.east, bounds.south],
          [bounds.east, bounds.north],
          [bounds.west, bounds.north],
          [bounds.west, bounds.south],
        ]],
      },
    }],
  };
}

function directionToLabel(dir: string): string {
  switch (dir) {
    case 'l':
    case 'L':
      return '左转';
    case 'r':
    case 'R':
      return '右转';
    case 's':
      return '直行';
    case 't':
      return '掉头';
    default:
      return '';
  }
}

function directionToArrow(dir: string): string {
  switch (dir) {
    case 'l':
    case 'L':
      return '←';
    case 'r':
    case 'R':
      return '→';
    case 's':
      return '↑';
    case 't':
      return '↺';
    default:
      return '';
  }
}

function composeArrowLabel(dirs: string[]): string {
  const unique = Array.from(new Set(dirs));
  const order = ['l', 'L', 's', 'r', 'R', 't'];
  return order
    .filter((dir) => unique.includes(dir))
    .map((dir) => directionToArrow(dir))
    .join('');
}

function getMidPoint(coords: Array<[number, number]>): [number, number] | null {
  if (coords.length === 0) {
    return null;
  }

  if (coords.length === 1) {
    return coords[0];
  }

  let total = 0;
  const lengths: number[] = [];
  for (let i = 0; i < coords.length - 1; i += 1) {
    const dx = coords[i + 1][0] - coords[i][0];
    const dy = coords[i + 1][1] - coords[i][1];
    const len = Math.hypot(dx, dy);
    lengths.push(len);
    total += len;
  }

  if (total <= 0) {
    return coords[Math.floor(coords.length / 2)];
  }

  const half = total / 2;
  let acc = 0;
  for (let i = 0; i < lengths.length; i += 1) {
    const len = lengths[i];
    if (acc + len >= half) {
      const ratio = (half - acc) / len;
      return [
        coords[i][0] + (coords[i + 1][0] - coords[i][0]) * ratio,
        coords[i][1] + (coords[i + 1][1] - coords[i][1]) * ratio,
      ];
    }
    acc += len;
  }

  return coords[coords.length - 1];
}

function getPointAlongLineFromEnd(
  coords: Array<[number, number]>,
  distanceMeters: number,
): [number, number] | null {
  if (coords.length === 0) {
    return null;
  }
  if (coords.length === 1) {
    return coords[0];
  }

  let remaining = distanceMeters;
  for (let i = coords.length - 1; i > 0; i -= 1) {
    const end = coords[i];
    const start = coords[i - 1];
    const latScale = metersPerDegreeLat();
    const lonScale = metersPerDegreeLon((start[1] + end[1]) / 2);
    const dx = (end[0] - start[0]) * lonScale;
    const dy = (end[1] - start[1]) * latScale;
    const len = Math.hypot(dx, dy);

    if (len <= 1e-6) {
      continue;
    }

    if (remaining <= len) {
      const ratio = (len - remaining) / len;
      return [
        start[0] + (end[0] - start[0]) * ratio,
        start[1] + (end[1] - start[1]) * ratio,
      ];
    }

    remaining -= len;
  }

  return coords[0];
}

function buildStopLineSegment(
  coords: Array<[number, number]>,
  widthMeters: number,
  offsetFromEndMeters = 2,
): Array<[number, number]> | null {
  if (coords.length < 2) {
    return null;
  }

  const center = getPointAlongLineFromEnd(coords, offsetFromEndMeters);
  const anchor = getPointAlongLineFromEnd(coords, Math.max(offsetFromEndMeters + 1.5, 3.5));
  if (!center || !anchor) {
    return null;
  }

  const latScale = metersPerDegreeLat();
  const lonScale = metersPerDegreeLon(center[1]);
  const dx = (center[0] - anchor[0]) * lonScale;
  const dy = (center[1] - anchor[1]) * latScale;
  const len = Math.hypot(dx, dy);
  if (len <= 1e-6) {
    return null;
  }

  const nx = -dy / len;
  const ny = dx / len;
  const halfWidth = widthMeters * 0.48;

  return [
    [center[0] + (nx * halfWidth) / lonScale, center[1] + (ny * halfWidth) / latScale],
    [center[0] - (nx * halfWidth) / lonScale, center[1] - (ny * halfWidth) / latScale],
  ];
}

function getHeadingRadians(
  coords: Array<[number, number]>,
  distanceFromEndMeters = 5,
): number | null {
  if (coords.length < 2) {
    return null;
  }

  const end = getPointAlongLineFromEnd(coords, 0);
  const anchor = getPointAlongLineFromEnd(coords, distanceFromEndMeters);
  if (!end || !anchor) {
    return null;
  }

  const latScale = metersPerDegreeLat();
  const lonScale = metersPerDegreeLon((end[1] + anchor[1]) / 2);
  const dx = (end[0] - anchor[0]) * lonScale;
  const dy = (end[1] - anchor[1]) * latScale;
  if (Math.hypot(dx, dy) <= 1e-6) {
    return null;
  }

  return Math.atan2(dy, dx);
}

function toLonLatOffset(origin: [number, number], xMeters: number, yMeters: number): [number, number] {
  const lonScale = metersPerDegreeLon(origin[1]);
  const latScale = metersPerDegreeLat();
  return [
    origin[0] + xMeters / lonScale,
    origin[1] + yMeters / latScale,
  ];
}

function rotateLocalPoint(point: [number, number], heading: number): [number, number] {
  const cos = Math.cos(heading);
  const sin = Math.sin(heading);
  return [
    point[0] * cos - point[1] * sin,
    point[0] * sin + point[1] * cos,
  ];
}

function localPolygonToGeo(origin: [number, number], heading: number, points: Array<[number, number]>): Array<[number, number]> {
  const adjustedHeading = heading - Math.PI / 2;
  return points.map((point) => {
    const rotated = rotateLocalPoint(point, adjustedHeading);
    return toLonLatOffset(origin, rotated[0], rotated[1]);
  });
}

function buildArrowPolygonPoints(kind: string): Array<[number, number]> {
  switch (kind) {
    case 'left':
      return [
        [-0.25, -3.2],
        [0.25, -3.2],
        [0.25, 0.4],
        [-1.2, 0.4],
        [-1.2, 1.1],
        [-3.0, 0],
        [-1.2, -1.1],
        [-1.2, -0.4],
        [-0.25, -0.4],
      ];
    case 'right':
      return [
        [-0.25, -3.2],
        [0.25, -3.2],
        [0.25, -0.4],
        [1.2, -0.4],
        [1.2, -1.1],
        [3.0, 0],
        [1.2, 1.1],
        [1.2, 0.4],
        [-0.25, 0.4],
      ];
    case 'uturn':
      return [
        [-0.25, -3.0],
        [0.25, -3.0],
        [0.25, 0.6],
        [1.1, 0.6],
        [1.1, -0.1],
        [2.6, 1.1],
        [1.1, 2.2],
        [1.1, 1.4],
        [-1.1, 1.4],
        [-1.1, 3.2],
        [-1.6, 3.2],
        [-1.6, 0.9],
        [-0.25, 0.9],
      ];
    case 'straight':
    default:
      return [
        [-0.35, -3.2],
        [0.35, -3.2],
        [0.35, 0.8],
        [1.1, 0.8],
        [0, 3.0],
        [-1.1, 0.8],
        [-0.35, 0.8],
      ];
  }
}

function buildArrowShapeFeature(
  center: [number, number],
  heading: number,
  kind: 'straight' | 'left' | 'right' | 'uturn',
  offsetXMeters = 0,
): GeoJSON.Feature {
  const points = buildArrowPolygonPoints(kind).map(([x, y]) => [x + offsetXMeters, y] as [number, number]);
  const coords = localPolygonToGeo(center, heading, points);
  return {
    type: 'Feature',
    properties: { kind },
    geometry: {
      type: 'Polygon',
      coordinates: [[...coords, coords[0]]],
    },
  };
}

function buildNetworkCollections(edges: Edge[], junctions: Junction[], connections: Connection[]): NetworkCollections {
  const roads: GeoJSON.Feature[] = [];
  const centerLines: GeoJSON.Feature[] = [];
  const laneLines: GeoJSON.Feature[] = [];
  const junctionFeatures: GeoJSON.Feature[] = [];
  const turnPaths: GeoJSON.Feature[] = [];
  const turnMarkers: GeoJSON.Feature[] = [];
  const stopLines: GeoJSON.Feature[] = [];
  const laneDirectionMarkers: GeoJSON.Feature[] = [];
  const laneDirectionShapes: GeoJSON.Feature[] = [];

  const connectionByVia = new Map(
    connections.map((connection) => [connection.via, connection]),
  );
  const connectionsByFromLane = new Map<string, Connection[]>();
  connections.forEach((connection) => {
    const key = `${connection.from}|${connection.fromLane}`;
    const items = connectionsByFromLane.get(key) ?? [];
    items.push(connection);
    connectionsByFromLane.set(key, items);
  });

  edges.forEach((edge) => {
    if (edge.function === 'internal') {
      edge.lanes.forEach((lane) => {
        const coords = lane.shape
          .filter((point) => Number.isFinite(point.x) && Number.isFinite(point.y))
          .map((point) => [point.x, point.y] as [number, number]);

        if (coords.length < 2) {
          return;
        }

        turnPaths.push({
          type: 'Feature',
          properties: {
            edgeId: edge.id,
            laneId: (lane as { laneId?: string }).laneId,
            dir: connectionByVia.get((lane as { laneId?: string }).laneId ?? '')?.dir ?? '',
            label: directionToLabel(connectionByVia.get((lane as { laneId?: string }).laneId ?? '')?.dir ?? ''),
          },
          geometry: {
            type: 'LineString',
            coordinates: coords,
          },
        });

        const connection = connectionByVia.get((lane as { laneId?: string }).laneId ?? '');
        const label = directionToLabel(connection?.dir ?? '');
        const midPoint = getMidPoint(coords);
        if (label && midPoint) {
          turnMarkers.push({
            type: 'Feature',
            properties: {
              edgeId: edge.id,
              laneId: (lane as { laneId?: string }).laneId,
              dir: connection?.dir ?? '',
              label,
            },
            geometry: {
              type: 'Point',
              coordinates: midPoint,
            },
          });
        }
      });
      return;
    }

    edge.lanes.forEach((lane, index) => {
      const coords = lane.shape
        .filter((point) => Number.isFinite(point.x) && Number.isFinite(point.y))
        .map((point) => [point.x, point.y] as [number, number]);

      if (coords.length < 2) {
        return;
      }

      centerLines.push({
        type: 'Feature',
        properties: {
          edgeId: edge.id,
          laneId: (lane as { laneId?: string }).laneId,
        },
        geometry: {
          type: 'LineString',
          coordinates: coords,
        },
      });

      const leftBoundary = buildOffsetBoundary(coords, lane.width, 'left');
      const rightBoundary = buildOffsetBoundary(coords, lane.width, 'right');

      if (leftBoundary.length >= 2 && rightBoundary.length >= 2) {
        roads.push({
          type: 'Feature',
          properties: {
            edgeId: edge.id,
            laneId: (lane as { laneId?: string }).laneId,
            laneWidth: lane.width,
          },
          geometry: {
            type: 'Polygon',
            coordinates: [[
              ...leftBoundary,
              ...rightBoundary.slice().reverse(),
              leftBoundary[0],
            ]],
          },
        });
      }

      laneLines.push({
        type: 'Feature',
        properties: { kind: index === 0 ? 'solid' : 'dashed', edgeId: edge.id },
        geometry: { type: 'LineString', coordinates: leftBoundary },
      });

      if (index === edge.lanes.length - 1) {
        laneLines.push({
          type: 'Feature',
          properties: { kind: 'solid', edgeId: edge.id },
          geometry: { type: 'LineString', coordinates: rightBoundary },
        });
      }

      const laneIndex = Number.parseInt(((lane as { laneId?: string }).laneId ?? '').split('_').pop() ?? `${index}`, 10);
      const laneConnections = connectionsByFromLane.get(`${edge.id}|${Number.isNaN(laneIndex) ? index : laneIndex}`) ?? [];
      const arrowLabel = composeArrowLabel(laneConnections.map((item) => item.dir));
      const stopLineOffset = Math.max(2.4, Math.min(4.2, lane.width * 0.75));
      const arrowOffset = stopLineOffset + 5.5;
      const arrowPoint = getPointAlongLineFromEnd(coords, arrowOffset);
      const stopLine = buildStopLineSegment(coords, lane.width, stopLineOffset);
      const heading = getHeadingRadians(coords, arrowOffset + 2.5);

      if (stopLine) {
        stopLines.push({
          type: 'Feature',
          properties: {
            edgeId: edge.id,
            laneId: (lane as { laneId?: string }).laneId,
          },
          geometry: {
            type: 'LineString',
            coordinates: stopLine,
          },
        });
      }

      if (arrowLabel && arrowPoint) {
        laneDirectionMarkers.push({
          type: 'Feature',
          properties: {
            edgeId: edge.id,
            laneId: (lane as { laneId?: string }).laneId,
            label: arrowLabel,
          },
          geometry: {
            type: 'Point',
            coordinates: arrowPoint,
          },
        });
      }

      if (heading !== null && arrowPoint) {
        const uniqueDirs = Array.from(new Set(laneConnections.map((item) => item.dir)));
        const normalizedKinds = uniqueDirs.flatMap((dir) => {
          if (dir === 'l' || dir === 'L') return ['left'] as const;
          if (dir === 'r' || dir === 'R') return ['right'] as const;
          if (dir === 's') return ['straight'] as const;
          if (dir === 't') return ['uturn'] as const;
          return [];
        });

        const count = normalizedKinds.length;
        const spacing = 2.6;
        normalizedKinds.forEach((kind, kindIndex) => {
          const offset = count <= 1 ? 0 : (kindIndex - (count - 1) / 2) * spacing;
          laneDirectionShapes.push(
            buildArrowShapeFeature(
              arrowPoint,
              heading,
              kind,
              offset,
            ),
          );
        });
      }
    });
  });

  junctions.forEach((junction) => {
    const coords = junction.shape
      .filter((point) => Number.isFinite(point.x) && Number.isFinite(point.y))
      .map((point) => [point.x, point.y] as [number, number]);

    if (coords.length < 3) {
      return;
    }

    junctionFeatures.push({
      type: 'Feature',
      properties: {},
      geometry: { type: 'Polygon', coordinates: [[...coords, coords[0]]] },
    });
  });

  return {
    roads: { type: 'FeatureCollection', features: roads },
    centerLines: { type: 'FeatureCollection', features: centerLines },
    laneLines: { type: 'FeatureCollection', features: laneLines },
    junctions: { type: 'FeatureCollection', features: junctionFeatures },
    turnPaths: { type: 'FeatureCollection', features: turnPaths },
    turnMarkers: { type: 'FeatureCollection', features: turnMarkers },
    stopLines: { type: 'FeatureCollection', features: stopLines },
    laneDirectionMarkers: { type: 'FeatureCollection', features: laneDirectionMarkers },
    laneDirectionShapes: { type: 'FeatureCollection', features: laneDirectionShapes },
  };
}

function getBoundsFromEdges(edges: Edge[]): LngLatBoundsLike | null {
  let minLon = Number.POSITIVE_INFINITY;
  let maxLon = Number.NEGATIVE_INFINITY;
  let minLat = Number.POSITIVE_INFINITY;
  let maxLat = Number.NEGATIVE_INFINITY;

  edges.forEach((edge) => {
    edge.lanes.forEach((lane) => {
      lane.shape.forEach((point) => {
        if (!Number.isFinite(point.x) || !Number.isFinite(point.y)) {
          return;
        }
        minLon = Math.min(minLon, point.x);
        maxLon = Math.max(maxLon, point.x);
        minLat = Math.min(minLat, point.y);
        maxLat = Math.max(maxLat, point.y);
      });
    });
  });

  if (!Number.isFinite(minLon) || !Number.isFinite(maxLon) || !Number.isFinite(minLat) || !Number.isFinite(maxLat)) {
    return null;
  }

  return [[minLon, minLat], [maxLon, maxLat]];
}

function updateGeoJSONSource(map: maplibregl.Map | null, sourceId: string, data: FeatureCollection) {
  if (!map) {
    return;
  }
  const source = map.getSource(sourceId) as GeoJSONSource | undefined;
  source?.setData(data);
}
function addNetworkLayers(map: maplibregl.Map) {
  if (!map.getSource(ROAD_SOURCE_ID)) {
    map.addSource(ROAD_SOURCE_ID, { type: 'geojson', data: emptyCollection() });
  }
  if (!map.getSource(ROAD_CENTER_SOURCE_ID)) {
    map.addSource(ROAD_CENTER_SOURCE_ID, { type: 'geojson', data: emptyCollection() });
  }
  if (!map.getSource(LANE_LINE_SOURCE_ID)) {
    map.addSource(LANE_LINE_SOURCE_ID, { type: 'geojson', data: emptyCollection() });
  }
  if (!map.getSource(JUNCTION_SOURCE_ID)) {
    map.addSource(JUNCTION_SOURCE_ID, { type: 'geojson', data: emptyCollection() });
  }
  if (!map.getSource(TURN_PATH_SOURCE_ID)) {
    map.addSource(TURN_PATH_SOURCE_ID, { type: 'geojson', data: emptyCollection() });
  }
  if (!map.getSource(TURN_MARKER_SOURCE_ID)) {
    map.addSource(TURN_MARKER_SOURCE_ID, { type: 'geojson', data: emptyCollection() });
  }
  if (!map.getSource(STOP_LINE_SOURCE_ID)) {
    map.addSource(STOP_LINE_SOURCE_ID, { type: 'geojson', data: emptyCollection() });
  }
  if (!map.getSource(LANE_DIRECTION_SOURCE_ID)) {
    map.addSource(LANE_DIRECTION_SOURCE_ID, { type: 'geojson', data: emptyCollection() });
  }
  if (!map.getSource(LANE_DIRECTION_SHAPE_SOURCE_ID)) {
    map.addSource(LANE_DIRECTION_SHAPE_SOURCE_ID, { type: 'geojson', data: emptyCollection() });
  }
  if (!map.getSource(SELECTION_SOURCE_ID)) {
    map.addSource(SELECTION_SOURCE_ID, { type: 'geojson', data: emptyCollection() });
  }

  if (!map.getLayer('road-fill')) {
    map.addLayer({
      id: 'road-fill',
      type: 'fill',
      source: ROAD_SOURCE_ID,
      paint: { 'fill-color': '#2c3138', 'fill-opacity': 0.96 },
    });
  }

  if (!map.getLayer('road-center-line')) {
    map.addLayer({
      id: 'road-center-line',
      type: 'line',
      source: ROAD_CENTER_SOURCE_ID,
      paint: {
        'line-color': '#ff6b6b',
        'line-width': ['interpolate', ['linear'], ['zoom'], 13, 1.2, 16, 2.4, 18, 4],
        'line-opacity': 0.95,
      },
    });
  }

  if (!map.getLayer('junction-fill')) {
    map.addLayer({
      id: 'junction-fill',
      type: 'fill',
      source: JUNCTION_SOURCE_ID,
      paint: { 'fill-color': '#c8d0d9', 'fill-opacity': 0.98 },
    });
  }

  if (!map.getLayer('junction-outline')) {
    map.addLayer({
      id: 'junction-outline',
      type: 'line',
      source: JUNCTION_SOURCE_ID,
      paint: {
        'line-color': '#8b95a1',
        'line-width': ['interpolate', ['linear'], ['zoom'], 13, 0.8, 16, 1.5, 18, 2.2],
      },
    });
  }

  if (!map.getLayer('lane-line-solid')) {
    map.addLayer({
      id: 'lane-line-solid',
      type: 'line',
      source: LANE_LINE_SOURCE_ID,
      filter: ['==', ['get', 'kind'], 'solid'],
      paint: {
        'line-color': '#fff9c4',
        'line-width': ['interpolate', ['linear'], ['zoom'], 14, 1.2, 16, 2.2, 18, 3.2],
      },
    });
  }

  if (!map.getLayer('stop-line')) {
    map.addLayer({
      id: 'stop-line',
      type: 'line',
      source: STOP_LINE_SOURCE_ID,
      paint: {
        'line-color': '#ffffff',
        'line-width': ['interpolate', ['linear'], ['zoom'], 14, 1.4, 16, 2.6, 18, 4],
      },
    });
  }

  if (!map.getLayer('lane-direction-label')) {
    map.addLayer({
      id: 'lane-direction-label',
      type: 'symbol',
      source: LANE_DIRECTION_SOURCE_ID,
      layout: {
        'text-field': ['get', 'label'],
        'text-size': ['interpolate', ['linear'], ['zoom'], 14, 12, 16, 18, 18, 24],
        'text-allow-overlap': true,
        'text-ignore-placement': true,
      },
      paint: {
        'text-color': '#ffffff',
        'text-halo-color': '#111827',
        'text-halo-width': 1.2,
      },
    });
  }

  if (!map.getLayer('lane-direction-shape')) {
    map.addLayer({
      id: 'lane-direction-shape',
      type: 'fill',
      source: LANE_DIRECTION_SHAPE_SOURCE_ID,
      paint: {
        'fill-color': '#ffffff',
        'fill-opacity': 0.96,
      },
    });
  }

  if (!map.getLayer('turn-path-line')) {
    map.addLayer({
      id: 'turn-path-line',
      type: 'line',
      source: TURN_PATH_SOURCE_ID,
      paint: {
        'line-color': '#f8fafc',
        'line-width': ['interpolate', ['linear'], ['zoom'], 14, 1.2, 16, 2.2, 18, 3],
        'line-opacity': 0.9,
        'line-dasharray': [1.2, 1.6],
      },
    });
  }

  if (!map.getLayer('turn-path-arrow')) {
    map.addLayer({
      id: 'turn-path-arrow',
      type: 'symbol',
      source: TURN_PATH_SOURCE_ID,
      layout: {
        'symbol-placement': 'line',
        'symbol-spacing': 80,
        'text-field': '➜',
        'text-size': ['interpolate', ['linear'], ['zoom'], 14, 10, 16, 14, 18, 18],
        'text-keep-upright': false,
        'text-rotation-alignment': 'map',
        'text-allow-overlap': true,
      },
      paint: {
        'text-color': '#f8fafc',
        'text-halo-color': '#1f2937',
        'text-halo-width': 1,
      },
    });
  }

  if (!map.getLayer('turn-marker-label')) {
    map.addLayer({
      id: 'turn-marker-label',
      type: 'symbol',
      source: TURN_MARKER_SOURCE_ID,
      layout: {
        'text-field': ['get', 'label'],
        'text-size': ['interpolate', ['linear'], ['zoom'], 14, 10, 16, 13, 18, 16],
        'text-allow-overlap': true,
        'text-ignore-placement': true,
      },
      paint: {
        'text-color': '#111827',
        'text-halo-color': '#ffffff',
        'text-halo-width': 1.4,
      },
    });
  }

  if (!map.getLayer('lane-line-dashed')) {
    map.addLayer({
      id: 'lane-line-dashed',
      type: 'line',
      source: LANE_LINE_SOURCE_ID,
      filter: ['==', ['get', 'kind'], 'dashed'],
      paint: {
        'line-color': '#ffffff',
        'line-width': ['interpolate', ['linear'], ['zoom'], 14, 1, 16, 1.8, 18, 2.8],
        'line-dasharray': [2, 2],
      },
    });
  }

  if (!map.getLayer('selection-fill')) {
    map.addLayer({
      id: 'selection-fill',
      type: 'fill',
      source: SELECTION_SOURCE_ID,
      paint: { 'fill-color': '#22c55e', 'fill-opacity': 0.16 },
    });
  }

  if (!map.getLayer('selection-outline')) {
    map.addLayer({
      id: 'selection-outline',
      type: 'line',
      source: SELECTION_SOURCE_ID,
      paint: { 'line-color': '#22c55e', 'line-width': 2 },
    });
  }
}

export default function RoadEditorWithImport({ xmlUrl }: RoadEditorWithImportProps) {
  const mapContainerRef = useRef<HTMLDivElement>(null);
  const mapRef = useRef<maplibregl.Map | null>(null);
  const [mapLoaded, setMapLoaded] = useState(false);

  const [edges, setEdges] = useState<Edge[]>([]);
  const [junctions, setJunctions] = useState<Junction[]>([]);
  const [connections, setConnections] = useState<Connection[]>([]);
  const [loading, setLoading] = useState(false);
  const [importing, setImporting] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [selectionEnabled, setSelectionEnabled] = useState(false);
  const [selectedBounds, setSelectedBounds] = useState<Bounds | null>(null);
  const [generatedNetFile, setGeneratedNetFile] = useState<string | null>(null);
  const [dragStart, setDragStart] = useState<ScreenPoint | null>(null);
  const [dragCurrent, setDragCurrent] = useState<ScreenPoint | null>(null);
  const [projects, setProjects] = useState<ProjectListItem[]>([]);
  const [tasks, setTasks] = useState<TaskListItem[]>([]);
  const [selectedProjectId, setSelectedProjectId] = useState<number | null>(null);
  const [simulationDuration, setSimulationDuration] = useState(60);
  const [simulationSpeed, setSimulationSpeed] = useState(1);
  const [panelMessage, setPanelMessage] = useState<string | null>(null);
  const [projectsLoading, setProjectsLoading] = useState(false);
  const [tasksLoading, setTasksLoading] = useState(false);
  const [runSubmitting, setRunSubmitting] = useState(false);
  const [showSimulationPanel, setShowSimulationPanel] = useState(false);

  const networkCollections = useMemo(
    () => buildNetworkCollections(edges, junctions, connections),
    [connections, edges, junctions],
  );
  const dragBoxStyle = useMemo(() => {
    if (!dragStart || !dragCurrent) {
      return null;
    }
    return buildSelectionBox(dragStart, dragCurrent);
  }, [dragCurrent, dragStart]);

  const applyXML = useCallback(async (xmlString: string) => {
    const { edges: parsedEdges, junctions: parsedJunctions, connections: parsedConnections } = await parseNetXML(xmlString);
    setEdges(parsedEdges);
    setJunctions(parsedJunctions);
    setConnections(parsedConnections);
  }, []);

  useEffect(() => {
    if (!xmlUrl) {
      return;
    }

    setLoading(true);
    setError(null);
    loadXMLFile(xmlUrl)
      .then(applyXML)
      .catch((err: Error) => setError(err.message))
      .finally(() => setLoading(false));
  }, [applyXML, xmlUrl]);

  const handleFileChange = useCallback(async (event: React.ChangeEvent<HTMLInputElement>) => {
    const file = event.target.files?.[0];
    if (!file) {
      return;
    }

    setLoading(true);
    setError(null);
    try {
      await applyXML(await file.text());
      setShowSimulationPanel(true);
    } catch (err) {
      setError((err as Error).message);
    } finally {
      setLoading(false);
      event.target.value = '';
    }
  }, [applyXML]);

  const resetSelection = useCallback(() => {
    setSelectionEnabled(false);
    setSelectedBounds(null);
    setGeneratedNetFile(null);
    setDragStart(null);
    setDragCurrent(null);
    setError(null);
    updateGeoJSONSource(mapRef.current, SELECTION_SOURCE_ID, emptyCollection());
  }, []);

  const toGeoCoordinate = useCallback((point: ScreenPoint) => {
    const map = mapRef.current;
    if (!map) {
      return null;
    }
    const lngLat = map.unproject([point.x, point.y]);
    return { lon: lngLat.lng, lat: lngLat.lat };
  }, []);

  const handleSelectionPointerDown = useCallback((event: React.PointerEvent<HTMLDivElement>) => {
    if (!selectionEnabled) {
      return;
    }

    const rect = event.currentTarget.getBoundingClientRect();
    const point = { x: event.clientX - rect.left, y: event.clientY - rect.top };
    mapRef.current?.dragPan.disable();
    setDragStart(point);
    setDragCurrent(point);
  }, [selectionEnabled]);

  const handleSelectionPointerMove = useCallback((event: React.PointerEvent<HTMLDivElement>) => {
    if (!dragStart) {
      return;
    }
    const rect = event.currentTarget.getBoundingClientRect();
    setDragCurrent({ x: event.clientX - rect.left, y: event.clientY - rect.top });
  }, [dragStart]);

  const handleSelectionPointerUp = useCallback((event: React.PointerEvent<HTMLDivElement>) => {
    if (!dragStart) {
      return;
    }

    const rect = event.currentTarget.getBoundingClientRect();
    const endPoint = { x: event.clientX - rect.left, y: event.clientY - rect.top };
    const startCoord = toGeoCoordinate(dragStart);
    const endCoord = toGeoCoordinate(endPoint);

    mapRef.current?.dragPan.enable();
    setDragStart(null);
    setDragCurrent(null);

    if (!startCoord || !endCoord) {
      setError('框选区域获取失败，请调整视图后重试。');
      return;
    }

    const bounds = {
      south: Math.min(startCoord.lat, endCoord.lat),
      west: Math.min(startCoord.lon, endCoord.lon),
      north: Math.max(startCoord.lat, endCoord.lat),
      east: Math.max(startCoord.lon, endCoord.lon),
    };

    setSelectedBounds(bounds);
    setSelectionEnabled(false);
    setError(null);
    updateGeoJSONSource(mapRef.current, SELECTION_SOURCE_ID, bboxPolygon(bounds));
  }, [dragStart, toGeoCoordinate]);
  const importBounds = useCallback(async () => {
    if (!selectedBounds) {
      setError('请先框选一个区域。');
      return;
    }

    setImporting(true);
    setLoading(true);
    setError(null);

    try {
      const authorization = getAuthorizationHeader();
      const query = buildOverpassQuery(selectedBounds);
      let osmBlob: Blob | null = null;
      let lastError = '所有 Overpass 节点都请求失败。';

      for (const server of OVERPASS_SERVERS) {
        try {
          const overpassResponse = await fetch(server, {
            method: 'POST',
            headers: { 'Content-Type': 'text/plain' },
            body: query,
          });

          if (!overpassResponse.ok) {
            lastError = `Overpass 请求失败: ${overpassResponse.status} ${overpassResponse.statusText}`;
            continue;
          }

          const xmlText = await overpassResponse.text();
          if (!xmlText.trim()) {
            lastError = 'Overpass 返回了空的 OSM 数据。';
            continue;
          }

          osmBlob = new Blob([xmlText], { type: 'text/xml' });
          break;
        } catch (err) {
          lastError = (err as Error).message;
        }
      }

      if (!osmBlob) {
        throw new Error(lastError);
      }

      const formData = new FormData();
      formData.append('osmFile', osmBlob, 'selection.osm');
      formData.append('mapName', `bbox_${Date.now()}`);

      const response = await fetch('/simulation-api/map/convert-osm', {
        method: 'POST',
        headers: { ...(authorization ? { Authorization: authorization } : {}) },
        body: formData,
      });

      if (!response.ok) {
        let message = `导入失败: ${response.status}`;
        try {
          const payload = await response.json() as { error?: string; detail?: string };
          message = payload.detail ? `${payload.error}\n${payload.detail}` : payload.error || message;
        } catch {
          message = (await response.text()) || message;
        }
        throw new Error(message);
      }

      const payload = await response.json() as BoundsImportResponse;
      if (!payload.netXml) {
        throw new Error('net.xml 生成成功，但响应中没有返回 XML 内容。');
      }

      await applyXML(payload.netXml);
      setGeneratedNetFile(payload.netFile);
      setShowSimulationPanel(true);
    } catch (err) {
      setError((err as Error).message);
    } finally {
      setImporting(false);
      setLoading(false);
    }
  }, [applyXML, selectedBounds]);

  const fetchProjects = useCallback(async () => {
    const authorization = getAuthorizationHeader();
    if (!authorization) {
      setProjects([]);
      setSelectedProjectId(null);
      setPanelMessage('请先框选一个区域。请先框选一个区域。?');
      return;
    }

    setProjectsLoading(true);
    try {
      const response = await fetch('/simulation-api/project/projects', {
        headers: { Authorization: authorization },
      });
      if (!response.ok) {
        throw new Error(`??导入失败: ${response.status}`);
      }

      const payload = await response.json() as { projects?: ProjectListItem[] };
      const projectItems = payload.projects ?? [];
      setProjects(projectItems);
      setSelectedProjectId((current) => {
        if (current && projectItems.some((item) => item.id === current)) {
          return current;
        }
        return projectItems[0]?.id ?? null;
      });
      setPanelMessage(projectItems.length > 0 ? null : '请先框选一个区域。仿真任务已提交?');
    } catch (err) {
      setPanelMessage((err as Error).message);
    } finally {
      setProjectsLoading(false);
    }
  }, []);

  const fetchTasks = useCallback(async (projectId: number) => {
    const authorization = getAuthorizationHeader();
    if (!authorization) {
      setTasks([]);
      return;
    }

    setTasksLoading(true);
    try {
      const response = await fetch(`/simulation-api/project/${projectId}/tasks`, {
        headers: { Authorization: authorization },
      });
      if (!response.ok) {
        throw new Error(`??导入失败: ${response.status}`);
      }

      const payload = await response.json() as { tasks?: TaskListItem[] };
      setTasks(payload.tasks ?? []);
    } catch (err) {
      setPanelMessage((err as Error).message);
    } finally {
      setTasksLoading(false);
    }
  }, []);

  const runSimulation = useCallback(async () => {
    if (!selectedProjectId) {
      setPanelMessage('请先框选一个区域。个路口');
      return;
    }

    if (!getAuthorizationHeader()) {
      setPanelMessage('当前未登录，无法启动仿真。');
      return;
    }

    setRunSubmitting(true);
    setPanelMessage(null);

    try {
      const response = await fetch(`/simulation-api/project/${selectedProjectId}/run`, {
        method: 'POST',
        headers: getJsonHeaders(),
        body: JSON.stringify({ simulationTime: simulationDuration, speed: simulationSpeed }),
      });
      if (!response.ok) {
        const text = await response.text();
        throw new Error(text || `??导入失败: ${response.status}`);
      }

      const payload = await response.json() as { taskId?: number; message?: string };
      setPanelMessage(payload.taskId ? `仿真任务已提交，任务 ID: ${payload.taskId}` : (payload.message ?? '仿真任务已提交'));
      await fetchTasks(selectedProjectId);
    } catch (err) {
      setPanelMessage((err as Error).message);
    } finally {
      setRunSubmitting(false);
    }
  }, [fetchTasks, selectedProjectId, simulationDuration, simulationSpeed]);

  useEffect(() => {
    const container = mapContainerRef.current;
    if (!container || mapRef.current) {
      return;
    }

    const map = new maplibregl.Map({
      container,
      style: {
        version: 8,
        sources: {
          osm: {
            type: 'raster',
            tiles: ['https://tile.openstreetmap.org/{z}/{x}/{y}.png'],
            tileSize: 256,
            attribution: '© OpenStreetMap contributors',
          },
        },
        layers: [{ id: 'osm-base', type: 'raster', source: 'osm' }],
      },
      center: [104.0125, 30.4751],
      zoom: 15,
      dragRotate: false,
      pitchWithRotate: false,
      touchPitch: false,
    });

    map.addControl(new maplibregl.NavigationControl({ showCompass: false }), 'top-right');
    mapRef.current = map;

    map.on('load', () => {
      addNetworkLayers(map);
      setMapLoaded(true);
      updateGeoJSONSource(map, ROAD_SOURCE_ID, networkCollections.roads);
      updateGeoJSONSource(map, ROAD_CENTER_SOURCE_ID, networkCollections.centerLines);
      updateGeoJSONSource(map, LANE_LINE_SOURCE_ID, networkCollections.laneLines);
      updateGeoJSONSource(map, JUNCTION_SOURCE_ID, networkCollections.junctions);
      updateGeoJSONSource(map, TURN_PATH_SOURCE_ID, networkCollections.turnPaths);
      updateGeoJSONSource(map, TURN_MARKER_SOURCE_ID, networkCollections.turnMarkers);
      updateGeoJSONSource(map, STOP_LINE_SOURCE_ID, networkCollections.stopLines);
      updateGeoJSONSource(map, LANE_DIRECTION_SOURCE_ID, networkCollections.laneDirectionMarkers);
      updateGeoJSONSource(map, LANE_DIRECTION_SHAPE_SOURCE_ID, networkCollections.laneDirectionShapes);
      if (selectedBounds) {
        updateGeoJSONSource(map, SELECTION_SOURCE_ID, bboxPolygon(selectedBounds));
      }
    });

    return () => {
      setMapLoaded(false);
      map.remove();
      mapRef.current = null;
    };
  }, []);

  useEffect(() => {
    const map = mapRef.current;
    if (!map || !mapLoaded || !map.isStyleLoaded()) {
      return;
    }
    updateGeoJSONSource(map, ROAD_SOURCE_ID, networkCollections.roads);
    updateGeoJSONSource(map, ROAD_CENTER_SOURCE_ID, networkCollections.centerLines);
    updateGeoJSONSource(map, LANE_LINE_SOURCE_ID, networkCollections.laneLines);
    updateGeoJSONSource(map, JUNCTION_SOURCE_ID, networkCollections.junctions);
    updateGeoJSONSource(map, TURN_PATH_SOURCE_ID, networkCollections.turnPaths);
    updateGeoJSONSource(map, TURN_MARKER_SOURCE_ID, networkCollections.turnMarkers);
    updateGeoJSONSource(map, STOP_LINE_SOURCE_ID, networkCollections.stopLines);
    updateGeoJSONSource(map, LANE_DIRECTION_SOURCE_ID, networkCollections.laneDirectionMarkers);
    updateGeoJSONSource(map, LANE_DIRECTION_SHAPE_SOURCE_ID, networkCollections.laneDirectionShapes);
  }, [mapLoaded, networkCollections]);

  useEffect(() => {
    const map = mapRef.current;
    if (!map || !mapLoaded || edges.length === 0) {
      return;
    }

    const bounds = getBoundsFromEdges(edges);
    if (bounds) {
      map.fitBounds(bounds, { padding: 80, duration: 600, maxZoom: 18 });
    }
  }, [edges, mapLoaded]);

  useEffect(() => {
    updateGeoJSONSource(mapRef.current, SELECTION_SOURCE_ID, selectedBounds ? bboxPolygon(selectedBounds) : emptyCollection());
  }, [selectedBounds]);

  useEffect(() => {
    if (!showSimulationPanel) {
      return;
    }
    fetchProjects();
  }, [fetchProjects, showSimulationPanel]);

  useEffect(() => {
    if (!showSimulationPanel || !selectedProjectId) {
      setTasks([]);
      return;
    }
    fetchTasks(selectedProjectId);
    const timer = window.setInterval(() => fetchTasks(selectedProjectId), 5000);
    return () => window.clearInterval(timer);
  }, [fetchTasks, selectedProjectId, showSimulationPanel]);

  const latestTask = tasks[0] ?? null;
  const wsAddress = latestTask?.monitorPort
    ? `${window.location.protocol === 'https:' ? 'wss' : 'ws'}://${window.location.hostname}:${latestTask.monitorPort}`
    : null;
  return (
    <div style={{ position: 'fixed', inset: 0, background: '#edf2f7' }}>
      <div ref={mapContainerRef} id="map-editor" style={{ width: '100%', height: '100%' }} />

      {selectionEnabled && (
        <div
          onPointerDown={handleSelectionPointerDown}
          onPointerMove={handleSelectionPointerMove}
          onPointerUp={handleSelectionPointerUp}
          style={{ position: 'absolute', inset: 0, zIndex: 20, cursor: 'crosshair' }}
        >
          {dragBoxStyle && (
            <div
              style={{
                position: 'absolute',
                left: dragBoxStyle.left,
                top: dragBoxStyle.top,
                width: dragBoxStyle.width,
                height: dragBoxStyle.height,
                border: '2px solid #22c55e',
                background: 'rgba(34, 197, 94, 0.16)',
                pointerEvents: 'none',
              }}
            />
          )}
        </div>
      )}

      <div style={{ position: 'absolute', top: 20, left: 20, zIndex: 30, display: 'flex', gap: 12, alignItems: 'center', flexWrap: 'wrap', maxWidth: 'calc(100vw - 40px)' }}>
        <button onClick={() => window.history.back()} style={{ padding: '8px 16px', cursor: 'pointer' }}>退出编辑器</button>

        <label style={{ padding: '8px 16px', background: loading ? '#9ca3af' : '#1677ff', color: '#fff', borderRadius: 4, cursor: loading ? 'not-allowed' : 'pointer' }}>
          {loading ? '加载中...' : '导入 net.xml'}
          <input type="file" accept=".xml" style={{ display: 'none' }} disabled={loading} onChange={handleFileChange} />
        </label>

        <button
          onClick={() => {
            setSelectionEnabled(true);
            setSelectedBounds(null);
            setGeneratedNetFile(null);
            setError(null);
          }}
          style={{ padding: '8px 16px', cursor: 'pointer' }}
          disabled={loading || importing}
        >
          {selectionEnabled ? '仿真任务已提交' : '退出编辑器? OSM'}
        </button>

        <button onClick={importBounds} style={{ padding: '8px 16px', cursor: selectedBounds ? 'pointer' : 'not-allowed' }} disabled={!selectedBounds || loading || importing}>
          {importing ? '加载中...' : '导入 net.xml'}
        </button>

        <button onClick={resetSelection} style={{ padding: '8px 16px', cursor: 'pointer' }} disabled={!selectedBounds && !generatedNetFile && !selectionEnabled}>
          清空选择
        </button>

        {edges.length > 0 && (
          <span style={{ color: '#fff', fontSize: 13, background: 'rgba(15, 23, 42, 0.7)', padding: '4px 10px', borderRadius: 4 }}>
            {edges.length} 条路段 · {junctions.length} 个路口
          </span>
        )}

        {selectedBounds && (
          <span style={{ color: '#fff', fontSize: 13, background: 'rgba(15, 23, 42, 0.7)', padding: '4px 10px', borderRadius: 4 }}>
            SW {selectedBounds.south.toFixed(5)}, {selectedBounds.west.toFixed(5)} · NE {selectedBounds.north.toFixed(5)}, {selectedBounds.east.toFixed(5)}
          </span>
        )}

        {generatedNetFile && (
          <span style={{ color: '#fff', fontSize: 13, background: 'rgba(15, 23, 42, 0.7)', padding: '4px 10px', borderRadius: 4, maxWidth: 420, whiteSpace: 'nowrap', overflow: 'hidden', textOverflow: 'ellipsis' }} title={generatedNetFile}>
            个路口: {generatedNetFile}
          </span>
        )}

        {error && <span style={{ color: '#ef4444', fontSize: 13 }}>? {error}</span>}
      </div>

      {showSimulationPanel && (
        <div style={{ position: 'absolute', top: 20, right: 20, zIndex: 30, width: 340, maxHeight: 'calc(100vh - 40px)', overflow: 'auto', padding: 18, borderRadius: 16, background: 'rgba(8, 12, 20, 0.82)', color: '#f5f7fa', boxShadow: '0 16px 40px rgba(0, 0, 0, 0.28)', backdropFilter: 'blur(14px)' }}>
          <div style={{ fontSize: 18, fontWeight: 700, marginBottom: 6 }}>退出编辑器?</div>
          <div style={{ fontSize: 12, opacity: 0.78, marginBottom: 16 }}>仿真任务已提交 net.xml 个路口,请先框选一个区域。仿真任务已提交?</div>

          <div style={{ display: 'grid', gap: 12 }}>
            <div style={{ padding: '12px 14px', borderRadius: 12, background: 'rgba(255,255,255,0.06)', fontSize: 12, lineHeight: 1.6 }}>
              <div>退出编辑器?: {projects[0]?.name ?? '退出编辑器?'}</div>
              <div>清空选择: {projects[0]?.status ?? '-'}</div>
            </div>

            <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 10 }}>
              <label style={{ display: 'grid', gap: 6, fontSize: 13 }}>
                <span>清空选择(?)</span>
                <input type="number" min={1} value={simulationDuration} onChange={(event) => setSimulationDuration(Math.max(1, Number(event.target.value) || 60))} style={{ padding: '10px 12px', borderRadius: 10, border: '1px solid rgba(255,255,255,0.14)', background: 'rgba(255,255,255,0.08)', color: '#fff' }} />
              </label>

              <label style={{ display: 'grid', gap: 6, fontSize: 13 }}>
                <span>??</span>
                <input type="number" min={0.1} step={0.1} value={simulationSpeed} onChange={(event) => setSimulationSpeed(Math.max(0.1, Number(event.target.value) || 1))} style={{ padding: '10px 12px', borderRadius: 10, border: '1px solid rgba(255,255,255,0.14)', background: 'rgba(255,255,255,0.08)', color: '#fff' }} />
              </label>
            </div>

            <div style={{ display: 'flex', gap: 10 }}>
              <button onClick={runSimulation} disabled={!selectedProjectId || runSubmitting} style={{ flex: 1, padding: '10px 14px', borderRadius: 10, border: 'none', background: runSubmitting ? '#6b7280' : '#16a34a', color: '#fff', cursor: !selectedProjectId || runSubmitting ? 'not-allowed' : 'pointer' }}>
                {runSubmitting ? '加载中...' : '清空选择'}
              </button>
              <button onClick={fetchProjects} disabled={projectsLoading} style={{ padding: '10px 14px', borderRadius: 10, border: '1px solid rgba(255,255,255,0.16)', background: 'rgba(255,255,255,0.08)', color: '#fff', cursor: projectsLoading ? 'not-allowed' : 'pointer' }}>
                {projectsLoading ? '加载中...' : '清空选择'}
              </button>
            </div>

            <div style={{ padding: '12px 14px', borderRadius: 12, background: 'rgba(255,255,255,0.06)', fontSize: 12, lineHeight: 1.6 }}>
              <div>清空选择: {latestTask ? `#${latestTask.id} · ${latestTask.status}` : '??'}</div>
              <div>清空选择: {latestTask?.monitorPort || '-'}</div>
              <div>TraCI ??: {latestTask?.traCIPort || '-'}</div>
              <div>?? WebSocket: {wsAddress ?? '-'}</div>
            </div>

            <div style={{ borderRadius: 12, background: 'rgba(255,255,255,0.05)', padding: 12 }}>
              <div style={{ fontSize: 13, fontWeight: 600, marginBottom: 10 }}>清空选择 {tasksLoading ? '· 个路口' : ''}</div>
              <div style={{ display: 'grid', gap: 8 }}>
                {tasks.length === 0 && <div style={{ fontSize: 12, opacity: 0.72 }}>请先框选一个区域。个路口</div>}
                {tasks.slice(0, 6).map((task) => (
                  <div key={task.id} style={{ padding: '10px 12px', borderRadius: 10, background: 'rgba(0,0,0,0.18)', fontSize: 12 }}>
                    <div>?? #{task.id} · {task.status}</div>
                    <div>?? {task.durationSeconds}s · ?? {task.speed}</div>
                    <div>?? {task.monitorPort || '-'} · TraCI {task.traCIPort || '-'}</div>
                    {task.lastError && <div style={{ color: '#fca5a5', marginTop: 4 }}>{task.lastError}</div>}
                  </div>
                ))}
              </div>
            </div>

            {panelMessage && <div style={{ fontSize: 12, color: panelMessage.includes('??') || panelMessage.includes('个路口') ? '#fca5a5' : '#bbf7d0', lineHeight: 1.5 }}>{panelMessage}</div>}
          </div>
        </div>
      )}
    </div>
  );
}
