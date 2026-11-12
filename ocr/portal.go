package ocr

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/godbus/dbus/v5"
)

const (
	portalBusName    = "org.freedesktop.portal.Desktop"
	portalObject     = dbus.ObjectPath("/org/freedesktop/portal/desktop")
	portalScreenshot = "org.freedesktop.portal.Screenshot"
	portalRequest    = "org.freedesktop.portal.Request"
)

var errScreenshotCancelled = errors.New("截图已取消")

// capturePortalScreenshot uses XDG Desktop Portal instead of GNOME Shell's
// private screenshot D-Bus API. The portal handles the user interaction and
// returns a URI that the application is allowed to read.
func capturePortalScreenshot(destination string) error {
	conn, err := dbus.ConnectSessionBus()
	if err != nil {
		return fmt.Errorf("连接 session D-Bus 失败: %w", err)
	}
	defer conn.Close()

	match := []dbus.MatchOption{
		dbus.WithMatchInterface(portalRequest),
		dbus.WithMatchMember("Response"),
		dbus.WithMatchPathNamespace(dbus.ObjectPath(string(portalObject) + "/request")),
	}
	if err := conn.AddMatchSignal(match...); err != nil {
		return fmt.Errorf("订阅截图 portal 响应失败: %w", err)
	}
	defer conn.RemoveMatchSignal(match...)

	signals := make(chan *dbus.Signal, 8)
	conn.Signal(signals)
	defer conn.RemoveSignal(signals)

	tokenBytes := make([]byte, 12)
	if _, err := rand.Read(tokenBytes); err != nil {
		return fmt.Errorf("生成截图请求标识失败: %w", err)
	}
	token := "ezetranslate_" + hex.EncodeToString(tokenBytes)
	options := map[string]dbus.Variant{
		"handle_token": dbus.MakeVariant(token),
		"interactive":  dbus.MakeVariant(true),
	}

	// Version 3 portals can explicitly request an area. Older portal versions
	// still get interactive=true, which lets the desktop choose its supported
	// interactive screenshot flow without passing an unsupported target.
	var targets dbus.Variant
	if call := conn.Object(portalBusName, portalObject).Call(
		"org.freedesktop.DBus.Properties.Get", 0, portalScreenshot, "AvailableTargets",
	); call.Err == nil && call.Store(&targets) == nil {
		if mask, ok := targets.Value().(uint32); ok && mask&4 != 0 {
			options["target"] = dbus.MakeVariant(uint32(4))
		}
	}

	var requestPath dbus.ObjectPath
	call := conn.Object(portalBusName, portalObject).Call(
		portalScreenshot+".Screenshot", 0, "", options,
	)
	if call.Err != nil {
		return fmt.Errorf("调用截图 portal 失败: %w", call.Err)
	}
	if err := call.Store(&requestPath); err != nil {
		return fmt.Errorf("读取截图 portal 请求句柄失败: %w", err)
	}

	timer := time.NewTimer(5 * time.Minute)
	defer timer.Stop()
	for {
		select {
		case signal, ok := <-signals:
			if !ok {
				return errors.New("截图 portal 的 D-Bus 连接已关闭")
			}
			if signal.Path != requestPath || signal.Name != portalRequest+".Response" || len(signal.Body) < 2 {
				continue
			}
			response, ok := signal.Body[0].(uint32)
			if !ok {
				return errors.New("截图 portal 返回了无法识别的响应")
			}
			if response == 1 {
				return errScreenshotCancelled
			}
			if response != 0 {
				return fmt.Errorf("截图 portal 请求失败（response=%d）", response)
			}
			results, ok := signal.Body[1].(map[string]dbus.Variant)
			if !ok {
				return errors.New("截图 portal 未返回结果")
			}
			uriVariant, ok := results["uri"]
			if !ok {
				return errors.New("截图 portal 未返回图片 URI")
			}
			uri, ok := uriVariant.Value().(string)
			if !ok {
				return errors.New("截图 portal 返回了无效图片 URI")
			}
			return copyScreenshotURI(uri, destination)
		case <-timer.C:
			_ = conn.Object(portalBusName, requestPath).Call(portalRequest+".Close", 0)
			return errors.New("等待截图 portal 响应超时")
		}
	}
}

func copyScreenshotURI(rawURI, destination string) error {
	parsed, err := url.Parse(rawURI)
	if err != nil || parsed.Scheme != "file" {
		return fmt.Errorf("截图 portal 返回了不支持的 URI: %q", rawURI)
	}
	data, err := os.ReadFile(parsed.Path)
	if err != nil {
		return fmt.Errorf("读取 portal 截图失败: %w", err)
	}
	if len(data) == 0 {
		return errors.New("portal 截图文件为空")
	}
	if err := os.WriteFile(destination, data, 0600); err != nil {
		return fmt.Errorf("保存 portal 截图失败: %w", err)
	}
	return nil
}

func shouldUseScreenshotPortal() bool {
	return os.Getenv("WAYLAND_DISPLAY") != "" || strings.EqualFold(os.Getenv("XDG_SESSION_TYPE"), "wayland")
}
