import asyncio
import json
import websockets
import msgpack

clients = set()

async def handle_ws(websocket):
    clients.add(websocket)
    try:
        await websocket.wait_closed()
    finally:
        clients.remove(websocket)

async def start_ws_server(port):
    # 正确使用 async with 启动 WebSocket 服务
    try:
        await websockets.serve(handle_ws, "0.0.0.0", port)
        print("WebSocket server started on port :{port}".format(port=port))
        #await asyncio.Future()  # 让它一直运行
    except OSError as e:
        print(f"❌ Failed to start WebSocket server: {e}")   

async def push_to_clients(queue):
    while True:
        data = await queue.get()
        dead = set()
        for ws in clients:
            try:                
                packed_data = msgpack.packb(data)
                await ws.send(packed_data)
            except:
                dead.add(ws)
        clients.difference_update(dead)

async def main():
    queue = asyncio.Queue()

    # 启动客户端推送协程
    asyncio.create_task(push_to_clients(queue))

    # 启动 WebSocket 服务
    await start_ws_server()

if __name__ == '__main__':
    asyncio.run(main())
