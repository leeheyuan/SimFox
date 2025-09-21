import argparse
import asyncio
from sim_runner import run_simulation
from db_writer import write_to_db
from ws_push import push_to_clients, start_ws_server



def parse_args():
    parser = argparse.ArgumentParser(description="SUMO Simulation Server")
    parser.add_argument("--config", required=True, help="SUMO .sumocfg file")
    parser.add_argument("--duration", type=int, default=60, help="Simulation seconds")
    parser.add_argument("--speed", type=float, default=1.0, help="Simulation speed multiplier")
    parser.add_argument("--ports", required=int, help="Ports")
    parser.add_argument("--projectId", required=int, help="ProjectId")
    return parser.parse_args()

async def main():
    args = parse_args()
    print(f"simul: config={args.config}, duration={args.duration}, speed={args.speed}")
    ports = [int(p.strip()) for p in args.ports.split(",") if p.strip()]
    print(f"Received ports: {ports}")
    data_queue = asyncio.Queue()
    await start_ws_server(ports[0])  # 启动 WebSocket 服务
    await asyncio.gather(
        run_simulation(args.config, args.duration, args.speed, data_queue,ports[1]),
        #write_to_db(data_queue),
        push_to_clients(data_queue)
    )

if __name__ == "__main__":
    try:
      asyncio.run(main())
    except KeyboardInterrupt:
      print("服务器被中断，正在退出...")