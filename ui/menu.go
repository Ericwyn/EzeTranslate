package ui

import (
	"fyne.io/fyne/v2"
)

/*
菜单结构是这样的

- MainMenu
	Menu1
		MenuItem-1
		MenuItem-2
		MenuItem-3
	Menu2
	Menu3

在 Menu1 中, 如果又没有个 MenuItem 的 IsQuit 为 True 的话, 那么会加一个 Quit 的 MenuItem

*/
func createAppMenu() *fyne.MainMenu {

	configSetMenuItem := fyne.NewMenuItem("参数设置", func() {
		showSetUi()
	})

	openLogMenuItem := fyne.NewMenuItem("日志查看", func() {
		showLogUi()
	})
	openLogMenuItem.IsQuit = true

	logMenu := fyne.NewMenu("设置", configSetMenuItem, openLogMenuItem)

	mainMenu := fyne.NewMainMenu(
		logMenu,
		fyne.NewMenu("关于",
			fyne.NewMenuItem("版本信息 ", func() {
				showAboutUi()
			}),
		),
	)

	return mainMenu
}
