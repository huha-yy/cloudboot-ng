#!/bin/bash
# 钉钉机器人进展推送脚本
# 用法: ./notify.sh "消息内容"

message=${1:-"默认消息"}

curl -L 'https://oapi.dingtalk.com/robot/send?access_token=058c83912e519059586f0666160ed5af3678edf4ad8cb82bcc9e336d10771606' \
  -H 'Content-Type: application/json' \
  -d "{
    \"msgtype\": \"text\",
    \"text\": {
        \"content\": \"StarryTalk $message\"
    }
}"
