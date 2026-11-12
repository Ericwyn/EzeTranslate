package ui

import (
	"sync"

	"github.com/vivaui/viva/app"
	"github.com/vivaui/viva/flex"
)

// transState 是完整模式与迷你模式共享的翻译数据。
// 后台 goroutine 与窗口线程都会访问它，所以用 mu 保护可变字段。
type transState struct {
	mu sync.Mutex

	input  *editBox // 完整模式的输入框（可编辑）
	result *editBox // 翻译结果框（只读，home/mini 共用一份文本）

	note    string // 结果说明，如 "zh->en"
	loading bool

	// 后台翻译只发布字符串，editor/排版/鼠标选区在 OnWake 中更新。
	pendingResult string
	resultPending bool

	selectNow string // 最近一次划词/OCR 文本，用于模式切换时回填输入框

	homeInv *app.Invalidator
	miniInv *app.Invalidator

	// 下拉开合状态需跨帧保留
	engineSel flex.SelectState
	langSel   flex.SelectState
}

func newTransState() *transState {
	return &transState{
		input:   newEditBox(homeInputID, false),
		result:  newEditBox(homeResultID, true),
		homeInv: &app.Invalidator{},
		miniInv: &app.Invalidator{},
	}
}

func (t *transState) setLoading() {
	t.mu.Lock()
	t.loading = true
	t.pendingResult = ""
	t.resultPending = true
	t.mu.Unlock()
}

func (t *transState) setResult(result, note string) {
	t.mu.Lock()
	t.loading = false
	t.note = note
	t.pendingResult = result
	t.resultPending = true
	t.mu.Unlock()
}

// applyResult 仅由窗口线程调用，避免翻译回调与鼠标拖选同时修改 editor。
func (t *transState) applyResult() {
	t.mu.Lock()
	result, pending := t.pendingResult, t.resultPending
	t.pendingResult = ""
	t.resultPending = false
	t.mu.Unlock()
	if pending {
		t.result.SetText(result)
	}
}

func (t *transState) setNote(note string) {
	t.mu.Lock()
	t.note = note
	t.mu.Unlock()
}

func (t *transState) isLoading() bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.loading
}

func (t *transState) getNote() string {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.note
}
