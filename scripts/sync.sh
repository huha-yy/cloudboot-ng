#!/bin/bash
# CloudBoot NG - 自动化同步脚本
# 功能: 测试 → 更新CHANGELOG → Git Commit → Git Push
# 用法: ./scripts/sync.sh "提交信息"

set -e  # 遇到错误立即退出

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 项目根目录
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$PROJECT_ROOT"

# 打印带颜色的消息
info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 检查参数
if [ $# -eq 0 ]; then
    error "缺少提交信息"
    echo "用法: $0 \"提交信息\""
    echo "示例: $0 \"feat: 完成Provider沙箱隔离\""
    exit 1
fi

COMMIT_MSG="$1"

# Banner
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${BLUE}  CloudBoot NG - 自动化同步脚本${NC}"
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""

# ============================================
# Step 1: 运行测试
# ============================================
info "Step 1/5: 运行单元测试..."
echo ""

if ! go test ./... -v -cover; then
    error "测试失败！请修复后再提交"
    exit 1
fi

echo ""
success "所有测试通过"
echo ""

# ============================================
# Step 2: 更新 CHANGELOG
# ============================================
info "Step 2/5: 更新 CHANGELOG.md..."
echo ""

CHANGELOG_FILE="CHANGELOG.md"
TODAY=$(date +"%Y-%m-%d")
TIME_NOW=$(date +"%H:%M")

# 如果CHANGELOG不存在，创建初始文件
if [ ! -f "$CHANGELOG_FILE" ]; then
    cat > "$CHANGELOG_FILE" <<EOF
# Changelog

All notable changes to CloudBoot NG will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

EOF
fi

# 解析提交信息类型
COMMIT_TYPE=""
COMMIT_DESC="$COMMIT_MSG"

if [[ "$COMMIT_MSG" =~ ^feat:\ (.+)$ ]]; then
    COMMIT_TYPE="### Added"
    COMMIT_DESC="${BASH_REMATCH[1]}"
elif [[ "$COMMIT_MSG" =~ ^fix:\ (.+)$ ]]; then
    COMMIT_TYPE="### Fixed"
    COMMIT_DESC="${BASH_REMATCH[1]}"
elif [[ "$COMMIT_MSG" =~ ^refactor:\ (.+)$ ]]; then
    COMMIT_TYPE="### Changed"
    COMMIT_DESC="${BASH_REMATCH[1]}"
elif [[ "$COMMIT_MSG" =~ ^docs:\ (.+)$ ]]; then
    COMMIT_TYPE="### Documentation"
    COMMIT_DESC="${BASH_REMATCH[1]}"
elif [[ "$COMMIT_MSG" =~ ^test:\ (.+)$ ]]; then
    COMMIT_TYPE="### Testing"
    COMMIT_DESC="${BASH_REMATCH[1]}"
elif [[ "$COMMIT_MSG" =~ ^chore:\ (.+)$ ]]; then
    COMMIT_TYPE="### Maintenance"
    COMMIT_DESC="${BASH_REMATCH[1]}"
else
    COMMIT_TYPE="### Changed"
fi

# 在Unreleased段落下添加条目
# 使用awk在特定位置插入
awk -v date="$TODAY" -v time="$TIME_NOW" -v type="$COMMIT_TYPE" -v desc="$COMMIT_DESC" '
BEGIN { done=0; }
/## \[Unreleased\]/ {
    print;
    if (!done) {
        # 检查下一行是否已经有同类型的标题
        getline;
        if ($0 == type) {
            print;
            print "- " desc " (" date " " time ")";
        } else {
            print "";
            print type;
            print "- " desc " (" date " " time ")";
            print "";
            print;
        }
        done=1;
    }
    next;
}
{ print }
' "$CHANGELOG_FILE" > "${CHANGELOG_FILE}.tmp" && mv "${CHANGELOG_FILE}.tmp" "$CHANGELOG_FILE"

success "CHANGELOG.md 已更新"
echo ""

# ============================================
# Step 3: 检查Git状态
# ============================================
info "Step 3/5: 检查 Git 状态..."
echo ""

if ! git diff-index --quiet HEAD --; then
    info "检测到未提交的更改"
    git status --short
else
    warn "没有检测到更改，跳过提交"
    exit 0
fi

echo ""

# ============================================
# Step 4: Git Commit
# ============================================
info "Step 4/5: 提交更改到 Git..."
echo ""

git add -A

# 添加Co-Authored-By
FULL_COMMIT_MSG="$(cat <<EOF
$COMMIT_MSG

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>
EOF
)"

git commit -m "$FULL_COMMIT_MSG"

COMMIT_HASH=$(git rev-parse --short HEAD)
success "已提交: $COMMIT_HASH - $COMMIT_MSG"
echo ""

# ============================================
# Step 5: Git Push
# ============================================
info "Step 5/5: 推送到远程仓库..."
echo ""

CURRENT_BRANCH=$(git rev-parse --abbrev-ref HEAD)
info "当前分支: $CURRENT_BRANCH"

if git push origin "$CURRENT_BRANCH"; then
    success "推送成功到 origin/$CURRENT_BRANCH"
else
    error "推送失败，请手动执行: git push origin $CURRENT_BRANCH"
    exit 1
fi

echo ""
echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${GREEN}  ✅ 同步完成${NC}"
echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""
info "查看远程仓库: https://github.com/huha-yy/cloudboot-ng"
echo ""
