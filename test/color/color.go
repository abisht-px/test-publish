// color is used for colorizing terminal output
package color

type color string

const (
	reset  color = "\033[0m"
	red    color = "\033[31m"
	green  color = "\033[32m"
	yellow color = "\033[33m"
	blue   color = "\033[34m"
	purple color = "\033[35m"
	cyan   color = "\033[36m"
	gray   color = "\033[37m"
	white  color = "\033[97m"
)

var (
	Red    = newColor(red)
	Green  = newColor(green)
	Yellow = newColor(yellow)
	Blue   = newColor(blue)
	Purple = newColor(purple)
	Cyan   = newColor(cyan)
	Gray   = newColor(gray)
	White  = newColor(white)
)

func newColor(c color) func(string) string {
	return func(s string) string {
		return colorize(s, c)
	}
}

func colorize(input string, c color) string {
	return string(c) + input + string(reset)
}
