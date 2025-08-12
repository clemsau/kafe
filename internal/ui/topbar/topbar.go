package topbar

import (
	"strings"

	"github.com/clemsau/kafe/internal/ui/controls"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// TopBar represents the top bar with controls and logo
type TopBar struct {
	*tview.Flex
	controlsView *tview.TextView
	logoView     *tview.TextView
	controls     []controls.Control
}

// NewTopBar creates a new top bar with controls on left and ASCII logo on right
func NewTopBar(controlsList []controls.Control) *TopBar {
	topBar := &TopBar{
		Flex:     tview.NewFlex(),
		controls: controlsList,
	}

	// Create controls view
	topBar.controlsView = tview.NewTextView()
	topBar.setupControls()

	// Create logo view
	topBar.logoView = tview.NewTextView()
	topBar.setupLogo()

	// Set up the flex layout: controls on left, logo on right
	topBar.SetDirection(tview.FlexColumn).
		AddItem(topBar.controlsView, 0, 1, false).
		AddItem(topBar.logoView, 40, 0, false) // Fixed width for logo

	return topBar
}

// setupControls configures the controls display in vertical layout
func (t *TopBar) setupControls() {
	var controlsText strings.Builder

	// Find the maximum key length for alignment
	maxKeyLength := 0
	for _, control := range t.controls {
		if len(control.Key) > maxKeyLength {
			maxKeyLength = len(control.Key)
		}
	}

	// Add some padding at the top
	controlsText.WriteString("\n")

	// Create vertical list of controls with aligned descriptions
	for i, control := range t.controls {
		if i > 0 {
			controlsText.WriteString("\n")
		}
		controlsText.WriteString("  ") // Left padding

		// Format with aligned descriptions
		padding := strings.Repeat(" ", maxKeyLength-len(control.Key))
		controlsText.WriteString("[yellow]<" + control.Key + ">" + padding + "[white]    " + control.Description)
	}

	t.controlsView.SetText(controlsText.String())
	t.controlsView.SetTextColor(tcell.ColorWhite)
	t.controlsView.SetBackgroundColor(tcell.ColorBlack)
	t.controlsView.SetDynamicColors(true)
	t.controlsView.SetBorder(false)
	t.controlsView.SetWordWrap(false)
	t.controlsView.SetScrollable(false)
}

// setupLogo configures the ASCII art logo display
func (t *TopBar) setupLogo() {
	logo := strings.Join([]string{
		"██╗  ██╗ █████╗ ███████╗███████╗",
		"██║ ██╔╝██╔══██╗██╔════╝██╔════╝",
		"█████╔╝ ███████║█████╗  █████╗  ",
		"██╔═██╗ ██╔══██║██╔══╝  ██╔══╝  ",
		"██║  ██╗██║  ██║██║     ███████╗",
		"╚═╝  ╚═╝╚═╝  ╚═╝╚═╝     ╚══════╝",
	}, "\n")

	t.logoView.SetText(logo)
	t.logoView.SetTextAlign(tview.AlignRight)
	t.logoView.SetTextColor(tcell.ColorDarkCyan)
	t.logoView.SetBackgroundColor(tcell.ColorBlack)
	t.logoView.SetBorder(false)
	t.logoView.SetWordWrap(false)
	t.logoView.SetScrollable(false)
}

// SetControls updates the controls displayed in the bar
func (t *TopBar) SetControls(controlsList []controls.Control) {
	t.controls = controlsList
	t.setupControls()
}

// AddControl adds a single control to the bar
func (t *TopBar) AddControl(key, description string) {
	t.controls = append(t.controls, controls.Control{Key: key, Description: description})
	t.setupControls()
}

// GetHeight returns the recommended height for the top bar
func (t *TopBar) GetHeight() int {
	return 6 // Height needed for the ASCII logo
}
