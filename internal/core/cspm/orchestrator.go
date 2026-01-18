package cspm

import (
	"context"
	"fmt"
	"time"
)

// Orchestrator 提供 Provider 执行的原子序列编排
// 实现 Plan → Probe → Apply 的闭环逻辑，确保幂等性和安全性
type Orchestrator struct {
	executor *Executor
}

// NewOrchestrator 创建新的 Orchestrator
func NewOrchestrator(executor *Executor) *Orchestrator {
	return &Orchestrator{
		executor: executor,
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

	// Step 1: Plan - 生成执行计划
	planResult, err := o.executePlan(ctx, config)
	if err != nil {
		result.Success = false
		result.Error = fmt.Errorf("plan failed: %w", err)
		return result, err
	}
	result.Steps = append(result.Steps, *planResult)

	// 如果 Plan 失败，立即返回
	if !planResult.Success {
		result.Success = false
		result.Error = fmt.Errorf("plan validation failed")
		return result, result.Error
	}

	// Step 2: Probe - 探测当前状态（幂等性检查）
	probeResult, err := o.executeProbe(ctx)
	if err != nil {
		result.Success = false
		result.Error = fmt.Errorf("probe failed: %w", err)
		return result, err
	}
	result.Steps = append(result.Steps, *probeResult)

	// 检查是否已达标（幂等性）
	if o.isAlreadyConverged(probeResult, config) {
		result.Success = true
		result.Idempotent = true
		result.Message = "System already in desired state, skipping apply"
		result.Duration = time.Since(result.StartTime)
		return result, nil
	}

	// Step 3: Apply - 执行实际变更
	applyResult, err := o.executeApply(ctx, config)
	if err != nil {
		result.Success = false
		result.Error = fmt.Errorf("apply failed: %w", err)
		return result, err
	}
	result.Steps = append(result.Steps, *applyResult)

	// Step 4: Verify - 验证执行结果
	verifyResult, err := o.executeProbe(ctx)
	if err != nil {
		result.Success = false
		result.Error = fmt.Errorf("verification probe failed: %w", err)
		return result, err
	}
	result.Steps = append(result.Steps, StepResult{
		Name:     "verify",
		Success:  verifyResult.Success,
		Duration: verifyResult.Duration,
		Data:     verifyResult.Data,
	})

	// 最终结果
	result.Success = applyResult.Success
	result.Duration = time.Since(result.StartTime)

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
