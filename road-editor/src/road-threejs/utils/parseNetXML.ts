/**
 * parseNetXML.ts
 * 解析 SUMO .net.xml 路网文件，返回 Edge[] 和 Junction[]。
 *
 * 坐标转换链：
 *   SUMO局部坐标 (x, y)
 *     → UTM (sumoXY - netOffset，注意是减法)
 *     → WGS84 lon/lat  (proj4)
 *     → iTowns 场景 THREE.Vector3  (地心笛卡尔)
 *
 * 关键：SUMO netOffset 的含义是「原点的UTM坐标取反」，即：
 *   UTM = SUMO - netOffset
 * 不是加法。可用 origBoundary 验证：
 *   netOffset="-408129.03,-3378286.95"
 *   SUMO(0,0) → UTM(0-(-408129), 0-(-3378286)) = (408129, 3378286) → 约 104.04°E 30.53°N ✓
 */

import xml2js from 'xml2js';
import proj4 from 'proj4';
import * as THREE from 'three';
import * as itowns from 'itowns';
import type { Edge, Junction } from '@/road-threejs';

// ── 投影注册（从 net.xml projParameter 动态注册，兜底 UTM48）────

function registerProj(projParam: string): string {
  const key = 'SUMO_NET_PROJ';
  proj4.defs(key, projParam);
  return key;
}

// ── 坐标转换 ──────────────────────────────────────────────────

/**
 * SUMO局部坐标 → UTM（世界坐标）
 *
 * SUMO 定义：netOffset = -(UTM坐标系原点)
 * 因此：UTM = SUMO - netOffset
 */
function sumoToUTM(
  sumoX: number,
  sumoY: number,
  offsetX: number,
  offsetY: number
): [number, number] {
  return [sumoX - offsetX, sumoY - offsetY];
}

/**
 * UTM → WGS84 lon/lat
 */
function utmToWGS84(
  x: number,
  y: number,
  projKey: string
): [number, number] {
  return proj4(projKey, 'WGS84', [x, y]) as [number, number];
}

/**
 * WGS84 → iTowns 地心笛卡尔 THREE.Vector3
 */
export function wgs84ToItowns(lon: number, lat: number, altitude = 500): THREE.Vector3 {

  const coord = new itowns.Coordinates('EPSG:4326', lon, lat, altitude)
  const position = coord.as('EPSG:4978')
  const posVec = new THREE.Vector3().copy(position) 
  return posVec;
}

export function wgs84ToENU(point:any, origin:any) {
  // 1. 参考点转 EPSG:4978
  const originCoord = new itowns.Coordinates('EPSG:4326', origin.lon, origin.lat, origin.alt || 0);
  const originECEF = originCoord.as('EPSG:4978');

  // 2. 目标点转 EPSG:4978
  const pointCoord = new itowns.Coordinates('EPSG:4326', point.lon, point.lat, point.alt || 0);
  const pointECEF = pointCoord.as('EPSG:4978');

  // 3. 计算 ENU 基向量（单位向量）
  const lonRad = THREE.MathUtils.degToRad(origin.lon);
  const latRad = THREE.MathUtils.degToRad(origin.lat);

  const east = new THREE.Vector3(-Math.sin(lonRad), Math.cos(lonRad), 0).normalize();
  const north = new THREE.Vector3(
      -Math.sin(latRad) * Math.cos(lonRad),
      -Math.sin(latRad) * Math.sin(lonRad),
      Math.cos(latRad)
  ).normalize();

  const up = new THREE.Vector3().crossVectors(east, north).normalize(); 
  // 4. 计算差值向量
  const diff = new THREE.Vector3().subVectors(pointECEF, originECEF); 
  // 5. 投影到 ENU 轴
  return new THREE.Vector3(
      diff.dot(east),   // x: East
      diff.dot(north),  // y: North
      diff.dot(up)      // z: Up
  );
}

/**
 * 一步完成：SUMO局部坐标 → iTowns THREE.Vector3
 */
function sumoToItowns(
  sumoX: number,
  sumoY: number,
  offsetX: number,
  offsetY: number,
  projKey: string,
  altitude = 0
): THREE.Vector3 {
  const [utmX, utmY] = sumoToUTM(sumoX, sumoY, offsetX, offsetY);
  const [lon, lat] = utmToWGS84(utmX, utmY, projKey);
  return wgs84ToItowns(lon, lat, altitude);
}


/**
 * 一步完成：SUMO局部坐标 → wgs84 THREE.Vector3
 */
function sumoTowgs84(
  sumoX: number,
  sumoY: number,
  offsetX: number,
  offsetY: number,
  projKey: string,
  altitude = 0
): THREE.Vector3 {
  const [utmX, utmY] = sumoToUTM(sumoX, sumoY, offsetX, offsetY);
  const [lon, lat] = utmToWGS84(utmX, utmY, projKey);
  return new THREE.Vector3(lon, lat, altitude);
}

// ── 辅助解析 ──────────────────────────────────────────────────

function parseOffset(raw: string): [number, number] {
  const parts = raw.split(',').map(parseFloat);
  if (parts.length < 2 || parts.some(isNaN)) {
    console.warn('[parseNetXML] 无法解析 netOffset，使用 [0,0]:', raw);
    return [0, 0];
  }
  return [parts[0], parts[1]];
}

function parseShape(
  shapeRaw: string,
  offsetX: number,
  offsetY: number,
  projKey: string,
  altitude = 0
): THREE.Vector3[] {
  return shapeRaw
    .split(' ')
    .filter(Boolean)
    .map((point) => {
      const [x, y] = point.split(',').map(parseFloat);
      return sumoTowgs84(x, y, offsetX, offsetY, projKey, altitude);
    });
}

// ── 主导出 ────────────────────────────────────────────────────

export interface ParseNetXMLResult {
  edges: Edge[];
  junctions: Junction[];
  /** 路网中心点（WGS84），供 iTowns 定位用 */
  center: { lon: number; lat: number };
}

export async function parseNetXML(xmlString: string): Promise<ParseNetXMLResult> {
  const parser = new xml2js.Parser();
  const result = await parser.parseStringPromise(xmlString);

  if (!result?.net?.edge) {
    throw new Error('Invalid net.xml structure: missing net.edge');
  }

  // ── 1. 读取 <location> ─────────────────────────────────────
  const loc = result.net.location?.[0]?.$;

  // 优先使用 net.xml 自带的投影参数，没有则兜底 UTM48
  const projParam = loc?.projParameter
    ?? '+proj=utm +zone=48 +ellps=WGS84 +datum=WGS84 +units=m +no_defs';
  const projKey = registerProj(projParam);

  // netOffset：注意是减法，见文件头注释
  const [offsetX, offsetY] = loc?.netOffset
    ? parseOffset(loc.netOffset)
    : [0, 0];

  // 用 origBoundary（原始经纬度范围）直接算中心，最准确
  let center = { lon: 104.0125, lat: 30.4751 };
  if (loc?.origBoundary) {
    const [minLon, minLat, maxLon, maxLat] = loc.origBoundary
      .split(',')
      .map(parseFloat);
    center = {
      lon: (minLon + maxLon) / 2,
      lat: (minLat + maxLat) / 2,
    };
  } else if (loc?.convBoundary) {
    // 兜底：用 convBoundary（SUMO局部坐标范围）转换
    const [minX, minY, maxX, maxY] = loc.convBoundary
      .split(',')
      .map(parseFloat);
    const [utmX, utmY] = sumoToUTM((minX + maxX) / 2, (minY + maxY) / 2, offsetX, offsetY);
    const [lon, lat] = utmToWGS84(utmX, utmY, projKey);
    center = { lon, lat };
  }

  // ── 2. 解析 edges ──────────────────────────────────────────
  const edges: Edge[] = result.net.edge.map((edge: any) => {
    const lanes = (edge.lane ?? []).map((lane: any) => ({
      laneId: lane.$.id,
      width: parseFloat(lane.$.width ?? '3.2'),
      shape: parseShape(lane.$?.shape ?? '', offsetX, offsetY, projKey, 0),
    }));

    return {
      id: edge.$.id,
      from: edge.$.from,
      to: edge.$.to,
      function: edge.$.function,
      spreadType: edge.$.spreadType ?? 'right',
      lanes,
    };
  });

  // ── 3. 解析 junctions ─────────────────────────────────────
  const junctions: Junction[] = (result.net.junction ?? []).map((junction: any) => ({
    shape: parseShape(junction.$?.shape ?? '', offsetX, offsetY, projKey, -0.1),
  }));

  return { edges, junctions, center };
}
