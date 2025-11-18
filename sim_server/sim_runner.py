import traci
import subprocess, asyncio

async def run_simulation(config, duration, speed, queue,prot):    
    sumo_cmd = ["sumo", "-c", config, "--remote-port", str(prot)]
    proc = subprocess.Popen(sumo_cmd)
    await asyncio.sleep(1)  # 等待 SUMO 启动
    traci.init(prot)
    steps = int(duration / speed) 
    last_veh_ids = set() 
    for t in range(steps):
        traci.simulationStep()
        veh_ids = traci.vehicle.getIDList() 
        add_vehicles = []
        remove_vehicles = []
        vehicles = []    
         # 比较新增车辆和移除车辆
       # 计算新增和移除车辆
        added_vehicles = set(veh_ids) - last_veh_ids  # 新增的车辆
        removed_vehicles = last_veh_ids - set(veh_ids)  # 移除的车辆

        for vid in veh_ids:
            pos = traci.vehicle.getPosition3D(vid)
            vehicles.append({ 
                "id": vid,
                "x": round(pos[0], 2),
                "y": round(pos[1], 2),
                "z": round(pos[2], 2),
                "a": round(traci.vehicle.getAngle(vid), 2),
                "s": round(traci.vehicle.getSlope(vid), 2),
            }) 
        for vid in added_vehicles:
            type = traci.vehicle.getTypeID(vid) 
            add_vehicles.append({ 
                "id": vid, 
                "type":type
            })
        for vid in removed_vehicles: 
            remove_vehicles.append({ 
                "id": vid, 
            })    
        await queue.put({ 
                "adds": add_vehicles,
                "removes": remove_vehicles,
                "vehicles": vehicles, 
            })    
        await asyncio.sleep(1 / speed)

    traci.close()
    proc.terminate()
