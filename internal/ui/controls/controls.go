package controls

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// Control represents a single keyboard control
type Control struct {
	Key         string
	Description string
}

// ControlsBar displays available keyboard controls at the top of a view
type ControlsBar struct {
	*tview.TextView
	controls []Control
}

// NewControlsBar creates a new controls display bar
func NewControlsBar(controls []Control) *ControlsBar {
	controlsBar := &ControlsBar{
		TextView: tview.NewTextView(),
		controls: controls,
	}

	controlsBar.SetBorder(false)
	controlsBar.SetWordWrap(false)
	controlsBar.SetScrollable(false)
	controlsBar.SetDynamicColors(true)

	controlsBar.updateDisplay()
	return controlsBar
}

// SetControls updates the controls displayed in the bar
func (c *ControlsBar) SetControls(controls []Control) {
	c.controls = controls
	c.updateDisplay()
}

// AddControl adds a single control to the bar
func (c *ControlsBar) AddControl(key, description string) {
	c.controls = append(c.controls, Control{Key: key, Description: description})
	c.updateDisplay()
}

// updateDisplay refreshes the controls display
func (c *ControlsBar) updateDisplay() {
	var controlsText strings.Builder

	// Create a formatted string with all controls
	for i, control := range c.controls {
		if i > 0 {
			controlsText.WriteString("  ")
		}
		controlsText.WriteString(fmt.Sprintf("[yellow]%s[white]: %s", control.Key, control.Description))
	}

	c.SetText(controlsText.String()).
		SetTextColor(tcell.ColorWhite).
		SetBackgroundColor(tcell.ColorBlack)
}

// GetHeight returns the recommended height for the controls bar
func (c *ControlsBar) GetHeight() int {
	return 1
}
