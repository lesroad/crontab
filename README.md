# 项目介绍
这是一个结合 Etcd 与 MongoDB 实现一个基于 Master-Worker 分布式 架构的 crontab 系统。用户可以通过前端页面配置任务和 cron 表达式和命令来执行定时任务，也可以手动的编辑和删除正在执行的任务，还可以通过页面查看执行的情况，操作方便。

# 项目架构

## master
**配置管理config** 静态页面根目录，读写超时，服务端口，集群列表等
**前端** 用于展示web界面 bootstrap+ajax前后端分离
**服务接口** 给前端提供接口，比如查看，强杀，保存任务，提供日志
**任务管理JobMgr** 提供对job的管理，新建，删除，查看都是这个类实现的，具体实现方法就是依靠etcd，插入的话就是向 /cron/jobs/ 目录下存储key，然后worker节点对这个目录进行监听，监听到了有变化就拿出来，进行value的解析和执行。其他的功能都一样，强杀也是写一个key进去，不过是在 /cron/killer/ 目录下，当然有另一个killer的后台负责监听这个目录的变化，然后就解析，执行。
**服务发现WorkerMgr** WorkerMgr 实现服务的发现，就是查看能有几个还能继续工作的worker节点。实现方法也是基于etcd。简单来说，就是启动一个worker就向 /cron/workers/ 目录下注册一个自己的IP，master就通过对这个目录下所有key的读取来实现发现工作节点的。
**LogMgr日志服务** 查询mongodb中的日志。


## worker
**Register 注册服务** 目的是向etcd注册一个key，主要为了通知master我有一个节点上线，可以执行任务了。
**JobMgr监听job** 负责监听 /cron/jobs/ 目录下的变化，里面有一个loop。执行流程是这样的，每当有一个新任务从master加入到etcd，那么这个类就会把这个任务放入etcd任务事件队列
**任务调度Scheduler** 获取到任务事件后放入到jobPlanTable里面（map[jobname] jobPlan）。loop不断调用定时器，这个类的作用是从jobPlanTable里面遍历，找到时间到了的那个任务，尝试执行。
**任务执行** 执行的时候构建任务执行信息保存到任务执行表(jobExcutingTable)，先进行任务的抢锁（乐观锁），如果抢到锁，就执行，抢不到的话的等待下一轮。任务执行完成之后会通LogSink这个类存储到mongodb中去。
