// src/road-threejs/RoadLayerBridge.tsx
import { useEffect } from "react";
import { useThree } from "@react-three/fiber";
import { useRoadLayer } from "@/road-threejs";
import type { UseRoadLayerOptions } from "@/road-threejs";

export function RoadLayerBridge(props: Omit<UseRoadLayerOptions, "scene">) {
  const { scene } = useThree();
  const { addToScene, removeFromScene } = useRoadLayer(props);

  useEffect(() => {
    addToScene(scene);
    return () => removeFromScene();
  }, [scene, addToScene, removeFromScene]);

  return null;
}