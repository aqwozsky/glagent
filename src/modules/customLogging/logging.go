/*
How to use:
customlogging.SetTag("[GLAGENT]")
customlogging.SetShowTime(true)
customlogging.SetTimeFormat("15:04:05")

customlogging.Log(
	customlogging.INFO,
	"System started",
	customlogging.NoAnimation,
	customlogging.AnimationSpeedMedium,
	customlogging.ColorBlue,
	customlogging.ColorCyan,
)

customlogging.Log(
	customlogging.ERROR,
	"Connecting...",
	customlogging.Spinner,
	customlogging.AnimationSpeedFast,
	customlogging.ColorRed,
	customlogging.ColorYellow,
)

customlogging.TimeAnimated(
	"Clock running",
	5000,
	customlogging.AnimationSpeedMedium,
	customlogging.ColorGreen,
	customlogging.ColorMagenta,
)

*/

package customlogging

import (
	"fmt"
	"strings"
	"time"

	"github.com/fatih/color"
)

var tag = "[GLAGENT]"

var showTime = true
var timeFormat = "15:04:05"

type TextColor int

const (
	ColorDefault TextColor = iota
	ColorRed
	ColorYellow
	ColorGreen
	ColorCyan
	ColorBlue
	ColorMagenta
	ColorWhite
)

func (c TextColor) String() string {
	switch c {
	case ColorRed:
		return "Red"
	case ColorYellow:
		return "Yellow"
	case ColorGreen:
		return "Green"
	case ColorCyan:
		return "Cyan"
	case ColorBlue:
		return "Blue"
	case ColorMagenta:
		return "Magenta"
	case ColorWhite:
		return "White"
	default:
		return "Default"
	}
}

type LogType int

const (
	INFO LogType = iota
	SETUP
	WARN
	ERROR
	DEBUG
	RAINBOW
)

func (l LogType) String() string {
	switch l {
	case INFO:
		return "INFO"
	case SETUP:
		return "SETUP"
	case WARN:
		return "WARN"
	case ERROR:
		return "ERROR"
	case DEBUG:
		return "DEBUG"
	case RAINBOW:
		return "RAINBOW"
	default:
		return "UNKNOWN"
	}
}

type AnimationType int

const (
	NoAnimation AnimationType = iota
	Dots
	Spinner
	RainbowWave
	TimeFlow
)

func (a AnimationType) String() string {
	switch a {
	case NoAnimation:
		return "NoAnimation"
	case Dots:
		return "Dots"
	case Spinner:
		return "Spinner"
	case RainbowWave:
		return "RainbowWave"
	case TimeFlow:
		return "TimeFlow"
	default:
		return "UnknownAnimation"
	}
}

type AnimationSpeed int

const (
	AnimationSpeedSlow AnimationSpeed = iota
	AnimationSpeedMedium
	AnimationSpeedFast
)

func (s AnimationSpeed) String() string {
	switch s {
	case AnimationSpeedSlow:
		return "Slow"
	case AnimationSpeedMedium:
		return "Medium"
	case AnimationSpeedFast:
		return "Fast"
	default:
		return "UnknownSpeed"
	}
}

func speedToDuration(speed AnimationSpeed) time.Duration {
	switch speed {
	case AnimationSpeedSlow:
		return 150 * time.Millisecond
	case AnimationSpeedMedium:
		return 80 * time.Millisecond
	case AnimationSpeedFast:
		return 35 * time.Millisecond
	default:
		return 80 * time.Millisecond
	}
}

func SetTag(newTag string) {
	tag = newTag
}

func SetShowTime(enabled bool) {
	showTime = enabled
}

func SetTimeFormat(format string) {
	timeFormat = format
}

func currentTimeString() string {
	return time.Now().Format(timeFormat)
}

func buildPrefix(level LogType, timeColor TextColor) string {
	parts := []string{}

	if showTime {
		parts = append(parts, colorByChoice(timeColor, "["+currentTimeString()+"]"))
	}

	parts = append(parts, tag)
	parts = append(parts, "["+level.String()+"]")

	return strings.Join(parts, " ")
}

func Log(level LogType, message string, animType AnimationType, speed AnimationSpeed, animationColor TextColor, timeColor TextColor) {
	if level == RAINBOW {
		switch animType {
		case NoAnimation:
			fmt.Println(rainbow(message))
		case Dots:
			logWithDots(level, message, 3, speedToDuration(speed), animationColor)
		case Spinner:
			logWithSpinner(level, message, 12, speedToDuration(speed), animationColor)
		case RainbowWave:
			logWithRainbowWave(message, 18, speedToDuration(speed))
		case TimeFlow:
			logWithTimeFlow(message, 3000, speedToDuration(speed), animationColor, timeColor)
		default:
			fmt.Println(rainbow(message))
		}
		return
	}

	full := fmt.Sprintf("%s %s", buildPrefix(level, timeColor), message)

	switch animType {
	case NoAnimation:
		printByLevel(level, full)
	case Dots:
		logWithDots(level, full, 3, speedToDuration(speed), animationColor)
	case Spinner:
		logWithSpinner(level, full, 12, speedToDuration(speed), animationColor)
	case RainbowWave:
		logWithRainbowWave(full, 18, speedToDuration(speed))
	case TimeFlow:
		logWithTimeFlow(full, 3000, speedToDuration(speed), animationColor, timeColor)
	default:
		printByLevel(level, full)
	}
}

func Info(message string) {
	Log(INFO, message, NoAnimation, AnimationSpeedMedium, ColorDefault, ColorCyan)
}

func Setup(message string) {
	Log(SETUP, message, NoAnimation, AnimationSpeedMedium, ColorDefault, ColorCyan)
}

func Warn(message string) {
	Log(WARN, message, NoAnimation, AnimationSpeedMedium, ColorDefault, ColorYellow)
}

func Error(message string) {
	Log(ERROR, message, NoAnimation, AnimationSpeedMedium, ColorDefault, ColorRed)
}

func Debug(message string) {
	Log(DEBUG, message, NoAnimation, AnimationSpeedMedium, ColorDefault, ColorMagenta)
}

func Rainbow(message string) {
	Log(RAINBOW, message, NoAnimation, AnimationSpeedMedium, ColorDefault, ColorCyan)
}

func RainbowAnimated(message string, speed AnimationSpeed) {
	Log(RAINBOW, message, RainbowWave, speed, ColorDefault, ColorCyan)
}

func TimeAnimated(message string, durationMs int, speed AnimationSpeed, animationColor TextColor, timeColor TextColor) {
	logWithTimeFlow(message, durationMs, speedToDuration(speed), animationColor, timeColor)
}

func PrintBanner() {
	banner := `   ________    ___                    __ 
  / ____/ /   /   | ____ ____  ____  / /_
 / / __/ /   / /| |/ __ ` + "`" + `/ _ \/ __ \/ __/
/ /_/ / /___/ ___ / /_/ /  __/ / / / /_  
\____/_____/_/  |_\__, /\___/_/ /_/\__/  
                 /____/                   
`

	lines := strings.Split(banner, "\n")
	for _, line := range lines {
		fmt.Println(rainbow(line))
	}
}

func PrintAnimatedBanner(speed AnimationSpeed) {
	banner := `   ________    ___                    __ 
  / ____/ /   /   | ____ ____  ____  / /_
 / / __/ /   / /| |/ __ ` + "`" + `/ _ \/ __ \/ __/
/ /_/ / /___/ ___ / /_/ /  __/ / / / /_  
\____/_____/_/  |_\__, /\___/_/ /_/\__/  
                 /____/                   
`

	lines := strings.Split(banner, "\n")
	for _, line := range lines {
		logWithRainbowWave(line, 10, speedToDuration(speed))
	}
}

func RainbowFlow(text string, durationMs int, speed AnimationSpeed) {
	delay := speedToDuration(speed)
	offset := 0

	if durationMs == -1 {
		for {
			fmt.Print("\r" + rainbowShift(text, offset))
			time.Sleep(delay)
			offset++
		}
	}

	end := time.Now().Add(time.Duration(durationMs) * time.Millisecond)

	for time.Now().Before(end) {
		fmt.Print("\r" + rainbowShift(text, offset))
		time.Sleep(delay)
		offset++
	}

	fmt.Print("\r" + rainbow(text) + "\n")
}

func printByLevel(level LogType, text string) {
	fmt.Println(colorize(level, text))
}

func logWithDots(level LogType, text string, dotCount int, delay time.Duration, animationColor TextColor) {
	fmt.Print(colorize(level, text))

	for i := 0; i < dotCount; i++ {
		fmt.Print(colorByChoice(animationColor, "."))
		time.Sleep(delay)
	}

	fmt.Println()
}

func logWithSpinner(level LogType, text string, steps int, delay time.Duration, animationColor TextColor) {
	frames := []rune{'|', '/', '-', '\\'}

	for i := 0; i < steps; i++ {
		frame := frames[i%len(frames)]
		base := colorize(level, text)
		anim := colorByChoice(animationColor, string(frame))
		fmt.Print("\r" + base + " " + anim)
		time.Sleep(delay)
	}

	done := colorize(level, fmt.Sprintf("\r%s ✓\n", text))
	fmt.Print(done)
}

func logWithRainbowWave(text string, steps int, delay time.Duration) {
	for offset := 0; offset < steps; offset++ {
		fmt.Print("\r" + rainbowShift(text, offset))
		time.Sleep(delay)
	}
	fmt.Print("\r" + rainbow(text) + "\n")
}

func logWithTimeFlow(text string, durationMs int, delay time.Duration, animationColor TextColor, timeColor TextColor) {
	end := time.Now().Add(time.Duration(durationMs) * time.Millisecond)

	for time.Now().Before(end) {
		timePart := colorByChoice(timeColor, "["+currentTimeString()+"]")
		msgPart := colorByChoice(animationColor, text)
		fmt.Print("\r" + timePart + " " + msgPart)
		time.Sleep(delay)
	}

	timePart := colorByChoice(timeColor, "["+currentTimeString()+"]")
	msgPart := colorByChoice(animationColor, text)
	fmt.Print("\r" + timePart + " " + msgPart + "\n")
}

func colorize(level LogType, text string) string {
	switch level {
	case INFO:
		return color.GreenString(text)
	case SETUP:
		return color.CyanString(text)
	case WARN:
		return color.YellowString(text)
	case ERROR:
		return color.RedString(text)
	case DEBUG:
		return color.MagentaString(text)
	case RAINBOW:
		return rainbow(text)
	default:
		return text
	}
}

func colorByChoice(choice TextColor, text string) string {
	switch choice {
	case ColorRed:
		return color.RedString(text)
	case ColorYellow:
		return color.YellowString(text)
	case ColorGreen:
		return color.GreenString(text)
	case ColorCyan:
		return color.CyanString(text)
	case ColorBlue:
		return color.BlueString(text)
	case ColorMagenta:
		return color.MagentaString(text)
	case ColorWhite:
		return color.WhiteString(text)
	default:
		return text
	}
}

func rainbow(text string) string {
	colors := rainbowColors()

	var result strings.Builder
	visibleIndex := 0

	for _, r := range text {
		if r == ' ' {
			result.WriteRune(r)
			continue
		}

		colorFunc := colors[visibleIndex%len(colors)]
		result.WriteString(colorFunc("%c", r))
		visibleIndex++
	}

	return result.String()
}

func rainbowShift(text string, shift int) string {
	colors := rainbowColors()

	var result strings.Builder
	visibleIndex := 0

	for _, r := range text {
		if r == ' ' {
			result.WriteRune(r)
			continue
		}

		colorFunc := colors[(visibleIndex+shift)%len(colors)]
		result.WriteString(colorFunc("%c", r))
		visibleIndex++
	}

	return result.String()
}

func rainbowColors() []func(format string, a ...interface{}) string {
	return []func(format string, a ...interface{}) string{
		color.RedString,
		color.YellowString,
		color.GreenString,
		color.CyanString,
		color.BlueString,
		color.MagentaString,
	}
}