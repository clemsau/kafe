package utils

import "github.com/gdamore/tcell/v2"

func GetStatusColor(status string) tcell.Color {
	switch status {
	case "Ready":
		return tcell.ColorGreen
	case "Warning":
		return tcell.ColorYellow
	case "Error":
		return tcell.ColorRed
	default:
		return tcell.ColorWhite
	}
}
