package main

import (
	"fmt"
	"strings"
	"time"
)

func main() {
	const total = 100
	const barWidth = 30

	spinChars := []rune{'|', '/', '-', '\\'}
	secondarySpinChars := []rune{'.', 'o', 'O', '@', '*'}

	spinIndex := 0
	secondaryIndex := 0

	for i := 0; i <= total; i++ {
		// Primary spinner + progress bar
		percent := float64(i) / float64(total)
		filled := int(percent * barWidth)
		bar := fmt.Sprintf("[%s%s]", strings.Repeat("=", filled), strings.Repeat(" ", barWidth-filled))
		line1 := fmt.Sprintf("%c %s %3d%%", spinChars[spinIndex%len(spinChars)], bar, i)

		// Secondary spinner with bullet style and separate percentage
		subtaskPercent := i / 2 // e.g., simulate slower progress
		line2 := fmt.Sprintf("  %c Subtask: %3d%%", secondarySpinChars[secondaryIndex%len(secondarySpinChars)], subtaskPercent)

		// Clear and reprint both lines
		fmt.Print("\033[2K\r")  // clear current line
		fmt.Print(line1 + "\n") // print line 1
		fmt.Print("\033[2K\r")  // clear second line
		fmt.Print(line2 + "\n") // print line 2

		// Move cursor up to redraw in-place next iteration
		fmt.Print("\033[2A")

		spinIndex++
		if i%2 == 0 {
			secondaryIndex++
		}
		time.Sleep(50 * time.Millisecond)
	}

	// Move down and print done message
	fmt.Print("\033[2B")
	fmt.Println("✅ Done!")
}
