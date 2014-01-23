#!/bin/bash

chmod +x $1

while read line
do
    echo $line
    scp $1 root@$line:/root/
done<server_list.txt