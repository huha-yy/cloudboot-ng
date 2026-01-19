package models

import (
	"time"
)

// Job 表示一个异步任务（如RAID配置、OS安装）
type Job struct {
	ID          string    `gorm:"primaryKey" json:"id"`
	MachineID   string    `gorm:"index;column:machine_id" json:"machine_id"`
	Type        JobType   `gorm:"type:varchar(50)" json:"type"`
	Status      JobStatus `gorm:"type:varchar(20);index" json:"status"`
	ProfileID   string    `gorm:"type:varchar(36);index" json:"profile_id"` // OS Profile ID (for install_os jobs)
	StepCurrent string    `gorm:"type:varchar(100)" json:"step_current"`
	LogsPath    string    `gorm:"type:varchar(255)" json:"logs_path"`
	Error       string    `gorm:"type:text" json:"error,omitempty"`
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime" json:"updated_at"`

	// 关联
	Machine *Machine   `gorm:"foreignKey:MachineID" json:"machine,omitempty"`
	Profile *OSProfile `gorm:"foreignKey:ProfileID" json:"profile,omitempty"`
}

// JobType 任务类型枚举
type JobType string

const (
	// JobTypeAudit 硬件审计（探测）
	JobTypeAudit JobType = "audit"
	// JobTypeConfigRAID RAID配置
	JobTypeConfigRAID JobType = "config_raid"
	// JobTypeInstallOS 操作系统安装
	JobTypeInstallOS JobType = "install_os"
)

// JobStatus 任务状态枚举
type JobStatus string

const (
	// JobStatusPending 待执行
	JobStatusPending JobStatus = "pending"
	// JobStatusRunning 执行中
	JobStatusRunning JobStatus = "running"
	// JobStatusSuccess 成功
	JobStatusSuccess JobStatus = "success"
	// JobStatusFailed 失败
	JobStatusFailed JobStatus = "failed"
)

// TableName 指定表名
func (Job) TableName() string {
	return "jobs"
}

// IsTerminal 检查任务是否已终止（成功或失败）
func (j *Job) IsTerminal() bool {
	return j.Status == JobStatusSuccess || j.Status == JobStatusFailed
}

// IsPending 检查任务是否待执行
func (j *Job) IsPending() bool {
	return j.Status == JobStatusPending
}

// IsRunning 检查任务是否正在执行
func (j *Job) IsRunning() bool {
	return j.Status == JobStatusRunning
}

// SetError 设置错误信息并标记为失败
func (j *Job) SetError(err error) {
	j.Error = err.Error()
	j.Status = JobStatusFailed
}

// SetSuccess 标记任务成功
func (j *Job) SetSuccess() {
	j.Status = JobStatusSuccess
	j.Error = ""
}

// UpdateStep 更新当前执行步骤
func (j *Job) UpdateStep(step string) {
	j.StepCurrent = step
}
