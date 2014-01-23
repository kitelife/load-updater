#!/bin/bash

chmod +x $1

while read line
do
    ip=$(cut $line -f1)
    cpuNum=$(cut $line -f2)
    loadLevel=1
    if [ cpuNum -gt 4 ]; then
        loadLevel=2
    fi
    if [ cpuNum -gt 8 ]; then
        loadLevel=3
    fi
    ssh -q -n root@$line '/root/load_updater -load_level=$loadLevel -run_duration=10 > /dev/null 2>&1 &'
done<server_with_cpunum.txt