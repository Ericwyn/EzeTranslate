package ui

import (
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/Ericwyn/EzeTranslate/conf"
	"github.com/Ericwyn/EzeTranslate/ipc"
	"github.com/Ericwyn/EzeTranslate/log"
	"github.com/Ericwyn/EzeTranslate/ocr"
	"github.com/Ericwyn/EzeTranslate/strutils"
	"github.com/Ericwyn/EzeTranslate/trans"
	"github.com/Ericwyn/EzeTranslate/trans/baidu"
	"github.com/Ericwyn/EzeTranslate/trans/google"
	"github.com/Ericwyn/EzeTranslate/trans/openai"
	"github.com/Ericwyn/EzeTranslate/trans/youdao"
	"github.com/Ericwyn/EzeTranslate/ui/resource"
	"github.com/Ericwyn/GoTools/shell"
	"github.com/spf13/viper"

	"github.com/vivaui/viva/app"
	"github.com/vivaui/viva/ui"
)

// ezeApp 持有整个应用的运行期状态：Viva 多窗口句柄与共享的翻译面板。
// 它取代了原 Fyne 版里散落的一堆全局 window / entry 指针。
type ezeApp struct {
	application *app.Application

	homeWin  *app.Window
	miniWin  *app.Window
	setWin   *app.Window
	aboutWin *app.Window
	logWin   *app.Window

	// 翻译面板状态（home 与 mini 共享同一份结果数据，二者不会同时打开）
	trans *transState

	// switching 为 true 时窗口关闭不退出应用（用于完整<->迷你切换）
	switching bool

	// 启动参数对应的动作在首个窗口创建后执行；划词文本需在开窗抢焦点前捕获。
	startupSelection string
	startupOCR       bool

	// pendingShow 是后台线程（IPC/OCR）请求窗口线程执行的动作队列，
	// 通过 Invalidate 唤醒、在对应窗口的 OnWake 里取出执行。
	pendMu   sync.Mutex
	homeWake []func()
	miniWake []func()
}

var eze *ezeApp

// appIcon 是应用图标，openNewApp 时初始化一次。
var appIcon app.WindowIcon

// Tree 与控件标识。Viva 要求用稳定的身份树命名控件。
var (
	tree = ui.Tree{}

	homeFeat = tree.Feature("home")
	miniFeat = tree.Feature("mini")
	setFeat  = tree.Feature("set")
	logFeat  = tree.Feature("log")

	homeInputID  = tree.ID(homeFeat, "input")
	homeResultID = tree.ID(homeFeat, "result")
	miniResultID = tree.ID(miniFeat, "result")
	logResultID  = tree.ID(logFeat, "log")

	// 引擎选择 / 目标语言选择下拉各自需要独立 scope
	engineScope = tree.Feature("engineSel")
	langScope   = tree.Feature("langSel")

	// 输入优化勾选框标识
	ckAnnotationID = tree.ID(homeFeat, "ckAnno")
	ckSpaceID      = tree.ID(homeFeat, "ckSpace")
	ckReturnID     = tree.ID(homeFeat, "ckReturn")
	ckCamelID      = tree.ID(homeFeat, "ckCamel")

	// 按钮 / 菜单项标识
	homeTransBtnID = tree.ID(homeFeat, "transBtn")
	homeMiniBtnID  = tree.ID(homeFeat, "miniBtn")
	homeOcrBtnID   = tree.ID(homeFeat, "ocrBtn")
	homeSetBtnID   = tree.ID(homeFeat, "setBtn")
	homeLogBtnID   = tree.ID(homeFeat, "logBtn")
	homeAboutBtnID = tree.ID(homeFeat, "aboutBtn")

	miniTransBtnID = tree.ID(miniFeat, "transBtn")
	miniFullBtnID  = tree.ID(miniFeat, "fullBtn")
	miniOcrBtnID   = tree.ID(miniFeat, "ocrBtn")

	setSaveBtnID = tree.ID(setFeat, "saveBtn")

	logRefreshID = tree.ID(logFeat, "refresh")
	logClearID   = tree.ID(logFeat, "clear")

	// 关于窗按钮
	aboutFeat      = tree.Feature("about")
	aboutGithubID  = tree.ID(aboutFeat, "github")
	aboutIssueID   = tree.ID(aboutFeat, "issue")
	aboutReleaseID = tree.ID(aboutFeat, "release")
)

// StartApp 是 ui 层入口，签名与原 Fyne 版保持一致。
func StartApp(xclip bool, ocrFlag bool) {
	shell.Debug(true)

	conf.InitConfig()

	if trySendMessage(ipc.IpcMessagePing) {
		// 已有翻译进程在运行：把本次启动参数转发过去，然后退出。
		sendSocketMessage(xclip, ocrFlag, false)
		return
	}

	// 作为主进程：先读取划词，避免首个窗口获得焦点后 PRIMARY selection 丢失。
	startupSelection := ""
	if xclip {
		startupSelection = strutils.FormatInputBoxText(trans.GetSelection())
	}

	// 主进程不通过 IPC 给自己补发启动参数；开窗后直接执行缓存的启动动作。
	startUnixSocketServer()
	openNewApp(startupSelection, ocrFlag && !xclip)
}

func sendSocketMessage(xclip bool, ocrFlag bool, sleep bool) {
	if xclip {
		if sleep {
			time.Sleep(time.Millisecond * 500)
		}
		trySendMessage(ipc.IpcMessageNewSelection)
	} else if ocrFlag {
		if sleep {
			time.Sleep(time.Millisecond * 500)
		}
		trySendMessage(ipc.IpcMessageOcr)
	}
}

// openNewApp 启动 Viva 应用并打开首个窗口（完整 / 迷你）。
func openNewApp(startupSelection string, startupOCR bool) {
	eze = &ezeApp{
		trans:            newTransState(),
		startupSelection: startupSelection,
		startupOCR:       startupOCR,
	}
	appIcon = resource.WindowIcon()

	err := app.RunApplication(func(application *app.Application) error {
		eze.application = application

		if viper.GetBool(conf.ConfigKeyMiniMode) {
			if err := eze.openMiniWindow(); err != nil {
				return err
			}
		} else if err := eze.openHomeWindow(); err != nil {
			return err
		}
		eze.runStartupAction()
		return nil
	})
	if err != nil {
		log.E("Viva 应用运行失败: ", err)
	}
}

func (a *ezeApp) runStartupAction() {
	if a.startupSelection != "" {
		selection := a.startupSelection
		a.startupSelection = ""
		if a.homeWin != nil && !a.homeWin.Closed() {
			a.postHome(func() {
				a.trans.selectNow = selection
				a.trans.input.SetText(selection)
				a.startTransFormattedText(selection)
			})
		} else {
			a.postMini(func() {
				a.trans.selectNow = selection
				a.startTransFormattedText(selection)
			})
		}
		return
	}
	if a.startupOCR {
		a.startupOCR = false
		go a.runOcr(false)
	}
}

// postHome / postMini 把一个动作投递到对应窗口线程执行并请求重绘。
func (a *ezeApp) postHome(fn func()) {
	a.pendMu.Lock()
	a.homeWake = append(a.homeWake, fn)
	a.pendMu.Unlock()
	a.trans.homeInv.Invalidate()
}

func (a *ezeApp) postMini(fn func()) {
	a.pendMu.Lock()
	a.miniWake = append(a.miniWake, fn)
	a.pendMu.Unlock()
	a.trans.miniInv.Invalidate()
}

// drainHome / drainMini 在窗口线程的 OnWake 里调用，执行投递来的动作。
func (a *ezeApp) drainHome() {
	a.trans.applyResult()
	a.pendMu.Lock()
	fns := a.homeWake
	a.homeWake = nil
	a.pendMu.Unlock()
	for _, fn := range fns {
		fn()
	}
}

func (a *ezeApp) drainMini() {
	a.trans.applyResult()
	a.pendMu.Lock()
	fns := a.miniWake
	a.miniWake = nil
	a.pendMu.Unlock()
	for _, fn := range fns {
		fn()
	}
}

// ---------------- 翻译调度 ----------------

// startTrans 根据当前打开的窗口获取待翻译文本并异步翻译。
// 它可能被按钮点击（窗口线程）或 IPC 回调（后台线程）触发，内部只读写
// 受保护的 transState，真正的 UI 刷新交给 Invalidate + OnWake。
func (a *ezeApp) startTrans() {
	var text string
	if a.homeWin != nil && !a.homeWin.Closed() {
		text = a.trans.input.Text()
	} else {
		// 迷你模式下没有输入框，直接取划词
		text = trans.GetSelection()
	}
	a.startTransText(text)
}

// startTransText 使用已捕获的文本翻译。划词/OCR 回调可能已经让窗口取得焦点，
// 此时不能再从系统选区读取一次，否则会读到空内容或另一段文本。
func (a *ezeApp) startTransText(text string) {
	a.startTransFormattedText(strutils.FormatInputBoxText(text))
}

func (a *ezeApp) startTransFormattedText(formatText string) {
	if strings.Trim(formatText, " ") == "" {
		a.trans.setNote("请输入需要翻译的内容")
		a.invalidateActive()
		return
	}

	a.trans.setLoading()
	a.invalidateActive()

	handleTransResult := func(result string, note string) {
		log.D("翻译结果: " + result)
		a.trans.setResult(result, note)
		a.invalidateActive()
	}

	toLang := strutils.Lang(conf.ToLang)
	switch viper.GetString(conf.ConfigKeyTranslateSelect) {
	case "google":
		go google.Translate(formatText, toLang, handleTransResult)
	case "baidu":
		go baidu.Translate(formatText, toLang, handleTransResult)
	case "youdao":
		go youdao.Translate(formatText, toLang, handleTransResult)
	case "openai":
		go openai.Translate(formatText, toLang, handleTransResult)
	default:
		go openai.Translate(formatText, toLang, handleTransResult)
	}
}

// invalidateActive 请求当前打开的翻译窗口重绘。
func (a *ezeApp) invalidateActive() {
	if a.homeWin != nil && !a.homeWin.Closed() {
		a.trans.homeInv.Invalidate()
	}
	if a.miniWin != nil && !a.miniWin.Closed() {
		a.trans.miniInv.Invalidate()
	}
}

// ---------------- IPC ----------------

func trySendMessage(message ipc.IpcMessage) bool {
	if runtime.GOOS != "linux" {
		log.D("not linux, don't send socket msg")
		return false
	}
	if err := ipc.SendMessage(message); err == nil {
		log.D("已发送给其他翻译进程 : " + string(message))
		return true
	}
	log.D("IPC 消息发送失败")
	return false
}

func startUnixSocketServer() {
	if runtime.GOOS != "linux" {
		log.D("not linux, don't start socket server")
		return
	}
	go ipc.StartUnixSocketListener(func(message ipc.IpcMessage) {
		log.D("接收到 IPC 消息 : " + string(message))
		switch message {
		case ipc.IpcMessagePing:
		case ipc.IpcMessageOcr:
			eze.runOcr(false)
		case ipc.IpcMessageOcrAndTrans:
			eze.runOcr(true)
		case ipc.IpcMessageNewSelection:
			eze.onNewSelection()
		}
	})
}

// onNewSelection 处理划词翻译：读取划词填入输入框，必要时触发翻译。
func (a *ezeApp) onNewSelection() {
	a.onNewSelectionText(trans.GetSelection())
}

func (a *ezeApp) onNewSelectionText(selectText string) {
	selectText = strutils.FormatInputBoxText(selectText)
	log.D("获取的划词: " + selectText)
	a.trans.selectNow = selectText

	a.focusActive()

	// 若划词与当前输入一致且已有翻译结果，则不重复翻译
	if a.homeWin != nil && !a.homeWin.Closed() {
		if strings.Trim(a.trans.input.Text(), " ") == strings.Trim(selectText, " ") &&
			strings.Trim(a.trans.result.Text(), " ") != "" {
			log.D("划词与当前输入一致，不重复翻译")
			return
		}
		a.postHome(func() {
			a.trans.input.SetText(selectText)
			a.startTrans()
		})
		return
	}
	// 迷你模式：直接翻译划词
	a.startTransFormattedText(selectText)
}

// runOcr 处理 OCR：隐藏窗口 -> 截图识别 -> 恢复窗口 -> 填入输入框（可选翻译）。
func (a *ezeApp) runOcr(alsoTranslate bool) {
	post := a.activePost()

	// 隐藏当前窗口，避免遮挡截图。
	// Viva 的 Window.Hide/Show 会自行投递到窗口事件循环（源码注释
	// "may be called anywhere"），可从任意 goroutine 直接调用；
	// 这里必须直接调而不是走 post(Invalidate+OnWake)，否则窗口空闲时
	// 隐藏会被推迟到截图之后，导致截图里仍有主窗口。
	a.hideActive()
	// 留出时间让窗口从屏幕移除后再截图
	time.Sleep(400 * time.Millisecond)

	ocr.RunOcrWithHook(
		// 截图一完成就恢复窗口，不等后续的 mogrify + tesseract 识别，
		// 把窗口隐藏时间压到只覆盖「框选截图」这一小段。
		func() {
			a.showActive()
		},
		func(ocrRes string, success bool) {
			// 识别结果涉及 model 状态，放到窗口线程（post）里处理以保证 Paint 线程安全。
			post(func() {
				a.focusActive()
				if !success {
					log.E("ocr 识别失败")
					a.trans.setNote("OCR 未完成：请检查桌面截图权限、Tesseract 语言包，并查看运行日志")
					a.invalidateActive()
					return
				}
				ocrRes = strutils.FormatInputBoxText(ocrRes)
				a.trans.setNote("")
				log.D("获取的 OCR: " + ocrRes)
				a.trans.selectNow = ocrRes
				if a.homeWin != nil && !a.homeWin.Closed() {
					a.trans.input.SetText(ocrRes)
				}
				if alsoTranslate {
					a.startTransFormattedText(ocrRes)
				}
			})
		},
	)
}

// activePost 返回当前活动窗口对应的投递函数。
func (a *ezeApp) activePost() func(func()) {
	if a.miniWin != nil && !a.miniWin.Closed() {
		return a.postMini
	}
	return a.postHome
}

func (a *ezeApp) hideActive() {
	if a.homeWin != nil && !a.homeWin.Closed() {
		a.homeWin.Hide()
	}
	if a.miniWin != nil && !a.miniWin.Closed() {
		a.miniWin.Hide()
	}
}

func (a *ezeApp) showActive() {
	if a.homeWin != nil && !a.homeWin.Closed() {
		a.homeWin.Show()
	}
	if a.miniWin != nil && !a.miniWin.Closed() {
		a.miniWin.Show()
	}
}

func (a *ezeApp) focusActive() {
	// Viva 没有显式 RequestFocus，Show 已把窗口带到前台
	a.showActive()
}
