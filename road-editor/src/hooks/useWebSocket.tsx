import { useState, useEffect, useRef } from 'react';

interface Message {
  type: string;
  payload: any;
}

export function useWebSocket(url: string) {
  const [messages, setMessages] = useState<string[]>([]);
  const [isConnected, setIsConnected] = useState<boolean>(false);
  const ws = useRef<WebSocket | null>(null);

  // 当组件挂载时创建 WebSocket 连接
  useEffect(() => {
    ws.current = new WebSocket(url);

    // 监听 WebSocket 连接打开
    ws.current.onopen = () => {
      setIsConnected(true);
      console.log('WebSocket connected');
    };

    // 监听 WebSocket 接收到的消息
    ws.current.onmessage = (event: MessageEvent) => {
      console.log('Received:', event.data);
      setMessages((prev) => [...prev, event.data]);
    };

    // 监听 WebSocket 连接关闭
    ws.current.onclose = () => {
      setIsConnected(false);
      console.log('WebSocket disconnected');
    };

    // 监听错误
    ws.current.onerror = (error) => {
      console.error('WebSocket error:', error);
    };

    // 组件卸载时关闭 WebSocket 连接
    return () => {
      ws.current?.close();
    };
  }, [url]);

  // 发送消息的函数
  const sendMessage = (message: Message) => {
    if (ws.current && ws.current.readyState === WebSocket.OPEN) {
      ws.current.send(JSON.stringify(message));
    } else {
      console.log('WebSocket is not connected');
    }
  };

  return {
    messages,
    isConnected,
    sendMessage,
  };
}
