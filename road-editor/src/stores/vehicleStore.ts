// src/stores/vehicleStore.ts
import { create } from "zustand";

export interface Vehicle {
  id: string;
  type: string;
  x: number;
  y: number;
  z: number;
  s: number;
  a?: number; 
}

interface VehicleState {
  vehicles: Record<string, Vehicle>;
  addVehicle: (v: Vehicle) => void;
  updateVehicle: (v: Vehicle) => void;
  removeVehicle: (id: string) => void;
  bulkUpdate: (data: {
    vehicles?: Vehicle[];
    adds?: { id: string; type: string }[];
    removes?: { id: string }[];
  }) => void;
}

export const useVehicleStore = create<VehicleState>((set) => ({
  vehicles: {},
  addVehicle: (v) => set((s) => ({ vehicles: { ...s.vehicles, [v.id]: v } })),
  updateVehicle: (v) =>
    set((s) => ({
      vehicles: { ...s.vehicles, [v.id]: { ...s.vehicles[v.id], ...v } },
    })),
  removeVehicle: (id) =>
    set((s) => {
      const { [id]: _, ...rest } = s.vehicles;
      return { vehicles: rest };
    }), 
  bulkUpdate: ({
    vehicles,
    adds,
    removes,
  }: {
    vehicles?: Vehicle[];
    adds?: { id: string; type: string }[];
    removes?: { id: string }[];
  }) =>
    set((s) => {
      let updated = { ...s.vehicles };

      // 添加新车
      adds?.forEach(({ id, type }) => {
        updated[id] = { id, x: 0, y: 0, z: 0, s: 0, a: 0, type };
      });

      // 更新位置
     vehicles?.forEach((v) => {
        updated[v.id] = { 
          ...updated[v.id],  // 保留已有字段，比如 type
          ...v,              // 覆盖位置、角度等
          type: updated[v.id]?.type || v.type, // 优先已有的 type，没有的话用新传的
        };
      }); ``

      // 删除车辆
      removes?.forEach(({ id }) => {
        delete updated[id];
      });

      // console.log("✅ store updated, vehicles count:", Object.keys(updated).length);
      return { vehicles: updated };
    }),
}));
