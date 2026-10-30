# viva-spike

A desktop application built with Viva (Go, drawn on the GPU, no web view). This file is
the short form of how to write it; read it before changing the UI. The commands:

```sh
viva dev .          # run with hot reload: save a file, the window reloads and keeps its state
go test ./...      # tests drive the real application with no window (see Testing)
viva build          # a release program in build/
viva help           # every command
```

## How a window works

`Config.Paint(canvas, size)` is called again after every event that could change the
picture. It draws the whole window from your state, and it may be called many times, so
it must be cheap and must not change state, with one exception: a click is handled where
its control is declared (below). Nothing draws unless an event, `Invalidate`, or a running
animation asks for a frame; there is no frame loop to write.

State lives in a struct (`model` in main.go). Paint reads it. Never keep a `flex.Box`,
a `ui.Canvas` or a closure given to `Children` after Paint returns.

## Laying a page out (package github.com/vivaui/viva/flex)

Describe the page; do not compute coordinates.

```go
m.view.Draw(canvas, theme, bounds, func(f *flex.Frame) {
	f.Column().Fill().Pad(24).Gap(12).Children(func() {
		f.Text("Title").FontSize(20).Bold()
		f.Paragraph(description, 3).Muted()          // wraps; at most 3 lines
		f.Row().Gap(8).Items(flex.Center).Children(func() {
			f.Text("Notify me")
			f.Spacer()                               // pushes what follows to the far edge
			f.Switch(notify, m.notify).Small()
		})
		f.ScrollArea(listID).Grow(1).Gap(4).Children(func() { /* long content */ })
		f.Row().Gap(8).Justify(flex.End).Children(func() {
			f.Button(cancel, "Cancel").Outline()
			if f.Button(save, "Save").Clicked() { m.save() }
		})
	})
})
```

- `m.view` is a `flex.View` in your model: keep one per page painted every frame. It
  makes the page allocation-free once warm.
- Containers: `Row` (left to right), `Column` (top to bottom), `Stack` (on top of each
  other), `Grid(minCell)` (as many equal columns as fit), `Scroll(offset)` (a clipped
  column taller than itself). Leaves: `Text`, `Paragraph`, `Wrap`, `Spacer`, `Divider`, `Icon`,
  `Button`, `Checkbox`, `Switch`, `Slot`, `Progress(fraction)` (a bar; `.Indeterminate()` for unknown length),
  `Spinner()`, `Splitter(id, &state)` (see Keys, drag and the mouse), `Slider(id, &state)` (a number picker, see below).
- Size: nothing = as big as its content; `Width`, `Height`, `Size` fix it; `FillWidth`,
  `FillHeight`, `Fill` take the room; `Grow(n)` shares free room along the container;
  `Shrink(n)`; `Basis(px)` is what an element starts from before free room is shared;
  `WidthPercent`, `HeightPercent` take a share of the container; `MinWidth`, `MaxWidth`,
  `MinHeight`, `MaxHeight` bound it (room a capped element cannot take goes to the others,
  as in CSS: `Grow(1).MaxWidth(470)` beside a `Spacer()` works); `MaxWidthPercent(70)` and `MaxHeightPercent` cap at a share of the container (a chat bubble).
- Space: `Pad`, `PadX`, `PadY`, `PadSides(top,right,bottom,left)`; `Margin`, `MarginX`,
  `MarginY`, `MarginSides`; `Gap` between children. `Pad`, `Border` and `Background` work on a `Text` or `Paragraph` too
  (a chip, a badge, a chat bubble). Every length is a `float32`: `Width(float32(n))`, never an `int`.
- Place: across a container children stretch by default; `Items(flex.Center)` or
  `flex.Start`/`flex.End`/`flex.Stretch` changes that; `flex.Baseline` lines a Row's children up
  on their first line of text; `Self(...)` for one child;
  `Justify(flex.Center|End|Between)` places what is left along it; `Center()` does both.
- Look: `Background`, `HoverBackground`, `Radius`, `Border(width, color)`, `Shadow(ui.ShadowMD)` (a drop shadow under the fill; `ShadowXS` to `ShadowLG`), `Clip`. A circle (an avatar, a dot) is a
  square box with `Radius(size/2)` and `Center()`: `f.Row().Size(32, 32).Radius(16).Background(c).Center().Children(func() { f.Text("AB") })`.
- Window: in a `WindowCustomTitlebar` window nothing moves the window unless you say where: `f.Row().Height(36).MovesWindow().Children(...)` on the top bar
  (controls inside it still take their own clicks; a menu or dialog open over it is dismissed by a click there and the window moves once the pointer drags).
- Text: `FontSize`, `Weight`, `Bold`, `Color`, `Muted`, `Mono`, `MonoFace`, `TextCenter`, `TextEnd`, `TextDirection`, `LineHeight`;
  `Ellipsis()` keeps a Text to one line and ends it with … when it does not fit; `MaxLines(n)` lets a Text
  break into at most n lines (the last ends with …), `f.Wrap(text)` into as many as the width needs. Lines
  break where Unicode (UAX #14) allows: at spaces, after hyphens and slashes, between CJK characters (kinsoku
  rules apply), between the syllables of Thai, Lao, Khmer and Myanmar, at the spaces of Korean; a word wider than
  the line breaks after a `/ - _ .` or between characters, never inside an emoji, a conjunct or a letter with its
  marks. Hebrew, Arabic and other right-to-left text is reordered (UAX #9), aligned to the start of its paragraph
  (the right edge; `TextEnd()` picks the other) and given the paragraph's direction on every line; `TextDirection`
  overrides it. The height follows the width. Viva measures text with its own metrics,
  the same on every platform, and draws the platform's glyphs at that width: a button as wide as its label, a tab, a
  table column sized to content and every line break are the same on macOS, Windows and Linux, so do not size by a
  platform's font (or special-case a platform) to make text fit.
- Buttons (the plain one is the primary; there is no `Primary()`): `Secondary`, `Outline`, `Ghost`, `Link`, `Destructive`, `DestructiveOutline`,
  `XSmall`, `Small`, `Large`, `IconOnly`, `Disabled(cond bool)` (so `f.Button(minus, "-").Disabled(m.n <= 0)`, no `if`).
- A control call returns a `flex.Box` (there is no `f.Box`); keep it when you need it later: `filter := f.Button(id, "Filter")`.
  Checkbox and Switch only draw: they never flip themselves, and a click you do not read does nothing. Toggling is yours:
  `if f.Switch(auto, m.auto).Label("Auto-scroll").Clicked() { m.auto = !m.auto }` (same for `f.Checkbox(id, m.agree)`).
  `Label` is the name a screen reader or a test finds the control by, so give one to every control that has no text of
  its own, and put the visible caption in an `f.Text` beside it. A header row of named controls is a `Row` of `Button`s
  (`.Ghost()`) or `Hit` boxes; each is found by its label.
- Text fields that just work: keep a `flex.Field` per field in the model (`type model struct{ name, password flex.Field }`,
  `m.password.Secure = true`), set `config.TextInput = flex.Fields(&m.name, &m.password)`, and declare
  `f.Field(&m.name, nameID, "Name", "placeholder")`. Typing, caret, selection, clipboard and input methods work;
  read it with `m.name.Text()`, set it with `m.name.SetText("x")`. Do not build `text.Editor`s or snapshots by hand. In a `Row` a field has no width of its own: give it `Grow(1)` or `Width`, or it is invisible.
  Enter in a field: `if f.Field(&m.draft, id, "Message", "Type").Submitted() { m.send(); m.draft.SetText("") }` (true only in the
  repaint that delivers Enter, not for a click; `OnKey` is not needed). The label (first string) is never empty.
- Other fields, painted by their components in tree order: `f.Input(components.InputSnapshot{...})` (a control tall,
  as wide as its container), `f.InputGroup(...)`, `f.Textarea(...)` (at least 64 tall), and
  `f.Segmented(options, selectedID)`. Give them `Width`, `Grow` or `Height` like anything else.
- Interactive rows and cards: `Hit(id, "label")` or `HitWith(id, label, ui.HitOptions)` on a container box (`f.Row()`,
  `f.Column()`, `f.Slot()`; there is no `f.Hit`), then `Children`, then `.Clicked()`. `Hit` only makes a pointer target and an
  accessibility name; it paints nothing, so a container with `Hit` and no `Children` (and no size or background) panics
  at the first paint: `if f.Row().Pad(8).Hit(rows.At(i), name).Children(func() { f.Text(name) }).Clicked() { m.selected = i }`.
  `HitOptions` has `Cursor` (`ui.PointerCursorResizeHorizontal`, `PointerCursorHand`, ...), `Selected`, `Disabled`, `Role`.
- A segmented choice among labels, with no identities to invent: declare a scope once
  (`themeScope = tree.Sub(page, tree.Name("theme"))`), then
  `if i, ok := f.Choice(themeScope, []string{"System", "Light", "Dark"}, m.theme).ChosenIndex(); ok { m.theme = i }`
  (`m.theme` is the selected index). `f.Segmented(options, selectedID)` with `Chosen()` is the form where you
  make the `components.SegmentOption` IDs yourself. A vertical divider is just `f.Divider()` inside a `Row`
  (there is no `.Vertical()`).
- A table: the data is a `flex.TableRows` in the model (`[][]string`, passed as `&m.data`) or your own pointer type with
  `Len() int` and `Cell(row, col int) string`; not a plain slice of structs. Declare scopes for rows and header cells:

  ```go
  t := f.Table(&m.table, flex.TableSpec{Rows: rowScope, Header: headScope, Headers: []string{"Name", "Age"},
  	Sortable: []bool{true, true}, Source: &m.data, Striped: true,
  	Columns: []components.TableColumn{ {Width: 160}, {} }}) // a column with no width, the last, takes the rest
  if i, ok := t.RowClicked(); ok { m.table.Select(i) }       // Selected() is the row or -1, ClearSort() unsorts
  if c, ok := t.HeaderClicked(); ok { m.table.SortBy(c); m.sortData() } // Sort() returns (column, descending, ok)
  ```

  A header cell is a control labelled by its title only when `Header` is set and `Sortable[c]` is true; `Headers` alone
  draws a title nobody can click. In a `ScrollArea` the table is as tall as its rows and only visible rows are read;
  with `Height`, `Fill` or `Grow` it scrolls its rows under a fixed header (`config.OnScroll = m.table.Wheel`);
  `config.OnKey = m.table.Key` adds Up/Down/Page/Home/End selection.
- Tabs: a scope for the tab strip, an index in the model:
  `if i, ok := f.Tabs(tabs, []string{"General", "Keys"}, m.tab).Line().TabClicked(); ok { m.tab = i }`, then declare the
  active tab's content under it with an ordinary `switch m.tab`. `Line()` is the underlined variant (it draws 5 px
  under the strip), `Vertical()` stacks the tabs, `Disabled(true)` makes them inert.
- A drop-down select: a `flex.SelectState` in the model and a scope of its own for it (it numbers the options, the trigger,
  the dismiss target and the scroll arrows): `if i, ok := f.Select(sizes, &m.sizeSel, labels, m.size).Placeholder("Size").Picked(); ok { m.size = i }`.
  `Picked` opens and closes the list for you; the list is placed under the trigger (above it when there is no
  room), painted over everything, and scrolls when it is taller than the window; with the list open, route the wheel to it
  (`if m.sizeSel.Open { return m.sizeSel.Wheel(e) }` first in `config.OnScroll`, then `m.view.Wheel(e)`).
- A modal dialog: declare it only while it is open, `d := f.Dialog(flex.DialogSpec{Title: "Delete host?", Description: "...", Width: 400, Scope: dlgFocus, DismissID: cancel, CloseID: closeBtn})`
  (all three are `ui.ID`s made with `tree.ID`, despite the name `Scope`),
  then `d.Body(func() { ... })` for its content and `d.Footer(func() { f.Button(cancel, "Cancel").Outline(); f.Button(del, "Delete").Destructive() })`
  for its buttons (laid out against the right edge); `d.Dismissed()` is true in the repaint that delivers a click on the
  scrim, the close button or Escape (the dialog routes Escape itself). It is centered over a scrim, as tall as its body needs (`Height` fixes it) and
  traps focus. A `Select` inside its `Body` opens above it.
- Menus, tooltips and select lists: `f.Popup()` takes no room in its container, is as big as its content, is painted
  last and is kept inside the frame. A menu opened from a button: keep the button's `Box`, then declare the popup
  (anywhere after it, inside the page) with `BelowBox`:

  ```go
  filter := f.Button(filterBtn, "Filter").Outline()   // a Box, not an ID
  if filter.Clicked() { m.menuOpen = !m.menuOpen }
  if m.menuOpen {
  	f.Popup().BelowBox(filter, 4, flex.Start).Pad(4).Background(theme.Popover).Children(func() { /* items with Hit */ })
  }
  ```

  Place by rectangle with `At(x, y)` (a context menu at the pointer), `Below(rect, gap, flex.Start|Center|End)` or
  `Above(...)`; for an element of the page `BelowBox(box, gap, align)` (`AboveBox`, `BesideBox`): `Box.Rect()` is zero
  before `End`, so never pass it to `Below`. `Beside(anchor, gap, align)` opens on the right (a submenu), `BesideLeft`
  (`BesideLeftBox`) on the left. The requested side is used; if it does not fit, the side with more room;
  `Start|Center|End` aligns across (shifted, never flipped); then it is moved the least that keeps it inside the
  frame. `Within(rect)` keeps it inside rect instead. A menu that may be long needs a `MaxHeight` and a `ScrollArea`.
- Grid children can cover several cells: `.ColSpan(n)` and `.RowSpan(n)` (columns are equal; the next child goes in
  the first free cell after the previous one, skipping cells a span has taken).
- Take a child out of the flow with `.Absolute()` and place it by edges in its container: `.Top(8).Right(8)`
  (`Bottom`, `Left`; a negative edge reaches outside), `.Inset(0)` for all four, with `Width/Height/Size` or both
  edges of an axis to stretch between them. It takes no room, does not move its siblings, is painted and clicked
  above them, and edges do nothing without `.Absolute()`. A badge on an avatar, a close button on a card, an overlay
  (`f.Stack().Children(...)` holding it), a toast stack at the bottom right of the page
  (`f.Column().Absolute().Right(16).Bottom(16).Gap(8)`).
- Your own painting inside a laid-out page: `Slot()`, then `Rect()` after `f.End()`. Use
  `f := view.Begin(canvas, theme, bounds)` ... `f.End()` instead of `Draw` when you need
  the rectangles before you paint into them.
- A slider (volume, zoom, a threshold): a `flex.SliderState` in the model holds the number in `Value` (set it before the first frame), and
  `if f.Slider(volumeID, &m.volume).Range(0, 1).Step(0.05).Label("Volume").Changed() { m.player.SetVolume(m.volume.Value) }` draws it
  (`Range` defaults to 0..100, `Step` is optional, `Vertical()` makes it tall, `Disabled(cond)` dims it, 16 high and as wide as its column
  or 160 in a row unless given a `Width`). A drag, a click on the track and `config.OnKey = m.volume.Key` (arrows, PageUp/PageDown, Home/End)
  all keep `Value` inside the range; `Changed()` is true in the one paint that delivers a user's change, so call it last in the chain.
  Do not draw a track from `Slot` and `Progress` by hand.
- Scrolling needs nothing in the model: `f.ScrollArea(listID).Grow(1).Children(...)` with an identity (`listID = tree.ID(page, "list")`, or
  `rows.At(i)` for several) keeps its offset in the `flex.View`, and `config.OnScroll = m.view.Wheel` gives the wheel to the area under
  the pointer (the innermost, or the next one out when that is at its end), so side-by-side columns and nested lists each scroll alone.
  A log or chat that grows at the bottom: `.Follow(true)` keeps the area at its end until the user scrolls up (and again when they
  scroll back down; `.Follow(m.autoScroll)` makes it a switch). To move it by program, `s := m.view.ScrollState(listID)` has `Offset`,
  `Max()`, `AtEnd()`, `ScrollTo(px)` and `ScrollToEnd()` (call that after adding the item: it holds in the frame that grows the content);
  ask for it each time and do not keep the pointer. An area a frame does not declare forgets its offset (a closed dialog, another
  tab). `f.Scroll(offset)` is the lower-level form where you own the offset (`ui.ScrollEvent` has `DeltaY float32`, positive reveals
  content below, and `Precise` for trackpads).
- Do not write `ui.Rect{X: ..., Y: ...}` arithmetic for ordinary parts of a page. It breaks
  when a label gets longer or the window gets narrower; flex does not.

## Identities and clicks

Every clickable thing has a stable identity: a name in a tree, declared once at the top.

```go
var (
	tree = ui.Tree{}
	home = tree.Feature("home")
	save = tree.ID(home, "save")
	rows = tree.Sub(home, tree.Name("rows"))   // a list
)
// row i of the list:  rows.At(i)
// a second control per row (a Delete button): del := tree.Sub(home, tree.Name("delete")); del.At(i)
```

`tree.ID(scope, "word")` takes a string; `tree.Name("rows")` is for `tree.Sub`. Both are declared once at
the top, never inside Paint. A `ui.ID` names one control (`Button`, `Checkbox`, `Hit`, `Field`); a `ui.Scope` names a
list of them (`Table`, `Tabs`, `Select`, `Choice`), and `scope.At(i)` is the ID of item i. Never pass `nil`,
a string or your own data where either is wanted. `f.Checkbox(id, checked)` and `f.Switch(id, on)` take the identity first;
`.Label("...")` names a checkbox or switch for accessibility (it is not a caption: put a `f.Text` beside it).

All of a page's identities are made from its feature scope, menu items and popups included
(`tree.ID(page, "filter-today")`): `tree.ID` takes a Scope (`page`, `rows`), never another ID.
Pass `&tree` as `Config.Tree`. Handle a click where the control is declared:
`if f.Button(save, "Save").Clicked() { ... }`. `Clicked` is true only in the one repaint that
delivers the click. For a list, `f.Row().Hit(rows.ID(ui.Index(i)), name).Clicked()`.
Hover is `canvas.Hovered() == id` (flex's `HoverBackground` does this). From inside the frame callback,
`f.Canvas().Clicked(id)` asks the same for any identity.
An app with one central click switch may instead read `if id, ok := canvas.ClickTarget(); ok { canvas.AfterPaint(func() { m.Click(id) }) }`
once at the start of its paint. `AfterPaint` runs the handler after the paint, outside it, then repaints: use it
whenever a handler may block or open a native dialog. Prefer `Clicked` at each control.

## Keys, drag and the mouse

`Config.OnKey func(ui.KeyEvent) ui.EventResult` is called only for the control that has keyboard focus, so a key
handler on a page where nothing is focused never runs. For "Escape closes this" use the dismiss target below, not
`OnKey`. `ui.EventResult{Repaint: true}` asks for a repaint, `Handled: true` stops the default handling; it is
a struct, there is no `ui.Handled` constant.

A drag you handle yourself is `Config.OnPointer func(ui.PointerEvent) ui.EventResult`: `e.Kind` is `ui.PointerDown`, `PointerMove`,
`PointerUp` or `PointerExit`, `e.X`/`e.Y` are window pixels, `e.Target` is the identity pressed (it stays the pressed one while a
drag runs). Two panes with a draggable divider need none of it: keep a `flex.SplitterState` in the model, give the first pane its
`Pos` and declare `Splitter` between them (the pointer is captured, the resize cursor shows, `Mins` bounds both panes, and
`config.OnKey = m.split.Key` adds the arrow keys, `m.split.Dragging()` says a drag is running):

```go
m.split = flex.SplitterState{Pos: 240} // once, in newModel
f.Row().Fill().Children(func() {
	f.Column().Width(m.split.Pos).Children(func() { /* left pane */ })
	f.Splitter(splitID, &m.split).Mins(120, 200)
	f.Column().Grow(1).Children(func() { /* right pane */ })
})
```

(`Splitter` in a `Column` divides top and bottom; the first pane then gets `Height(m.split.Pos)`.)

## A modal dialog

Use `f.Dialog` (see Laying a page out): `d.Dismissed()` is true for the scrim, the close button and Escape alike
(`OnKey` never sees Escape for an unfocused page, and needs not). For a modal of your own look, build it as the last
child of a `Stack`, call `f.Canvas().SetEscapeTarget(scrim)` while it is open (Escape then arrives as a click on
`scrim`), and give the panel a `Hit` of its own so a click inside does not reach the scrim.

## Animation and time

Nothing runs by itself. A paint that must be followed by another (a progress bar, a spinner) calls
`canvas.RequestFrame()` (or `RequestFrameIn(d)`) and keeps asking while it is still moving; when it stops asking the
window goes idle again. `canvas.Now()` is the animation clock (a `time.Duration`): store `began := canvas.Now()` when
the thing starts and derive the state from `canvas.Now()-began` in Paint, never from a ticker or goroutine.
`canvas.Animate(id, slot, target, ui.DefaultMotion())` eases a value toward a target for you (hover fades, sliding
indicators).

```go
if m.running {
	fraction = min(float32(canvas.Now()-m.began)/float32(3*time.Second), 1)
	if fraction < 1 { canvas.RequestFrame() } else { m.running, m.done = false, true }
}
f.Column().Fill().Pad(24).Gap(12).Children(func() {
	if f.Button(start, "Start").Clicked() { m.running, m.began = true, canvas.Now(); canvas.RequestFrame() }
	f.Row().Gap(12).Items(flex.Center).Children(func() {
		f.Progress(fraction).Grow(1).Height(8)   // f.Progress(0).Indeterminate() when the length is unknown
		if m.running { f.Spinner() }             // turns by itself; asks for frames only while on screen
	})
})
```

`Progress` and `Spinner` need no `RequestFrame` of yours; the animation you drive from your own state (the fraction above) does.

## Several windows

`app.RunCanvas` runs one window. For more, `app.RunApplication(func(a *app.Application) error {...})` calls its
function once on the event-loop thread; open windows there or from a click, each with its own config and canvas:
`w, err := a.OpenWindow(config, new(ui.Canvas))`. `w.Close()` closes one (and `w.Closed()` says whether it is gone;
it is true for a nil window); `a.Quit()` ends everything. Keep `newConfig(m)` for the main window so tests still drive
it. To repaint another window after the shared model changed, give it an `app.Invalidator` (`config.Invalidator =
inv`, declared as `&app.Invalidator{}`, the zero value works) and call `inv.Invalidate()`.

```go
err := app.RunApplication(func(a *app.Application) error {
	m := newModel()
	var inspector *app.Window
	m.openInspector = func() { // set before the window opens; called from a click on the event-loop thread
		if inspector.Closed() { inspector, _ = a.OpenWindow(inspectorConfig(m), new(ui.Canvas)) }
	}
	m.closeInspector = func() { inspector.Close() }
	config := newConfig(m)
	config.OnClosed = a.Quit // closing the main window ends the program
	_, err := a.OpenWindow(config, new(ui.Canvas))
	return err
})
```

## The theme (ui.Theme; `ui.LightTheme()`, `ui.DarkTheme()`)

Colors, all `ui.Color`: `Background`, `Foreground`, `Card`, `CardForeground`, `CardElevated`, `Popover`,
`PopoverForeground`, `Muted`, `MutedForeground`, `Border`, `Hover`, `Input`, `Primary`, `PrimaryForeground`,
`Secondary`, `SecondaryForeground`, `Accent`, `AccentForeground`, `Ring`, `Success`, `Danger`, `DangerForeground`, and
`Chart` (an array of five). There is no `Surface`, `Text` or `Error`: use `Card`, `Foreground`, `Danger`.
A `ui.Color` is `{Red, Green, Blue, Alpha float32}` with each in 0..1 (so `ui.Color{Alpha: 0.5}` is a black scrim);
it has `Mix(other, t)`, `Over(backdrop)` and `WithAlpha(a)`, and no `R` or `A`. To write a color of your own, `ui.RGB24(0x0a84ff)`
(an integer constant, no string to parse), `ui.RGB(10, 132, 255)` or `ui.RGBA(10, 132, 255, 0.5)`; take theme colors where you can. A dark window is `ui.DarkTheme()` in place
of `ui.LightTheme()`: keep the theme in the model (`m.dark`) and pick it at the top of Paint; `Config.Paint` receives
no theme, you fill `theme.Background` yourself.
`theme.Metrics` has `ControlHeight`, `RadiusControl`, `RadiusSurface`, `FontSizeS`, `FontSizeM`, `FontSizeL`,
`PadControlX`, `PadSurface`, `GapItem`. Take colors and sizes from the theme rather than writing literals.

## Components (package github.com/vivaui/viva/components)

Use them from `github.com/vivaui/viva/components` directly: Button, Checkbox, Switch, Input, Table, Tabs,
Dialog, Popover, Command, Calendar, Pagination, Progress, Tooltip and more; importing the
package costs nothing until you call one. To change a component beyond its options, copy it
into your project with `viva add <name>` (`viva list` shows every name; what it calls into is
copied with it). They take a
`ui.Theme` (`ui.LightTheme()`, `ui.DarkTheme()`) and a rectangle; inside flex, put one in a
`Slot()` and paint it at `Rect()`.

Painting your own thing in a `Slot`: `canvas.Fill(rect, color)`, `Rounded(rect, radius, color)`, `Border`, `RoundedBorder`, `Elevate(rect, radius, ui.ShadowMD)` (a drop shadow: call it before the surface),
`DrawPath` and `canvas.Label(text, rect, ui.TextStyle{Size: 14, Color: c})` (`LabelCenter` is the same, centered) are all of
it; there is no arc, circle or line call (a circle is a `Rounded` square, or a flex box with `Radius`).
For tabs, tables and text fields use `f.Tabs`, `f.Table` and `f.Field` above, not the components by hand.

## Background work

A goroutine never touches the window. It stores its result under a lock, then calls
`invalidator.Invalidate()` (`Config.Invalidator`); `Config.OnWake` copies the result into the
model on the window's thread before the repaint.

## Testing

`app.NewDriver(config, &canvas)` runs the real application with no window:
`driver.Click(x, y)`, `Type`, `Key`, `Move`, `Scroll`, `Resize`, then `driver.Snapshot()` for the
labels, hits and rectangles painted. main_test.go clicks a control by its identity and reads
the labels; copy it. Test behaviour a user sees (the count changed, the row opened), not
pixel positions.

## Rules

- Allocation-conscious: no `fmt.Sprintf` per frame for text that has not changed if it is in a
  hot path; keep the strings in the model.
- No reflection, no `map[string]any` for UI state, no goroutines started from Paint.

## Limits to know

- Controls and containers laid out by flex are painted when you call `End` (or at the end of
  `Draw`), before whatever you paint into a `Slot()` afterwards. Hit order follows paint order, so
  declare overlapping interactive things in the order you want them to win.
- A `Border` is inside the box (as CSS `border-box`): it takes its width out of the padding area.
  `Stroke(width, color)` draws the same line but takes no room. `HitOutset(top, right, bottom, left)` makes the
  click region reach beyond the painted rectangle (an inset rounded row that clicks across its whole width).
