/**
 * parseNetXML.ts
 * Ã¨Â§Â£Ã¦Å¾Â SUMO .net.xml Ã¨Â·Â¯Ã§Â½â€˜Ã¦â€“â€¡Ã¤Â»Â¶Ã¯Â¼Å’Ã¨Â¿â€Ã¥â€ºÅ¾ Edge[] Ã¥â€™Å’ Junction[]Ã£â‚¬â€š
 *
 * Ã¥ÂÂÃ¦Â â€¡Ã¨Â½Â¬Ã¦ÂÂ¢Ã©â€œÂ¾Ã¯Â¼Å¡
 *   SUMOÃ¥Â±â‚¬Ã©Æ’Â¨Ã¥ÂÂÃ¦Â â€¡ (x, y)
 *     Ã¢â€ â€™ UTM (sumoXY - netOffsetÃ¯Â¼Å’Ã¦Â³Â¨Ã¦â€žÂÃ¦ËœÂ¯Ã¥â€¡ÂÃ¦Â³â€¢)
 *     Ã¢â€ â€™ WGS84 lon/lat  (proj4)
 *     Ã¢â€ â€™ iTowns Ã¥Å“ÂºÃ¦â„¢Â¯ THREE.Vector3  (Ã¥Å“Â°Ã¥Â¿Æ’Ã§Â¬â€ºÃ¥ÂÂ¡Ã¥Â°â€)
 *
 * Ã¥â€¦Â³Ã©â€Â®Ã¯Â¼Å¡SUMO netOffset Ã§Å¡â€žÃ¥ÂÂ«Ã¤Â¹â€°Ã¦ËœÂ¯Ã£â‚¬Å’Ã¥Å½Å¸Ã§â€šÂ¹Ã§Å¡â€žUTMÃ¥ÂÂÃ¦Â â€¡Ã¥Ââ€“Ã¥ÂÂÃ£â‚¬ÂÃ¯Â¼Å’Ã¥ÂÂ³Ã¯Â¼Å¡
 *   UTM = SUMO - netOffset
 * Ã¤Â¸ÂÃ¦ËœÂ¯Ã¥Å Â Ã¦Â³â€¢Ã£â‚¬â€šÃ¥ÂÂ¯Ã§â€Â¨ origBoundary Ã©ÂªÅ’Ã¨Â¯ÂÃ¯Â¼Å¡
 *   netOffset="-408129.03,-3378286.95"
 *   SUMO(0,0) Ã¢â€ â€™ UTM(0-(-408129), 0-(-3378286)) = (408129, 3378286) Ã¢â€ â€™ Ã§ÂºÂ¦ 104.04Ã‚Â°E 30.53Ã‚Â°N Ã¢Å“â€œ
 */

import xml2js from 'xml2js';
import proj4 from 'proj4';
import * as THREE from 'three';
import * as itowns from 'itowns';
import type { Edge, Junction } from '@/road-threejs';

export interface Connection {
  from: string;
  to: string;
  fromLane: number;
  toLane: number;
  via: string;
  dir: string;
  shape: THREE.Vector3[];
}

// Ã¢â€â‚¬Ã¢â€â‚¬ Ã¦Å â€¢Ã¥Â½Â±Ã¦Â³Â¨Ã¥â€ Å’Ã¯Â¼Ë†Ã¤Â»Å½ net.xml projParameter Ã¥Å Â¨Ã¦â‚¬ÂÃ¦Â³Â¨Ã¥â€ Å’Ã¯Â¼Å’Ã¥â€¦Å“Ã¥Âºâ€¢ UTM48Ã¯Â¼â€°Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬

function registerProj(projParam: string): string {
  const key = 'SUMO_NET_PROJ';
  proj4.defs(key, projParam);
  return key;
}

// Ã¢â€â‚¬Ã¢â€â‚¬ Ã¥ÂÂÃ¦Â â€¡Ã¨Â½Â¬Ã¦ÂÂ¢ Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬

/**
 * SUMOÃ¥Â±â‚¬Ã©Æ’Â¨Ã¥ÂÂÃ¦Â â€¡ Ã¢â€ â€™ UTMÃ¯Â¼Ë†Ã¤Â¸â€“Ã§â€¢Å’Ã¥ÂÂÃ¦Â â€¡Ã¯Â¼â€°
 *
 * SUMO Ã¥Â®Å¡Ã¤Â¹â€°Ã¯Â¼Å¡netOffset = -(UTMÃ¥ÂÂÃ¦Â â€¡Ã§Â³Â»Ã¥Å½Å¸Ã§â€šÂ¹)
 * Ã¥â€ºÂ Ã¦Â­Â¤Ã¯Â¼Å¡UTM = SUMO - netOffset
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
 * UTM Ã¢â€ â€™ WGS84 lon/lat
 */
function utmToWGS84(
  x: number,
  y: number,
  projKey: string
): [number, number] {
  return proj4(projKey, 'WGS84', [x, y]) as [number, number];
}

/**
 * WGS84 Ã¢â€ â€™ iTowns Ã¥Å“Â°Ã¥Â¿Æ’Ã§Â¬â€ºÃ¥ÂÂ¡Ã¥Â°â€ THREE.Vector3
 */
export function wgs84ToItowns(lon: number, lat: number, altitude = 500): THREE.Vector3 {

  const coord = new itowns.Coordinates('EPSG:4326', lon, lat, altitude)
  const position = coord.as('EPSG:4978')
  const posVec = new THREE.Vector3().copy(position) 
  return posVec;
}

export function wgs84ToENU(point:any, origin:any) {
  // 1. Ã¥Ââ€šÃ¨â‚¬Æ’Ã§â€šÂ¹Ã¨Â½Â¬ EPSG:4978
  const originCoord = new itowns.Coordinates('EPSG:4326', origin.lon, origin.lat, origin.alt || 0);
  const originECEF = originCoord.as('EPSG:4978');

  // 2. Ã§â€ºÂ®Ã¦Â â€¡Ã§â€šÂ¹Ã¨Â½Â¬ EPSG:4978
  const pointCoord = new itowns.Coordinates('EPSG:4326', point.lon, point.lat, point.alt || 0);
  const pointECEF = pointCoord.as('EPSG:4978');

  // 3. Ã¨Â®Â¡Ã§Â®â€” ENU Ã¥Å¸ÂºÃ¥Ââ€˜Ã©â€¡ÂÃ¯Â¼Ë†Ã¥Ââ€¢Ã¤Â½ÂÃ¥Ââ€˜Ã©â€¡ÂÃ¯Â¼â€°
  const lonRad = THREE.MathUtils.degToRad(origin.lon);
  const latRad = THREE.MathUtils.degToRad(origin.lat);

  const east = new THREE.Vector3(-Math.sin(lonRad), Math.cos(lonRad), 0).normalize();
  const north = new THREE.Vector3(
      -Math.sin(latRad) * Math.cos(lonRad),
      -Math.sin(latRad) * Math.sin(lonRad),
      Math.cos(latRad)
  ).normalize();

  const up = new THREE.Vector3().crossVectors(east, north).normalize(); 
  // 4. Ã¨Â®Â¡Ã§Â®â€”Ã¥Â·Â®Ã¥â‚¬Â¼Ã¥Ââ€˜Ã©â€¡Â
  const diff = new THREE.Vector3().subVectors(pointECEF, originECEF); 
  // 5. Ã¦Å â€¢Ã¥Â½Â±Ã¥Ë†Â° ENU Ã¨Â½Â´
  return new THREE.Vector3(
      diff.dot(east),   // x: East
      diff.dot(north),  // y: North
      diff.dot(up)      // z: Up
  );
}

/**
 * Ã¤Â¸â‚¬Ã¦Â­Â¥Ã¥Â®Å’Ã¦Ë†ÂÃ¯Â¼Å¡SUMOÃ¥Â±â‚¬Ã©Æ’Â¨Ã¥ÂÂÃ¦Â â€¡ Ã¢â€ â€™ iTowns THREE.Vector3
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
 * Ã¤Â¸â‚¬Ã¦Â­Â¥Ã¥Â®Å’Ã¦Ë†ÂÃ¯Â¼Å¡SUMOÃ¥Â±â‚¬Ã©Æ’Â¨Ã¥ÂÂÃ¦Â â€¡ Ã¢â€ â€™ wgs84 THREE.Vector3
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

// Ã¢â€â‚¬Ã¢â€â‚¬ Ã¨Â¾â€¦Ã¥Å Â©Ã¨Â§Â£Ã¦Å¾Â Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬

function parseOffset(raw: string): [number, number] {
  const parts = raw.split(',').map(parseFloat);
  if (parts.length < 2 || parts.some(isNaN)) {
    console.warn('[parseNetXML] Ã¦â€”Â Ã¦Â³â€¢Ã¨Â§Â£Ã¦Å¾Â netOffsetÃ¯Â¼Å’Ã¤Â½Â¿Ã§â€Â¨ [0,0]:', raw);
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

// Ã¢â€â‚¬Ã¢â€â‚¬ Ã¤Â¸Â»Ã¥Â¯Â¼Ã¥â€¡Âº Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬

export interface ParseNetXMLResult {
  edges: Edge[];
  junctions: Junction[];
  /** Ã¨Â·Â¯Ã§Â½â€˜Ã¤Â¸Â­Ã¥Â¿Æ’Ã§â€šÂ¹Ã¯Â¼Ë†WGS84Ã¯Â¼â€°Ã¯Â¼Å’Ã¤Â¾â€º iTowns Ã¥Â®Å¡Ã¤Â½ÂÃ§â€Â¨ */
  center: { lon: number; lat: number };
}

export async function parseNetXML(xmlString: string): Promise<ParseNetXMLResult> {
  const parser = new xml2js.Parser();
  const result = await parser.parseStringPromise(xmlString);

  if (!result?.net?.edge) {
    throw new Error('Invalid net.xml structure: missing net.edge');
  }

  // Ã¢â€â‚¬Ã¢â€â‚¬ 1. Ã¨Â¯Â»Ã¥Ââ€“ <location> Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬
  const loc = result.net.location?.[0]?.$;

  // Ã¤Â¼ËœÃ¥â€¦Ë†Ã¤Â½Â¿Ã§â€Â¨ net.xml Ã¨â€¡ÂªÃ¥Â¸Â¦Ã§Å¡â€žÃ¦Å â€¢Ã¥Â½Â±Ã¥Ââ€šÃ¦â€¢Â°Ã¯Â¼Å’Ã¦Â²Â¡Ã¦Å“â€°Ã¥Ë†â„¢Ã¥â€¦Å“Ã¥Âºâ€¢ UTM48
  const projParam = loc?.projParameter
    ?? '+proj=utm +zone=48 +ellps=WGS84 +datum=WGS84 +units=m +no_defs';
  const projKey = registerProj(projParam);

  // netOffsetÃ¯Â¼Å¡Ã¦Â³Â¨Ã¦â€žÂÃ¦ËœÂ¯Ã¥â€¡ÂÃ¦Â³â€¢Ã¯Â¼Å’Ã¨Â§ÂÃ¦â€“â€¡Ã¤Â»Â¶Ã¥Â¤Â´Ã¦Â³Â¨Ã©â€¡Å 
  const [offsetX, offsetY] = loc?.netOffset
    ? parseOffset(loc.netOffset)
    : [0, 0];

  // Ã§â€Â¨ origBoundaryÃ¯Â¼Ë†Ã¥Å½Å¸Ã¥Â§â€¹Ã§Â»ÂÃ§ÂºÂ¬Ã¥ÂºÂ¦Ã¨Å’Æ’Ã¥â€ºÂ´Ã¯Â¼â€°Ã§â€ºÂ´Ã¦Å½Â¥Ã§Â®â€”Ã¤Â¸Â­Ã¥Â¿Æ’Ã¯Â¼Å’Ã¦Å“â‚¬Ã¥â€¡â€ Ã§Â¡Â®
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
    // Ã¥â€¦Å“Ã¥Âºâ€¢Ã¯Â¼Å¡Ã§â€Â¨ convBoundaryÃ¯Â¼Ë†SUMOÃ¥Â±â‚¬Ã©Æ’Â¨Ã¥ÂÂÃ¦Â â€¡Ã¨Å’Æ’Ã¥â€ºÂ´Ã¯Â¼â€°Ã¨Â½Â¬Ã¦ÂÂ¢
    const [minX, minY, maxX, maxY] = loc.convBoundary
      .split(',')
      .map(parseFloat);
    const [utmX, utmY] = sumoToUTM((minX + maxX) / 2, (minY + maxY) / 2, offsetX, offsetY);
    const [lon, lat] = utmToWGS84(utmX, utmY, projKey);
    center = { lon, lat };
  }

  // Ã¢â€â‚¬Ã¢â€â‚¬ 2. Ã¨Â§Â£Ã¦Å¾Â edges Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬
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

  const laneShapeByID = new Map<string, THREE.Vector3[]>();
  result.net.edge.forEach((edge: any) => {
    (edge.lane ?? []).forEach((lane: any) => {
      const laneId = lane.$?.id;
      if (!laneId) {
        return;
      }

      laneShapeByID.set(
        laneId,
        parseShape(lane.$?.shape ?? '', offsetX, offsetY, projKey, 0),
      );
    });
  });

  // Ã¢â€â‚¬Ã¢â€â‚¬ 3. Ã¨Â§Â£Ã¦Å¾Â junctions Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬Ã¢â€â‚¬
  const junctions: Junction[] = (result.net.junction ?? []).map((junction: any) => ({
    shape: parseShape(junction.$?.shape ?? '', offsetX, offsetY, projKey, -0.1),
  }));

  const connections: Connection[] = (result.net.connection ?? []).map((connection: any) => {
    const attrs = connection.$ ?? {};
    return {
      from: attrs.from ?? '',
      to: attrs.to ?? '',
      fromLane: parseInt(attrs.fromLane ?? '0', 10),
      toLane: parseInt(attrs.toLane ?? '0', 10),
      via: attrs.via ?? '',
      dir: attrs.dir ?? '',
      shape: laneShapeByID.get(attrs.via ?? '') ?? [],
    };
  }).filter((item: Connection) => item.via && item.shape.length >= 2);

  return { edges, junctions, connections, center };
}
