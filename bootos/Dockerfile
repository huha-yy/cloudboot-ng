# CloudBoot BootOS v4 - OpenEuler Based
# 按照文档要求：Layer 0: 基础底座 (OpenEuler 22.03 LTS SP3)

FROM openeuler/openeuler:22.03-lts AS builder

# 配置国内镜像源（阿里云）
RUN sed -i 's|https://repo.openeuler.org|https://mirrors.aliyun.com/openeuler|g' /etc/yum.repos.d/openEuler.repo

# 安装构建依赖
RUN dnf install -y \
    go \
    gcc \
    git \
    && dnf clean all

# 构建 cb-agent（静态编译）
WORKDIR /build
COPY cb-agent /build/cb-agent
WORKDIR /build/cb-agent

# 设置 Go 环境变量
ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=amd64

RUN go build -ldflags="-s -w" -o /usr/local/bin/cb-agent ./main.go

# 运行时阶段 - Layer 1: 用户空间
FROM openeuler/openeuler:22.03-lts

# 配置国内镜像源（阿里云）
RUN sed -i 's|https://repo.openeuler.org|https://mirrors.aliyun.com/openeuler|g' /etc/yum.repos.d/openEuler.repo

# 安装 Systemd 和运行时依赖
RUN dnf install -y \
    # Init System (Systemd 精简版)
    systemd \
    systemd-udev \
    # Network Stack (NetworkManager + iproute2)
    NetworkManager \
    iproute \
    # Hardware Tools (完整工具链)
    pciutils \
    dmidecode \
    ethtool \
    nvme-cli \
    lsscsi \
    smartmontools \
    hdparm \
    # 加密支持（国密）
    openssl \
    # 其他必要工具
    curl \
    util-linux \
    kexec-tools \
    && dnf clean all

# 复制 cb-agent 二进制
COPY --from=builder /usr/local/bin/cb-agent /usr/local/bin/cb-agent

# 创建运行时空间 - Layer 2: Tmpfs 沙箱
RUN mkdir -p /opt/cloudboot/runtime/{bin,lib,configs,logs,pipe}

# 创建 Systemd 服务单元
COPY cloudboot.service /etc/systemd/system/cloudboot.service

# 创建初始化脚本
COPY init.sh /usr/local/bin/cloudboot-init.sh
RUN chmod +x /usr/local/bin/cloudboot-init.sh

# 配置 Systemd
RUN systemctl enable cloudboot.service

# 设置环境变量
ENV CB_SERVER_URL=http://10.0.0.1:8080
ENV CB_POLL_INTERVAL=5s

# 配置 Systemd 作为 PID 1
STOPSIGNAL SIGRTMIN+3
CMD ["/sbin/init"]