import asyncio
import sqlite3

conn = sqlite3.connect("sim.db")
cur = conn.cursor()
cur.execute("CREATE TABLE IF NOT EXISTS vehicle (time INT, id TEXT, x REAL, y REAL)")

async def write_to_db(queue):
    while True:
        data = await queue.get()
        cur.execute("INSERT INTO vehicle VALUES (?, ?, ?, ?)", 
                    (data["time"], data["id"], data["x"], data["y"]))
        conn.commit()
