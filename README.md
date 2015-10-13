**注意：只支持Linux！**

- `fetch_cpu_num.sh` 批量获取服务器的CPU核数，需要服务器ip列表文件server_list.txt（一行一个ip），输出server_with_cpunum.txt
- `distributing.sh` 批量分发load-updater程序，运行方式：`./distributing.sh load-updater`
- `task_run.sh` 启动各服务器上的load-updater程序，依赖于文件server_with_cpunum.txt，根据CPU核数设置不同的负载，默认运行10分钟
- `check_run.sh` 检测各服务器上的load-updater程序是否已启动。
