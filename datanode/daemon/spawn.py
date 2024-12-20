"""
this file contains definitions to expose a spawn port.
if daemon is started as ghost mode,
then whenever data is received on this port by the namenode,
if it is a marker message it saves the store as the file name
then it boots a datanode

if it is in data mode and the datanode fails,
it will automatically convert the node into a ghost mode
and connect to the namenode.
"""

from typing import Protocol
import proto.ghost_pb2 as ghost
import proto.ghost_pb2_grpc as ghostservice
import grpc
import sqlite3

class ISpawner(Protocol):
    def run(): None

class Spawner:
    def __init__(self, name: str, port: int, db: str):
        self.namenode = port
        self.name = name
        self.db = db

    def run(self: "Spawner") -> int:
        # Create a channel to the server
        print("running spawner")
        with grpc.insecure_channel(f'localhost:{self.namenode}') as channel:
            stub = ghostservice.GhostServiceStub(channel)
            request = ghost.SpawnWait(name=self.name)

            try:
                response_stream = stub.Spawn(request)
                for spawn_command in response_stream:
                    print(f"received spawncommand: status={spawn_command.status}, commands={spawn_command.commands}")
                    self.save(spawn_command.commands)
                    return spawn_command.lamport
            except grpc.RpcError as e:
                print(f"gRPC error occurred: {e}")

    def save(self: "Spawner", data):
        try:
            with sqlite3.connect(self.db) as conn:
                cursor = conn.cursor()
                cursor.execute("""DROP TABLE IF EXISTS datanode;""")
                cursor.execute("""
                    CREATE TABLE IF NOT EXISTS datanode (
                        object TEXT PRIMARY KEY,
                        data BLOB NOT NULL,
                        sequence INTEGER DEFAULT 1
                    );
                """)

                for objectInfo in data:
                    if len(objectInfo.objectName) == 0:
                        continue
                    cursor.execute("""
                        INSERT INTO datanode (object, data)
                        VALUES (?, ?)
                    """, (objectInfo.objectName, objectInfo.objectData))

                print('commiting')
                conn.commit()
        except sqlite3.Error as e:
            print(f"SQLite error occurred: {e}")