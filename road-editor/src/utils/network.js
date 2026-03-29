import xml2js from 'xml2js';




// 解析 SUMO 路网配置的函数
export async  function parseNetXML(xmlString){
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