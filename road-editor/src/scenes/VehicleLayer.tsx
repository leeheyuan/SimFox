// src/scenes/VehicleLayer.tsx
import { useVehicleStore } from "@/stores/vehicleStore";
import VehicleMesh from "./Vehicle";

export default function VehicleLayer() {
  const vehicles = useVehicleStore((s) => s.vehicles);

  return (
    <>
      {Object.values(vehicles).map((v) => (
        <VehicleMesh key={v.id} vehicle={v} />
      ))}
    </>
  );
}
