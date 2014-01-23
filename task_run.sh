#!/bin/bash

while read line
do
    ip=$(echo -n $line | cut -d ' ' -f1)
    cpuNum=$(echo -n $line | cut -d ' ' -f2)
    loadLevel=1
    if [ 4 -lt $cpuNum ]; then
        loadLevel=2
    fi
    if [ 8 -lt $cpuNum ]; then
        loadLevel=3
    fi
    echo $ip
    echo $loadLevel
    command="/root/load_updater -load_level=$loadLevel -run_duration=10 > /dev/null 2>&1 &"
    ssh -q -n root@$ip $command
done<server_with_cpunum.txt