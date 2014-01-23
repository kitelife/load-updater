#!/bin/bash

while read line
do
    ip=$(echo -n $line | cut -d ' ' -f1)
    cpuNum=$(echo -n $line | cut -d ' ' -f2)
    pauseInterval=500000
    if [ 4 -lt $cpuNum ]; then
        pauseInterval=1000000
    fi
    if [ 8 -lt $cpuNum ]; then
        pauseInterval=2000000
    fi
    echo $ip
    echo $pauseInterval
    command="/root/load_updater -pause_interval=$pauseInterval -run_duration=10 > /dev/null 2>&1 &"
    ssh -q -n root@$ip $command
done<server_with_cpunum.txt