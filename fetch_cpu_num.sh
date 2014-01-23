#!/bin/bash

while read line
do
    cpu_num=$(ssh -q -n root@$line 'grep cpu /proc/stat | wc -l')
    let cpu_num=cpu_num-1
    echo $line" "$cpu_num >> server_with_cpunum.txt
done < server_list.txt