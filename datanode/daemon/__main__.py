from pathlib import Path
import logging
import subprocess
import os
import argparse
from arg import Args, validate_mode, Mode
import spawn

# Get the current working directory
daemon_dir = Path.cwd()

def main(*, args: Args):
    print(f"daemon was started with the following args: {args}")
    env = os.environ.copy()
    exitCode = 1
    try:
        process = subprocess.Popen(
            ["go", "run", ".", f"-store={args['store']}", f"-process={args['process']}", f"-lease={args['lease']}" ],
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
        exitCode = process.returncode

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
    # main(args=args)