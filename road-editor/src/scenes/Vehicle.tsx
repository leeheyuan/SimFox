// src/scenes/Vehicle.tsx
import { Mesh } from "three";
import { useRef } from "react";
import { useFrame } from "@react-three/fiber";
import { Vehicle } from "@/stores/vehicleStore";

interface Props {
  vehicle: Vehicle;
}

export default function VehicleMesh({ vehicle }: Props) {
  const ref = useRef<Mesh>(null);

  useFrame(() => {
    if (ref.current) {
      ref.current.position.set(vehicle.x, 0, vehicle.y);
      if (vehicle.a) {
        ref.current.rotation.y = -vehicle.a;
      }
    }
  });

  return (
    <mesh ref={ref}>
      <boxGeometry args={[1, 0.5, 2]} />
      <meshStandardMaterial color="orange" />
    </mesh>
  );
}


