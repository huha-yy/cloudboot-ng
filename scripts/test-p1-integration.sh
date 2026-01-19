#!/bin/bash
#
# CloudBoot NG - P1 功能集成测试
# 测试PXE/iPXE启动和Kickstart/AutoYaST模板生成
#

set -e

SERVER_URL="http://localhost:8080"
TEST_MAC="aa:bb:cc:dd:ee:ff"
TEST_MACHINE_ID=""

echo "=========================================="
echo "CloudBoot NG - P1 Integration Test"
echo "=========================================="
echo

# 颜色定义
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

function pass() {
    echo -e "${GREEN}✓${NC} $1"
}

function fail() {
    echo -e "${RED}✗${NC} $1"
    exit 1
}

function info() {
    echo -e "${YELLOW}→${NC} $1"
}

# 测试1: Agent注册
info "Test 1: Agent Registration"
REGISTER_RESP=$(curl -s -X POST ${SERVER_URL}/api/boot/v1/register \
  -H "Content-Type: application/json" \
  -d '{
    "mac_address": "'"${TEST_MAC}"'",
    "ip_address": "10.0.0.100",
    "hostname": "test-server-001",
    "hardware_spec": {
      "schema_version": "1.0",
      "system": {"manufacturer": "Dell", "product_name": "PowerEdge R740"},
      "cpu": {"cores": 32, "sockets": 2},
      "memory": {"total_bytes": 68719476736}
    }
  }')

TEST_MACHINE_ID=$(echo $REGISTER_RESP | jq -r '.machine_id')

if [ "$TEST_MACHINE_ID" != "null" ] && [ "$TEST_MACHINE_ID" != "" ]; then
    pass "Agent registered successfully: machine_id=${TEST_MACHINE_ID}"
else
    fail "Agent registration failed"
fi

# 测试2: iPXE脚本生成 (discovery模式)
info "Test 2: iPXE Script Generation (Discovery Mode)"
IPXE_RESP=$(curl -s ${SERVER_URL}/boot/ipxe/${TEST_MAC})

if echo "$IPXE_RESP" | grep -q "#!ipxe"; then
    pass "iPXE script generated"
    if echo "$IPXE_RESP" | grep -q "discovery"; then
        pass "Boot mode is 'discovery'"
    else
        fail "Boot mode is not 'discovery'"
    fi
else
    fail "Invalid iPXE script"
fi

# 测试3: 创建OS Profile (CentOS 7)
info "Test 3: Create OS Profile (CentOS 7)"
PROFILE_RESP=$(curl -s -X POST ${SERVER_URL}/api/v1/profiles \
  -H "Content-Type: application/json" \
  -d '{
    "name": "CentOS 7.9 Standard",
    "distro": "centos7",
    "version": "7.9",
    "config": {
      "root_password_hash": "$6$rounds=656000$YQKJWPTd7USGake4$CmMFQe/ppJPPGYnRsVAlyuPmcxgRfWWNJ5S0AXBX5HcOPY/Y2DZVFB6OwDKk0Fmb2a/C.T77i2VZCpNzqwCB.",
      "timezone": "America/New_York",
      "repo_url": "http://mirror.centos.org/centos/7/os/x86_64/",
      "partitions": [
        {"mount_point": "/boot", "size_mb": 1024, "file_system": "ext4"},
        {"mount_point": "swap", "size_mb": 8192, "file_system": "swap"},
        {"mount_point": "/", "size_mb": 0, "file_system": "ext4", "grow": true}
      ],
      "network_config": {
        "boot_proto": "dhcp",
        "device": "eth0"
      },
      "packages": ["vim", "wget", "curl", "net-tools"],
      "install_agent": true
    }
  }')

PROFILE_ID=$(echo $PROFILE_RESP | jq -r '.id')

if [ "$PROFILE_ID" != "null" ] && [ "$PROFILE_ID" != "" ]; then
    pass "OS Profile created: profile_id=${PROFILE_ID}"
else
    fail "OS Profile creation failed"
fi

# 测试4: 创建安装任务
info "Test 4: Create Installation Job"
JOB_RESP=$(curl -s -X POST ${SERVER_URL}/api/v1/machines/${TEST_MACHINE_ID}/provision \
  -H "Content-Type: application/json" \
  -d '{
    "profile_id": "'"${PROFILE_ID}"'",
    "type": "install_os"
  }')

JOB_ID=$(echo $JOB_RESP | jq -r '.id')

if [ "$JOB_ID" != "null" ] && [ "$JOB_ID" != "" ]; then
    pass "Installation job created: job_id=${JOB_ID}"
else
    fail "Installation job creation failed"
fi

# 测试5: iPXE脚本生成 (install模式)
info "Test 5: iPXE Script Generation (Install Mode)"
sleep 1  # 等待数据库更新
IPXE_RESP2=$(curl -s ${SERVER_URL}/boot/ipxe/${TEST_MAC})

if echo "$IPXE_RESP2" | grep -q "install"; then
    pass "Boot mode changed to 'install'"
    if echo "$IPXE_RESP2" | grep -q "inst.ks"; then
        pass "Kickstart URL found in iPXE script"
    else
        fail "Kickstart URL not found"
    fi
else
    fail "Boot mode is not 'install'"
fi

# 测试6: Kickstart配置生成
info "Test 6: Kickstart Configuration Generation"
KS_RESP=$(curl -s ${SERVER_URL}/boot/kickstart/${TEST_MACHINE_ID})

if echo "$KS_RESP" | grep -q "# CloudBoot NG"; then
    pass "Kickstart configuration generated"

    # 验证关键字段
    if echo "$KS_RESP" | grep -q "url --url="; then
        pass "Repo URL configured"
    fi

    if echo "$KS_RESP" | grep -q "part /boot"; then
        pass "Partitions configured"
    fi

    if echo "$KS_RESP" | grep -q "network --bootproto=dhcp"; then
        pass "Network configured"
    fi

    if echo "$KS_RESP" | grep -q "vim"; then
        pass "Packages configured"
    fi
else
    fail "Invalid Kickstart configuration"
fi

echo
echo "=========================================="
echo -e "${GREEN}All tests passed!${NC}"
echo "=========================================="
echo
echo "Summary:"
echo "  - Machine ID: ${TEST_MACHINE_ID}"
echo "  - Profile ID: ${PROFILE_ID}"
echo "  - Job ID: ${JOB_ID}"
echo "  - iPXE URL: ${SERVER_URL}/boot/ipxe/${TEST_MAC}"
echo "  - Kickstart URL: ${SERVER_URL}/boot/kickstart/${TEST_MACHINE_ID}"
echo

# 可选：保存生成的文件用于检查
echo "Saving generated files..."
curl -s ${SERVER_URL}/boot/ipxe/${TEST_MAC} > /tmp/test.ipxe
curl -s ${SERVER_URL}/boot/kickstart/${TEST_MACHINE_ID} > /tmp/test.ks

echo "  - iPXE script: /tmp/test.ipxe"
echo "  - Kickstart config: /tmp/test.ks"
echo

echo "✅ P1 Integration Test Completed!"
