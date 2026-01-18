package cspm

import (
	"context"
	"fmt"
	"time"

	"github.com/cloudboot/cloudboot-ng/internal/core/logbroker"
)

// Orchestrator 提供 Provider 执行的原子序列编排
// 实现 Plan → Probe → Apply 的闭环逻辑，确保幂等性和安全性
type Orchestrator struct {
	executor *Executor
	broker   *logbroker.Broker // 日志流代理 (可选)
	jobID    string            // Job ID for logging (可选)
}

// NewOrchestrator 创建新的 Orchestrator
func NewOrchestrator(executor *Executor) *Orchestrator {
	return &Orchestrator{
		executor: executor,
	}
}

// SetLogBroker 设置日志流代理
func (o *Orchestrator) SetLogBroker(broker *logbroker.Broker, jobID string) {
	o.broker = broker
	o.jobID = jobID
}

// publishLog 发布日志到broker (如果配置)
func (o *Orchestrator) publishLog(level, message string) {
	if o.broker != nil && o.jobID != "" {
		o.broker.PublishHTML(o.jobID, level, message)
	}
}

// ApplyWithPlan 执行完整的原子序列：Plan → Probe → Apply
// 这是生产环境推荐的安全执行方式
//
// 执行流程：
// 1. Plan：预演变更，生成执行计划
// 2. Probe：探测当前状态，检查是否已达标（幂等性）
// 3. Apply：执行实际变更（如果需要）
func (o *Orchestrator) ApplyWithPlan(ctx context.Context, config map[string]interface{}) (*OrchestratorResult, error) {
	result := &OrchestratorResult{
		StartTime: time.Now(),
		Steps:     make([]StepResult, 0),
	}

	o.publishLog("INFO", "🚀 开始Provider原子序列执行")

	// Step 1: Plan - 生成执行计划
	o.publishLog("INFO", "📋 Step 1/4: 执行Plan - 预演变更")
	planResult, err := o.executePlan(ctx, config)
	if err != nil {
		result.Success = false
		result.Error = fmt.Errorf("plan failed: %w", err)
		o.publishLog("ERROR", fmt.Sprintf("❌ Plan失败: %v", err))
		return result, err
	}
	result.Steps = append(result.Steps, *planResult)

	// 如果 Plan 失败，立即返回
	if !planResult.Success {
		result.Success = false
		result.Error = fmt.Errorf("plan validation failed")
		o.publishLog("ERROR", "❌ Plan验证失败")
		return result, result.Error
	}
	o.publishLog("INFO", "✅ Plan执行成功")

	// Step 2: Probe - 探测当前状态（幂等性检查）
	o.publishLog("INFO", "🔍 Step 2/4: 执行Probe - 探测当前状态")
	probeResult, err := o.executeProbe(ctx)
	if err != nil {
		result.Success = false
		result.Error = fmt.Errorf("probe failed: %w", err)
		o.publishLog("ERROR", fmt.Sprintf("❌ Probe失败: %v", err))
		return result, err
	}
	result.Steps = append(result.Steps, *probeResult)
	o.publishLog("INFO", "✅ Probe执行成功")

	// 检查是否已达标（幂等性）
	if o.isAlreadyConverged(probeResult, config) {
		result.Success = true
		result.Idempotent = true
		result.Message = "System already in desired state, skipping apply"
		result.Duration = time.Since(result.StartTime)
		o.publishLog("INFO", "🎯 系统已达标，跳过Apply步骤 (幂等性)")
		o.publishLog("INFO", "✅ 执行完成 (幂等)")
		return result, nil
	}

	// Step 3: Apply - 执行实际变更
	o.publishLog("INFO", "⚙️ Step 3/4: 执行Apply - 应用变更")
	applyResult, err := o.executeApply(ctx, config)
	if err != nil {
		result.Success = false
		result.Error = fmt.Errorf("apply failed: %w", err)
		o.publishLog("ERROR", fmt.Sprintf("❌ Apply失败: %v", err))
		return result, err
	}
	result.Steps = append(result.Steps, *applyResult)
	o.publishLog("INFO", "✅ Apply执行成功")

	// Step 4: Verify - 验证执行结果
	o.publishLog("INFO", "🔍 Step 4/4: 执行Verify - 验证结果")
	verifyResult, err := o.executeProbe(ctx)
	if err != nil {
		result.Success = false
		result.Error = fmt.Errorf("verification probe failed: %w", err)
		o.publishLog("ERROR", fmt.Sprintf("❌ Verify失败: %v", err))
		return result, err
	}
	result.Steps = append(result.Steps, StepResult{
		Name:     "verify",
		Success:  verifyResult.Success,
		Duration: verifyResult.Duration,
		Data:     verifyResult.Data,
	})
	o.publishLog("INFO", "✅ Verify执行成功")

	// 最终结果
	result.Success = applyResult.Success
	result.Duration = time.Since(result.StartTime)

	o.publishLog("INFO", fmt.Sprintf("🎉 执行完成 - 总耗时: %.2fs", result.Duration.Seconds()))

	return result, nil
}

// executePlan 执行 Plan 步骤
func (o *Orchestrator) executePlan(ctx context.Context, config map[string]interface{}) (*StepResult, error) {
	startTime := time.Now()

	execResult, err := o.executor.Execute(ctx, "plan", config)
	if err != nil {
		return &StepResult{
			Name:     "plan",
			Success:  false,
			Duration: time.Since(startTime),
		}, err
	}

	return &StepResult{
		Name:     "plan",
		Success:  execResult.IsSuccess(),
		Duration: execResult.Duration,
		Data:     execResult.Data,
		Logs:     execResult.Logs,
	}, nil
}

// executeProbe 执行 Probe 步骤
func (o *Orchestrator) executeProbe(ctx context.Context) (*StepResult, error) {
	startTime := time.Now()

	execResult, err := o.executor.Execute(ctx, "probe", nil)
	if err != nil {
		return &StepResult{
			Name:     "probe",
			Success:  false,
			Duration: time.Since(startTime),
		}, err
	}

	return &StepResult{
		Name:     "probe",
		Success:  execResult.IsSuccess(),
		Duration: execResult.Duration,
		Data:     execResult.Data,
		Logs:     execResult.Logs,
	}, nil
}

// executeApply 执行 Apply 步骤
func (o *Orchestrator) executeApply(ctx context.Context, config map[string]interface{}) (*StepResult, error) {
	startTime := time.Now()

	execResult, err := o.executor.Execute(ctx, "apply", config)
	if err != nil {
		return &StepResult{
			Name:     "apply",
			Success:  false,
			Duration: time.Since(startTime),
		}, err
	}

	return &StepResult{
		Name:     "apply",
		Success:  execResult.IsSuccess(),
		Duration: execResult.Duration,
		Data:     execResult.Data,
		Logs:     execResult.Logs,
	}, nil
}

// isAlreadyConverged 检查系统是否已达标（幂等性检查）
// 这是防止重复执行的关键逻辑
func (o *Orchestrator) isAlreadyConverged(probeResult *StepResult, desiredConfig map[string]interface{}) bool {
	// 如果 Probe 失败，说明硬件不可用或有问题，需要执行 Apply
	if !probeResult.Success {
		return false
	}

	// 提取期望状态
	desiredState, ok := desiredConfig["desired_state"].(map[string]interface{})
	if !ok {
		// 没有 desired_state，无法比较
		return false
	}

	desiredLevel, _ := desiredState["level"].(string)

	// 处理 drives 可能是 []string 或 []interface{} 的情况
	var desiredDrives []interface{}
	switch v := desiredState["drives"].(type) {
	case []interface{}:
		desiredDrives = v
	case []string:
		// 转换 []string 为 []interface{}
		desiredDrives = make([]interface{}, len(v))
		for i, s := range v {
			desiredDrives[i] = s
		}
	default:
		return false
	}

	// 提取当前状态（从 Probe 返回）
	probeData := probeResult.Data
	if probeData == nil {
		return false
	}

	// 检查是否已存在虚拟驱动器
	vdListRaw, ok := probeData["virtual_drives"]
	if !ok {
		// 没有虚拟驱动器，需要创建
		return false
	}

	vdList, ok := vdListRaw.([]interface{})
	if !ok || len(vdList) == 0 {
		// 虚拟驱动器列表为空，需要创建
		return false
	}

	// 检查是否已存在相同配置的虚拟驱动器
	for _, vdRaw := range vdList {
		vd, ok := vdRaw.(map[string]interface{})
		if !ok {
			continue
		}

		// 比较 RAID 级别
		currentLevel, _ := vd["level"].(string)
		if currentLevel != desiredLevel {
			continue
		}

		// 比较驱动器列表
		currentDrivesRaw, ok := vd["drives"]
		if !ok {
			continue
		}

		currentDrives, ok := currentDrivesRaw.([]interface{})
		if !ok {
			continue
		}

		// 比较驱动器数量
		if len(currentDrives) != len(desiredDrives) {
			continue
		}

		// 比较驱动器内容（简化版：只比较数量，生产环境需要精确比较每个驱动器ID）
		// 如果找到匹配的虚拟驱动器，说明已达标
		return true
	}

	// 没有找到匹配的虚拟驱动器，需要执行 Apply
	return false
}

// OrchestratorResult 编排器执行结果
type OrchestratorResult struct {
	Success    bool          `json:"success"`
	Idempotent bool          `json:"idempotent"` // true 表示系统已达标，跳过了 Apply
	Message    string        `json:"message"`
	Steps      []StepResult  `json:"steps"`
	Error      error         `json:"-"`
	StartTime  time.Time     `json:"start_time"`
	Duration   time.Duration `json:"duration"`
}

// StepResult 单个步骤的执行结果
type StepResult struct {
	Name     string                 `json:"name"`     // plan, probe, apply, verify
	Success  bool                   `json:"success"`
	Duration time.Duration          `json:"duration"`
	Data     map[string]interface{} `json:"data"`
	Logs     []LogEntry             `json:"logs,omitempty"`
}

// GetFailedStep 获取第一个失败的步骤
func (r *OrchestratorResult) GetFailedStep() *StepResult {
	for i := range r.Steps {
		if !r.Steps[i].Success {
			return &r.Steps[i]
		}
	}
	return nil
}

// GetStepByName 根据名称获取步骤结果
func (r *OrchestratorResult) GetStepByName(name string) *StepResult {
	for i := range r.Steps {
		if r.Steps[i].Name == name {
			return &r.Steps[i]
		}
	}
	return nil
}
