// Package tui provides a terminal user interface for synthtribe2midi
package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/filepicker"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/james-see/synthtribe2midi/pkg/converter"
	"github.com/james-see/synthtribe2midi/pkg/converter/devices"
)

// Acid-inspired color scheme (303/acid aesthetic)
var (
	// Primary colors - acid green and silver
	acidGreen  = lipgloss.Color("#39FF14")
	acidYellow = lipgloss.Color("#FFFF00")
	silverGray = lipgloss.Color("#C0C0C0")
	darkGray   = lipgloss.Color("#333333")
	
	// Styles
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(acidGreen).
			Background(darkGray).
			Padding(0, 2).
			MarginBottom(1)
	
	menuStyle = lipgloss.NewStyle().
			Foreground(silverGray).
			PaddingLeft(2)
	
	selectedStyle = lipgloss.NewStyle().
			Foreground(acidGreen).
			Bold(true).
			PaddingLeft(2)
	
	statusStyle = lipgloss.NewStyle().
			Foreground(acidYellow).
			PaddingTop(1)
	
	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF0000")).
			Bold(true)
	
	successStyle = lipgloss.NewStyle().
			Foreground(acidGreen).
			Bold(true)
	
	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#666666")).
			MarginTop(1)
	
	boxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(acidGreen).
			Padding(1, 2)
)

// State represents the current TUI state
type State int

const (
	StateMenu State = iota
	StateFilePicker
	StateConverting
	StateResult
)

// MenuItem represents a menu option
type MenuItem struct {
	Title       string
	Description string
	FromFormat  string
	ToFormat    string
}

var menuItems = []MenuItem{
	{Title: "MIDI → SEQ", Description: "Convert MIDI file to SynthTribe .seq pattern", FromFormat: "midi", ToFormat: "seq"},
	{Title: "SEQ → MIDI", Description: "Convert SynthTribe .seq pattern to MIDI file", FromFormat: "seq", ToFormat: "midi"},
	{Title: "MIDI → SYX", Description: "Convert MIDI file to SysEx dump", FromFormat: "midi", ToFormat: "syx"},
	{Title: "SYX → MIDI", Description: "Convert SysEx dump to MIDI file", FromFormat: "syx", ToFormat: "midi"},
	{Title: "SEQ → SYX", Description: "Convert .seq pattern to SysEx dump", FromFormat: "seq", ToFormat: "syx"},
	{Title: "SYX → SEQ", Description: "Convert SysEx dump to .seq pattern", FromFormat: "syx", ToFormat: "seq"},
	{Title: "Exit", Description: "Exit the application", FromFormat: "", ToFormat: ""},
}

// Model represents the TUI model
type Model struct {
	state        State
	menuIndex    int
	filePicker   filepicker.Model
	spinner      spinner.Model
	selectedFile string
	outputFile   string
	conversion   MenuItem
	err          error
	width        int
	height       int
}

// conversionDoneMsg signals conversion completion
type conversionDoneMsg struct {
	outputFile string
	err        error
}

// Init initializes the TUI model
func (m Model) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick)
}

// New creates a new TUI model
func New() Model {
	// Initialize file picker
	fp := filepicker.New()
	fp.AllowedTypes = []string{".mid", ".midi", ".seq", ".syx"}
	fp.CurrentDirectory, _ = os.Getwd()
	
	// Initialize spinner
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(acidGreen)
	
	return Model{
		state:      StateMenu,
		menuIndex:  0,
		filePicker: fp,
		spinner:    s,
	}
}

// Update handles TUI updates
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Handle file picker state first - it needs to receive all messages
	if m.state == StateFilePicker {
		// Check for escape/quit keys first
		if keyMsg, ok := msg.(tea.KeyMsg); ok {
			switch keyMsg.String() {
			case "esc":
				m.state = StateMenu
				return m, nil
			case "q", "ctrl+c":
				return m, tea.Quit
			}
		}

		// Pass all other messages to the file picker
		var cmd tea.Cmd
		m.filePicker, cmd = m.filePicker.Update(msg)

		// Check if file was selected
		if didSelect, path := m.filePicker.DidSelectFile(msg); didSelect {
			m.selectedFile = path
			m.state = StateConverting
			return m, tea.Batch(m.spinner.Tick, m.performConversion())
		}

		return m, cmd
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.filePicker.SetHeight(msg.Height - 10)
		return m, nil

	case tea.KeyMsg:
		switch m.state {
		case StateMenu:
			return m.updateMenu(msg)
		case StateResult:
			return m.updateResult(msg)
		}

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case conversionDoneMsg:
		m.state = StateResult
		m.outputFile = msg.outputFile
		m.err = msg.err
		return m, nil
	}

	return m, nil
}

func (m Model) updateMenu(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.menuIndex > 0 {
			m.menuIndex--
		}
	case "down", "j":
		if m.menuIndex < len(menuItems)-1 {
			m.menuIndex++
		}
	case "enter":
		if m.menuIndex == len(menuItems)-1 {
			return m, tea.Quit
		}
		m.conversion = menuItems[m.menuIndex]
		m.state = StateFilePicker
		
		// Set file picker filter based on input format
		switch m.conversion.FromFormat {
		case "midi":
			m.filePicker.AllowedTypes = []string{".mid", ".midi"}
		case "seq":
			m.filePicker.AllowedTypes = []string{".seq"}
		case "syx":
			m.filePicker.AllowedTypes = []string{".syx"}
		}
		
		return m, m.filePicker.Init()
	case "q", "ctrl+c":
		return m, tea.Quit
	}
	return m, nil
}

func (m Model) updateResult(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter", "esc":
		m.state = StateMenu
		m.err = nil
		m.selectedFile = ""
		m.outputFile = ""
		return m, nil
	case "q", "ctrl+c":
		return m, tea.Quit
	}
	return m, nil
}

func (m Model) performConversion() tea.Cmd {
	return func() tea.Msg {
		device := devices.NewTD3()
		conv := converter.New(device)
		
		data, err := os.ReadFile(m.selectedFile)
		if err != nil {
			return conversionDoneMsg{err: err}
		}
		
		var result []byte
		var outputExt string
		
		switch m.conversion.FromFormat + "2" + m.conversion.ToFormat {
		case "midi2seq":
			result, err = conv.MIDIToSeq(data)
			outputExt = ".seq"
		case "seq2midi":
			result, err = conv.SeqToMIDI(data)
			outputExt = ".mid"
		case "midi2syx":
			result, err = conv.MIDIToSyx(data)
			outputExt = ".syx"
		case "syx2midi":
			result, err = conv.SyxToMIDI(data)
			outputExt = ".mid"
		case "seq2syx":
			result, err = conv.SeqToSyx(data)
			outputExt = ".syx"
		case "syx2seq":
			result, err = conv.SyxToSeq(data)
			outputExt = ".seq"
		}
		
		if err != nil {
			return conversionDoneMsg{err: err}
		}
		
		// Generate output filename
		base := strings.TrimSuffix(m.selectedFile, filepath.Ext(m.selectedFile))
		outputFile := base + outputExt
		
		err = os.WriteFile(outputFile, result, 0644)
		if err != nil {
			return conversionDoneMsg{err: err}
		}
		
		return conversionDoneMsg{outputFile: outputFile}
	}
}

// View renders the TUI
func (m Model) View() string {
	var s strings.Builder
	
	// Header
	header := asciiLogo()
	s.WriteString(header)
	s.WriteString("\n")
	
	switch m.state {
	case StateMenu:
		s.WriteString(m.viewMenu())
	case StateFilePicker:
		s.WriteString(m.viewFilePicker())
	case StateConverting:
		s.WriteString(m.viewConverting())
	case StateResult:
		s.WriteString(m.viewResult())
	}
	
	// Footer help
	s.WriteString("\n")
	s.WriteString(helpStyle.Render("↑/↓: navigate • enter: select • q: quit"))
	
	return s.String()
}

func (m Model) viewMenu() string {
	var s strings.Builder
	
	s.WriteString(titleStyle.Render(" SELECT CONVERSION "))
	s.WriteString("\n\n")
	
	for i, item := range menuItems {
		if i == m.menuIndex {
			s.WriteString(selectedStyle.Render(fmt.Sprintf("▸ %s", item.Title)))
			s.WriteString("\n")
			s.WriteString(lipgloss.NewStyle().Foreground(acidYellow).PaddingLeft(4).Render(item.Description))
		} else {
			s.WriteString(menuStyle.Render(fmt.Sprintf("  %s", item.Title)))
		}
		s.WriteString("\n")
	}
	
	return boxStyle.Render(s.String())
}

func (m Model) viewFilePicker() string {
	var s strings.Builder
	
	s.WriteString(titleStyle.Render(fmt.Sprintf(" SELECT %s FILE ", strings.ToUpper(m.conversion.FromFormat))))
	s.WriteString("\n\n")
	s.WriteString(m.filePicker.View())
	s.WriteString("\n")
	s.WriteString(helpStyle.Render("esc: back to menu"))
	
	return s.String()
}

func (m Model) viewConverting() string {
	var s strings.Builder
	
	s.WriteString(titleStyle.Render(" CONVERTING "))
	s.WriteString("\n\n")
	s.WriteString(fmt.Sprintf("%s Converting %s...\n", m.spinner.View(), filepath.Base(m.selectedFile)))
	s.WriteString(statusStyle.Render(fmt.Sprintf("  %s → %s", m.conversion.FromFormat, m.conversion.ToFormat)))
	
	return boxStyle.Render(s.String())
}

func (m Model) viewResult() string {
	var s strings.Builder
	
	if m.err != nil {
		s.WriteString(titleStyle.Render(" ERROR "))
		s.WriteString("\n\n")
		s.WriteString(errorStyle.Render(fmt.Sprintf("✗ Conversion failed: %s", m.err.Error())))
	} else {
		s.WriteString(titleStyle.Render(" SUCCESS "))
		s.WriteString("\n\n")
		s.WriteString(successStyle.Render("✓ Conversion complete!"))
		s.WriteString("\n\n")
		s.WriteString(fmt.Sprintf("Input:  %s\n", filepath.Base(m.selectedFile)))
		s.WriteString(fmt.Sprintf("Output: %s", filepath.Base(m.outputFile)))
	}
	
	s.WriteString("\n\n")
	s.WriteString(helpStyle.Render("Press enter to continue"))
	
	return boxStyle.Render(s.String())
}

func asciiLogo() string {
	logo := `
   _____ _   _ _   _ _____ _   _ _____ ____  ___ ____  _____ ____  __  __ ___ ____ ___ 
  / ____| \ | | \ | |_   _| | | |_   _|  _ \|_ _| __ )| ____|___ \|  \/  |_ _|  _ \_ _|
  \___ \|  \| |  \| | | | | |_| | | | | |_) || ||  _ \|  _|   __) | |\/| || || | | | | 
   ___) | |\  | |\  | | | |  _  | | | |  _ < | || |_) | |___ / __/| |  | || || |_| | | 
  |____/|_| \_|_| \_| |_| |_| |_| |_| |_| \_\___|____/|_____|_____|_|  |_|___|____/___|
`
	return lipgloss.NewStyle().Foreground(acidGreen).Render(logo)
}

// Run starts the TUI application
func Run() error {
	p := tea.NewProgram(New(), tea.WithAltScreen())
	_, err := p.Run()
	return err
}

