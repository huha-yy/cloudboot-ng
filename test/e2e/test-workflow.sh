#!/bin/bash
# CloudBoot E2E Workflow Test
# Tests the complete provisioning workflow using Docker

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Test results
TESTS_PASSED=0
TESTS_FAILED=0

echo "=========================================="
echo " CloudBoot E2E Workflow Test"
echo "=========================================="
echo ""

# Helper functions
pass() {
    echo -e "${GREEN}[✓]${NC} $1"
    ((TESTS_PASSED++))
}

fail() {
    echo -e "${RED}[✗]${NC} $1"
    ((TESTS_FAILED++))
}

info() {
    echo -e "${YELLOW}[*]${NC} $1"
}

check_service() {
    local service=$1
    local url=$2
    if curl -sf "$url" >/dev/null 2>&1; then
        pass "$service is running"
        return 0
    else
        fail "$service is not responding at $url"
        return 1
    fi
}

# Step 1: Start CloudBoot Server
info "Step 1: Starting CloudBoot Server..."
cd "$PROJECT_ROOT"

# Build server if needed
if [ ! -f "bin/cloudboot-server" ]; then
    info "Building server..."
    go build -o bin/cloudboot-server ./cmd/server
fi

# Start server in background
info "Starting server on port 8080..."
./bin/cloudboot-server &
SERVER_PID=$!
sleep 3

# Check health
check_service "CloudBoot Server" "http://localhost:8080/health"

# Step 2: Seed Database
info "Step 2: Seeding database..."
cd "$PROJECT_ROOT/tools/seed"
go run main.go --db="$PROJECT_ROOT/cloudboot.db"
pass "Database seeded"

# Step 3: Verify API Endpoints
info "Step 3: Testing API endpoints..."

# Test machines endpoint
if curl -sf http://localhost:8080/api/v1/machines | jq -e '.machines' >/dev/null 2>&1; then
    pass "GET /api/v1/machines works"
else
    fail "GET /api/v1/machines failed"
fi

# Test profiles endpoint
if curl -sf http://localhost:8080/api/v1/profiles | jq -e '.profiles' >/dev/null 2>&1; then
    pass "GET /api/v1/profiles works"
else
    fail "GET /api/v1/profiles failed"
fi

# Test jobs endpoint
if curl -sf http://localhost:8080/api/v1/jobs | jq -e '.jobs' >/dev/null 2>&1; then
    pass "GET /api/v1/jobs works"
else
    fail "GET /api/v1/jobs failed"
fi

# Step 4: Test Agent Registration
info "Step 4: Testing agent registration..."
REGISTER_RESPONSE=$(curl -sf -X POST http://localhost:8080/api/boot/v1/register \
    -H "Content-Type: application/json" \
    -d '{
        "mac_address": "aa:bb:cc:dd:ee:ff",
        "ip_address": "192.168.1.200",
        "hardware_spec": {
            "system_manufacturer": "Test Vendor",
            "system_product": "Test Model",
            "cpu": {"model": "Test CPU", "cores": 8},
            "memory": {"total_gb": 64}
        }
    }')

if echo "$REGISTER_RESPONSE" | jq -e '.machine_id' >/dev/null 2>&1; then
    MACHINE_ID=$(echo "$REGISTER_RESPONSE" | jq -r '.machine_id')
    pass "Agent registration successful (Machine ID: $MACHINE_ID)"
else
    fail "Agent registration failed"
fi

# Step 5: Test Task Polling
info "Step 5: Testing task polling..."
TASK_RESPONSE=$(curl -sf "http://localhost:8080/api/boot/v1/task?machine_id=$MACHINE_ID")

if echo "$TASK_RESPONSE" | jq -e '.no_task' >/dev/null 2>&1; then
    pass "Task polling works (no task available)"
else
    fail "Task polling failed"
fi

# Step 6: Test Provision Request
info "Step 6: Testing provision request..."

# Get first profile
PROFILE_ID=$(curl -sf http://localhost:8080/api/v1/profiles | jq -r '.profiles[0].id')

if [ -n "$PROFILE_ID" ] && [ "$PROFILE_ID" != "null" ]; then
    # Provision machine
    PROVISION_RESPONSE=$(curl -sf -X POST "http://localhost:8080/api/v1/machines/$MACHINE_ID/provision" \
        -H "Content-Type: application/json" \
        -d "{\"profile_id\": \"$PROFILE_ID\"}")

    if echo "$PROVISION_RESPONSE" | jq -e '.id' >/dev/null 2>&1; then
        JOB_ID=$(echo "$PROVISION_RESPONSE" | jq -r '.id')
        pass "Provisioning job created (Job ID: $JOB_ID)"
    else
        fail "Provision request failed"
    fi
else
    fail "No profiles available for testing"
fi

# Step 7: Test Log Upload
info "Step 7: Testing log upload..."
if [ -n "$JOB_ID" ]; then
    LOG_RESPONSE=$(curl -sf -X POST http://localhost:8080/api/boot/v1/logs \
        -H "Content-Type: application/json" \
        -d "{
            \"job_id\": \"$JOB_ID\",
            \"logs\": [
                {
                    \"ts\": \"$(date -u +%Y-%m-%dT%H:%M:%SZ)\",
                    \"level\": \"INFO\",
                    \"component\": \"test\",
                    \"msg\": \"Test log entry\"
                }
            ]
        }")

    if [ $? -eq 0 ]; then
        pass "Log upload successful"
    else
        fail "Log upload failed"
    fi
else
    fail "Skipping log upload (no job ID)"
fi

# Step 8: Test Status Report
info "Step 8: Testing status report..."
if [ -n "$JOB_ID" ]; then
    STATUS_RESPONSE=$(curl -sf -X POST http://localhost:8080/api/boot/v1/status \
        -H "Content-Type: application/json" \
        -d "{
            \"job_id\": \"$JOB_ID\",
            \"status\": \"running\",
            \"step_current\": \"Installing packages\"
        }")

    if [ $? -eq 0 ]; then
        pass "Status report successful"
    else
        fail "Status report failed"
    fi
else
    fail "Skipping status report (no job ID)"
fi

# Step 9: Test SSE Log Streaming
info "Step 9: Testing SSE log streaming..."
if [ -n "$JOB_ID" ]; then
    # Start SSE stream in background and capture for 2 seconds
    timeout 2 curl -sf "http://localhost:8080/api/stream/logs/$JOB_ID" > /tmp/sse-test.log 2>&1 || true

    if [ -s /tmp/sse-test.log ]; then
        pass "SSE log streaming works"
    else
        fail "SSE log streaming failed"
    fi
    rm -f /tmp/sse-test.log
else
    fail "Skipping SSE test (no job ID)"
fi

# Step 10: Test Profile Preview
info "Step 10: Testing profile config preview..."
if [ -n "$PROFILE_ID" ]; then
    PREVIEW_RESPONSE=$(curl -sf "http://localhost:8080/api/v1/profiles/$PROFILE_ID/preview")

    if echo "$PREVIEW_RESPONSE" | grep -q "Kickstart"; then
        pass "Profile config preview works"
    else
        fail "Profile config preview failed"
    fi
else
    fail "Skipping preview test (no profile ID)"
fi

# Cleanup
info "Cleanup: Stopping server..."
kill $SERVER_PID 2>/dev/null || true
pass "Server stopped"

# Summary
echo ""
echo "=========================================="
echo " Test Summary"
echo "=========================================="
echo -e "${GREEN}Passed: $TESTS_PASSED${NC}"
echo -e "${RED}Failed: $TESTS_FAILED${NC}"
echo ""

if [ $TESTS_FAILED -eq 0 ]; then
    echo -e "${GREEN}All tests passed!${NC}"
    exit 0
else
    echo -e "${RED}Some tests failed.${NC}"
    exit 1
fi
