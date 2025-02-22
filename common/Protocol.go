package common

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/gorhill/cronexpr"
)

// 定时任务
type Job struct {
	Name     string `json:"name"`     // 任务名
	Command  string `json:"command"`  // shell命令
	CronExpr string `json:"cronExpr"` // cron表达式
}

// 任务调度计划
type JobSchedulePlan struct {
	Job      *Job                 // 要调度的任务信息
	Expr     *cronexpr.Expression // 解析好的cronexpr表达式
	NextTime time.Time            // 下次调度时间
}

// 任务执行状态
type JobExecuteInfo struct {
	Job        *Job               // 任务信息
	PlanTIme   time.Time          // 理论调度时间
	RealTime   time.Time          // 实际调度时间
	CancelCtx  context.Context    // command执行的上下文
	CancelFunc context.CancelFunc // 取消command执行的函数
}

// HTTP接口应答
type Response struct {
	Errno int         `json:"errno"`
	Msg   string      `json:"msg"`
	Data  interface{} `json:"data"` // interface 万能容器
}

// 变化事件
type JobEvent struct {
	EventType int // SAVE,DELETE
	Job       *Job
}

// 任务执行结果
type JobExecuteResult struct {
	ExcuteInfo *JobExecuteInfo // 执行状态
	Output     []byte          // 脚本输出
	Err        error           // 脚本错误原因
	StartTime  time.Time       // 启动时间
	EndTime    time.Time       // 结束时间
}

// 任务执行日志
type JobLog struct {
	JobName      string `json:"jobName" bson:"jobName"`           //	任务名字
	Command      string `json:"command" bson:"command"`           // 脚本命令
	Err          string `json:"err" bson:"err"`                   // 错误原因
	Output       string `json:"output" bson:"output"`             // 脚本输出
	PlanTime     int64  `json:"planTime" bson:"planTime"`         // 计划开始时间
	ScheduleTime int64  `json:"scheduleTime" bson:"scheduleTime"` // 实际调度时间
	StartTime    int64  `json:"startTime" bson:"startTime"`       // 任务执行开始时间
	EndTime      int64  `json:"endTime" bson:"endTime"`           // 任务执行结束时间
}

// 日志批次
type LogBatch struct {
	Logs []interface{} // 多条日志
}

// 任务日志过滤条件
type JobLogFilter struct {
	JobName string `bson:"jobName"`
}

// 任务日志排序
type SortLogByStartTime struct {
	SortOrder int `bson:"startTime"` // {startTime:-1}
}

// 应答方法
func BuildResponse(errno int, msg string, data interface{}) (resp []byte, err error) {
	// 1 定义一个response对象
	var (
		response Response
	)

	response.Errno = errno
	response.Msg = msg
	response.Data = data

	// 2 序列化
	resp, err = json.Marshal(response)

	return
}

// 反序列化Job
func UnpackJob(value []byte) (ret *Job, err error) {
	var (
		job *Job
	)

	job = &Job{}
	if err = json.Unmarshal(value, job); err != nil {
		return
	}
	ret = job

	return
}

// 从etcd的key中提取任务名
// /cron/jobs/job10 提取出 job10
func ExtractJobName(jobKey string) string {
	return strings.TrimPrefix(jobKey, JOB_SAVE_DIR)
}

// /cron/killer/job10 提取出 job10
func ExtractKillerName(killerKey string) string {
	return strings.TrimPrefix(killerKey, JOB_KILLER_DIR)
}

//  任务变化事件有两种，更新/删除
func BuildJobEvent(eventType int, job *Job) (jobEvent *JobEvent) {
	return &JobEvent{EventType: eventType, Job: job}
}

// 构造任务执行计划
func BuildJobSchedulePlan(job *Job) (jobSchedulePlan *JobSchedulePlan, err error) {
	var (
		expr *cronexpr.Expression
	)

	// 解析表达式
	if job == nil {
		fmt.Println("job == nil")
	}
	if expr, err = cronexpr.Parse(job.CronExpr); err != nil {
		return
	}

	// 生成调度计划
	jobSchedulePlan = &JobSchedulePlan{
		job,
		expr,
		expr.Next(time.Now()),
	}
	return
}

func BuildJobExcuteInfo(jobSchedulePlan *JobSchedulePlan) (jobExecuteInfo *JobExecuteInfo) {
	jobExecuteInfo = &JobExecuteInfo{
		Job:      jobSchedulePlan.Job,
		PlanTIme: jobSchedulePlan.NextTime, // 计划调度时间
		RealTime: time.Now(),
	}
	jobExecuteInfo.CancelCtx, jobExecuteInfo.CancelFunc = context.WithCancel(context.TODO())
	return
}

// 提取worker的ip
func ExtractWorkerIp(regKey string) string {
	return strings.TrimPrefix(regKey, JOB_WORKER_DIR)
}
