import React, { useRef, useMemo } from 'react';
import { Canvas, useFrame } from '@react-three/fiber';
import { OrbitControls } from '@react-three/drei';
import * as THREE from 'three';

/**
 * 创建具有指定宽度的道路几何体（矩形带）并返回几何体和中心点
 */
function createRoadGeometry(pathPoints: THREE.Vector3[], width: number): { geometry: THREE.BufferGeometry, center: THREE.Vector3 } {
  const leftSide: THREE.Vector3[] = [];
  const rightSide: THREE.Vector3[] = [];

  const base = pathPoints[0].clone();
  const shiftedPoints = pathPoints.map(p => p.clone().sub(base));

  for (let i = 0; i < shiftedPoints.length - 1; i++) {
    const p1 = shiftedPoints[i];
    const p2 = shiftedPoints[i + 1];
    const direction = new THREE.Vector3().subVectors(p2, p1).normalize();
    const normal = new THREE.Vector3(-direction.z, 0, direction.x).multiplyScalar(width / 2);

    leftSide.push(p1.clone().add(normal));
    rightSide.push(p1.clone().sub(normal));

    if (i === shiftedPoints.length - 2) {
      leftSide.push(p2.clone().add(normal));
      rightSide.push(p2.clone().sub(normal));
    }
  }

  const shapePoints = [...leftSide, ...rightSide.reverse()];
  const shape2D = shapePoints.map(p => new THREE.Vector2(p.x, p.z));
  const shape = new THREE.Shape(shape2D);
  const geometry = new THREE.ShapeGeometry(shape);

  geometry.translate(base.x, 0, base.z);

  // 计算几何中心点
  geometry.computeBoundingBox();
  const bbox = geometry.boundingBox!;
  const center = new THREE.Vector3();
  bbox.getCenter(center);

  return { geometry, center };
}

const TestRoad = () => {
  const testPoints = [
    new THREE.Vector3(2037.28, 0, 2275.95),
    new THREE.Vector3(2036.95, 0, 2274.26),
    new THREE.Vector3(2037.64, 0, 2272.81),
    new THREE.Vector3(2039.34, 0, 2271.61),
    new THREE.Vector3(2042.07, 0, 2270.66),
  ];

  const { geometry, center } = useMemo(() => createRoadGeometry(testPoints, 3.2), []);

  return (
    <Canvas camera={{ position: [center.x, center.y + 20, center.z + 20], fov: 60 }}>
      <ambientLight intensity={0.5} />
      <directionalLight position={[center.x, center.y + 50, center.z]} intensity={1.0} />

      <mesh geometry={geometry} rotation={[0, 0, 0]}>
        <meshStandardMaterial color="gray" side={THREE.DoubleSide} />
      </mesh>

      <gridHelper args={[50, 50]} />
      <axesHelper args={[10]} />

      <OrbitControls
        target={[center.x, center.y, center.z]}
        rotateSpeed={1.0}
        zoomSpeed={1.2}
        panSpeed={0.8}
      />
    </Canvas>
  );
};

export default TestRoad;
