package ui

import (
	"testing"

	"github.com/Ericwyn/EzeTranslate/conf"

	"github.com/vivaui/viva/app"
	"github.com/vivaui/viva/text"
	"github.com/vivaui/viva/ui"
)

// TestSetFieldSelectAll 验证设置窗单行字段经 Ctrl+A（OnShortcut 路径）能全选。
func TestSetFieldSelectAll(t *testing.T) {
	conf.InitConfig()
	a := newTestApp()
	appIcon = app.WindowIcon{}

	// 手动构造 setState（不开真实窗口）
	set = &setState{inv: &app.Invalidator{}}
	set.fields = []*setFieldDef{
		{newSingleLineEditBox(setOpenaiKeyID), "OpenAI Key", "key", conf.ConfigKeyOpenAiKey},
	}
	set.fields[0].box.SetText("secret-value-123")

	theme := ui.LightTheme()
	cfg := app.Config{
		Tree: &tree, Width: 520, Height: 600,
		ClearColor: clearColor(theme),
		TextInput: func(ctx app.TextInputContext) *app.TextInput {
			var holder app.TextInput
			for _, fd := range set.fields {
				if in := fd.box.textInput(ctx, &holder); in != nil {
					return in
				}
			}
			return nil
		},
		OnShortcut: a.setOnShortcut,
		Paint: func(canvas *ui.Canvas, size ui.Size) {
			a.paintSet(canvas, size, theme)
		},
	}
	var canvas ui.Canvas
	driver, err := app.NewDriver(cfg, &canvas)
	if err != nil {
		t.Fatalf("NewDriver: %v", err)
	}
	t.Cleanup(driver.Close)
	_ = driver.Resize(520, 600)

	// 聚焦该字段
	snap := driver.Snapshot()
	for i := range snap.Hits {
		if snap.Hits[i].ID == setOpenaiKeyID {
			h := snap.Hits[i]
			_ = driver.Click(h.Bounds.X+h.Bounds.Width/2, h.Bounds.Y+h.Bounds.Height/2)
			break
		}
	}
	// 折叠选区后发 Ctrl+A
	set.fields[0].box.editor.SetSelection(text.Range{Start: 0, End: 0})
	if _, err := driver.ShortcutChord(ui.ModifierControl, 'a'); err != nil {
		t.Fatalf("ShortcutChord: %v", err)
	}
	runes := len([]rune(set.fields[0].box.Text()))
	sel := set.fields[0].box.editor.Selection()
	if sel.Start != 0 || sel.End != runes {
		t.Errorf("Ctrl+A 后选区=%+v，期望 {0,%d}", sel, runes)
	}
	t.Logf("选区=%+v 文本长度=%d", sel, runes)
}
