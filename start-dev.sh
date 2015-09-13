#!/bin/bash

prepare_node() {
	local name=$1

	rm -rf ./data/$name
	mkdir -p ./data/$name
	mkdir -p ./data/$name/my-show/
	
	dd if=/dev/zero of=./data/$name/my-show/my-show-$RANDOM.mkv bs=1024 count=5240
}

start_node() {
	local name=$1
	local port=$2
	local httpPort=$3

	shift 2

	./mediasyncer --name=$name --bind-port=$port --http-port=$httpPort --volume=./data/$name $*
}

go build -o ./mediasyncer ./server

prepare_node node1
prepare_node node2
prepare_node node3
dd if=/dev/zero of=./data/node1/my-show/10mb bs=1024 count=10240 && touch -A -120000 ./data/node1/my-show/10mb

start_node node1 9000 8000 &
pid_node1=$!
sleep 1
start_node node2 9001 8001 localhost:9000 &
pid_node2=$!
sleep 2

start_node node3 9002 8002 localhost:9000 --price-formula=random
pid_node3=$1

kill -9 $pid_node1 $pid_node2 $pid_node3
wait $pid_node1 $pid_node2 $pid_node3