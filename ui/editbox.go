package ui

import (
	"unicode/utf8"

	"github.com/vivaui/viva/app"
	"github.com/vivaui/viva/components"
	"github.com/vivaui/viva/flex"
	"github.com/vivaui/viva/text"
	"github.com/vivaui/viva/ui"
)

// editBox 封装一个多行可编辑文本框所需的全部状态。
// Viva 的原生文本服务只认识“一行”，多行的换行、按视觉行移动光标、
// 滚动都要靠 TextareaLayout 协助，这里把这套接线集中起来，供输入框复用。
type editBox struct {
	editor     *text.Editor
	layout     *components.TextareaLayout
	id         ui.ID
	readOnly   bool
	singleLine bool // true 时用单行 Input 渲染（设置窗各字段），而非多行 Textarea
	selecting  bool // 只读框是否正在鼠标拖选；可编辑框的选区由 Viva 管理
}

func newEditBox(id ui.ID, readOnly bool) *editBox {
	return &editBox{
		editor:   text.NewEditor(""),
		layout:   new(components.TextareaLayout),
		id:       id,
		readOnly: readOnly,
	}
}

// newSingleLineEditBox 创建一个单行可编辑框（设置窗各字段用），
// 它用 f.Input 渲染，持有公开 editor 以便支持 Ctrl+A 全选。
func newSingleLineEditBox(id ui.ID) *editBox {
	return &editBox{
		editor:     text.NewEditor(""),
		id:         id,
		singleLine: true,
	}
}

// paintSingle 画出单行输入框，返回 Box 以便设置尺寸。
func (e *editBox) paintSingle(f *flex.Frame, canvas *ui.Canvas, label, placeholder string) flex.Box {
	value := e.editor.String()
	focused := canvas != nil && canvas.Focused() == e.id
	snapshot := components.InputSnapshot{
		ID:           e.id,
		Label:        label,
		Value:        value,
		DisplayValue: value,
		Placeholder:  placeholder,
		Focused:      focused,
	}
	if focused {
		selection := e.editor.Selection()
		snapshot.SelectionStart = byteOffsetOfRune(value, selection.Start)
		snapshot.SelectionEnd = byteOffsetOfRune(value, selection.End)
		snapshot.Caret = byteOffsetOfRune(value, e.editor.Caret())
		snapshot.ShowCaret = selection.Start == selection.End
		snapshot.PlatformEditing = value != ""
		if marked, composing := e.editor.Composition(); composing {
			snapshot.CompositionStart = byteOffsetOfRune(value, marked.Start)
			snapshot.CompositionEnd = byteOffsetOfRune(value, marked.End)
		}
	}
	return f.Input(snapshot)
}

func (e *editBox) Text() string {
	return e.editor.String()
}

func (e *editBox) SetText(value string) {
	e.editor.SetSelection(text.Range{Start: 0, End: e.editor.RuneCount()})
	e.editor.Insert(value)
	if e.readOnly {
		e.editor.MoveTo(0, false)
		e.selecting = false
		*e.layout = components.TextareaLayout{}
	}
}

// SelectAll 选中全部文本（供 Ctrl+A 使用）。
func (e *editBox) selectAll() {
	e.editor.SetSelection(text.Range{Start: 0, End: e.editor.RuneCount()})
}

// paint 在 flex 布局里画出这个多行框，返回它的 Box 以便继续设置尺寸。
func (e *editBox) paint(f *flex.Frame, canvas *ui.Canvas, label, placeholder string, height float32) flex.Box {
	value := e.editor.String()
	focused := canvas != nil && canvas.Focused() == e.id

	snapshot := components.TextareaSnapshot{
		ID:             e.id,
		Label:          label,
		Placeholder:    placeholder,
		Value:          value,
		DisplayValue:   value,
		ReadOnly:       e.readOnly,
		Focused:        focused,
		HideEditorText: true,
		ShowCaret:      focused && !e.readOnly,
		Layout:         e.layout,
	}

	if focused {
		selection := e.editor.Selection()
		snapshot.SelectionStart = byteOffsetOfRune(value, selection.Start)
		snapshot.SelectionEnd = byteOffsetOfRune(value, selection.End)
		snapshot.Caret = byteOffsetOfRune(value, e.editor.Caret())
		snapshot.ShowCaret = !e.readOnly && selection.Start == selection.End
		if marked, composing := e.editor.Composition(); composing {
			snapshot.CompositionStart = byteOffsetOfRune(value, marked.Start)
			snapshot.CompositionEnd = byteOffsetOfRune(value, marked.End)
		}
	}

	return f.Textarea(snapshot).Height(height)
}

// textInput 给 Config.TextInput 用：把原生文本服务直接接到这个 editor。
func (e *editBox) textInput(ctx app.TextInputContext, holder *app.TextInput) *app.TextInput {
	if e.readOnly || ctx.ID != e.id || ctx.ID.IsZero() {
		return nil
	}
	holder.Editor = e.editor
	return holder
}

// handleKey 处理按视觉行移动光标（上下/行首尾/翻页）。
// 可编辑框的输入、退格、左右、IME 由原生 editor 完成；只读框额外处理左右选区移动。
func (e *editBox) handleKey(event ui.KeyEvent) ui.EventResult {
	if event.Target != e.id || e.layout == nil {
		return ui.EventResult{}
	}
	if e.readOnly && (event.Key == ui.KeyLeft || event.Key == ui.KeyRight) {
		extend := event.Modifiers&ui.ModifierShift != 0
		if event.Key == ui.KeyLeft {
			e.editor.MoveLeft(extend)
		} else {
			e.editor.MoveRight(extend)
		}
		return ui.EventResult{Handled: true, Repaint: true}
	}
	movement, extend := components.TextareaMovementForKey(event.Key, event.Modifiers)
	if movement == components.TextareaMoveNone {
		return ui.EventResult{}
	}
	value := e.editor.String()
	to := e.layout.Move(value, byteOffsetOfRune(value, e.editor.Caret()), movement)
	e.editor.MoveTo(utf8.RuneCountInString(value[:to]), extend)
	return ui.EventResult{Handled: true, Repaint: true}
}

// handlePointer 给只读结果框补上选区；Viva 不为 ReadOnly 命中项放置原生光标。
// OffsetAt 使用实际排版的字节偏移，转为 editor 的 rune 索引后可正确选择中文和换行。
func (e *editBox) handlePointer(event ui.PointerEvent) ui.EventResult {
	if !e.readOnly || event.Target != e.id || e.layout == nil {
		return ui.EventResult{}
	}
	extend := true
	switch event.Kind {
	case ui.PointerDown:
		if event.Button != ui.PointerButtonPrimary && event.Button != ui.PointerButtonNone {
			return ui.EventResult{}
		}
		extend = event.Modifiers&ui.ModifierShift != 0
	case ui.PointerMove:
		if !e.selecting || !event.Captured {
			return ui.EventResult{}
		}
	case ui.PointerUp:
		if !e.selecting || (event.Button != ui.PointerButtonPrimary && event.Button != ui.PointerButtonNone) {
			return ui.EventResult{}
		}
	default:
		return ui.EventResult{}
	}
	offset, ok := e.layout.OffsetAt(ui.Point{X: event.X, Y: event.Y})
	if !ok {
		return ui.EventResult{}
	}
	e.editor.MoveTo(utf8.RuneCountInString(e.Text()[:offset]), extend)
	e.selecting = event.Kind != ui.PointerUp
	return ui.EventResult{Handled: true, Repaint: true, Focus: e.id}
}

// handleShortcut 统一处理全选及只读框复制；输入框的复制仍交给原生文本服务。
func (e *editBox) handleShortcut(event ui.KeyEvent) ui.EventResult {
	if event.Target != e.id || event.Modifiers&(ui.ModifierControl|ui.ModifierMeta) == 0 {
		return ui.EventResult{}
	}
	switch event.Text {
	case "a", "A":
		e.selectAll()
		return ui.EventResult{Handled: true, Repaint: true}
	case "c", "C":
		if e.readOnly {
			value := e.Text()
			selection := e.editor.Selection()
			return ui.EventResult{
				Handled:       true,
				ClipboardText: value[byteOffsetOfRune(value, selection.Start):byteOffsetOfRune(value, selection.End)],
			}
		}
	}
	return ui.EventResult{}
}

// handleScroll 处理滚轮。
func (e *editBox) handleScroll(event ui.ScrollEvent) ui.EventResult {
	if e.layout == nil {
		return ui.EventResult{}
	}
	e.layout.ScrollBy(event.DeltaY)
	return ui.EventResult{Handled: true, Repaint: true}
}

// byteOffsetOfRune 返回 value 第 n 个 rune 的字节偏移。
func byteOffsetOfRune(value string, n int) int {
	for offset := range value {
		if n == 0 {
			return offset
		}
		n--
	}
	return len(value)
}
