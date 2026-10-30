package ui

import (
	"os/exec"
	"regexp"
	"strconv"

	"github.com/Ericwyn/EzeTranslate/log"
	"github.com/vivaui/viva/ui"
)

// monitor 是一块显示器的像素几何（来自 xrandr --listmonitors）。
type monitor struct {
	x, y, w, h int
}

// monitorGeom 匹配 xrandr --listmonitors 里的几何串，例如 "2560/597x1440/336+0+0"
var monitorGeom = regexp.MustCompile(`(\d+)/\d+x(\d+)/\d+\+(\d+)\+(\d+)`)

// centeredOrigin 返回让 w×h 窗口在“光标所在显示器”上居中的左上角坐标。
// 仅 Linux/X11 下通过 xrandr + xdotool 查询；拿不到信息时返回 nil，
// 交给窗口管理器自行摆放（退化为原生行为，不会出错）。
func centeredOrigin(w, h int) *ui.Point {
	monitors := listMonitors()
	if len(monitors) == 0 {
		return nil
	}

	target := monitors[0]
	if cx, cy, ok := cursorPos(); ok {
		for _, m := range monitors {
			if cx >= m.x && cx < m.x+m.w && cy >= m.y && cy < m.y+m.h {
				target = m
				break
			}
		}
	}

	ox := target.x + (target.w-w)/2
	oy := target.y + (target.h-h)/2
	if ox < target.x {
		ox = target.x
	}
	if oy < target.y {
		oy = target.y
	}
	return &ui.Point{X: float32(ox), Y: float32(oy)}
}

// listMonitors 解析 xrandr --listmonitors 的输出。
func listMonitors() []monitor {
	out, err := exec.Command("xrandr", "--listmonitors").Output()
	if err != nil {
		log.D("xrandr --listmonitors 失败: ", err)
		return nil
	}
	var res []monitor
	for _, line := range splitLines(string(out)) {
		m := monitorGeom.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		w, _ := strconv.Atoi(m[1])
		h, _ := strconv.Atoi(m[2])
		x, _ := strconv.Atoi(m[3])
		y, _ := strconv.Atoi(m[4])
		if w > 0 && h > 0 {
			res = append(res, monitor{x: x, y: y, w: w, h: h})
		}
	}
	return res
}

// cursorPos 通过 xdotool 读取鼠标位置，用于判断窗口该落在哪块屏幕。
func cursorPos() (int, int, bool) {
	out, err := exec.Command("xdotool", "getmouselocation", "--shell").Output()
	if err != nil {
		return 0, 0, false
	}
	var x, y int
	var gotX, gotY bool
	for _, line := range splitLines(string(out)) {
		if v, ok := cutPrefix(line, "X="); ok {
			x, _ = strconv.Atoi(v)
			gotX = true
		} else if v, ok := cutPrefix(line, "Y="); ok {
			y, _ = strconv.Atoi(v)
			gotY = true
		}
	}
	return x, y, gotX && gotY
}

func splitLines(s string) []string {
	var res []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			res = append(res, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		res = append(res, s[start:])
	}
	return res
}

func cutPrefix(s, prefix string) (string, bool) {
	if len(s) >= len(prefix) && s[:len(prefix)] == prefix {
		return s[len(prefix):], true
	}
	return "", false
}
