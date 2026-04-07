import React, { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import * as itowns from 'itowns';
import * as THREE from 'three';
import { useRoadLayer } from '@/road-threejs';
import { parseNetXML } from '@/road-threejs/utils/parseNetXML';
import type { Edge, Junction } from '@/road-threejs';

const roadTexture = new THREE.TextureLoader().load('/textures/asphalt.png');

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

const OVERPASS_SERVERS = [
  'https://overpass.kumi.systems/api/interpreter',
  'https://overpass.openstreetmap.fr/api/interpreter',
  'https://overpass-api.de/api/interpreter',
];

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

function buildOverpassQuery(bounds: Bounds): string {
  return `[out:xml];
(
  way["highway"](${bounds.south},${bounds.west},${bounds.north},${bounds.east});
);
(._;>;);
out body;
`;
}

export default function RoadEditorWithImport({ xmlUrl }: RoadEditorWithImportProps) {
  const viewerRef = useRef<HTMLDivElement>(null);
  const viewInstance = useRef<itowns.GlobeView | null>(null);
  const addToSceneRef = useRef<((scene: THREE.Scene) => void) | null>(null);

  const [edges, setEdges] = useState<Edge[]>([]);
  const [junctions, setJunctions] = useState<Junction[]>([]);
  const [loading, setLoading] = useState(false);
  const [importing, setImporting] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [selectionEnabled, setSelectionEnabled] = useState(false);
  const [selectedBounds, setSelectedBounds] = useState<Bounds | null>(null);
  const [generatedNetFile, setGeneratedNetFile] = useState<string | null>(null);
  const [dragStart, setDragStart] = useState<ScreenPoint | null>(null);
  const [dragCurrent, setDragCurrent] = useState<ScreenPoint | null>(null);
  const [selectedBoxStyle, setSelectedBoxStyle] = useState<SelectionBoxStyle | null>(null);

  const applyXML = useCallback(async (xmlString: string) => {
    const { edges: parsedEdges, junctions: parsedJunctions, center } =
      await parseNetXML(xmlString);

    setEdges(parsedEdges);
    setJunctions(parsedJunctions);

    const view = viewInstance.current;
    if (view && center) {
      view.controls?.lookAtCoordinate(
        {
          coord: new itowns.Coordinates('EPSG:4326', center.lon, center.lat),
          range: 2000,
          tilt: 45,
        },
        true,
      );
    }
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

  const handleFileChange = useCallback(
    async (e: React.ChangeEvent<HTMLInputElement>) => {
      const file = e.target.files?.[0];
      if (!file) {
        return;
      }

      setLoading(true);
      setError(null);
      try {
        await applyXML(await file.text());
      } catch (err) {
        setError((err as Error).message);
      } finally {
        setLoading(false);
        e.target.value = '';
      }
    },
    [applyXML],
  );

  const buildSelectionBox = useCallback((start: ScreenPoint, end: ScreenPoint): SelectionBoxStyle => {
    const left = Math.min(start.x, end.x);
    const top = Math.min(start.y, end.y);
    const width = Math.abs(end.x - start.x);
    const height = Math.abs(end.y - start.y);

    return { left, top, width, height };
  }, []);

  const dragBoxStyle = useMemo(() => {
    if (!dragStart || !dragCurrent) {
      return null;
    }

    return buildSelectionBox(dragStart, dragCurrent);
  }, [buildSelectionBox, dragCurrent, dragStart]);

  const resetSelection = useCallback(() => {
    setSelectionEnabled(false);
    setSelectedBounds(null);
    setGeneratedNetFile(null);
    setDragStart(null);
    setDragCurrent(null);
    setSelectedBoxStyle(null);
    setError(null);
  }, []);

  const toGeoCoordinate = useCallback((point: ScreenPoint) => {
    const controls = viewInstance.current?.controls as {
      pickGeoPosition?: (coords: THREE.Vector2) => { longitude: number; latitude: number } | undefined;
    } | undefined;

    if (!controls?.pickGeoPosition) {
      return null;
    }

    const coords = controls.pickGeoPosition(new THREE.Vector2(point.x, point.y));
    if (!coords) {
      return null;
    }

    return {
      lon: coords.longitude,
      lat: coords.latitude,
    };
  }, []);

  const handleSelectionPointerDown = useCallback(
    (event: React.PointerEvent<HTMLDivElement>) => {
      if (!selectionEnabled) {
        return;
      }

      const rect = event.currentTarget.getBoundingClientRect();
      const point = {
        x: event.clientX - rect.left,
        y: event.clientY - rect.top,
      };

      setDragStart(point);
      setDragCurrent(point);
      setSelectedBoxStyle(null);
    },
    [selectionEnabled],
  );

  const handleSelectionPointerMove = useCallback(
    (event: React.PointerEvent<HTMLDivElement>) => {
      if (!dragStart) {
        return;
      }

      const rect = event.currentTarget.getBoundingClientRect();
      setDragCurrent({
        x: event.clientX - rect.left,
        y: event.clientY - rect.top,
      });
    },
    [dragStart],
  );

  const handleSelectionPointerUp = useCallback(
    (event: React.PointerEvent<HTMLDivElement>) => {
      if (!dragStart) {
        return;
      }

      const rect = event.currentTarget.getBoundingClientRect();
      const endPoint = {
        x: event.clientX - rect.left,
        y: event.clientY - rect.top,
      };

      const startCoord = toGeoCoordinate(dragStart);
      const endCoord = toGeoCoordinate(endPoint);

      setDragStart(null);
      setDragCurrent(null);

      if (!startCoord || !endCoord) {
        setSelectedBoxStyle(null);
        setError('框选区域超出地球表面，可调整视角后重试。');
        return;
      }

      setSelectedBoxStyle(buildSelectionBox(dragStart, endPoint));
      setSelectedBounds({
        south: Math.min(startCoord.lat, endCoord.lat),
        west: Math.min(startCoord.lon, endCoord.lon),
        north: Math.max(startCoord.lat, endCoord.lat),
        east: Math.max(startCoord.lon, endCoord.lon),
      });
      setSelectionEnabled(false);
      setError(null);
    },
    [buildSelectionBox, dragStart, toGeoCoordinate],
  );

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
      let lastError = 'All Overpass servers failed.';

      for (const server of OVERPASS_SERVERS) {
        try {
          const overpassResponse = await fetch(server, {
            method: 'POST',
            headers: {
              'Content-Type': 'text/plain',
            },
            body: query,
          });

          if (!overpassResponse.ok) {
            lastError = `Overpass request failed: ${overpassResponse.status} ${overpassResponse.statusText}`;
            continue;
          }

          const xmlText = await overpassResponse.text();
          if (!xmlText.trim()) {
            lastError = 'Overpass returned an empty OSM response.';
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
        headers: {
          ...(authorization ? { Authorization: authorization } : {}),
        },
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
        throw new Error('net.xml generation succeeded but the response did not include XML content.');
      }
      await applyXML(payload.netXml);
      setGeneratedNetFile(payload.netFile);
    } catch (err) {
      setError((err as Error).message);
    } finally {
      setImporting(false);
      setLoading(false);
    }
  }, [applyXML, selectedBounds]);

  const { addToScene } = useRoadLayer({
    edges,
    junctions,
    roadTexture,
    container: viewerRef.current,
    onNeedRender: () => viewInstance.current?.notifyChange(),
  });

  useEffect(() => {
    addToSceneRef.current = addToScene;
  }, [addToScene]);

  useEffect(() => {
    const container = viewerRef.current;
    if (!container) {
      return;
    }

    let rafId: number;

    const init = () => {
      if (container.clientWidth === 0 || container.clientHeight === 0) {
        rafId = requestAnimationFrame(init);
        return;
      }

      const view = new itowns.GlobeView(container, {
        coord: new itowns.Coordinates('EPSG:4326', 104.0125, 30.4751),
        range: 5000,
      });
      viewInstance.current = view;

      const canvas = container.querySelector('canvas');
      if (canvas) {
        Object.assign(canvas.style, {
          position: 'absolute',
          top: '0',
          left: '0',
          width: '100%',
          height: '100%',
        });
      }

      window.dispatchEvent(new Event('resize'));

      view.addLayer(
        new itowns.ColorLayer('Ortho', {
          source: new itowns.TMSSource({
            crs: 'EPSG:3857',
            isInverted: true,
            format: 'image/png',
            url: 'https://tile.openstreetmap.org/${z}/${x}/${y}.png',
          }),
        }),
      );

      const scene = (view as any).scene as THREE.Scene;
      const ambientLight = new THREE.AmbientLight(0xffffff, 1.35);
      const hemiLight = new THREE.HemisphereLight(0xf3f7ff, 0x5f666d, 1.1);
      const directionalLight = new THREE.DirectionalLight(0xffffff, 1.9);
      directionalLight.position.set(800, 1200, 900);
      scene.add(ambientLight);
      scene.add(hemiLight);
      scene.add(directionalLight);
      addToSceneRef.current?.(scene);
    };

    rafId = requestAnimationFrame(init);

    return () => {
      cancelAnimationFrame(rafId);
      viewInstance.current?.dispose();
      viewInstance.current = null;
    };
  }, []);

  useEffect(() => {
    const view = viewInstance.current;
    if (!view || edges.length === 0) {
      return;
    }

    const scene = (view as any).scene as THREE.Scene;
    addToScene(scene);
  }, [addToScene, edges, junctions]);

  return (
    <div style={{ position: 'fixed', inset: 0 }}>
      <div ref={viewerRef} id="viewerDiv" style={{ width: '100%', height: '100%' }} />

      {selectionEnabled && (
        <div
          onPointerDown={handleSelectionPointerDown}
          onPointerMove={handleSelectionPointerMove}
          onPointerUp={handleSelectionPointerUp}
          style={{
            position: 'absolute',
            inset: 0,
            zIndex: 9998,
            cursor: 'crosshair',
          }}
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
                background: 'rgba(34, 197, 94, 0.18)',
                pointerEvents: 'none',
              }}
            />
          )}
        </div>
      )}

      {selectedBoxStyle && !selectionEnabled && (
        <div
          style={{
            position: 'absolute',
            inset: 0,
            zIndex: 9997,
            pointerEvents: 'none',
          }}
        >
          <div
            style={{
              position: 'absolute',
              left: selectedBoxStyle.left,
              top: selectedBoxStyle.top,
              width: selectedBoxStyle.width,
              height: selectedBoxStyle.height,
              border: '2px solid #22c55e',
              background: 'rgba(34, 197, 94, 0.12)',
              boxShadow: '0 0 0 1px rgba(255,255,255,0.35) inset',
            }}
          />
        </div>
      )}

      <div
        style={{
          position: 'absolute',
          top: 20,
          left: 20,
          zIndex: 10000,
          display: 'flex',
          gap: 12,
          alignItems: 'center',
          flexWrap: 'wrap',
          maxWidth: 'calc(100vw - 40px)',
        }}
      >
        <button
          onClick={() => window.history.back()}
          style={{ padding: '8px 16px', cursor: 'pointer' }}
        >
          退出编辑器
        </button>

        <label
          style={{
            padding: '8px 16px',
            background: loading ? '#aaa' : '#1677ff',
            color: '#fff',
            borderRadius: 4,
            cursor: loading ? 'not-allowed' : 'pointer',
          }}
        >
          {loading ? '加载中...' : '导入 net.xml'}
          <input
            type="file"
            accept=".xml"
            style={{ display: 'none' }}
            disabled={loading}
            onChange={handleFileChange}
          />
        </label>

        <button
          onClick={() => {
            setSelectionEnabled(true);
            setSelectedBounds(null);
            setGeneratedNetFile(null);
            setSelectedBoxStyle(null);
            setError(null);
          }}
          style={{ padding: '8px 16px', cursor: 'pointer' }}
          disabled={loading || importing}
        >
          {selectionEnabled ? '请拖拽框选区域' : '框选区域导入 OSM'}
        </button>

        <button
          onClick={importBounds}
          style={{ padding: '8px 16px', cursor: selectedBounds ? 'pointer' : 'not-allowed' }}
          disabled={!selectedBounds || loading || importing}
        >
          {importing ? '转换中...' : '生成 net.xml'}
        </button>

        <button
          onClick={resetSelection}
          style={{ padding: '8px 16px', cursor: 'pointer' }}
          disabled={!selectedBounds && !generatedNetFile && !selectionEnabled}
        >
          清空选择
        </button>

        {edges.length > 0 && (
          <span
            style={{
              color: '#fff',
              fontSize: 13,
              background: 'rgba(0,0,0,.45)',
              padding: '4px 10px',
              borderRadius: 4,
            }}
          >
            {edges.length} 条路段 · {junctions.length} 个路口
          </span>
        )}

        {selectedBounds && (
          <span
            style={{
              color: '#fff',
              fontSize: 13,
              background: 'rgba(0,0,0,.45)',
              padding: '4px 10px',
              borderRadius: 4,
            }}
          >
            SW {selectedBounds.south.toFixed(5)}, {selectedBounds.west.toFixed(5)} · NE {selectedBounds.north.toFixed(5)}, {selectedBounds.east.toFixed(5)}
          </span>
        )}

        {generatedNetFile && (
          <span
            style={{
              color: '#fff',
              fontSize: 13,
              background: 'rgba(0,0,0,.45)',
              padding: '4px 10px',
              borderRadius: 4,
              maxWidth: 420,
              whiteSpace: 'nowrap',
              overflow: 'hidden',
              textOverflow: 'ellipsis',
            }}
            title={generatedNetFile}
          >
            已生成: {generatedNetFile}
          </span>
        )}

        {error && (
          <span style={{ color: '#ff4d4f', fontSize: 13 }}>⚠ {error}</span>
        )}
      </div>
    </div>
  );
}
