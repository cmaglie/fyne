package widget

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/internal/widget"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
)

var _ fyne.Focusable = (*CMButton)(nil)

// CMButton widget has a text label and triggers an event func when clicked
type CMButton struct {
	DisableableWidget
	Text string
	Icon fyne.Resource
	// Specify how prominent the button should be, High will highlight the button and Low will remove some decoration.
	//
	// Since: 1.4
	Importance    ButtonImportance
	Alignment     ButtonAlign
	IconPlacement ButtonIconPlacement
	ColorEnabled  color.Color
	ColorDisabled color.Color
	ColorFocused  color.Color
	ColorPrimary  color.Color
	ColorHover    color.Color

	OnTapped func() `json:"-"`

	hovered, focused bool
	tapAnim          *fyne.Animation
}

// NewCMButton creates a new button widget with the set label and tap handler
func NewCMButton(label string, tapped func()) *CMButton {
	button := &CMButton{
		Text:          label,
		OnTapped:      tapped,
		ColorEnabled:  theme.ButtonColor(),
		ColorDisabled: theme.DisabledButtonColor(),
		ColorFocused:  theme.FocusColor(),
		ColorPrimary:  theme.PrimaryColor(),
		ColorHover:    theme.HoverColor(),
	}

	button.ExtendBaseWidget(button)
	return button
}

// NewCMButtonWithIcon creates a new button widget with the specified label, themed icon and tap handler
func NewCMButtonWithIcon(label string, icon fyne.Resource, tapped func()) *CMButton {
	button := &CMButton{
		Text:     label,
		Icon:     icon,
		OnTapped: tapped,
	}

	button.ExtendBaseWidget(button)
	return button
}

// CreateRenderer is a private method to Fyne which links this widget to its renderer
func (b *CMButton) CreateRenderer() fyne.WidgetRenderer {
	b.ExtendBaseWidget(b)
	seg := &TextSegment{Text: b.Text, Style: RichTextStyleStrong}
	seg.Style.Alignment = fyne.TextAlignCenter
	text := NewRichText(seg)
	text.inset = fyne.NewSize(theme.Padding()*2, theme.Padding()*2)

	background := canvas.NewRectangle(b.ColorEnabled)
	tapBG := canvas.NewRectangle(color.Transparent)
	b.tapAnim = newButtonTapAnimation(tapBG, b)
	b.tapAnim.Curve = fyne.AnimationEaseOut
	objects := []fyne.CanvasObject{
		background,
		tapBG,
		text,
	}
	shadowLevel := widget.ButtonLevel
	if b.Importance == LowImportance {
		shadowLevel = widget.BaseLevel
	}
	r := &cmButtonRenderer{
		ShadowingRenderer: widget.NewShadowingRenderer(objects, shadowLevel),
		background:        background,
		tapBG:             tapBG,
		cmButton:          b,
		label:             text,
		layout:            layout.NewHBoxLayout(),
	}
	r.updateIconAndText()
	r.applyTheme()
	return r
}

// Cursor returns the cursor type of this widget
func (b *CMButton) Cursor() desktop.Cursor {
	return desktop.DefaultCursor
}

// FocusGained is a hook called by the focus handling logic after this object gained the focus.
func (b *CMButton) FocusGained() {
	b.focused = true
	b.Refresh()
}

// FocusLost is a hook called by the focus handling logic after this object lost the focus.
func (b *CMButton) FocusLost() {
	b.focused = false
	b.Refresh()
}

// MinSize returns the size that this widget should not shrink below
func (b *CMButton) MinSize() fyne.Size {
	b.ExtendBaseWidget(b)
	return b.BaseWidget.MinSize()
}

// MouseIn is called when a desktop pointer enters the widget
func (b *CMButton) MouseIn(*desktop.MouseEvent) {
	b.hovered = true
	b.Refresh()
}

// MouseMoved is called when a desktop pointer hovers over the widget
func (b *CMButton) MouseMoved(*desktop.MouseEvent) {
}

// MouseOut is called when a desktop pointer exits the widget
func (b *CMButton) MouseOut() {
	b.hovered = false
	b.Refresh()
}

// SetIcon updates the icon on a label - pass nil to hide an icon
func (b *CMButton) SetIcon(icon fyne.Resource) {
	b.Icon = icon

	b.Refresh()
}

// SetText allows the button label to be changed
func (b *CMButton) SetText(text string) {
	b.Text = text

	b.Refresh()
}

// Tapped is called when a pointer tapped event is captured and triggers any tap handler
func (b *CMButton) Tapped(*fyne.PointEvent) {
	if b.Disabled() {
		return
	}

	b.tapAnimation()
	b.Refresh()

	if b.OnTapped != nil {
		b.OnTapped()
	}
}

// TypedRune is a hook called by the input handling logic on text input events if this object is focused.
func (b *CMButton) TypedRune(rune) {
}

// TypedKey is a hook called by the input handling logic on key events if this object is focused.
func (b *CMButton) TypedKey(ev *fyne.KeyEvent) {
	if ev.Name == fyne.KeySpace {
		b.Tapped(nil)
	}
}

func (b *CMButton) tapAnimation() {
	if b.tapAnim == nil {
		return
	}
	b.tapAnim.Stop()
	b.tapAnim.Start()
}

type cmButtonRenderer struct {
	*widget.ShadowingRenderer

	icon       *canvas.Image
	label      *RichText
	background *canvas.Rectangle
	tapBG      *canvas.Rectangle
	cmButton   *CMButton
	layout     fyne.Layout
}

// Layout the components of the button widget
func (r *cmButtonRenderer) Layout(size fyne.Size) {
	var inset fyne.Position
	bgSize := size
	if r.cmButton.Importance != LowImportance {
		inset = fyne.NewPos(theme.Padding()/2, theme.Padding()/2)
		bgSize = size.Subtract(fyne.NewSize(theme.Padding(), theme.Padding()))
	}
	r.LayoutShadow(bgSize, inset)

	r.background.Move(inset)
	r.background.Resize(bgSize)

	hasIcon := r.icon != nil
	hasLabel := r.label.Segments[0].(*TextSegment).Text != ""
	if !hasIcon && !hasLabel {
		// Nothing to layout
		return
	}
	iconSize := fyne.NewSize(theme.IconInlineSize(), theme.IconInlineSize())
	labelSize := r.label.MinSize()
	padding := r.padding()
	if hasLabel {
		if hasIcon {
			// Both
			var objects []fyne.CanvasObject
			if r.cmButton.IconPlacement == ButtonIconLeadingText {
				objects = append(objects, r.icon, r.label)
			} else {
				objects = append(objects, r.label, r.icon)
			}
			r.icon.SetMinSize(iconSize)
			min := r.layout.MinSize(objects)
			r.layout.Layout(objects, min)
			pos := alignedPosition(r.cmButton.Alignment, padding, min, size)
			r.label.Move(r.label.Position().Add(pos))
			r.icon.Move(r.icon.Position().Add(pos))
		} else {
			// Label Only
			r.label.Move(alignedPosition(r.cmButton.Alignment, padding, labelSize, size))
			r.label.Resize(labelSize)
		}
	} else {
		// Icon Only
		r.icon.Move(alignedPosition(r.cmButton.Alignment, padding, iconSize, size))
		r.icon.Resize(iconSize)
	}
}

// MinSize calculates the minimum size of a button.
// This is based on the contained text, any icon that is set and a standard
// amount of padding added.
func (r *cmButtonRenderer) MinSize() (size fyne.Size) {
	hasIcon := r.icon != nil
	hasLabel := r.label.Segments[0].(*TextSegment).Text != ""
	iconSize := fyne.NewSize(theme.IconInlineSize(), theme.IconInlineSize())
	labelSize := r.label.MinSize()
	if hasLabel {
		size.Width = labelSize.Width
	}
	if hasIcon {
		if hasLabel {
			size.Width += theme.Padding()
		}
		size.Width += iconSize.Width
	}
	size.Height = fyne.Max(labelSize.Height, iconSize.Height)
	size = size.Add(r.padding())
	return
}

func (r *cmButtonRenderer) Refresh() {
	r.label.inset = fyne.NewSize(theme.Padding()*2, theme.Padding()*2)
	r.label.Segments[0].(*TextSegment).Text = r.cmButton.Text
	r.updateIconAndText()
	r.applyTheme()
	r.background.Refresh()
	r.Layout(r.cmButton.Size())
	canvas.Refresh(r.cmButton.super())
}

// applyTheme updates this button to match the current theme
func (r *cmButtonRenderer) applyTheme() {
	r.background.FillColor = r.buttonColor()
	r.label.Segments[0].(*TextSegment).Style.ColorName = theme.ColorNameForeground
	switch {
	case r.cmButton.disabled:
		r.label.Segments[0].(*TextSegment).Style.ColorName = theme.ColorNameDisabled
	case r.cmButton.Importance == HighImportance:
		r.label.Segments[0].(*TextSegment).Style.ColorName = theme.ColorNameBackground
	}
	r.label.Refresh()
	if r.icon != nil && r.icon.Resource != nil {
		switch res := r.icon.Resource.(type) {
		case *theme.ThemedResource:
			if r.cmButton.Importance == HighImportance {
				r.icon.Resource = theme.NewInvertedThemedResource(res)
				r.icon.Refresh()
			}
		case *theme.InvertedThemedResource:
			if r.cmButton.Importance != HighImportance {
				r.icon.Resource = res.Original()
				r.icon.Refresh()
			}
		}
	}
	r.ShadowingRenderer.RefreshShadow()
}

func (r *cmButtonRenderer) buttonColor() color.Color {
	switch {
	case r.cmButton.Disabled():
		return r.cmButton.ColorDisabled
	case r.cmButton.focused:
		return blendColor(r.cmButton.ColorEnabled, r.cmButton.ColorFocused)
	case r.cmButton.hovered:
		bg := r.cmButton.ColorEnabled
		if r.cmButton.Importance == HighImportance {
			bg = r.cmButton.ColorPrimary
		}

		return blendColor(bg, r.cmButton.ColorHover)
	case r.cmButton.Importance == HighImportance:
		return r.cmButton.ColorPrimary
	default:
		return r.cmButton.ColorEnabled
	}
}

func (r *cmButtonRenderer) padding() fyne.Size {
	if r.cmButton.Text == "" {
		return fyne.NewSize(theme.Padding()*4, theme.Padding()*4)
	}
	return fyne.NewSize(theme.Padding()*6, theme.Padding()*4)
}

func (r *cmButtonRenderer) updateIconAndText() {
	if r.cmButton.Icon != nil && r.cmButton.Visible() {
		if r.icon == nil {
			r.icon = canvas.NewImageFromResource(r.cmButton.Icon)
			r.icon.FillMode = canvas.ImageFillContain
			r.SetObjects([]fyne.CanvasObject{r.background, r.tapBG, r.label, r.icon})
		}
		if r.cmButton.Disabled() {
			r.icon.Resource = theme.NewDisabledResource(r.cmButton.Icon)
		} else {
			r.icon.Resource = r.cmButton.Icon
		}
		r.icon.Refresh()
		r.icon.Show()
	} else if r.icon != nil {
		r.icon.Hide()
	}
	if r.cmButton.Text == "" {
		r.label.Hide()
	} else {
		r.label.Show()
	}
	r.label.Refresh()
}
