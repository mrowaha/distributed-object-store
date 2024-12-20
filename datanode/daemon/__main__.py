from pathlib import Path
import logging
import subprocess
import os
import argparse
from arg import Args, validate_mode, Mode
import spawn
import time
import threading
# Get the current working directory
daemon_dir = Path.cwd()

def main(*, args: Args):
    lamport = 0
    if args['mode'] == Mode.GHOST:
        spawner : spawn.ISpawner = spawn.Spawner(args['process'], args['port'], args['store'])
        lamport = spawner.run()
        print(f"lamport updated to {lamport}")

    while True:
        env = os.environ.copy()
        try:
            exec = ["go", "run", ".", f"-store={args['store']}", f"-process={args['process']}", f"-lease={args['lease']}", f"-lamport={lamport}"]
            print(f'execution command: {" ".join(exec)}')
            process = subprocess.Popen(
                exec,
                stdout=subprocess.PIPE,  # Capture standard output
                stderr=subprocess.PIPE,   # Capture standard error
                env=env,
                text=True
            )

            while True:
                if process.poll() is not None:
                    break
                error = process.stderr.readline()
                print(f"[daemon] {error}",end="")

            spawner : spawn.ISpawner = spawn.Spawner(args['process'], args['port'], args['store'])
            spawner.run()
        except Exception as e:
            print(f"An error occurred in daemon: {e}")

if __name__ == "__main__":
    ap = argparse.ArgumentParser()
    ap.add_argument("-s", "--store", type=str, help="store .db file name", required=True)
    ap.add_argument("-p", "--port", type=int, help="port of the name node service", default=50051)
    ap.add_argument("-proc", "--process", type=str, help="name of the process", required=True)
    ap.add_argument("-l", "--lease", type=str, help="lease address", required=True)
    ap.add_argument("-m", "--mode", type=validate_mode, help="ghost or data mode", default=Mode.GHOST)
    args : Args =  Args(**vars(ap.parse_args()))
    print(args)
    main(args=args)