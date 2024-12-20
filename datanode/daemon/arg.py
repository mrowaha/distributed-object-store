from typing import TypedDict
from enum import Enum
class Mode(Enum):
    GHOST = "ghost"
    DATA = "data"

# Custom type function to validate enum
def validate_mode(value):
    return Mode[value.upper()]  # Convert string to enum member
        

class Args(TypedDict):
    store: str
    port: int
    process: str
    lease: str
    mode: Mode