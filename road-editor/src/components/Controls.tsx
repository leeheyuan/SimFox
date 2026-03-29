// 引入需要的模块
import { OrbitControls } from '@react-three/drei';
import { useThree } from '@react-three/fiber';
import { useRef, useEffect } from 'react';
import * as THREE from 'three';

const Controls = ({ center }: { center: THREE.Vector3 }) => {
  const controlsRef = useRef<any>();
  const { camera } = useThree();

  useEffect(() => {
    if (controlsRef.current) {
      controlsRef.current.target.copy(center); // 设置 OrbitControls 的目标为 center
      controlsRef.current.update();            // 强制更新
      camera.lookAt(center);                   // 相机也朝向 center
    }
  }, [center]);

  return (
    <OrbitControls
      ref={controlsRef}
      enableZoom={true}
      enableRotate={false}
      enablePan={true}
    />
  );
};
