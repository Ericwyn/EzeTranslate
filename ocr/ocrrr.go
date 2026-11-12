package ocr

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/Ericwyn/EzeTranslate/log"
)

type OcrCallback func(text string, success bool)

// RunOcr 运行 OCR 识别文字并返回
func RunOcr(callback OcrCallback) {
	RunOcrWithHook(nil, callback)
}

// RunOcrWithHook 与 RunOcr 相同，但在「截图完成、即将开始识别」之间调用
// onCaptured（若非 nil）。调用方可借此在耗时的识别开始前先恢复窗口显示，
// 避免窗口隐藏时间过长。onCaptured 在截图失败时不会被调用。
func RunOcrWithHook(onCaptured func(), callback OcrCallback) {
	if callback == nil {
		return
	}

	// Use a fresh temporary file so cancelling/failing the screenshot cannot
	// accidentally OCR an old image left by a previous run.
	image, err := os.CreateTemp("", "ezetranslate-ocr-*.png")
	if err != nil {
		log.E("创建 OCR 临时图片失败: ", err)
		callback("", false)
		return
	}
	screenshotPath := image.Name()
	if err := image.Close(); err != nil {
		_ = os.Remove(screenshotPath)
		log.E("关闭 OCR 临时图片失败: ", err)
		callback("", false)
		return
	}
	_ = os.Remove(screenshotPath)
	defer os.Remove(screenshotPath)

	var captureErr error
	if shouldUseScreenshotPortal() {
		captureErr = capturePortalScreenshot(screenshotPath)
	} else {
		output, err := exec.Command("gnome-screenshot", "-a", "-f", screenshotPath).CombinedOutput()
		if err != nil {
			captureErr = fmt.Errorf("gnome-screenshot: %w: %s", err, strings.TrimSpace(string(output)))
		}
	}
	if captureErr != nil {
		log.E("截图失败: ", captureErr)
		callback("", false)
		return
	}
	info, err := os.Stat(screenshotPath)
	if err != nil || !info.Mode().IsRegular() || info.Size() == 0 {
		log.E("gnome-screenshot 未生成有效图片: ", err)
		callback("", false)
		return
	}

	// 截图已完成，后续 mogrify + tesseract 较慢，先让调用方恢复窗口
	if onCaptured != nil {
		onCaptured()
	}

	// ImageMagick is optional. Debian packages do not always include it; OCR
	// should still proceed with the original screenshot when mogrify is absent.
	if _, err := exec.LookPath("mogrify"); err == nil {
		if output, err := exec.Command("mogrify", "-modulate", "100,0", "-resize", "400%", screenshotPath).CombinedOutput(); err != nil {
			log.E("mogrify 失败，改用原始截图: ", err, " ", strings.TrimSpace(string(output)))
		}
	}

	language, err := tesseractLanguage()
	if err != nil {
		log.E("Tesseract 语言包不可用: ", err)
		callback("", false)
		return
	}
	output, err := exec.Command("tesseract", screenshotPath, "stdout", "-l", language).CombinedOutput()
	if err != nil {
		log.E("tesseract 识别失败: ", err, " ", strings.TrimSpace(string(output)))
		callback("", false)
		return
	}
	tranRes := strings.TrimSpace(string(output))

	if tranRes == "" {
		callback("", false)
		return
	}
	callback(tranRes, true)
}

// GetFileModTime 获取文件修改时间并返回 Unix 时间戳。
func GetFileModTime(path string) int64 {
	info, err := os.Stat(path)
	if err != nil {
		log.E("stat fileinfo error: ", err)
		return time.Now().Unix()
	}
	return info.ModTime().Unix()
}

func tesseractLanguage() (string, error) {
	output, err := exec.Command("tesseract", "--list-langs").CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("无法读取语言包列表: %w: %s", err, strings.TrimSpace(string(output)))
	}
	available := make(map[string]bool)
	for _, line := range strings.Split(string(output), "\n") {
		language := strings.TrimSpace(line)
		if language != "" && !strings.HasPrefix(language, "List of available languages") {
			available[language] = true
		}
	}
	return selectTesseractLanguage(available)
}

func selectTesseractLanguage(available map[string]bool) (string, error) {
	switch {
	case available["chi_sim"] && available["eng"]:
		return "chi_sim+eng", nil
	case available["eng"]:
		return "eng", nil
	case available["chi_sim"]:
		return "chi_sim", nil
	default:
		return "", fmt.Errorf("未安装 eng 或 chi_sim 语言包")
	}
}
