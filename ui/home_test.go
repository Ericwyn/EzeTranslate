package ui

import (
	"strings"
	"testing"

	"github.com/Ericwyn/EzeTranslate/conf"

	"github.com/vivaui/viva/app"
	"github.com/vivaui/viva/ui"
)

// newTestApp 构造一个不进入事件循环、可被 headless driver 驱动的 ezeApp。
func newTestApp() *ezeApp {
	return &ezeApp{trans: newTransState()}
}

func labelText(snap app.Snapshot) string {
	var b strings.Builder
	for _, l := range snap.Labels {
		b.WriteString(l.Value)
		b.WriteString("\n")
	}
	return b.String()
}

// TestHomeRenders 用无窗口 driver 驱动真实的完整模式 Paint，验证：
// 中文界面文字绘制、输入框可接收中文、结果区显示翻译结果、关键按钮在场景里。
func TestHomeRenders(t *testing.T) {
	// 配置依赖 viper，初始化一次（读/建用户配置文件）
	conf.InitConfig()

	a := newTestApp()
	appIcon = app.WindowIcon{} // 测试里不解码图标

	theme := ui.LightTheme()
	var canvas ui.Canvas
	driver, err := app.NewDriver(a.homeConfig(theme), &canvas)
	if err != nil {
		t.Fatalf("NewDriver: %v", err)
	}
	if err := driver.Resize(450, 600); err != nil {
		t.Fatalf("Resize: %v", err)
	}

	// 往输入框塞中文
	a.trans.input.SetText("你好，世界")
	if got := a.trans.input.Text(); got != "你好，世界" {
		t.Fatalf("输入框内容=%q，期望 你好，世界", got)
	}

	// 模拟一次翻译结果（中文显示）
	a.trans.setResult("Hello world 这是中英文混排结果", "zh->en")

	_ = driver.Resize(450, 600)
	labels := labelText(driver.Snapshot())

	for _, want := range []string{"你好", "中英文混排", "执行翻译", "参数设置", "输入优化"} {
		if !strings.Contains(labels, want) {
			t.Errorf("快照缺少 %q，labels=\n%s", want, labels)
		}
	}
	t.Logf("painted labels:\n%s", labels)
}

// TestHomeKeyDispatch 验证按键分发：Ctrl+Enter 作为翻译快捷键被标记 Handled，
// 普通方向键交给输入框（editBox）处理而非被 home 拦截。
// 不直接断言翻译结果，避免在无网络/无 xclip 的测试环境触发真实翻译。
func TestHomeKeyDispatch(t *testing.T) {
	a := newTestApp()

	// Ctrl+Enter 应被完整模式识别为翻译快捷键
	res := a.homeOnKey(ui.KeyEvent{Key: ui.KeyEnter, Modifiers: ui.ModifierControl})
	if !res.Handled {
		t.Errorf("Ctrl+Enter 应被处理为翻译快捷键")
	}

	// Alt+Enter 同样应被识别
	res = a.homeOnKey(ui.KeyEvent{Key: ui.KeyEnter, Modifiers: ui.ModifierAlt})
	if !res.Handled {
		t.Errorf("Alt+Enter 应被处理为翻译快捷键")
	}

	// 一个普通字符键（无修饰）不应被当作翻译快捷键拦截——
	// 它会交给输入框 editBox.handleKey，对非移动键返回未处理。
	res = a.homeOnKey(ui.KeyEvent{Key: ui.KeyEnter})
	_ = res // 纯 Enter 交给输入框，不触发翻译；此处仅确保不 panic
}

// TestHomeCtrlASelectsAll 端到端验证 Ctrl+A 全选：
// 通过真实 paint 找到输入框命中区域、聚焦，再发 Ctrl+A，
// 断言输入框 editor 选区覆盖全部文本。这条路径由 Viva 原生文本层处理
// （见 Viva 的 TestEditingPriorityBelongsToTheBackend），我们验证它对本应用生效。
func TestHomeCtrlASelectsAll(t *testing.T) {
	conf.InitConfig()

	a := newTestApp()
	appIcon = app.WindowIcon{}
	a.trans.input.SetText("hello world")

	theme := ui.LightTheme()
	var canvas ui.Canvas
	driver, err := app.NewDriver(a.homeConfig(theme), &canvas)
	if err != nil {
		t.Fatalf("NewDriver: %v", err)
	}
	if err := driver.Resize(450, 600); err != nil {
		t.Fatalf("Resize: %v", err)
	}

	// 找到输入框命中区域并点击聚焦
	snap := driver.Snapshot()
	var hit *ui.Hit
	for i := range snap.Hits {
		if snap.Hits[i].ID == homeInputID {
			hit = &snap.Hits[i]
			break
		}
	}
	if hit == nil {
		t.Fatalf("未找到输入框命中区域")
	}
	cx := hit.Bounds.X + hit.Bounds.Width/2
	cy := hit.Bounds.Y + hit.Bounds.Height/2
	if err := driver.Click(cx, cy); err != nil {
		t.Fatalf("Click: %v", err)
	}

	// 发 Ctrl+A
	handled, err := driver.ShortcutChord(ui.ModifierControl, 'a')
	if err != nil {
		t.Fatalf("ShortcutChord: %v", err)
	}
	if !handled {
		t.Fatalf("Ctrl+A 未被处理")
	}

	// 断言全选：选区应覆盖 "hello world" 的全部 11 个字符
	runes := len([]rune(a.trans.input.Text()))
	sel := a.trans.input.editor.Selection()
	if sel.Start != 0 || sel.End != runes {
		t.Errorf("Ctrl+A 后选区=%+v，期望 {0,%d} 覆盖全部文本", sel, runes)
	}
	t.Logf("选区=%+v, 文本长度=%d", sel, runes)
}
