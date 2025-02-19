package rui

import (
	"fmt"
	"strconv"
	"strings"
)

// Frame - the location and size of a rectangle area
type Frame struct {
	// Left - the left border
	Left float64
	// Top - the top border
	Top float64
	// Width - the width of a rectangle area
	Width float64
	// Height - the height of a rectangle area
	Height float64
}

// Right returns the right border
func (frame Frame) Right() float64 {
	return frame.Left + frame.Width
}

// Bottom returns the bottom border
func (frame Frame) Bottom() float64 {
	return frame.Top + frame.Height
}

// View - base view interface
type View interface {
	ViewStyle
	fmt.Stringer

	// Init initializes fields of View by default values
	Init(session Session)
	// Session returns the current Session interface
	Session() Session
	// Parent returns the parent view
	Parent() View
	parentHTMLID() string
	setParentID(parentID string)
	// Tag returns the tag of View interface
	Tag() string
	// ID returns the id of the view
	ID() string
	// Focusable returns true if the view receives the focus
	Focusable() bool
	// Frame returns the location and size of the view in pixels
	Frame() Frame
	// Scroll returns the location size of the scrolable view in pixels
	Scroll() Frame
	// SetAnimated sets the value (second argument) of the property with name defined by the first argument.
	// Return "true" if the value has been set, in the opposite case "false" are returned and
	// a description of the error is written to the log
	SetAnimated(tag string, value interface{}, animation Animation) bool
	// SetChangeListener set the function to track the change of the View property
	SetChangeListener(tag string, listener func(View, string))
	// HasFocus returns 'true' if the view has focus
	HasFocus() bool

	handleCommand(self View, command string, data DataObject) bool
	htmlClass(disabled bool) string
	htmlTag() string
	closeHTMLTag() bool
	htmlID() string
	htmlSubviews(self View, buffer *strings.Builder)
	htmlProperties(self View, buffer *strings.Builder)
	htmlDisabledProperties(self View, buffer *strings.Builder)
	cssStyle(self View, builder cssBuilder)
	addToCSSStyle(addCSS map[string]string)

	getTransitions() Params

	onResize(self View, x, y, width, height float64)
	onItemResize(self View, index string, x, y, width, height float64)
	setNoResizeEvent()
	isNoResizeEvent() bool
	setScroll(x, y, width, height float64)
}

// viewData - base implementation of View interface
type viewData struct {
	viewStyle
	session          Session
	tag              string
	viewID           string
	_htmlID          string
	parentID         string
	systemClass      string
	changeListener   map[string]func(View, string)
	singleTransition map[string]Animation
	addCSS           map[string]string
	frame            Frame
	scroll           Frame
	noResizeEvent    bool
	created          bool
	hasFocus         bool
	//animation map[string]AnimationEndListener
}

func newView(session Session) View {
	view := new(viewData)
	view.Init(session)
	return view
}

func setInitParams(view View, params Params) {
	if params != nil {
		session := view.Session()
		if !session.ignoreViewUpdates() {
			session.setIgnoreViewUpdates(true)
			defer session.setIgnoreViewUpdates(false)
		}
		for _, tag := range params.AllTags() {
			if value, ok := params[tag]; ok {
				view.Set(tag, value)
			}
		}
	}
}

// NewView create new View object and return it
func NewView(session Session, params Params) View {
	view := new(viewData)
	view.Init(session)
	setInitParams(view, params)
	return view
}

func (view *viewData) Init(session Session) {
	view.viewStyle.init()
	view.tag = "View"
	view.session = session
	view.changeListener = map[string]func(View, string){}
	view.addCSS = map[string]string{}
	//view.animation = map[string]AnimationEndListener{}
	view.singleTransition = map[string]Animation{}
	view.noResizeEvent = false
	view.created = false
}

func (view *viewData) Session() Session {
	return view.session
}

func (view *viewData) Parent() View {
	return view.session.viewByHTMLID(view.parentID)
}

func (view *viewData) parentHTMLID() string {
	return view.parentID
}

func (view *viewData) setParentID(parentID string) {
	view.parentID = parentID
}

func (view *viewData) Tag() string {
	return view.tag
}

func (view *viewData) ID() string {
	return view.viewID
}

func (view *viewData) ViewByID(id string) View {
	if id == view.ID() {
		if v := view.session.viewByHTMLID(view.htmlID()); v != nil {
			return v
		}
		return view
	}
	return nil
}

func (view *viewData) Focusable() bool {
	if focus, ok := boolProperty(view, Focusable, view.session); ok {
		return focus
	}
	return false
}

func (view *viewData) Remove(tag string) {
	view.remove(strings.ToLower(tag))
}

func (view *viewData) remove(tag string) {
	switch tag {
	case ID:
		view.viewID = ""

	case UserData:
		delete(view.properties, tag)

	case Style, StyleDisabled:
		if _, ok := view.properties[tag]; ok {
			delete(view.properties, tag)
			updateProperty(view.htmlID(), "class", view.htmlClass(IsDisabled(view, "")), view.session)
		}

	case FocusEvent, LostFocusEvent:
		view.removeFocusListener(tag)

	case KeyDownEvent, KeyUpEvent:
		view.removeKeyListener(tag)

	case ClickEvent, DoubleClickEvent, MouseDown, MouseUp, MouseMove, MouseOut, MouseOver, ContextMenuEvent:
		view.removeMouseListener(tag)

	case PointerDown, PointerUp, PointerMove, PointerOut, PointerOver, PointerCancel:
		view.removePointerListener(tag)

	case TouchStart, TouchEnd, TouchMove, TouchCancel:
		view.removeTouchListener(tag)

	case TransitionRunEvent, TransitionStartEvent, TransitionEndEvent, TransitionCancelEvent:
		view.removeTransitionListener(tag)

	case AnimationStartEvent, AnimationEndEvent, AnimationIterationEvent, AnimationCancelEvent:
		view.removeAnimationListener(tag)

	case ResizeEvent, ScrollEvent:
		delete(view.properties, tag)

	case Content:
		if _, ok := view.properties[Content]; ok {
			delete(view.properties, Content)
			updateInnerHTML(view.htmlID(), view.session)
		}

	default:
		view.viewStyle.remove(tag)
		viewPropertyChanged(view, tag)
	}

	view.propertyChangedEvent(tag)
}

func (view *viewData) propertyChangedEvent(tag string) {
	if listener, ok := view.changeListener[tag]; ok {
		listener(view, tag)
	}

	switch tag {
	case BorderLeft, BorderRight, BorderTop, BorderBottom,
		BorderStyle, BorderLeftStyle, BorderRightStyle, BorderTopStyle, BorderBottomStyle,
		BorderColor, BorderLeftColor, BorderRightColor, BorderTopColor, BorderBottomColor,
		BorderWidth, BorderLeftWidth, BorderRightWidth, BorderTopWidth, BorderBottomWidth:
		tag = Border

	case CellBorderStyle, CellBorderColor, CellBorderWidth,
		CellBorderLeft, CellBorderLeftStyle, CellBorderLeftColor, CellBorderLeftWidth,
		CellBorderRight, CellBorderRightStyle, CellBorderRightColor, CellBorderRightWidth,
		CellBorderTop, CellBorderTopStyle, CellBorderTopColor, CellBorderTopWidth,
		CellBorderBottom, CellBorderBottomStyle, CellBorderBottomColor, CellBorderBottomWidth:
		tag = CellBorder

	case OutlineColor, OutlineStyle, OutlineWidth:
		tag = Outline

	case RadiusX, RadiusY, RadiusTopLeft, RadiusTopLeftX, RadiusTopLeftY,
		RadiusTopRight, RadiusTopRightX, RadiusTopRightY,
		RadiusBottomLeft, RadiusBottomLeftX, RadiusBottomLeftY,
		RadiusBottomRight, RadiusBottomRightX, RadiusBottomRightY:
		tag = Radius

	case MarginTop, MarginRight, MarginBottom, MarginLeft,
		"top-margin", "right-margin", "bottom-margin", "left-margin":
		tag = Margin

	case PaddingTop, PaddingRight, PaddingBottom, PaddingLeft,
		"top-padding", "right-padding", "bottom-padding", "left-padding":
		tag = Padding

	case CellPaddingTop, CellPaddingRight, CellPaddingBottom, CellPaddingLeft:
		tag = CellPadding

	case ColumnSeparatorStyle, ColumnSeparatorWidth, ColumnSeparatorColor:
		tag = ColumnSeparator

	default:
		return
	}

	if listener, ok := view.changeListener[tag]; ok {
		listener(view, tag)
	}

}

func (view *viewData) Set(tag string, value interface{}) bool {
	return view.set(strings.ToLower(tag), value)
}

func (view *viewData) set(tag string, value interface{}) bool {
	if value == nil {
		view.remove(tag)
		return true
	}

	result := func(res bool) bool {
		if res {
			view.propertyChangedEvent(tag)
		}
		return res
	}

	switch tag {
	case ID:
		text, ok := value.(string)
		if !ok {
			notCompatibleType(ID, value)
			return false
		}
		view.viewID = text

	case UserData:
		view.properties[tag] = value

	case Style, StyleDisabled:
		text, ok := value.(string)
		if !ok {
			notCompatibleType(ID, value)
			return false
		}
		view.properties[tag] = text
		if view.created {
			updateProperty(view.htmlID(), "class", view.htmlClass(IsDisabled(view, "")), view.session)
		}

	case FocusEvent, LostFocusEvent:
		return result(view.setFocusListener(tag, value))

	case KeyDownEvent, KeyUpEvent:
		return result(view.setKeyListener(tag, value))

	case ClickEvent, DoubleClickEvent, MouseDown, MouseUp, MouseMove, MouseOut, MouseOver, ContextMenuEvent:
		return result(view.setMouseListener(tag, value))

	case PointerDown, PointerUp, PointerMove, PointerOut, PointerOver, PointerCancel:
		return result(view.setPointerListener(tag, value))

	case TouchStart, TouchEnd, TouchMove, TouchCancel:
		return result(view.setTouchListener(tag, value))

	case TransitionRunEvent, TransitionStartEvent, TransitionEndEvent, TransitionCancelEvent:
		return result(view.setTransitionListener(tag, value))

	case AnimationStartEvent, AnimationEndEvent, AnimationIterationEvent, AnimationCancelEvent:
		return result(view.setAnimationListener(tag, value))

	case ResizeEvent, ScrollEvent:
		return result(view.setFrameListener(tag, value))

	default:
		if !view.viewStyle.set(tag, value) {
			return false
		}
		if view.created {
			viewPropertyChanged(view, tag)
		}
	}

	view.propertyChangedEvent(tag)
	return true
}

func viewPropertyChanged(view *viewData, tag string) {
	if view.updateTransformProperty(tag) {
		return
	}

	htmlID := view.htmlID()
	session := view.session

	switch tag {
	case Disabled:
		updateInnerHTML(view.parentHTMLID(), session)
		return

	case Visibility:
		switch GetVisibility(view, "") {
		case Invisible:
			updateCSSProperty(htmlID, Visibility, "hidden", session)
			updateCSSProperty(htmlID, "display", "", session)

		case Gone:
			updateCSSProperty(htmlID, Visibility, "hidden", session)
			updateCSSProperty(htmlID, "display", "none", session)

		default:
			updateCSSProperty(htmlID, Visibility, "visible", session)
			updateCSSProperty(htmlID, "display", "", session)
		}
		return

	case Background:
		updateCSSProperty(htmlID, Background, view.backgroundCSS(session), session)
		return

	case Border:
		if getBorder(view, Border) == nil {
			updateCSSProperty(htmlID, BorderWidth, "", session)
			updateCSSProperty(htmlID, BorderColor, "", session)
			updateCSSProperty(htmlID, BorderStyle, "none", session)
			return
		}
		fallthrough

	case BorderLeft, BorderRight, BorderTop, BorderBottom:
		if border := getBorder(view, Border); border != nil {
			updateCSSProperty(htmlID, BorderWidth, border.cssWidthValue(session), session)
			updateCSSProperty(htmlID, BorderColor, border.cssColorValue(session), session)
			updateCSSProperty(htmlID, BorderStyle, border.cssStyleValue(session), session)
		}
		return

	case BorderStyle, BorderLeftStyle, BorderRightStyle, BorderTopStyle, BorderBottomStyle:
		if border := getBorder(view, Border); border != nil {
			updateCSSProperty(htmlID, BorderStyle, border.cssStyleValue(session), session)
		}
		return

	case BorderColor, BorderLeftColor, BorderRightColor, BorderTopColor, BorderBottomColor:
		if border := getBorder(view, Border); border != nil {
			updateCSSProperty(htmlID, BorderColor, border.cssColorValue(session), session)
		}
		return

	case BorderWidth, BorderLeftWidth, BorderRightWidth, BorderTopWidth, BorderBottomWidth:
		if border := getBorder(view, Border); border != nil {
			updateCSSProperty(htmlID, BorderWidth, border.cssWidthValue(session), session)
		}
		return

	case Outline, OutlineColor, OutlineStyle, OutlineWidth:
		updateCSSProperty(htmlID, Outline, GetOutline(view, "").cssString(), session)
		return

	case Shadow:
		updateCSSProperty(htmlID, "box-shadow", shadowCSS(view, Shadow, session), session)
		return

	case TextShadow:
		updateCSSProperty(htmlID, "text-shadow", shadowCSS(view, TextShadow, session), session)
		return

	case Radius, RadiusX, RadiusY, RadiusTopLeft, RadiusTopLeftX, RadiusTopLeftY,
		RadiusTopRight, RadiusTopRightX, RadiusTopRightY,
		RadiusBottomLeft, RadiusBottomLeftX, RadiusBottomLeftY,
		RadiusBottomRight, RadiusBottomRightX, RadiusBottomRightY:
		radius := GetRadius(view, "")
		updateCSSProperty(htmlID, "border-radius", radius.cssString(), session)
		return

	case Margin, MarginTop, MarginRight, MarginBottom, MarginLeft,
		"top-margin", "right-margin", "bottom-margin", "left-margin":
		margin := GetMargin(view, "")
		updateCSSProperty(htmlID, Margin, margin.cssString(), session)
		return

	case Padding, PaddingTop, PaddingRight, PaddingBottom, PaddingLeft,
		"top-padding", "right-padding", "bottom-padding", "left-padding":
		padding := GetPadding(view, "")
		updateCSSProperty(htmlID, Padding, padding.cssString(), session)
		return

	case AvoidBreak:
		if avoid, ok := boolProperty(view, AvoidBreak, session); ok {
			if avoid {
				updateCSSProperty(htmlID, "break-inside", "avoid", session)
			} else {
				updateCSSProperty(htmlID, "break-inside", "auto", session)
			}
		}
		return

	case Clip:
		if clip := getClipShape(view, Clip, session); clip != nil && clip.valid(session) {
			updateCSSProperty(htmlID, `clip-path`, clip.cssStyle(session), session)
		} else {
			updateCSSProperty(htmlID, `clip-path`, "none", session)
		}
		return

	case ShapeOutside:
		if clip := getClipShape(view, ShapeOutside, session); clip != nil && clip.valid(session) {
			updateCSSProperty(htmlID, ShapeOutside, clip.cssStyle(session), session)
		} else {
			updateCSSProperty(htmlID, ShapeOutside, "none", session)
		}
		return

	case Filter:
		text := ""
		if value := view.getRaw(tag); value != nil {
			if filter, ok := value.(ViewFilter); ok {
				text = filter.cssStyle(session)
			}
		}
		updateCSSProperty(htmlID, tag, text, session)
		return

	case BackdropFilter:
		text := ""
		if value := view.getRaw(tag); value != nil {
			if filter, ok := value.(ViewFilter); ok {
				text = filter.cssStyle(session)
			}
		}
		updateCSSProperty(htmlID, "-webkit-backdrop-filter", text, session)
		updateCSSProperty(htmlID, tag, text, session)
		return

	case FontName:
		if font, ok := stringProperty(view, FontName, session); ok {
			updateCSSProperty(htmlID, "font-family", font, session)
		} else {
			updateCSSProperty(htmlID, "font-family", "", session)
		}
		return

	case Italic:
		if state, ok := boolProperty(view, tag, session); ok {
			if state {
				updateCSSProperty(htmlID, "font-style", "italic", session)
			} else {
				updateCSSProperty(htmlID, "font-style", "normal", session)
			}
		} else {
			updateCSSProperty(htmlID, "font-style", "", session)
		}
		return

	case SmallCaps:
		if state, ok := boolProperty(view, tag, session); ok {
			if state {
				updateCSSProperty(htmlID, "font-variant", "small-caps", session)
			} else {
				updateCSSProperty(htmlID, "font-variant", "normal", session)
			}
		} else {
			updateCSSProperty(htmlID, "font-variant", "", session)
		}
		return

	case Strikethrough, Overline, Underline:
		updateCSSProperty(htmlID, "text-decoration", view.cssTextDecoration(session), session)
		for _, tag2 := range []string{TextLineColor, TextLineStyle, TextLineThickness} {
			viewPropertyChanged(view, tag2)
		}
		return

	case Transition:
		view.updateTransitionCSS()
		return

	case AnimationTag:
		updateCSSProperty(htmlID, AnimationTag, view.animationCSS(session), session)
		return

	case AnimationPaused:
		paused, ok := boolProperty(view, AnimationPaused, session)
		if !ok {
			updateCSSProperty(htmlID, `animation-play-state`, ``, session)
		} else if paused {
			updateCSSProperty(htmlID, `animation-play-state`, `paused`, session)
		} else {
			updateCSSProperty(htmlID, `animation-play-state`, `running`, session)
		}
		return

	case ZIndex:
		if i, ok := intProperty(view, ZIndex, session, 0); ok {
			updateCSSProperty(htmlID, ZIndex, strconv.Itoa(i), session)
		}
		return

	case Row, Column:
		if parent := view.parentHTMLID(); parent != "" {
			updateInnerHTML(parent, session)
		}
		return
	}

	if cssTag, ok := sizeProperties[tag]; ok {
		size, _ := sizeProperty(view, tag, session)
		updateCSSProperty(htmlID, cssTag, size.cssString(""), session)
		return
	}

	colorTags := map[string]string{
		BackgroundColor: BackgroundColor,
		TextColor:       "color",
		TextLineColor:   "text-decoration-color",
		CaretColor:      CaretColor,
	}
	if cssTag, ok := colorTags[tag]; ok {
		if color, ok := colorProperty(view, tag, session); ok {
			updateCSSProperty(htmlID, cssTag, color.cssString(), session)
		} else {
			updateCSSProperty(htmlID, cssTag, "", session)
		}
		return
	}

	if valuesData, ok := enumProperties[tag]; ok && valuesData.cssTag != "" {
		n, _ := enumProperty(view, tag, session, 0)
		updateCSSProperty(htmlID, valuesData.cssTag, valuesData.cssValues[n], session)
		return
	}

	for _, floatTag := range []string{Opacity, ScaleX, ScaleY, ScaleZ, RotateX, RotateY, RotateZ} {
		if tag == floatTag {
			if f, ok := floatProperty(view, floatTag, session, 0); ok {
				updateCSSProperty(htmlID, floatTag, strconv.FormatFloat(f, 'g', -1, 64), session)
			}
			return
		}
	}
}

func (view *viewData) Get(tag string) interface{} {
	return view.get(strings.ToLower(tag))
}

func (view *viewData) get(tag string) interface{} {
	if tag == ID {
		if view.viewID != "" {
			return view.viewID
		} else {
			return nil
		}
	}
	return view.viewStyle.get(tag)
}

func (view *viewData) htmlTag() string {
	if semantics := GetSemantics(view, ""); semantics > DefaultSemantics {
		values := enumProperties[Semantics].cssValues
		if semantics < len(values) {
			return values[semantics]
		}
	}
	return "div"
}

func (view *viewData) closeHTMLTag() bool {
	return true
}

func (view *viewData) htmlID() string {
	if view._htmlID == "" {
		view._htmlID = view.session.nextViewID()
	}
	return view._htmlID
}

func (view *viewData) htmlSubviews(self View, buffer *strings.Builder) {
}

func (view *viewData) addToCSSStyle(addCSS map[string]string) {
	view.addCSS = addCSS
}

func (view *viewData) cssStyle(self View, builder cssBuilder) {
	view.viewStyle.cssViewStyle(builder, view.session)
	switch GetVisibility(view, "") {
	case Invisible:
		builder.add(`visibility`, `hidden`)

	case Gone:
		builder.add(`display`, `none`)
	}

	if view.addCSS != nil {
		for tag, value := range view.addCSS {
			builder.add(tag, value)
		}
	}
}

func (view *viewData) htmlProperties(self View, buffer *strings.Builder) {
	view.created = true
	if view.frame.Left != 0 || view.frame.Top != 0 || view.frame.Width != 0 || view.frame.Height != 0 {
		buffer.WriteString(fmt.Sprintf(` data-left="%g" data-top="%g" data-width="%g" data-height="%g"`,
			view.frame.Left, view.frame.Top, view.frame.Width, view.frame.Height))
	}
}

func (view *viewData) htmlDisabledProperties(self View, buffer *strings.Builder) {
	if IsDisabled(self, "") {
		buffer.WriteString(` data-disabled="1"`)
	} else {
		buffer.WriteString(` data-disabled="0"`)
	}
}

func viewHTML(view View, buffer *strings.Builder) {
	viewHTMLTag := view.htmlTag()
	buffer.WriteRune('<')
	buffer.WriteString(viewHTMLTag)
	buffer.WriteString(` id="`)
	buffer.WriteString(view.htmlID())
	buffer.WriteRune('"')

	disabled := IsDisabled(view, "")

	if cls := view.htmlClass(disabled); cls != "" {
		buffer.WriteString(` class="`)
		buffer.WriteString(cls)
		buffer.WriteRune('"')
	}

	var cssBuilder viewCSSBuilder
	view.cssStyle(view, &cssBuilder)

	if style := cssBuilder.finish(); style != "" {
		buffer.WriteString(` style="`)
		buffer.WriteString(style)
		buffer.WriteRune('"')
	}

	buffer.WriteRune(' ')
	view.htmlProperties(view, buffer)
	buffer.WriteRune(' ')
	view.htmlDisabledProperties(view, buffer)

	if view.isNoResizeEvent() {
		buffer.WriteString(` data-noresize="1" `)
	} else {
		buffer.WriteRune(' ')
	}

	if view.Focusable() && !disabled {
		buffer.WriteString(`tabindex="0" `)
	}

	buffer.WriteString(`onscroll="scrollEvent(this, event)" `)

	keyEventsHtml(view, buffer)
	mouseEventsHtml(view, buffer)
	pointerEventsHtml(view, buffer)
	touchEventsHtml(view, buffer)
	focusEventsHtml(view, buffer)
	transitionEventsHtml(view, buffer)
	animationEventsHtml(view, buffer)

	buffer.WriteRune('>')
	view.htmlSubviews(view, buffer)
	if view.closeHTMLTag() {
		buffer.WriteString(`</`)
		buffer.WriteString(viewHTMLTag)
		buffer.WriteRune('>')
	}
}

func (view *viewData) htmlClass(disabled bool) string {
	cls := "ruiView"
	disabledStyle := false
	if disabled {
		if value, ok := stringProperty(view, StyleDisabled, view.Session()); ok && value != "" {
			cls += " " + value
			disabledStyle = true
		}
	}
	if !disabledStyle {
		if value, ok := stringProperty(view, Style, view.Session()); ok {
			cls += " " + value
		}
	}

	if view.systemClass != "" {
		cls = view.systemClass + " " + cls
	}

	return cls
}

func (view *viewData) handleCommand(self View, command string, data DataObject) bool {
	switch command {

	case KeyDownEvent, KeyUpEvent:
		if !IsDisabled(self, "") {
			handleKeyEvents(self, command, data)
		}

	case ClickEvent, DoubleClickEvent, MouseDown, MouseUp, MouseMove, MouseOut, MouseOver, ContextMenuEvent:
		handleMouseEvents(self, command, data)

	case PointerDown, PointerUp, PointerMove, PointerOut, PointerOver, PointerCancel:
		handlePointerEvents(self, command, data)

	case TouchStart, TouchEnd, TouchMove, TouchCancel:
		handleTouchEvents(self, command, data)

	case FocusEvent:
		view.hasFocus = true
		for _, listener := range getFocusListeners(view, "", command) {
			listener(self)
		}

	case LostFocusEvent:
		view.hasFocus = false
		for _, listener := range getFocusListeners(view, "", command) {
			listener(self)
		}

	case TransitionRunEvent, TransitionStartEvent, TransitionEndEvent, TransitionCancelEvent:
		view.handleTransitionEvents(command, data)

	case AnimationStartEvent, AnimationEndEvent, AnimationIterationEvent, AnimationCancelEvent:
		view.handleAnimationEvents(command, data)

	case "scroll":
		view.onScroll(view, dataFloatProperty(data, "x"), dataFloatProperty(data, "y"), dataFloatProperty(data, "width"), dataFloatProperty(data, "height"))

	case "widthChanged":
		if value, ok := data.PropertyValue("width"); ok {
			if width, ok := StringToSizeUnit(value); ok {
				self.setRaw(Width, width)
			}
		}

	case "heightChanged":
		if value, ok := data.PropertyValue("height"); ok {
			if height, ok := StringToSizeUnit(value); ok {
				self.setRaw(Height, height)
			}
		}

		/*
			case "resize":
				floatProperty := func(tag string) float64 {
					if value, ok := data.PropertyValue(tag); ok {
						if result, err := strconv.ParseFloat(value, 64); err == nil {
							return result
						}
					}
					return 0
				}

				self.onResize(self, floatProperty("x"), floatProperty("y"), floatProperty("width"), floatProperty("height"))
				return true
		*/
	default:
		return false
	}
	return true

}

func (view *viewData) SetChangeListener(tag string, listener func(View, string)) {
	if listener == nil {
		delete(view.changeListener, tag)
	} else {
		view.changeListener[tag] = listener
	}
}

func (view *viewData) HasFocus() bool {
	return view.hasFocus
}

func (view *viewData) String() string {
	return getViewString(view)
}
