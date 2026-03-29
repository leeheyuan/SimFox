import React, { useEffect, useState, useRef } from 'react';
import { Canvas, useFrame, useThree, ThreeElements, useLoader } from '@react-three/fiber';
import { OrbitControls, PerspectiveCamera, Stats, Detailed } from '@react-three/drei';
import * as THREE from 'three';
import { Line2 } from "three-stdlib";
import { LineMaterial,LineGeometry } from "three-stdlib"; 
import { useGLTF } from "@react-three/drei";
import { InstancedMesh, Object3D } from "three";
import { TextureLoader, CatmullRomCurve3, TubeGeometry, Vector3, RepeatWrapping } from 'three';
import xml2js from 'xml2js';
import { mergeGeometries } from 'three/examples/jsm/utils/BufferGeometryUtils.js';
import VehicleInstancedLayer from "@/scenes/VehicleInstancedLayer";
declare module 'react' {
  namespace JSX {
    interface IntrinsicElements extends ThreeElements { }
  }
}


// 解析 SUMO 路网配置的函数
const parseNetXML = async (xmlString: string) => {
  const parser = new xml2js.Parser();
  try {
    // 使用解析器解析 XML 数据
    const result = await parser.parseStringPromise(xmlString);
    if (!result || !result.net || !result.net.edge) {
      throw new Error("Invalid XML structure");
    }
    // 提取道路和车道数据
    const edges = result.net.edge.map((edge: any) => {

      const lanes = edge.lane ? edge.lane.map((lane: any) => {

        const shape = lane.$.shape && lane.$.shape ? lane.$.shape.split(' ') : [];
        const shapePoints = shape.map((point: string) => {
          const [x, y] = point.split(',').map(parseFloat);
          return new THREE.Vector3(x, 0, -y); // 将坐标转为 Vector3 格式
        });
        return {
          width: parseFloat(lane.$.width || '3.2'), // 默认宽度为 3.2
          laneId: lane.$.id,
          shape: shapePoints,
        };

      }) : [];
      return {
        id: edge.$.id, // 获取 edge 的 id
        from: edge.$.from, // 获取 from 属性
        to: edge.$.to, // 获取 to 属性
        function: edge.$.function,
        spreadType: edge.$.spreadType,
        lanes: lanes // 包含该 edge 下的所有 lanes
      };
    });


    // 提取道路和车道数据
    const junctions = result.net.junction.map((junction: any) => {
      const shape = junction.$.shape && junction.$.shape ? junction.$.shape.split(' ') : [];
      const shapePoints = shape.map((point: string) => {
        const [x, y] = point.split(',').map(parseFloat);
        return new THREE.Vector3(x, -0.1, -y); // 将坐标转为 Vector3 格式
      });
      return {
        shape: shapePoints,
      };
    });
    return { edges: edges, junctions: junctions };
  } catch (error) {
    console.error("Error parsing XML:", error);
    return { edges: null, junctions: null };
  }
};

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


const computeCenter = (edges: any[]) => {
  const points: THREE.Vector3[] = [];

  edges.forEach(edge =>
    edge.lanes.forEach((lane: any) =>
      lane.shape.forEach((point: THREE.Vector3) => points.push(point))
    )
  );

  if (points.length === 0) return new THREE.Vector3(0, 0, 0);

  const center = new THREE.Vector3();
  points.forEach(p => center.add(p));
  center.divideScalar(points.length);

  return center;
};


const CameraController = ({ center }: { center: THREE.Vector3 }) => {
  const { camera } = useThree();

  useEffect(() => {


    camera.position.set(center.x, center.y + 500, center.z); // 高空鸟瞰
    //camera.position.set(0, 100, 0); // 高空鸟瞰
    camera.lookAt(center);


    // camera.position.set(2290.75,  1000, 2232.04); // 高空鸟瞰
    // camera.lookAt(new THREE.Vector3(2290.75,0,2232.04));

  }, [camera, center]);

  return null;
};


function createJunctionGeometry(pathPoints: THREE.Vector3[], width: number, spreadType: string, scene: THREE.Scene) {

  const base = pathPoints[0].clone()
  const shiftedPoints = pathPoints.map(p => p.clone().sub(base))
  const shape2D = shiftedPoints.map(p => new THREE.Vector2(p.x, p.z))
  const shape = new THREE.Shape(shape2D)
  const geometry = new THREE.ShapeGeometry(shape);
  geometry.rotateX(Math.PI / 2);
  return { geometry, position: base };

}

function RoadJunctions({ junctions, roadTexture }: { junctions: any[], roadTexture: THREE.Texture }) {
  // ✅ 顶层获取 scene
  const { scene } = useThree();
  // 组合几何体
  const mergedGeometry = React.useMemo(() => {
    if (junctions.length === 0) return null;

    const geometries: THREE.BufferGeometry[] = [];
    junctions.forEach(junction => {
      if (junction.shape.length < 3) return;
      const { geometry, position } = createJunctionGeometry(junction.shape, 20, "", scene);
      if (geometry) {

        geometry.translate(position.x, position.y, position.z);
        geometries.push(geometry);

      }

    });
    if (geometries.length === 0) return null;
    return mergeGeometries(geometries, false);
  }, [junctions]);

  if (!mergedGeometry) return null;


  return (
    <mesh geometry={mergedGeometry}>
      <meshStandardMaterial
        map={roadTexture}
        side={THREE.DoubleSide}
        wireframe={false}
      />
      <axesHelper args={[10000]} />
    </mesh>
  );
}

function RoadMesh({ edges, roadTexture }: { edges: any[], roadTexture: THREE.Texture }) {
  // ✅ 顶层获取 scene
  const { scene } = useThree();
  // 组合几何体
  const mergedGeometry = React.useMemo(() => {
    if (edges.length === 0) return null;

    const geometries: THREE.BufferGeometry[] = [];
    edges.forEach(edge => {
      if (edge.function && edge.function === "internal") return;
      edge.lanes.forEach((lane: any) => {
        const points: THREE.Vector3[] = lane.shape;
        if (points.length < 2) return;
        const { geometry, position } = createRoadGeometry(points, lane.width, edge.spreadType, scene);
        if (geometry) {
          geometry.translate(position.x, position.y, position.z);
          geometries.push(geometry);
        }

      });
    });

    if (geometries.length === 0) return null;
    return mergeGeometries(geometries, false);
  }, [edges]);

  if (!mergedGeometry) return null;


  return (
    <mesh geometry={mergedGeometry} renderOrder={1}>
      <meshStandardMaterial
        map={roadTexture}
        side={THREE.DoubleSide}
        wireframe={false}
        stencilWrite={true}
        stencilRef={1}
        depthWrite={true}
        stencilFunc={THREE.AlwaysStencilFunc} // 总是通过测试
        stencilZPass={THREE.IncrementStencilOp} // 每画一层，模板值 +1
      />
      <axesHelper args={[10000]} />
    </mesh>
  );
}



function createRoadGeometry(pathPoints: THREE.Vector3[], width: number, spreadType: string, scene: THREE.Scene) {

  const base = pathPoints[0].clone()
  const shiftedPoints = pathPoints.map(p => p.clone().sub(base))
  let compensatesLeft: THREE.Vector3[] = [];
  let compensatesRight: THREE.Vector3[] = [];
  const shapes: THREE.Shape[] = [];
  for (let i = 0; i < shiftedPoints.length - 1; i++) {

    const leftSide: THREE.Vector3[] = []
    const rightSide: THREE.Vector3[] = []

    const p1 = shiftedPoints[i];
    const p2 = shiftedPoints[i + 1];
    const direction = new THREE.Vector3().subVectors(p2, p1).normalize();
    const normal = new THREE.Vector3(-direction.z, 0, direction.x).multiplyScalar(width / 2)

    let pointLeft = p2.clone().add(normal)
    let point1Left = p1.clone().add(normal)
    leftSide.push(point1Left)
    leftSide.push(pointLeft)

    let pointRight = p2.clone().sub(normal)
    let point1Right = p1.clone().sub(normal)
    rightSide.push(point1Right)
    rightSide.push(pointRight)

    const shapePoints = [...leftSide, ...rightSide.reverse()]
    const shape2D = shapePoints.map(p => new THREE.Vector2(p.x, p.z))
    const shape = new THREE.Shape(shape2D)
    shapes.push(shape)


    if (i > 0) {
      compensatesLeft.push(point1Left.clone())
      compensatesLeft.push(p1.clone())
      const shape2D = compensatesLeft.map(p => new THREE.Vector2(p.x, p.z))
      const shape = new THREE.Shape(shape2D)
      shapes.push(shape)

      compensatesRight.push(point1Right.clone())
      compensatesRight.push(p1.clone())
      const shape2D1 = compensatesRight.map(p => new THREE.Vector2(p.x, p.z))
      const shape1 = new THREE.Shape(shape2D1)
      shapes.push(shape1)
    }

    compensatesLeft = []
    compensatesRight = []
    compensatesLeft.push(pointLeft.clone())
    compensatesRight.push(pointRight.clone())

  }
  if (shapes.length > 0) {
    const geometry = new THREE.ShapeGeometry(shapes);
    geometry.rotateX(Math.PI / 2);
    return { geometry, position: base };
  }
  return { geometry: null, position: base };
}


const createMergedLane = (positions: number[], dashed: boolean = true) => {

  const lineGeometry = new LineGeometry();
  lineGeometry.setPositions(positions);

  const lineMaterial = new LineMaterial({
    color: 0xffffff,
    linewidth: 3,
    dashed: dashed,
    dashSize: 2,
    gapSize: 1,
    depthTest: true,
    transparent: true, // 确保透明度不影响判断
  });

  // 显式赋值确保生效
  lineMaterial.stencilWrite = true;
  lineMaterial.stencilRef = 1;
  lineMaterial.stencilFunc = THREE.EqualStencilFunc;
  lineMaterial.stencilFail = THREE.KeepStencilOp;
  lineMaterial.stencilZFail = THREE.KeepStencilOp;
  lineMaterial.stencilZPass = THREE.KeepStencilOp;

  lineMaterial.polygonOffset = true;
  lineMaterial.polygonOffsetFactor = -1;
  lineMaterial.polygonOffsetUnits = -1;

  const line = new Line2(lineGeometry, lineMaterial);
  line.renderOrder = 999; // 确保车道线远大于路面的 renderOrder(1)
  line.computeLineDistances();

  return line;

}

const createMergedLaneLines = (edges: any[]) => {
  const lines: Line2[] = [];

  edges.forEach(edge => {
    edge.lanes.forEach((lane: any, index: number) => {

      const points: THREE.Vector3[] = lane.shape;
      if (points.length < 2) return;
      if (edge.function && edge.function === "internal") return;

      const positions: number[] = [];
      let isdashed = true
      if (index == 0) {
        isdashed = false;
        
      } else {
      }

      if (index == edge.lanes.length - 1) {
        const numberRight: number[] = [];
        for (let i = 0; i < points.length - 1; i++) {
          const p1 = points[i]
          const p2 = points[i + 1]
          const direction = new THREE.Vector3().subVectors(p2, p1).normalize();
          const normal = new THREE.Vector3(-direction.z, 0, direction.x).multiplyScalar(lane.width / 2)
          let pointRight = p2.clone().sub(normal)
          let point1Right = p1.clone().sub(normal)
          numberRight.push(point1Right.x, point1Right.y + 0.01, point1Right.z)
          numberRight.push(pointRight.x, pointRight.y + 0.01, pointRight.z)
        } 

      }


      for (let i = 0; i < points.length - 1; i++) {
        const p1 = points[i]
        const p2 = points[i + 1]
        const direction = new THREE.Vector3().subVectors(p2, p1).normalize();
        const normal = new THREE.Vector3(-direction.z, 0, direction.x).multiplyScalar(lane.width / 2)
        let pointLeft = p2.clone().add(normal)
        let point1Left = p1.clone().add(normal)
        positions.push(point1Left.x, point1Left.y + 0.01, point1Left.z)
        positions.push(pointLeft.x, pointLeft.y + 0.01, pointLeft.z)
      }

      if (index != 0) {
        lines.push(createMergedLane(positions, isdashed));
      }
      
    });
  });

  return lines.length > 0 ? lines : null;
};

function LaneLines({ edges }: { edges: any[] }) {
  const { scene } = useThree();

  useEffect(() => {
    const lines = createMergedLaneLines(edges);
    if (lines) {
      lines.forEach(line => scene.add(line));
    }
    return () => {
      if (lines) {
        lines.forEach(line => scene.remove(line));
      }
    };
  }, [edges, scene]);

  return null;
}


function TestVehicle() {
  const meshRef = useRef<InstancedMesh>(null);
  const dummy = useRef(new Object3D()).current;

  // GLTF 模型
  const { nodes, materials } = useGLTF("/models/vehicle/PassengerCar.glb");
  const geometry = (nodes.SM_CP_Car_06 as any).geometry;
  const material = materials[Object.keys(materials)[0]];

  // 初始化 instancedMesh
  useEffect(() => {
    if (!meshRef.current) return;

    for (let i = 0; i < 200; i++) {
      dummy.position.set(1744.26 - Math.random() * 40, 0, -2855.63 + Math.random() * 40);
      dummy.updateMatrix();
      meshRef.current!.setMatrixAt(i, dummy.matrix);
      // meshRef.current.instanceMatrix.needsUpdate = true;
    }

  }, [dummy]);

  return <instancedMesh ref={meshRef} args={[geometry, material, 200]} />; // 1000 可替换成 vehiclesRef.current.length
}


const RoadNetwork = ({ netXML }: { netXML: string }) => {
  const [edges, setEdges] = useState<any[]>([]);
  const [junctions, setJunctions] = useState<any[]>([]);
  const [center, setCenter] = useState(new THREE.Vector3(0, 0, 0));
  const roadTexture = useLoader(TextureLoader, '/textures/asphalt.jpg');
  const lineTexture = useLoader(TextureLoader, '/textures/单虚线.png');
  const junctionTexture = useLoader(TextureLoader, '/textures/沥青6.jpg');

  useEffect(() => {
    const loadNet = async () => {
      const netData = await parseNetXML(netXML);
      setEdges(netData.edges);
      setJunctions(netData.junctions);
      const centerPoint = computeCenter(netData.edges);
      setCenter(centerPoint);
    };
    loadNet();
    roadTexture.wrapS = RepeatWrapping;
    roadTexture.wrapT = RepeatWrapping;
    roadTexture.repeat.set(1, 5);

    junctionTexture.wrapS = THREE.RepeatWrapping;
    junctionTexture.wrapT = THREE.RepeatWrapping;
    junctionTexture.repeat.set(10, 1);
    junctionTexture.repeat.set(10, 1);


  }, [netXML]);


  return (
    <Canvas
      camera={{ fov: 60, near: 0.1, far: 10000 }}
      style={{ width: '100%', height: '100vh' }}
      gl={{
        stencil: true,             // ✅ 必须开启
        antialias: true,
        alpha: true,
      }}
      onCreated={({ gl }) => {
        //gl.setClearColor(0xffffff, 1); // 设置背景色，方便观察
        gl.autoClearStencil = true;   // ✅ 确保每帧清除模板
      }}
    >
      <CameraController center={center} />
      <Stats />
      <ambientLight intensity={0.5} />
      <directionalLight position={[10, 20, 10]} intensity={1.5} castShadow />
      <Controls center={center} />

      {
        <RoadJunctions junctions={junctions} roadTexture={junctionTexture} />
      }
      {
        <RoadMesh edges={edges} roadTexture={roadTexture} />
      }
      {
        <LaneLines edges={edges} />
      }
      {
        <VehicleInstancedLayer />
      }
    </Canvas>
  );
};

export default RoadNetwork;
