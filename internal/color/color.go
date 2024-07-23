package color

import "fmt"

const (
	ColorReset  = "\033[0m"
	ColorRed    = "\033[31m"
	ColorGreen  = "\033[32m"
	ColorYellow = "\033[33m"
	ColorBlue   = "\033[34m"
	ColorPurple = "\033[35m"
	ColorCyan   = "\033[36m"
	ColorWhite  = "\033[37m"
)

func PrintRed(message string) {
	fmt.Println(ColorRed + message + ColorReset)
}

func PrintGreen(message string) {
	fmt.Println(ColorGreen + message + ColorReset)
}

func PrintYellow(message string) {
	fmt.Println(ColorYellow + message + ColorReset)
}

func PrintBlue(message string) {
	fmt.Println(ColorBlue + message + ColorReset)
}

func PrintPurple(message string) {
	fmt.Println(ColorPurple + message + ColorReset)
}

func PrintCyan(message string) {
	fmt.Println(ColorCyan + message + ColorReset)
}

func PrintWhite(message string) {
	fmt.Println(ColorWhite + message + ColorReset)
}
