package ui

import (
	"github.com/Ericwyn/EzeTranslate/conf"
	"github.com/Ericwyn/EzeTranslate/log"
	"github.com/spf13/viper"

	"github.com/vivaui/viva/app"
	"github.com/vivaui/viva/flex"
	"github.com/vivaui/viva/ui"
)

// setField 把一个单行输入框与它对应的配置键绑定起来。
type setFieldDef struct {
	box   *editBox
	label string
	hint  string
	key   string
}

// setState 持有设置窗口的全部输入框状态。
type setState struct {
	view       flex.View
	inv        *app.Invalidator
	fields     []*setFieldDef
	configPath string // 打开窗口时取一次，避免 Paint 每帧调用 GetConfigFilePath（有副作用/日志）
}

var set *setState

// 设置窗各字段标识
var (
	setBaiduIdID      = tree.ID(setFeat, "baiduId")
	setBaiduSecretID  = tree.ID(setFeat, "baiduSecret")
	setYoudaoIdID     = tree.ID(setFeat, "youdaoId")
	setYoudaoSecretID = tree.ID(setFeat, "youdaoSecret")
	setGoogleUrlID    = tree.ID(setFeat, "googleUrl")
	setGoogleProxyID  = tree.ID(setFeat, "googleProxy")
	setOpenaiUrlID    = tree.ID(setFeat, "openaiUrl")
	setOpenaiKeyID    = tree.ID(setFeat, "openaiKey")
	setOpenaiModelID  = tree.ID(setFeat, "openaiModel")
	setScrollID       = tree.ID(setFeat, "scroll")
)

// openSetWindow 打开参数设置窗口；已打开则前置。
func (a *ezeApp) openSetWindow() {
	if a.setWin != nil && !a.setWin.Closed() {
		a.setWin.Show()
		return
	}

	// 用当前配置初始化各输入框（单行 editBox，支持 Ctrl+A 全选）
	set = &setState{inv: &app.Invalidator{}}
	set.fields = []*setFieldDef{
		{newSingleLineEditBox(setBaiduIdID), "百度 AppId", "请填写从百度处申请到的 AppId", conf.ConfigKeyBaiduTransAppId},
		{newSingleLineEditBox(setBaiduSecretID), "百度 AppSecret", "请填写从百度处申请到的 AppSecret", conf.ConfigKeyBaiduTransAppSecret},
		{newSingleLineEditBox(setYoudaoIdID), "有道 AppId", "请填写从有道处申请到的 AppId", conf.ConfigKeyYouDaoTransAppId},
		{newSingleLineEditBox(setYoudaoSecretID), "有道 AppSecret", "请填写从有道处申请到的 AppSecret", conf.ConfigKeyYouDaoTransAppSecret},
		{newSingleLineEditBox(setGoogleUrlID), "Google Url", "默认 https://translate.googleapis.com", conf.ConfigKeyGoogleTranslateUrl},
		{newSingleLineEditBox(setGoogleProxyID), "Http 代理", "如 http://127.0.0.1:1080", conf.ConfigKeyGoogleTranslateProxy},
		{newSingleLineEditBox(setOpenaiUrlID), "OpenAI Url", "默认 https://api.openai.com/v1/chat/completions", conf.ConfigKeyOpenAIApiUrl},
		{newSingleLineEditBox(setOpenaiKeyID), "OpenAI Key", "OpenAI 的接口 Key", conf.ConfigKeyOpenAiKey},
		{newSingleLineEditBox(setOpenaiModelID), "OpenAI Model", "OpenAI 的接口模型", conf.ConfigKeyOpenAiModel},
	}
	for _, fd := range set.fields {
		fd.box.SetText(viper.GetString(fd.key))
	}
	set.configPath = conf.GetConfigFilePath()

	theme := ui.LightTheme()
	config := app.Config{
		Tree:        &tree,
		Title:       "程序设置",
		Icon:        appIcon,
		Width:       520,
		Height:      600,
		Origin:      centeredOrigin(520, 600),
		ClearColor:  clearColor(theme),
		Invalidator: set.inv,
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
		OnScroll: func(e ui.ScrollEvent) ui.EventResult {
			return set.view.Wheel(e)
		},
	}
	config.Paint = func(canvas *ui.Canvas, size ui.Size) {
		a.paintSet(canvas, size, theme)
	}
	config.OnClosed = func() {
		a.setWin = nil
	}

	win, err := a.application.OpenWindow(config, new(ui.Canvas))
	if err != nil {
		log.E("打开设置窗口失败: ", err)
		return
	}
	a.setWin = win
}

// setOnShortcut 处理设置窗的 Ctrl+A：对聚焦的字段全选。
func (a *ezeApp) setOnShortcut(e ui.KeyEvent) ui.EventResult {
	if (e.Text == "a" || e.Text == "A") && (e.Modifiers&ui.ModifierControl != 0 || e.Modifiers&ui.ModifierMeta != 0) {
		for _, fd := range set.fields {
			if e.Target == fd.box.id {
				fd.box.selectAll()
				return ui.EventResult{Handled: true, Repaint: true}
			}
		}
	}
	return ui.EventResult{}
}

func (a *ezeApp) paintSet(canvas *ui.Canvas, size ui.Size, theme ui.Theme) {
	bounds := ui.Rect{Width: size.Width, Height: size.Height}
	canvas.Fill(bounds, theme.Background)

	set.view.Draw(canvas, theme, bounds, func(f *flex.Frame) {
		f.Column().Fill().Pad(16).Gap(10).Children(func() {
			f.Text("程序设置").FontSize(20).Bold()

			f.ScrollArea(setScrollID).Grow(1).Gap(10).Children(func() {
				for _, fd := range set.fields {
					box := fd.box
					label := fd.label
					hint := fd.hint
					f.Column().Gap(3).Children(func() {
						f.Text(label)
						box.paintSingle(f, canvas, label, hint).FillWidth()
						f.Text(hint).FontSize(11).Muted()
					})
				}

				f.Text("配置文件地址").Muted()
				f.Text(set.configPath).FontSize(12)
			})

			f.Row().Children(func() {
				if f.Button(setSaveBtnID, "保存设置").Clicked() {
					a.saveSet()
				}
			})
		})
	})
}

// saveSet 保存设置并关闭窗口。
func (a *ezeApp) saveSet() {
	log.D("保存设置")
	for _, fd := range set.fields {
		viper.Set(fd.key, fd.box.Text())
	}
	if err := viper.WriteConfig(); err != nil {
		log.E("保存设置时候发生错误")
	}
	if a.setWin != nil {
		a.setWin.Close()
		a.setWin = nil
	}
}
