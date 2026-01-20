package cspm

import (
	"context"
	"os"
	"testing"
	"time"
)

// TestOrchestratorApplyWithPlan 测试完整的原子序列执行
func TestOrchestratorApplyWithPlan(t *testing.T) {
	// 清理状态文件，确保测试从干净状态开始
	os.Remove("/tmp/cloudboot-provider-mock-state.json")

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

// TestOrchestratorIdempotency 测试幂等性
func TestOrchestratorIdempotency(t *testing.T) {
	// 清理状态文件，确保测试从干净状态开始
	os.Remove("/tmp/cloudboot-provider-mock-state.json")

	// 使用 Mock Provider
	executor := NewExecutor("../../../cmd/provider-mock/provider-mock")
	executor.SetTimeout(15 * time.Second) // 延长超时，因为需要执行两次

	orchestrator := NewOrchestrator(executor)

	config := map[string]interface{}{
		"desired_state": map[string]interface{}{
			"level":  "10",
			"drives": []string{"0:0", "0:1", "0:2", "0:3"},
		},
	}

	ctx := context.Background()

	// 第一次执行：应该创建 RAID
	t.Log("First execution: should create RAID")
	result1, err := orchestrator.ApplyWithPlan(ctx, config)
	if err != nil {
		t.Fatalf("First ApplyWithPlan failed: %v", err)
	}

	if !result1.Success {
		t.Errorf("First execution should succeed")
		if failedStep := result1.GetFailedStep(); failedStep != nil {
			t.Logf("Failed step: %s", failedStep.Name)
		}
	}

	if result1.Idempotent {
		t.Errorf("First execution should NOT be idempotent (should apply changes)")
	}

	t.Logf("First execution: Success=%v, Idempotent=%v, Duration=%v",
		result1.Success, result1.Idempotent, result1.Duration)

	// 第二次执行：应该检测到已达标，跳过 Apply
	t.Log("Second execution: should detect convergence and skip apply")
	result2, err := orchestrator.ApplyWithPlan(ctx, config)
	if err != nil {
		t.Fatalf("Second ApplyWithPlan failed: %v", err)
	}

	if !result2.Success {
		t.Errorf("Second execution should succeed")
		if failedStep := result2.GetFailedStep(); failedStep != nil {
			t.Logf("Failed step: %s", failedStep.Name)
		}
	}

	// 关键验证：第二次应该是幂等的
	if !result2.Idempotent {
		t.Errorf("Second execution should be idempotent (system already converged)")
	}

	// 验证第二次执行跳过了 Apply 步骤
	// 应该只有 Plan 和 Probe，没有 Apply 和 Verify
	expectedSteps := []string{"plan", "probe"}
	if len(result2.Steps) != len(expectedSteps) {
		t.Errorf("Second execution should have %d steps (no apply/verify), got %d",
			len(expectedSteps), len(result2.Steps))
	}

	t.Logf("Second execution: Success=%v, Idempotent=%v, Duration=%v, Steps=%d",
		result2.Success, result2.Idempotent, result2.Duration, len(result2.Steps))

	// 验证第二次执行速度更快（因为跳过了 Apply）
	if result2.Duration >= result1.Duration {
		t.Logf("Warning: Second execution (%v) should be faster than first (%v)",
			result2.Duration, result1.Duration)
	}
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
