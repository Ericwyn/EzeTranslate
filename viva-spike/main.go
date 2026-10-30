package main

import (
	"log"
	"sync"

	"github.com/Ericwyn/EzeTranslate/conf"
	"github.com/Ericwyn/EzeTranslate/trans/openai"

	"github.com/vivaui/viva/app"
	"github.com/vivaui/viva/flex"
	"github.com/vivaui/viva/ui"
)

// ---- 控件标识 ----
var (
	tree      = ui.Tree{}
	home      = tree.Feature("home")
	inputID   = tree.ID(home, "input")
	transBtn  = tree.ID(home, "trans")
	clearBtn  = tree.ID(home, "clear")
	resultID  = tree.ID(home, "resultScroll")
)

// pending 是后台 goroutine 与 UI 线程之间的交接区。
// goroutine 写结果、置 done，再 Invalidate；OnWake 在 UI 线程把结果拷进 model。
type pending struct {
	mu     sync.Mutex
	result string
	note   string
	done   bool
}

type model struct {
	input   flex.Field // 输入框（测试中文输入/IME）
	view    flex.View
	inv     *app.Invalidator
	pend    *pending
	loading bool
	result  string // 翻译结果（测试中文显示）
	note    string
}

func newConfig(m *model) app.Config {
	theme := ui.LightTheme()

	config := app.Config{
		Tree:        &tree,
		Title:       "EzeTranslate · Viva 验证",
		Width:       460,
		Height:      560,
		Invalidator: m.inv,
		TextInput:   flex.Fields(&m.input),
	}

	// OnWake：UI 线程上把后台结果拷进 model，再重绘
	config.OnWake = func() {
		m.pend.mu.Lock()
		if m.pend.done {
			m.result = m.pend.result
			m.note = m.pend.note
			m.loading = false
			m.pend.done = false
		}
		m.pend.mu.Unlock()
	}

	config.Paint = func(canvas *ui.Canvas, size ui.Size) {
		bounds := ui.Rect{Width: size.Width, Height: size.Height}
		canvas.Fill(bounds, theme.Background)

		m.view.Draw(canvas, theme, bounds, func(f *flex.Frame) {
			f.Column().Fill().Pad(20).Gap(12).Children(func() {
				f.Text("EzeTranslate · Viva UI 验证").FontSize(20).Bold()
				f.Paragraph("输入中文或英文，点“翻译”，复用现有 trans/openai 引擎。用于验证：中文输入(IME)、中文显示、业务层复用。", 3).Muted()

				f.Text("输入文本").Muted()
				f.Field(&m.input, inputID, "输入", "在这里输入中文试试……").FillWidth()

				f.Row().Gap(8).Children(func() {
					label := "翻译"
					if m.loading {
						label = "翻译中…"
					}
					if f.Button(transBtn, label).Clicked() && !m.loading {
						m.startTranslate()
					}
					if f.Button(clearBtn, "清空").Outline().Clicked() {
						m.input.SetText("")
						m.result = ""
						m.note = ""
					}
				})

				f.Divider()

				f.Text("翻译结果").Muted()
				if m.note != "" {
					f.Text(m.note).FontSize(12).Muted()
				}
				f.ScrollArea(resultID).Grow(1).Children(func() {
					if m.result == "" {
						f.Paragraph("（结果会显示在这里）", 2).Muted()
					} else {
						f.Wrap(m.result)
					}
				})
			})
		})
	}
	return config
}

// startTranslate 在后台 goroutine 调用现有 openai.Translate
func (m *model) startTranslate() {
	text := m.input.Text()
	if text == "" {
		return
	}
	m.loading = true
	go func() {
		openai.Translate(text, "", func(result string, note string) {
			m.pend.mu.Lock()
			m.pend.result = result
			m.pend.note = note
			m.pend.done = true
			m.pend.mu.Unlock()
			m.inv.Invalidate()
		})
	}()
}

func main() {
	// 复用现有配置加载（读取 ~/.config/EzeTranslate/config.yaml 里的 openai key/url/model）
	conf.InitConfig()

	m := &model{
		inv:  &app.Invalidator{},
		pend: &pending{},
	}
	var canvas ui.Canvas
	if err := app.RunCanvas(newConfig(m), &canvas); err != nil {
		log.Fatal(err)
	}
}
