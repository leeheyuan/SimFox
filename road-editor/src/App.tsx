import React from 'react';
import RoadEditorWithImport from './components/RoadEditorWithImport';
import { useWebSocket } from './hooks/useWebSocket';
import { decodeMsgpackBlob } from './utils/msgpack'; 

import { BrowserRouter, Routes, Route, Link } from 'react-router-dom'; 

const loadXMLFile = async (filePath: string) => {
  const response = await fetch(filePath);
  const xmlText = await response.text();
  return xmlText;
};

import RoadNetwork from './components/RoadNetwork'; 
import { useVehicleStore } from "@/stores/vehicleStore";

export default function App() {
  const { messages, isConnected, sendMessage } = useWebSocket('ws://localhost:8010');
  const [netXML, setNetXML] = React.useState<string>(''); 
  const bulkUpdate = useVehicleStore((s) => s.bulkUpdate)
  messages.forEach(async msg => { 
    let data = await  decodeMsgpackBlob(msg)
    //console.log('WebSocket Message:', data) 
    bulkUpdate(data)
  });

/*  if (!isConnected) {
    return <div>Connecting to WebSocket...</div>;
  } */

  React.useEffect(() => {
    loadXMLFile('osm.net.xml').then(setNetXML);
  }, []);

  React.useEffect(() => {
    const url = new URL(window.location.href);
    const token = url.searchParams.get('token');
    if (!token) {
      return;
    }

    localStorage.setItem('token', token);
    url.searchParams.delete('token');
    window.history.replaceState({}, document.title, url.toString());
  }, []);

 // App.js
return (
  <BrowserRouter>
    <Routes>
      {/* 首页：包含导航栏 */}
      <Route path="/" element={
        <>
          <nav><Link to="/editor">进入道路编辑器</Link></nav>
          <div>这里是首页内容</div>

          <nav><Link to="/network">进入道路可视化</Link></nav>
          <div>这里是道路可视化</div>
        </>
      } />

      {/* 编辑器页面：独立渲染，不含 nav */}
      <Route path="/editor" element={<RoadEditorWithImport />} />
      <Route path="/network " element={<RoadNetwork netXML={netXML}/>} />
    </Routes>
  </BrowserRouter>
);

  /*return (
    <div>
      <h1>SUMO 路网可视化</h1>
      {netXML ? <RoadNetwork netXML={netXML} /> : <p>加载中...</p>}
    </div>
  );*/
}
