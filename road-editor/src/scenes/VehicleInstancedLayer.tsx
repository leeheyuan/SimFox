// src/scenes/VehicleInstancedLayer.tsx
import { useEffect, useRef ,useMemo} from "react";
import { InstancedMesh, Object3D } from "three";
import { useGLTF } from "@react-three/drei";
import { useVehicleStore, Vehicle } from "@/stores/vehicleStore"; 
import { useFrame } from "@react-three/fiber"; 
import * as THREE from "three"; 
import { SimplifyModifier } from "three-stdlib";

const vehicleModels: Record<string, string> = {
  veh_passenger: "/models/vehicle/PassengerCar.glb",
  small: "/models/vehicle/SmallCar.glb",
  special: "/models/vehicle/SpecialCar.glb",
  truck: "/models/vehicle/Truck.glb",
};

// 简单示例：根据 LOD 生成不同 geometry
function generateLODs(nodes: any) {
  const modifier = new SimplifyModifier();

  const origGeo = (nodes.SM_CP_Car_06 as any).geometry;

  // 高精度
  const highGeo = origGeo.clone();

  // 中精度（50% 面数）
  const midGeo = modifier.modify(origGeo.clone(), Math.floor(origGeo.attributes.position.count * 0.5));

  // 低精度（10% 面数）
  const lowGeo = modifier.modify(origGeo.clone(), Math.floor(origGeo.attributes.position.count * 0.1));

  return { highGeo, midGeo, lowGeo };
}

function VehicleInstances({ type }: { type: string }) {
  const meshHighRef = useRef<THREE.InstancedMesh>(null);
  const meshMidRef = useRef<THREE.InstancedMesh>(null);
  const meshLowRef = useRef<THREE.InstancedMesh>(null);
  const dummy = useMemo(() => new THREE.Object3D(), []);

  // GLTF 模型
  const { nodes, materials } = useGLTF(vehicleModels[type]);
  const material = materials[Object.keys(materials)[0]];

  const { highGeo, midGeo, lowGeo } = useMemo(() => generateLODs(nodes), [nodes]);

  // 车辆实例数据
  const vehiclesRef = useRef<Vehicle[]>([]);
  const maxCount = 1000; // 预估最大实例数

  // 每帧根据相机距离切换 LOD 可见性
  useFrame(({ camera }) => {
    if (!vehiclesRef.current.length) return;

    const camPos = camera.position;

    const updateMesh = (meshRef: any, filter: (dist: number) => boolean) => {
      if (!meshRef.current) return;
      let count = 0;
      vehiclesRef.current.forEach((v, i) => {
        const dist = camPos.distanceTo(new THREE.Vector3(v.x, 0, -v.y));
        if (filter(dist)) {
          dummy.position.set(v.x, 0, -v.y);
          dummy.rotation.y = v.a || 0;
          dummy.updateMatrix();
          meshRef.current.setMatrixAt(count, dummy.matrix);
          count++;
        }
      });
      meshRef.current.count = count;
      meshRef.current.instanceMatrix.needsUpdate = true;
    };

    // 高、中、低精度的距离阈值
    updateMesh(meshHighRef, (d) => d < 50);
    updateMesh(meshMidRef, (d) => d >= 50 && d < 150);
    updateMesh(meshLowRef, (d) => d >= 150);
  });

  // 订阅 store 变化
  useEffect(() => {
    const unsubscribe = useVehicleStore.subscribe(
      (state) => state.vehicles,
      (newVehicles) => {
        vehiclesRef.current = Object.values(newVehicles).filter((v) => v.type === type);
      }
    );
    return () => unsubscribe();
  }, [type]);

  return (
    <group>
      {highGeo && <instancedMesh ref={meshHighRef} args={[highGeo, material, maxCount]} />}
      {midGeo && <instancedMesh ref={meshMidRef} args={[midGeo, material, maxCount]} />}
      {lowGeo && <instancedMesh ref={meshLowRef} args={[lowGeo, material, maxCount]} />}
    </group>
  );
}

export default function VehicleInstancedLayer() {
  const vehicles = useVehicleStore.getState().vehicles;
  const types = Array.from(new Set(Object.values(vehicles).map((v) => v.type)));

  return (
    <>
      {types.map((type) => (
        <VehicleInstances key={type} type={type} />
      ))}
    </>
  );
}
