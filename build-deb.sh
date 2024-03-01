#!/bin/bash

echo "start build EzeTranslate debian package"
echo ""

go build EzeTranslate.go
echo "build EzeTranslate success"
echo ""

# 执行 ./EzeTranslate -v, 获取一个版本号 VER_CODE
VER_CODE=$("./EzeTranslate" -v)
echo "Version Code: $VER_CODE"

# 获取当前时间，格式为 yy/MM/dd_HH_mm_ss
CURRENT_TIME=$(date +"%y%m%d%H%M%S")
echo "Current Time: $CURRENT_TIME"

# 创建文件夹 ./build-target/deb/VER_CODE_{yy/MM/dd_HH_mm_ss}/EzeTranslate-deb
TARGET_DIR="./build-target/deb/${VER_CODE}_${CURRENT_TIME}"
mkdir -p "$TARGET_DIR"

echo "target build path: $TARGET_DIR"
echo ""

# 复制 deb-build-tpl 到 $TARGET_DIR/EzeTranslate-deb 里面
# 复制 EzeTranslate, config.yaml, res-static/ 到 $TARGET_DIR/EzeTranslate-deb/opt/EzeTranslate/ 里
cp -r "./deb-build-tpl" "$TARGET_DIR/eze-translate"
cp -r "./EzeTranslate" "./config.yaml" "./res-static" "$TARGET_DIR/eze-translate/opt/EzeTranslate"

# 开始 build deb
cd "$TARGET_DIR"

echo "start build debian package in $TARGET_DIR ..."
echo ""

dpkg-deb --build eze-translate

echo ""
echo "build success, target deb file: $TARGET_DIR/eze-translate.deb"
