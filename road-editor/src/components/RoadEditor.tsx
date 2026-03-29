/**
 * RoadEditor.tsx
 * iTowns 地图 + SUMO net.xml 道路渲染。
 *
 * 支持两种加载方式：
 *   1. 传入 xmlUrl prop → 组件挂载时自动从服务器拉取
 *   2. 点击「导入 net.xml」→ 手动上传本地文件（兜底）
 *
 * 坐标转换链：SUMO局部坐标 → UTM48 → WGS84 → iTowns 场景坐标
 *
 * 用法：
 *   <RoadEditor xmlUrl="/api/map/chengdu.net.xml" />
 *   <RoadEditor />   ← 纯手动上传
 */

import React, { useState, useEffect, useRef, useCallback } from 'react';
import * as itowns from 'itowns';
import * as THREE from 'three';
import { useRoadLayer } from '@/road-threejs';
import { parseNetXML } from '@/road-threejs/utils/parseNetXML';
import type { Edge, Junction } from '@/road-threejs';

// ── 纹理（替换成你的真实路径）─────────────────────────────
const roadTexture = new THREE.TextureLoader().load('/textures/asphalt.png'); 
 // const lineTexture = useLoader(TextureLoader, '/textures/单虚线.png');
  //const junctionTexture = useLoader(TextureLoader, '/textures/沥青6.jpg');
  

// ── Props ──────────────────────────────────────────────────
interface RoadEditorProps {
  /** 服务器上的 net.xml 地址，传入后组件挂载时自动加载 */
  xmlUrl?: string;
}

// ── 从服务器下载 XML 文本 ──────────────────────────────────
async function loadXMLFile(filePath: string): Promise<string> {
  const response = await fetch(filePath);
  if (!response.ok) {
    throw new Error(`下载失败：${response.status} ${response.statusText}`);
  }
  return response.text();
}

// ── 组件 ───────────────────────────────────────────────────
export default function RoadEditor({ xmlUrl }: RoadEditorProps) {
  const viewerRef    = useRef<HTMLDivElement>(null);
  const viewInstance = useRef<itowns.GlobeView | null>(null);

  const [edges,     setEdges]     = useState<Edge[]>([]);
  const [junctions, setJunctions] = useState<Junction[]>([]);
  const [loading,   setLoading]   = useState(false);
  const [error,     setError]     = useState<string | null>(null);

  // ── 公共：解析 XML 字符串并更新状态 ───────────────────────
  const applyXML = useCallback(async (xmlString: string) => {
    const { edges: parsedEdges, junctions: parsedJunctions, center } =
      await parseNetXML(xmlString);

    setEdges(parsedEdges);
    setJunctions(parsedJunctions);

    // 解析完成后把地图视角飞到路网中心
    const view = viewInstance.current;
    if (view && center) {
      view.controls?.lookAtCoordinate(
        {
          coord: new itowns.Coordinates('EPSG:4326', center.lon, center.lat),
          range: 2000,
          tilt:  45,
        },
        true // 动画过渡
      );
    }
  }, []);

  // ── 方式 1：从服务器拉取 ───────────────────────────────────
  useEffect(() => {
    if (!xmlUrl) return;
    setLoading(true);
    setError(null);
    loadXMLFile(xmlUrl)
      .then(applyXML)
      .catch((err: Error) => setError(err.message))
      .finally(() => setLoading(false));
  }, [xmlUrl, applyXML]);

  // ── 方式 2：手动上传本地文件 ───────────────────────────────
  const handleFileChange = useCallback(
    async (e: React.ChangeEvent<HTMLInputElement>) => {
      const file = e.target.files?.[0];
      if (!file) return;
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
    [applyXML]
  );

  // ── 道路层 hook ────────────────────────────────────────
  const { addToScene } = useRoadLayer({
    edges,
    junctions,
    roadTexture,
    container:    viewerRef.current,
    onNeedRender: () => viewInstance.current?.notifyChange(),
  });

  // ── iTowns 初始化 ──────────────────────────────────────
  useEffect(() => {
    const container = viewerRef.current;
    if (!container) return;

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
        })
      );

      const scene = (view as any).scene as THREE.Scene;
      addToScene(scene);
    };

    rafId = requestAnimationFrame(init);

    return () => {
      cancelAnimationFrame(rafId);
      viewInstance.current?.dispose();
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  // ── edges / junctions 变化后重新挂载道路层 ─────────────
  useEffect(() => {
    const view = viewInstance.current;
    if (!view || edges.length === 0) return;
    const scene = (view as any).scene as THREE.Scene;
    addToScene(scene);
  }, [edges, junctions, addToScene]);

  // ── UI ────────────────────────────────────────────────
  return (
    <div style={{ position: 'fixed', inset: 0 }}>
      <div ref={viewerRef} id="viewerDiv" style={{ width: '100%', height: '100%' }} />

      <div
        style={{
          position: 'absolute',
          top: 20,
          left: 20,
          zIndex: 10000,
          display: 'flex',
          gap: 12,
          alignItems: 'center',
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

        {error && (
          <span style={{ color: '#ff4d4f', fontSize: 13 }}>⚠ {error}</span>
        )}
      </div>
    </div>
  );
}
