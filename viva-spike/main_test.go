package main

import (
	"strings"
	"testing"

	"github.com/vivaui/viva/app"
	"github.com/vivaui/viva/ui"
)

func labelText(snap app.Snapshot) string {
	var b strings.Builder
	for _, l := range snap.Labels {
		b.WriteString(l.Value)
		b.WriteString("\n")
	}
	return b.String()
}

// 无窗口驱动真实应用：验证布局可绘制、输入框可接收文本（含中文）、中文可进入渲染场景。
func TestRenderAndField(t *testing.T) {
	m := &model{inv: &app.Invalidator{}, pend: &pending{}}
	var canvas ui.Canvas
	driver, err := app.NewDriver(newConfig(m), &canvas)
	if err != nil {
		t.Fatalf("NewDriver: %v", err)
	}
	if err := driver.Resize(460, 560); err != nil {
		t.Fatalf("Resize: %v", err)
	}

	// 1) 直接塞中文到输入框
	m.input.SetText("你好，世界")
	if got := m.input.Text(); got != "你好，世界" {
		t.Fatalf("field text = %q, want 你好，世界", got)
	}

	// 2) 模拟后台翻译结果（中文显示）
	m.result = "Hello world 这是一段中英文混排的结果"
	m.note = "zh->en"

	_ = driver.Resize(460, 560)
	labels := labelText(driver.Snapshot())

	if !strings.Contains(labels, "你好") {
		t.Errorf("snapshot 未包含输入中文，labels=\n%s", labels)
	}
	if !strings.Contains(labels, "中英文混排") {
		t.Errorf("snapshot 未包含中文结果，labels=\n%s", labels)
	}
	t.Logf("painted labels:\n%s", labels)
}

// 验证 driver.Type（文本输入事件路径，IME 在真实窗口下的 headless 近似）能进入字段。
func TestTypeIntoField(t *testing.T) {
	m := &model{inv: &app.Invalidator{}, pend: &pending{}}
	var canvas ui.Canvas
	driver, err := app.NewDriver(newConfig(m), &canvas)
	if err != nil {
		t.Fatalf("NewDriver: %v", err)
	}
	_ = driver.Resize(460, 560)

	// 聚焦输入框：找到它的命中区域点一下
	snap := driver.Snapshot()
	var hit *ui.Hit
	for i := range snap.Hits {
		if snap.Hits[i].ID == inputID {
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
	if err := driver.Type("hello"); err != nil {
		t.Fatalf("Type: %v", err)
	}
	if got := m.input.Text(); !strings.Contains(got, "hello") {
		t.Errorf("Type 后字段内容=%q，期望含 hello", got)
	}
	t.Logf("字段内容=%q, focused=%v", m.input.Text(), driver.Snapshot().Focused)
}
