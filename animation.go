package rui

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

const (
	// AnimationTag is the constant for the "animation" property tag.
	// The "animation" property sets and starts animations.
	// Valid types of value are []Animation and Animation
	AnimationTag = "animation"

	// AnimationPause is the constant for the "animation-pause" property tag.
	// The "animation-pause" property sets whether an animation is running or paused.
	AnimationPaused = "animation-paused"

	// TransitionTag is the constant for the "transition" property tag.
	// The "transition" property sets transition animation of view properties.
	// Valid type of "transition" property value is Params. Valid type of Params value is Animation.
	Transition = "transition"

	// PropertyTag is the constant for the "property" animation property tag.
	// The "property" property describes a scenario for changing a View property.
	// Valid types of value are []AnimatedProperty and AnimatedProperty
	PropertyTag = "property"

	// Duration is the constant for the "duration" animation property tag.
	// The "duration" float property sets the length of time in seconds that an animation takes to complete one cycle.
	Duration = "duration"

	// Delay is the constant for the "delay" animation property tag.
	// The "delay" float property specifies the amount of time in seconds to wait from applying
	// the animation to an element before beginning to perform the animation. The animation can start later,
	// immediately from its beginning, or immediately and partway through the animation.
	Delay = "delay"

	// TimingFunction is the constant for the "timing-function" animation property tag.
	// The "timing-function" property sets how an animation progresses through the duration of each cycle.
	TimingFunction = "timing-function"

	// IterationCount is the constant for the "iteration-count" animation property tag.
	// The "iteration-count" int property sets the number of times an animation sequence
	// should be played before stopping.
	IterationCount = "iteration-count"

	// AnimationDirection is the constant for the "animation-direction" animation property tag.
	//The "animation-direction" property sets whether an animation should play forward, backward,
	// or alternate back and forth between playing the sequence forward and backward.
	AnimationDirection = "animation-direction"

	// NormalAnimation is value of the "animation-direction" property.
	// The animation plays forwards each cycle. In other words, each time the animation cycles,
	// the animation will reset to the beginning state and start over again. This is the default value.
	NormalAnimation = 0
	// ReverseAnimation is value of the "animation-direction" property.
	// The animation plays backwards each cycle. In other words, each time the animation cycles,
	// the animation will reset to the end state and start over again. Animation steps are performed
	// backwards, and timing functions are also reversed.
	// For example, an "ease-in" timing function becomes "ease-out".
	ReverseAnimation = 1
	// AlternateAnimation is value of the "animation-direction" property.
	// The animation reverses direction each cycle, with the first iteration being played forwards.
	// The count to determine if a cycle is even or odd starts at one.
	AlternateAnimation = 2
	// AlternateReverseAnimation is value of the "animation-direction" property.
	// The animation reverses direction each cycle, with the first iteration being played backwards.
	// The count to determine if a cycle is even or odd starts at one.
	AlternateReverseAnimation = 3

	// EaseTiming - a timing function which increases in velocity towards the middle of the transition, slowing back down at the end
	EaseTiming = "ease"
	// EaseInTiming - a timing function which starts off slowly, with the transition speed increasing until complete
	EaseInTiming = "ease-in"
	// EaseOutTiming - a timing function which starts transitioning quickly, slowing down the transition continues.
	EaseOutTiming = "ease-out"
	// EaseInOutTiming - a timing function which starts transitioning slowly, speeds up, and then slows down again.
	EaseInOutTiming = "ease-in-out"
	// LinearTiming - a timing function at an even speed
	LinearTiming = "linear"
)

// StepsTiming return a timing function along stepCount stops along the transition, diplaying each stop for equal lengths of time
func StepsTiming(stepCount int) string {
	return "steps(" + strconv.Itoa(stepCount) + ")"
}

// CubicBezierTiming return a cubic-Bezier curve timing function. x1 and x2 must be in the range [0, 1].
func CubicBezierTiming(x1, y1, x2, y2 float64) string {
	if x1 < 0 {
		x1 = 0
	} else if x1 > 1 {
		x1 = 1
	}
	if x2 < 0 {
		x2 = 0
	} else if x2 > 1 {
		x2 = 1
	}
	return fmt.Sprintf("cubic-bezier(%g, %g, %g, %g)", x1, y1, x2, y2)
}

// AnimatedProperty describes the change script of one property
type AnimatedProperty struct {
	// Tag is the name of the property
	Tag string
	// From is the initial value of the property
	From interface{}
	// To is the final value of the property
	To interface{}
	// KeyFrames is intermediate property values
	KeyFrames map[int]interface{}
}

type animationData struct {
	propertyList
	keyFramesName string
}

// Animation interface is used to set animation parameters. Used properties:
// "property", "id", "duration", "delay", "timing-function", "iteration-count", and "animation-direction"
type Animation interface {
	Properties
	fmt.Stringer
	writeTransitionString(tag string, buffer *strings.Builder)
	animationCSS(session Session) string
	transitionCSS(buffer *strings.Builder, session Session)
	hasAnimatedPropery() bool
	animationName() string
}

func parseAnimation(obj DataObject) Animation {
	animation := new(animationData)
	animation.init()

	for i := 0; i < obj.PropertyCount(); i++ {
		if node := obj.Property(i); node != nil {
			if node.Type() == TextNode {
				animation.Set(node.Tag(), node.Text())
			} else {
				animation.Set(node.Tag(), node)
			}
		}
	}
	return animation
}

func NewAnimation(params Params) Animation {
	animation := new(animationData)
	animation.init()

	for tag, value := range params {
		animation.Set(tag, value)
	}
	return animation
}

func (animation *animationData) hasAnimatedPropery() bool {
	props := animation.getRaw(PropertyTag)
	if props == nil {
		ErrorLog("There are no animated properties.")
		return false
	}

	if _, ok := props.([]AnimatedProperty); !ok {
		ErrorLog("Invalid animated properties.")
		return false
	}

	return true
}

func (animation *animationData) animationName() string {
	return animation.keyFramesName
}

func (animation *animationData) normalizeTag(tag string) string {
	tag = strings.ToLower(tag)
	if tag == Direction {
		return AnimationDirection
	}
	return tag
}

func (animation *animationData) Set(tag string, value interface{}) bool {
	if value == nil {
		animation.Remove(tag)
		return true
	}

	switch tag = animation.normalizeTag(tag); tag {
	case ID:
		if text, ok := value.(string); ok {
			text = strings.Trim(text, " \t\n\r")
			if text == "" {
				delete(animation.properties, tag)
			} else {
				animation.properties[tag] = text
			}
			return true
		}
		notCompatibleType(tag, value)
		return false

	case PropertyTag:
		switch value := value.(type) {
		case AnimatedProperty:
			if value.From == nil && value.KeyFrames != nil {
				if val, ok := value.KeyFrames[0]; ok {
					value.From = val
					delete(value.KeyFrames, 0)
				}
			}
			if value.To == nil && value.KeyFrames != nil {
				if val, ok := value.KeyFrames[100]; ok {
					value.To = val
					delete(value.KeyFrames, 100)
				}
			}

			if value.From == nil {
				ErrorLog("AnimatedProperty.From is nil")
			} else if value.To == nil {
				ErrorLog("AnimatedProperty.To is nil")
			} else {
				animation.properties[tag] = []AnimatedProperty{value}
				return true
			}

		case []AnimatedProperty:
			props := []AnimatedProperty{}
			for _, val := range value {
				if val.From == nil && val.KeyFrames != nil {
					if v, ok := val.KeyFrames[0]; ok {
						val.From = v
						delete(val.KeyFrames, 0)
					}
				}
				if val.To == nil && val.KeyFrames != nil {
					if v, ok := val.KeyFrames[100]; ok {
						val.To = v
						delete(val.KeyFrames, 100)
					}
				}

				if val.From == nil {
					ErrorLog("AnimatedProperty.From is nil")
				} else if val.To == nil {
					ErrorLog("AnimatedProperty.To is nil")
				} else {
					props = append(props, val)
				}
			}
			if len(props) > 0 {
				animation.properties[tag] = props
				return true
			} else {
				ErrorLog("[]AnimatedProperty is empty")
			}

		case DataNode:
			parseObject := func(obj DataObject) (AnimatedProperty, bool) {
				result := AnimatedProperty{}
				for i := 0; i < obj.PropertyCount(); i++ {
					if node := obj.Property(i); node.Type() == TextNode {
						propTag := strings.ToLower(node.Tag())
						switch propTag {
						case "from", "0", "0%":
							result.From = node.Text()

						case "to", "100", "100%":
							result.To = node.Text()

						default:
							tagLen := len(propTag)
							if tagLen > 0 && propTag[tagLen-1] == '%' {
								propTag = propTag[:tagLen-1]
							}
							n, err := strconv.Atoi(propTag)
							if err != nil {
								ErrorLog(err.Error())
							} else if n < 0 || n > 100 {
								ErrorLogF(`key-frame "%d" is out of range`, n)
							} else {
								if result.KeyFrames == nil {
									result.KeyFrames = map[int]interface{}{n: node.Text()}
								} else {
									result.KeyFrames[n] = node.Text()
								}
							}
						}
					}
				}
				if result.From != nil && result.To != nil {
					return result, true
				}
				return result, false
			}

			switch value.Type() {
			case ObjectNode:
				if prop, ok := parseObject(value.Object()); ok {
					animation.properties[tag] = []AnimatedProperty{prop}
					return true
				}

			case ArrayNode:
				props := []AnimatedProperty{}
				for _, val := range value.ArrayElements() {
					if val.IsObject() {
						if prop, ok := parseObject(val.Object()); ok {
							props = append(props, prop)
						}
					} else {
						notCompatibleType(tag, val)
					}
				}
				if len(props) > 0 {
					animation.properties[tag] = props
					return true
				}

			default:
				notCompatibleType(tag, value)
			}

		default:
			notCompatibleType(tag, value)
		}

	case Duration:
		return animation.setFloatProperty(tag, value, 0, math.MaxFloat64)

	case Delay:
		return animation.setFloatProperty(tag, value, -math.MaxFloat64, math.MaxFloat64)

	case TimingFunction:
		if text, ok := value.(string); ok {
			animation.properties[tag] = text
			return true
		}

	case IterationCount:
		return animation.setIntProperty(tag, value)

	case AnimationDirection:
		return animation.setEnumProperty(AnimationDirection, value, enumProperties[AnimationDirection].values)

	default:
		ErrorLogF(`The "%s" property is not supported by Animation`, tag)
	}

	return false
}

func (animation *animationData) Remove(tag string) {
	delete(animation.properties, animation.normalizeTag(tag))
}

func (animation *animationData) Get(tag string) interface{} {
	return animation.getRaw(animation.normalizeTag(tag))
}

func (animation *animationData) String() string {
	buffer := allocStringBuilder()
	defer freeStringBuilder(buffer)

	buffer.WriteString("animation {")

	// TODO

	buffer.WriteString("}")
	return buffer.String()
}

func (animation *animationData) animationCSS(session Session) string {
	if animation.keyFramesName == "" {
		props := animation.getRaw(PropertyTag)
		if props == nil {
			ErrorLog("There are no animated properties.")
			return ""
		}

		animatedProps, ok := props.([]AnimatedProperty)
		if !ok {
			ErrorLog("Invalid animated properties.")
			return ""
		}

		animation.keyFramesName = session.registerAnimation(animatedProps)
	}

	buffer := allocStringBuilder()
	defer freeStringBuilder(buffer)

	buffer.WriteString(animation.keyFramesName)

	if duration, _ := floatProperty(animation, Duration, session, 1); duration > 0 {
		buffer.WriteString(fmt.Sprintf(" %gs ", duration))
	} else {
		buffer.WriteString(" 1s ")
	}

	buffer.WriteString(animation.timingFunctionCSS(session))

	if delay, _ := floatProperty(animation, Delay, session, 0); delay > 0 {
		buffer.WriteString(fmt.Sprintf(" %gs", delay))
	} else {
		buffer.WriteString(" 0s")
	}

	if iterationCount, _ := intProperty(animation, IterationCount, session, 0); iterationCount >= 0 {
		if iterationCount == 0 {
			iterationCount = 1
		}
		buffer.WriteString(fmt.Sprintf(" %d ", iterationCount))
	} else {
		buffer.WriteString(" infinite ")
	}

	direction, _ := enumProperty(animation, AnimationDirection, session, 0)
	values := enumProperties[AnimationDirection].cssValues
	if direction < 0 || direction >= len(values) {
		direction = 0
	}
	buffer.WriteString(values[direction])

	// TODO "animation-fill-mode"
	buffer.WriteString(" forwards")

	return buffer.String()
}

func (animation *animationData) transitionCSS(buffer *strings.Builder, session Session) {

	if duration, _ := floatProperty(animation, Duration, session, 1); duration > 0 {
		buffer.WriteString(fmt.Sprintf(" %gs ", duration))
	} else {
		buffer.WriteString(" 1s ")
	}

	buffer.WriteString(animation.timingFunctionCSS(session))

	if delay, _ := floatProperty(animation, Delay, session, 0); delay > 0 {
		buffer.WriteString(fmt.Sprintf(" %gs", delay))
	}
}

func (animation *animationData) writeTransitionString(tag string, buffer *strings.Builder) {
	buffer.WriteString(tag)
	buffer.WriteString("{")
	lead := " "

	writeFloatProperty := func(name string) bool {
		if value := animation.getRaw(name); value != nil {
			buffer.WriteString(lead)
			buffer.WriteString(name)
			buffer.WriteString(" = ")
			writePropertyValue(buffer, name, value, "")
			lead = ", "
			return true
		}
		return false
	}

	if !writeFloatProperty(Duration) {
		buffer.WriteString(" duration = 1")
		lead = ", "
	}

	writeFloatProperty(Delay)

	if value := animation.getRaw(TimingFunction); value != nil {
		if timingFunction, ok := value.(string); ok && timingFunction != "" {
			buffer.WriteString(lead)
			buffer.WriteString(TimingFunction)
			buffer.WriteString(" = ")
			if strings.ContainsAny(timingFunction, " ,()") {
				buffer.WriteRune('"')
				buffer.WriteString(timingFunction)
				buffer.WriteRune('"')
			}
		}
	}

	buffer.WriteString(" }")
}

func (animation *animationData) timingFunctionCSS(session Session) string {
	if timingFunction, ok := stringProperty(animation, TimingFunction, session); ok {
		if timingFunction, ok = session.resolveConstants(timingFunction); ok && validateTimingFunction(timingFunction) {
			return timingFunction
		}
	}
	return ("ease")
}

func validateTimingFunction(timingFunction string) bool {
	switch timingFunction {
	case "", EaseTiming, EaseInTiming, EaseOutTiming, EaseInOutTiming, LinearTiming:
		return true
	}

	size := len(timingFunction)
	if size > 0 && timingFunction[size-1] == ')' {
		if index := strings.IndexRune(timingFunction, '('); index > 0 {
			args := timingFunction[index+1 : size-1]
			switch timingFunction[:index] {
			case "steps":
				if _, err := strconv.Atoi(strings.Trim(args, " \t\n")); err == nil {
					return true
				}

			case "cubic-bezier":
				if params := strings.Split(args, ","); len(params) == 4 {
					for _, param := range params {
						if _, err := strconv.ParseFloat(strings.Trim(param, " \t\n"), 64); err != nil {
							return false
						}
					}
					return true
				}
			}
		}
	}

	return false
}

func (session *sessionData) registerAnimation(props []AnimatedProperty) string {

	session.animationCounter++
	name := fmt.Sprintf("kf%06d", session.animationCounter)

	var cssBuilder cssStyleBuilder

	cssBuilder.startAnimation(name)

	fromParams := Params{}
	toParams := Params{}
	frames := []int{}

	for _, prop := range props {
		fromParams[prop.Tag] = prop.From
		toParams[prop.Tag] = prop.To
		if len(prop.KeyFrames) > 0 {
			for frame := range prop.KeyFrames {
				needAppend := true
				for i, n := range frames {
					if n == frame {
						needAppend = false
						break
					} else if frame < n {
						needAppend = false
						frames = append(append(frames[:i], frame), frames[i+1:]...)
						break
					}
				}
				if needAppend {
					frames = append(frames, frame)
				}
			}
		}
	}

	cssBuilder.startAnimationFrame("from")
	NewViewStyle(fromParams).cssViewStyle(&cssBuilder, session)
	cssBuilder.endAnimationFrame()

	if len(frames) > 0 {
		for _, frame := range frames {
			params := Params{}
			for _, prop := range props {
				if prop.KeyFrames != nil {
					if value, ok := prop.KeyFrames[frame]; ok {
						params[prop.Tag] = value
					}
				}
			}

			if len(params) > 0 {
				cssBuilder.startAnimationFrame(strconv.Itoa(frame) + "%")
				NewViewStyle(params).cssViewStyle(&cssBuilder, session)
				cssBuilder.endAnimationFrame()
			}
		}
	}

	cssBuilder.startAnimationFrame("to")
	NewViewStyle(toParams).cssViewStyle(&cssBuilder, session)
	cssBuilder.endAnimationFrame()

	cssBuilder.endAnimation()

	style := cssBuilder.finish()
	session.animationCSS += style
	style = strings.ReplaceAll(style, "\n", `\n`)
	session.runScript(`document.querySelector('style').textContent += "` + style + `"`)

	return name
}

func (view *viewData) SetAnimated(tag string, value interface{}, animation Animation) bool {
	if animation == nil {
		return view.Set(tag, value)
	}

	updateProperty(view.htmlID(), "ontransitionend", "transitionEndEvent(this, event)", view.session)
	updateProperty(view.htmlID(), "ontransitioncancel", "transitionCancelEvent(this, event)", view.session)

	if prevAnimation, ok := view.transitions[tag]; ok {
		view.singleTransition[tag] = prevAnimation
	} else {
		view.singleTransition[tag] = nil
	}
	view.transitions[tag] = animation
	view.updateTransitionCSS()

	result := view.Set(tag, value)
	if !result {
		delete(view.singleTransition, tag)
		view.updateTransitionCSS()
	}

	return result
}

func (style *viewStyle) animationCSS(session Session) string {
	if value := style.getRaw(AnimationTag); value != nil {
		if animations, ok := value.([]Animation); ok {
			buffer := allocStringBuilder()
			defer freeStringBuilder(buffer)

			for _, animation := range animations {
				if css := animation.animationCSS(session); css != "" {
					if buffer.Len() > 0 {
						buffer.WriteString(", ")
					}
					buffer.WriteString(css)
				}
			}

			return buffer.String()
		}
	}

	return ""
}

func (style *viewStyle) transitionCSS(session Session) string {
	buffer := allocStringBuilder()
	defer freeStringBuilder(buffer)

	for tag, animation := range style.transitions {
		if buffer.Len() > 0 {
			buffer.WriteString(", ")
		}
		buffer.WriteString(tag)
		animation.transitionCSS(buffer, session)
	}
	return buffer.String()
}

func (view *viewData) updateTransitionCSS() {
	updateCSSProperty(view.htmlID(), "transition", view.transitionCSS(view.Session()), view.Session())
}

func (view *viewData) getTransitions() Params {
	result := Params{}
	for tag, animation := range view.transitions {
		result[tag] = animation
	}
	return result
}

// SetAnimated sets the property with name "tag" of the "rootView" subview with "viewID" id by value. Result:
// true - success,
// false - error (incompatible type or invalid format of a string value, see AppLog).
func SetAnimated(rootView View, viewID, tag string, value interface{}, animation Animation) bool {
	if view := ViewByID(rootView, viewID); view != nil {
		return view.SetAnimated(tag, value, animation)
	}
	return false
}

// IsAnimationPaused returns "true" if an animation of the subview is paused, "false" otherwise.
// If the second argument (subviewID) is "" then a value from the first argument (view) is returned.
func IsAnimationPaused(view View, subviewID string) bool {
	if subviewID != "" {
		view = ViewByID(view, subviewID)
	}
	if view != nil {
		if result, ok := boolStyledProperty(view, AnimationPaused); ok {
			return result
		}
	}
	return false
}

// GetTransition returns the subview transitions. The result is always non-nil.
// If the second argument (subviewID) is "" then transitions of the first argument (view) is returned
func GetTransition(view View, subviewID string) Params {
	if subviewID != "" {
		view = ViewByID(view, subviewID)
	}

	if view != nil {
		return view.getTransitions()
	}

	return Params{}
}

// AddTransition adds the transition for the subview property.
// If the second argument (subviewID) is "" then the transition is added to the first argument (view)
func AddTransition(view View, subviewID, tag string, animation Animation) bool {
	if tag == "" {
		return false
	}

	if subviewID != "" {
		view = ViewByID(view, subviewID)
	}

	if view == nil {
		return false
	}

	transitions := view.getTransitions()
	transitions[tag] = animation
	return view.Set(Transition, transitions)
}

// GetAnimation returns the subview animations. The result is always non-nil.
// If the second argument (subviewID) is "" then transitions of the first argument (view) is returned
func GetAnimation(view View, subviewID string) []Animation {
	if subviewID != "" {
		view = ViewByID(view, subviewID)
	}

	if view != nil {
		if value := view.getRaw(AnimationTag); value != nil {
			if animations, ok := value.([]Animation); ok && animations != nil {
				return animations
			}
		}
	}

	return []Animation{}
}
