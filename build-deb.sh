#!/bin/bash

set -e

echo "start build EzeTranslate debian package"
echo ""

go build EzeTranslate.go
echo "build EzeTranslate success"
echo ""

VER_CODE=$("./EzeTranslate" -v)
echo "Version Code: $VER_CODE"

CURRENT_TIME=$(date +"%y%m%d%H%M%S")
echo "Current Time: $CURRENT_TIME"

TARGET_DIR="./build-target/deb/${VER_CODE}_${CURRENT_TIME}"
mkdir -p "$TARGET_DIR"

echo "target build path: $TARGET_DIR"
echo ""

cp -r "./deb-build-tpl" "$TARGET_DIR/eze-translate-${VER_CODE}"
mkdir -p "$TARGET_DIR/eze-translate-$VER_CODE/opt/EzeTranslate"
mkdir -p "$TARGET_DIR/eze-translate-$VER_CODE/usr/share/pixmaps"

cp -r "./EzeTranslate" "$TARGET_DIR/eze-translate-${VER_CODE}/opt/EzeTranslate"
cp "./res-static/icon/icon.png" "$TARGET_DIR/eze-translate-${VER_CODE}/usr/share/pixmaps/EzeTranslate.png"

cd "$TARGET_DIR"

echo "start build debian package in $TARGET_DIR ..."
echo ""

dpkg-deb --build eze-translate-${VER_CODE}

echo ""
echo "build success, target deb file: $TARGET_DIR/eze-translate-${VER_CODE}.deb"
