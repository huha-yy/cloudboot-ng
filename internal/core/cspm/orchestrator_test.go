package cspm

import (
	"context"
	"testing"
	"time"
)

// TestOrchestratorApplyWithPlan 测试完整的原子序列执行
func TestOrchestratorApplyWithPlan(t *testing.T) {
	// 使用 Mock Provider
	executor := NewExecutor("../../../cmd/provider-mock/provider-mock")
	executor.SetTimeout(10 * time.Second)

	orchestrator := NewOrchestrator(executor)

	config := map[string]interface{}{
		"desired_state": map[string]interface{}{
			"level":  "10",
			"drives": []string{"0:0", "0:1", "0:2", "0:3"},
		},
	}

	ctx := context.Background()
	result, err := orchestrator.ApplyWithPlan(ctx, config)

	if err != nil {
		t.Fatalf("ApplyWithPlan failed: %v", err)
	}

	// 验证执行步骤
	expectedSteps := []string{"plan", "probe", "apply", "verify"}
	if len(result.Steps) != len(expectedSteps) {
		t.Errorf("Expected %d steps, got %d", len(expectedSteps), len(result.Steps))
	}

	for i, expectedName := range expectedSteps {
		if i >= len(result.Steps) {
			break
		}
		if result.Steps[i].Name != expectedName {
			t.Errorf("Step %d: expected name %s, got %s", i, expectedName, result.Steps[i].Name)
		}
	}

	// 验证总体成功
	if !result.Success {
		t.Errorf("Expected success, but got failure")
		if result.Error != nil {
			t.Logf("Error: %v", result.Error)
		}
		if failedStep := result.GetFailedStep(); failedStep != nil {
			t.Logf("Failed step: %s", failedStep.Name)
		}
	}

	// 验证执行时间
	if result.Duration == 0 {
		t.Error("Duration should not be zero")
	}

	t.Logf("Total duration: %v", result.Duration)
	t.Logf("Success: %v, Idempotent: %v", result.Success, result.Idempotent)
}

// TestOrchestratorIdempotency 测试幂等性（暂时跳过，因为需要Mock Provider支持状态检查）
func TestOrchestratorIdempotency(t *testing.T) {
	t.Skip("Idempotency check requires Mock Provider to support state comparison")

	// TODO: 实现完整的幂等性测试
	// 1. 第一次执行 ApplyWithPlan，创建 RAID
	// 2. 第二次执行 ApplyWithPlan，期望检测到已达标，跳过 Apply
	// 3. 验证 result.Idempotent == true
}

// TestOrchestratorPlanFailure 测试 Plan 失败时的行为
func TestOrchestratorPlanFailure(t *testing.T) {
	t.Skip("Requires Mock Provider to support plan failure simulation")

	// TODO: 测试当 Plan 返回 failed 时，流程应该立即停止
	// 不应该继续执行 Probe 和 Apply
}

// TestOrchestratorProbeFailure 测试 Probe 失败时的行为
func TestOrchestratorProbeFailure(t *testing.T) {
	t.Skip("Requires Mock Provider to support probe failure simulation")

	// TODO: 测试当 Probe 返回 failed 时，流程的处理逻辑
}

// TestOrchestratorTimeout 测试超时熔断
func TestOrchestratorTimeout(t *testing.T) {
	t.Skip("Requires Mock Provider to simulate long-running operations")

	// TODO: 测试当 Provider 执行超时时，context 能够正确取消
	// 验证超时错误能够正确返回
}

// TestOrchestratorGetStepByName 测试按名称获取步骤结果
func TestOrchestratorGetStepByName(t *testing.T) {
	result := &OrchestratorResult{
		Steps: []StepResult{
			{Name: "plan", Success: true},
			{Name: "probe", Success: true},
			{Name: "apply", Success: false},
		},
	}

	// 测试存在的步骤
	planStep := result.GetStepByName("plan")
	if planStep == nil {
		t.Error("Expected to find 'plan' step")
	} else if planStep.Name != "plan" {
		t.Errorf("Expected step name 'plan', got '%s'", planStep.Name)
	}

	// 测试不存在的步骤
	nonExistentStep := result.GetStepByName("nonexistent")
	if nonExistentStep != nil {
		t.Error("Expected nil for nonexistent step")
	}
}

// TestOrchestratorGetFailedStep 测试获取失败步骤
func TestOrchestratorGetFailedStep(t *testing.T) {
	// 测试有失败步骤的情况
	result := &OrchestratorResult{
		Steps: []StepResult{
			{Name: "plan", Success: true},
			{Name: "probe", Success: true},
			{Name: "apply", Success: false},
			{Name: "verify", Success: false},
		},
	}

	failedStep := result.GetFailedStep()
	if failedStep == nil {
		t.Error("Expected to find failed step")
	} else if failedStep.Name != "apply" {
		t.Errorf("Expected first failed step to be 'apply', got '%s'", failedStep.Name)
	}

	// 测试全部成功的情况
	successResult := &OrchestratorResult{
		Steps: []StepResult{
			{Name: "plan", Success: true},
			{Name: "probe", Success: true},
			{Name: "apply", Success: true},
		},
	}

	if failedStep := successResult.GetFailedStep(); failedStep != nil {
		t.Error("Expected nil for all successful steps")
	}
}
