package ui

import (
	"strings"
	"testing"
	"time"
	"unicode/utf8"

	"github.com/Ericwyn/EzeTranslate/conf"
	"github.com/Ericwyn/EzeTranslate/strutils"
	"github.com/spf13/viper"
	"github.com/vivaui/viva/app"
	"github.com/vivaui/viva/text"
	"github.com/vivaui/viva/ui"
)

func interactionDriver(t *testing.T, a *ezeApp, miniMode bool) *app.Driver {
	t.Helper()
	// flex.View 的缓存属于单个窗口，不让上一个测试的控件状态影响此窗口。
	home = &homeView{}
	mini = &miniView{}
	cfg := a.homeConfig(ui.LightTheme())
	if miniMode {
		cfg = a.miniConfig(ui.LightTheme())
	}
	driver, err := app.NewDriver(cfg, new(ui.Canvas))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(driver.Close)
	if err := driver.Resize(cfg.Width, cfg.Height); err != nil {
		t.Fatal(err)
	}
	return driver
}

func interactionHit(t *testing.T, driver *app.Driver, id ui.ID) ui.Hit {
	t.Helper()
	for _, hit := range driver.Snapshot().Hits {
		if hit.ID == id {
			return hit
		}
	}
	t.Fatalf("未找到控件 %v", id)
	return ui.Hit{}
}

func focusInteractionHit(t *testing.T, driver *app.Driver, id ui.ID) {
	t.Helper()
	hit := interactionHit(t, driver, id)
	if err := driver.Click(hit.Bounds.X+hit.Bounds.Width/2, hit.Bounds.Y+hit.Bounds.Height/2); err != nil {
		t.Fatal(err)
	}
}

func TestHomeEnterThroughTextInput(t *testing.T) {
	// 目标与源语言相同，真实 startTrans/异步回调可运行但不访问网络或 xclip。
	oldLang := conf.ToLang
	conf.ToLang = string(strutils.Chinese)
	t.Cleanup(func() { conf.ToLang = oldLang })
	for key, value := range map[string]interface{}{
		conf.ConfigKeyTranslateSelect:      "openai",
		conf.ConfigKeyFormatAnnotation:     false,
		conf.ConfigKeyFormatSpace:          false,
		conf.ConfigKeyFormatCarriageReturn: false,
		conf.ConfigKeyFormatCamelCase:      false,
	} {
		previous := viper.Get(key)
		viper.Set(key, value)
		t.Cleanup(func() { viper.Set(key, previous) })
	}

	for _, tc := range []struct {
		name string
		mods ui.Modifiers
	}{
		{"Ctrl+Enter", ui.ModifierControl},
		{"Meta+Enter", ui.ModifierMeta},
	} {
		t.Run(tc.name, func(t *testing.T) {
			a := newTestApp()
			a.homeWin = &app.Window{} // 仅标记完整模式，不打开系统窗口。
			a.trans.input.SetText("你好世界")
			driver := interactionDriver(t, a, false)
			focusInteractionHit(t, driver, homeInputID)
			a.trans.input.editor.MoveTo(a.trans.input.editor.RuneCount(), false)
			if err := driver.Key(ui.KeyEnter, tc.mods); err != nil {
				t.Fatal(err)
			}
			if got := a.trans.input.Text(); got != "你好世界" {
				t.Fatalf("翻译快捷键不应插入换行，得到 %q", got)
			}
			deadline := time.Now().Add(2 * time.Second)
			for a.trans.result.Text() != "你好世界" && time.Now().Before(deadline) {
				if err := driver.Pump(); err != nil {
					t.Fatal(err)
				}
				time.Sleep(time.Millisecond)
			}
			if got := a.trans.result.Text(); got != "你好世界" {
				t.Fatalf("快捷键没有触发真实翻译流程，结果=%q", got)
			}
			for _, mods := range []ui.Modifiers{0, ui.ModifierShift} {
				before := a.trans.input.Text()
				if err := driver.Key(ui.KeyEnter, mods); err != nil {
					t.Fatal(err)
				}
				if got := a.trans.input.Text(); got != before+"\n" {
					t.Fatalf("普通/Shift+Enter 应换行，得到 %q", got)
				}
			}
		})
	}
}

func TestHomeEnterWhileLoading(t *testing.T) {
	a := newTestApp()
	a.trans.input.SetText("你好")
	a.trans.setLoading()
	driver := interactionDriver(t, a, false)
	focusInteractionHit(t, driver, homeInputID)
	if err := driver.Key(ui.KeyEnter, ui.ModifierControl); err != nil {
		t.Fatal(err)
	}
	if a.trans.getNote() != "" || a.trans.input.Text() != "你好" || !a.trans.isLoading() {
		t.Fatal("翻译中应吞掉快捷键，既不重复请求也不插入换行")
	}
}

func copyResult(t *testing.T, driver *app.Driver, box *editBox) string {
	t.Helper()
	handled, err := driver.ShortcutChord(ui.ModifierControl, 'c')
	if err != nil || !handled {
		t.Fatalf("Ctrl+C: handled=%v err=%v", handled, err)
	}
	selection := box.editor.Selection()
	value := box.Text()
	want := value[byteOffsetOfRune(value, selection.Start):byteOffsetOfRune(value, selection.End)]
	if got := driver.Clipboard(); got != want {
		t.Fatalf("剪贴板=%q，期望选中内容 %q", got, want)
	}
	return want
}

func TestResultSelectionAndCopy(t *testing.T) {
	for _, miniMode := range []bool{false, true} {
		name := "home"
		if miniMode {
			name = "mini"
		}
		t.Run(name, func(t *testing.T) {
			a := newTestApp()
			value := "你好，世界\nHello, world 👋"
			a.trans.result.SetText(value)
			driver := interactionDriver(t, a, miniMode)
			box := a.trans.result
			hit := interactionHit(t, driver, box.id)
			if !hit.ReadOnly {
				t.Fatal("翻译结果必须保持只读")
			}

			// 从中文首行拖到第二行的中部，验证中文/换行/部分文本的复制。
			x1, y1 := hit.TextBounds.X, hit.TextBounds.Y+2
			x2, y2 := x1+35, y1+ui.LightTheme().Metrics.FontSizeM+7
			if err := driver.Press(x1, y1, ui.PointerButtonPrimary); err != nil {
				t.Fatal(err)
			}
			if err := driver.Move(x2, y2); err != nil {
				t.Fatal(err)
			}
			if err := driver.Release(x2, y2, ui.PointerButtonPrimary); err != nil {
				t.Fatal(err)
			}
			partial := copyResult(t, driver, box)
			if !strings.HasPrefix(partial, "你好，世界\n") || partial == value {
				t.Fatalf("跨行拖选应复制部分结果，得到 %q", partial)
			}
			// 反向拖选要保留正确锚点，复制相同文字。
			if err := driver.Press(x2, y2, ui.PointerButtonPrimary); err != nil {
				t.Fatal(err)
			}
			if err := driver.Move(x1, y1); err != nil {
				t.Fatal(err)
			}
			if err := driver.Release(x1, y1, ui.PointerButtonPrimary); err != nil {
				t.Fatal(err)
			}
			if got := copyResult(t, driver, box); got != partial {
				t.Fatalf("反向拖选=%q，期望 %q", got, partial)
			}
			if box.selecting {
				t.Fatal("鼠标释放后应停止拖选")
			}

			for _, mods := range []ui.Modifiers{ui.ModifierControl, ui.ModifierMeta} {
				if handled, err := driver.ShortcutChord(mods, 'a'); err != nil || !handled {
					t.Fatalf("全选: handled=%v err=%v", handled, err)
				}
				if got := copyResult(t, driver, box); got != value {
					t.Fatalf("全选复制=%q，期望 %q", got, value)
				}
			}
			// 原生文本服务不能编辑结果，剪切/粘贴也不能改变内容。
			if driver.Snapshot().Editor != nil {
				t.Fatal("只读结果不应挂载可编辑的原生输入")
			}
			_ = driver.Type("不应写入")
			_ = driver.Key(ui.KeyBackspace, 0)
			_ = driver.Key(ui.KeyDelete, 0)
			_, _ = driver.ShortcutChord(ui.ModifierControl, 'x')
			_, _ = driver.ShortcutChord(ui.ModifierControl, 'v')
			if box.Text() != value {
				t.Fatal("只读结果被编辑")
			}
		})
	}
}

func TestResultSelectionAfterScrollAndWake(t *testing.T) {
	a := newTestApp()
	driver := interactionDriver(t, a, false)
	value := strings.Repeat("你好，世界 hello\n", 40)
	// 模拟后台回调，只有发布结果，不应直接改 editor。
	a.trans.setResult(value, "en->zh")
	if a.trans.result.Text() != "" {
		t.Fatal("后台结果不应直接修改 editor")
	}
	a.trans.homeInv.Invalidate()
	if err := driver.Pump(); err != nil {
		t.Fatal(err)
	}
	hit := interactionHit(t, driver, homeResultID)
	if err := driver.Scroll(ui.ScrollEvent{X: hit.Bounds.X + 20, Y: hit.Bounds.Y + 20, DeltaY: 140}); err != nil {
		t.Fatal(err)
	}
	x, y := hit.TextBounds.X, hit.TextBounds.Y+2
	if err := driver.Press(x, y, ui.PointerButtonPrimary); err != nil {
		t.Fatal(err)
	}
	start := a.trans.result.editor.Caret()
	if start <= 0 {
		t.Fatal("滚动后点击顶部应命中后面的行")
	}
	if err := driver.Move(x+35, y); err != nil {
		t.Fatal(err)
	}
	if err := driver.Release(x+35, y, ui.PointerButtonPrimary); err != nil {
		t.Fatal(err)
	}
	if got := copyResult(t, driver, a.trans.result); got == "" || !utf8.ValidString(got) {
		t.Fatalf("滚动后应能选中并复制有效 UTF-8，得到 %q", got)
	}
	a.trans.setResult("新结果", "en->zh")
	a.trans.homeInv.Invalidate()
	if err := driver.Pump(); err != nil {
		t.Fatal(err)
	}
	if sel := a.trans.result.editor.Selection(); sel != (text.Range{}) {
		t.Fatalf("新翻译结果应重置旧选区，得到 %+v", sel)
	}
	if got := a.trans.result.Text(); got != "新结果" {
		t.Fatalf("OnWake 未更新结果: %q", got)
	}
}
