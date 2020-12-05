#!/bin/bash

ROOMS=1
CLIENTS=20

for i in $(seq 1 $ROOMS)
do
	for j in $(seq 1 $CLIENTS)
	do
		node test.js " room-$i client-$i-$j " & 
	done 
done

