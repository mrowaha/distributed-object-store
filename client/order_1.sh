#!/bin/bash


# first we are going to create two objects
go run . -cmd=1 -object=me -data=rowaha

# ... next object
go run . -cmd=1 -object=hello -data=world


echo completed order_1
