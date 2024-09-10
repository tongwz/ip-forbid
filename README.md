# ip-forbid
监控nginx日志从而来封禁ip

> 功能
> 
> 将nginx的log日志进行增量读取，读取后进行特殊访问的监控，遇到恶意的访问 将ip放入到
> 
> 黑名单中去 防止访问
> 
> 还要加入过期时间可以释放出黑名单 - 暂时未加 后续考虑

是cdn白名单的用户不进入黑名单

每次是按照文件大小指针进行读取，可以减少不必要的IO

读取按照每次30秒内和 之后进入全局map 之后尽心一分钟触发频次 + 连续分钟触发次数 进行封禁IP

读取的是nginx errorLog文件进行屏蔽

### 进行封禁需要准备的东西

> 1  将config/.config.toml 修改成 config/config.toml
> 
> 2  配置里面的需要监控的error文件
> 
> 3  设置环境变量 可以在/etc/profile中或者 ~/.bash_profile 中配置:
> 
> export IP_FORBID=/xxx/xxx/ipForbid
> 
> 4  进行source ~/.bash_profile 加入环境变量
> 
> 5 运行 - nohup ./ipForbid &>>/xxxx/xxxx/logs/fmt.log &
> 
> 6 查看有没有日志生成 一般继续生成就是没有问题
> 

### 建议

  使用root账户，因为这里会执行命令行 nginx -t 和 nginx -s reload 

  我的程序也会主动修改config.toml配置文件因此在程序运行过程中不要进行修改，因为它会忽略，并重写
