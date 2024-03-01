#!/bin/bash

echo "start build EzeTranslate"
echo ""

go build EzeTranslate.go

# 执行 ./EzeTranslate -v, 获取一个版本号 VER_CODE
VER_CODE=$("./EzeTranslate" -v)
echo "Version Code: $VER_CODE"

# 获取当前时间，格式为 yy/MM/dd_HH_mm_ss
CURRENT_TIME=$(date +"%y%m%d%H%M%S")
echo "Current Time: $CURRENT_TIME"

# 创建文件夹 ./build-target/VER_CODE_{yy/MM/dd_HH_mm_ss}
TARGET_DIR="./build-target/${VER_CODE}_linux_${CURRENT_TIME}"
mkdir -p "$TARGET_DIR"

# 将 ./EzeTranslate, ./res-static/ ,./config.yaml 都复制到 ./build-target/VER_CODE_{yy/MM/dd_HH_mm_ss}/
cp -r "./EzeTranslate" "./res-static" "./config.yaml" "$TARGET_DIR/"

echo "Build and packaging completed successfully."
echo ""
echo "target build path: $TARGET_DIR"
