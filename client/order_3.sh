#!/bin/bash

: '
  in this script, we are going to attempt to create two objects
  then we are going to attempt to update both of those objects to different values 
 '


# create first object
go run . -object=me -data=rowaha

# create second object
go run . -object=hello -data=world

echo [object_3] checkpoint: attempted creating objects


# attempt update first object
go run . -cmd=3 -object=me -data=not_rowaha

# attempt update second object
go run . -cmd=3 -object=hello -data=not_world
