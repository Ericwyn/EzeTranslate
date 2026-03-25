#!/bin/bash

set -e

echo "start build EzeTranslate debian package"
echo ""

go build EzeTranslate.go
echo "build EzeTranslate success"
echo ""

VER_CODE=$("./EzeTranslate" -v)
DEB_VERSION=$(printf '%s' "$VER_CODE" | sed -E 's/^[Vv]//; s/[^0-9A-Za-z.+~-]+/-/g')
echo "Version Code: $VER_CODE"
echo "Deb Version: $DEB_VERSION"

CURRENT_TIME=$(date +"%y%m%d%H%M%S")
echo "Current Time: $CURRENT_TIME"

TARGET_DIR="./build-target/deb/${VER_CODE}_${CURRENT_TIME}"
mkdir -p "$TARGET_DIR"

echo "target build path: $TARGET_DIR"
echo ""

cp -r "./deb-build-tpl" "$TARGET_DIR/eze-translate-${VER_CODE}"
mkdir -p "$TARGET_DIR/eze-translate-$VER_CODE/opt/EzeTranslate"
mkdir -p "$TARGET_DIR/eze-translate-$VER_CODE/usr/share/pixmaps"
sed -i "s/^Version: .*/Version: ${DEB_VERSION}/" "$TARGET_DIR/eze-translate-${VER_CODE}/DEBIAN/control"

cp -r "./EzeTranslate" "$TARGET_DIR/eze-translate-${VER_CODE}/opt/EzeTranslate"
cp "./res-static/icon/icon.png" "$TARGET_DIR/eze-translate-${VER_CODE}/usr/share/pixmaps/EzeTranslate.png"

cd "$TARGET_DIR"

echo "start build debian package in $TARGET_DIR ..."
echo ""

dpkg-deb --build eze-translate-${VER_CODE}

echo ""
echo "build success, target deb file: $TARGET_DIR/eze-translate-${VER_CODE}.deb"
