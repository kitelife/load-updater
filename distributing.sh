#!/bin/bash

chmod +x $1

while read line
do
    scp $1 root@$line:/root/
    #ssh root@$line '/root/load_updater > /dev/null 2>&1 &'
    # 自动添加crontab任务项，并重启cron
done<target_server_list.txt