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

class ISpawner(Protocol):
    def run(): None

class Spawner:
    def __init__(self):
        pass

    def run():
        # Create a channel to the server
        with grpc.insecure_channel('localhost:50051') as channel:
            stub = ghostservice.GhostServiceStub(channel)
            response = stub.Hello(ghost.HelloMsg(name="rowaha"))
            print("Greeting:", response.message)
