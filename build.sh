echo "start build"
fyne package -icon ./res-static/icon/icon.png

echo "copy files to build-target"
mkdir ./build-target

echo "move build target..."
mv ./EzeTranslate build-target/
echo ""

echo "copy resource"
cp -rf ./res-static/ build-target/
echo ""

echo "copy config"
cp ./config.yaml ./build-target/
cp -rf .conf ./build-target/
echo ""

echo "build success, you can open binary file in build-target"