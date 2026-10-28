// Copyright (C) 2025  T-Force I/O
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package tui

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"charm.land/bubbles/v2/progress"
	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/tforce-io/tf-golib/opx"
)

const (
	barWidth      = 50
	maxErrorLines = 3
)

// ProcessStatusMsg is used for updating state of ProcessStatus component.
type ProcessStatusMsg struct {
	action         string
	item           string
	itemPercent    float64
	totalInfinited bool
	totalPercent   float64
	errors         []string
}

// ProcessStatus is the Bubble Tea component that renders the 4-line progress UI.
type ProcessStatus struct {
	allowInterrupt bool

	action         string
	item           string
	itemPercent    float64
	totalPercent   float64
	totalInfinited bool
	errors         []string

	itemProgress  progress.Model
	itemSpinner   spinner.Model
	totalProgress progress.Model
}

// Return a new ProcessStatus instance.
func NewProcessStatus() *ProcessStatus {
	isDark := lipgloss.HasDarkBackground(os.Stdin, os.Stdout)
	textColor := lipgloss.Color("8")
	if isDark {
		textColor = lipgloss.Color("15")
	}

	item := progress.New(
		progress.WithColors(textColor),
		progress.WithFillCharacters(progress.DefaultFullCharFullBlock, progress.DefaultEmptyCharBlock),
		progress.WithWidth(barWidth),
	)

	spin := spinner.New(
		spinner.WithSpinner(spinner.Spinner{
			Frames: []string{".", "..", "...", "....", ".....", "......", ".......", ""},
			FPS:    time.Second / 4, //nolint:mnd
		}),
	)

	total := progress.New(
		progress.WithColors(textColor),
		progress.WithFillCharacters(progress.DefaultFullCharFullBlock, progress.DefaultEmptyCharBlock),
		progress.WithWidth(barWidth),
	)

	return &ProcessStatus{
		itemProgress:  item,
		itemSpinner:   spin,
		totalProgress: total,
	}
}

// Allow the ProcessStatus to signal the program to exit immediately.
func (m *ProcessStatus) WithAllowInterrupt(allow bool) *ProcessStatus {
	m.allowInterrupt = allow
	return m
}

// Display the ProcessStatus to the terminal.
func (m *ProcessStatus) Run(notifier *BubbleteaNotifier, onInterrupt func()) *TeaProgramHandle {
	p := tea.NewProgram(m)
	notifier.setProgram(p)

	done := make(chan error, 1)
	go func() {
		_, err := p.Run()
		if errors.Is(err, tea.ErrInterrupted) || errors.Is(err, tea.ErrProgramKilled) {
			if onInterrupt != nil {
				onInterrupt()
			}
		}
		done <- err
	}()
	return &TeaProgramHandle{
		program:  p,
		notifier: notifier,
		done:     done,
	}
}

// Bubbletea lifecycle implementation: Init.
func (m ProcessStatus) Init() tea.Cmd {
	return m.itemSpinner.Tick
}

// Bubbletea lifecycle implementation: Update.
func (m ProcessStatus) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		if msg.String() == "ctrl+c" || (msg.String() == "esc" && m.allowInterrupt) {
			return m, tea.Interrupt
		}

	case ProcessStatusMsg:
		m.action = msg.action
		m.item = msg.item
		m.itemPercent = msg.itemPercent
		m.totalInfinited = msg.totalInfinited
		m.totalPercent = msg.totalPercent
		m.errors = msg.errors

	case progress.FrameMsg:
		var cmdF, cmdO tea.Cmd
		fm, c := m.itemProgress.Update(msg)
		m.itemProgress = fm
		cmdF = c
		om, c2 := m.totalProgress.Update(msg)
		m.totalProgress = om
		cmdO = c2
		return m, tea.Batch(cmdF, cmdO)

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.itemSpinner, cmd = m.itemSpinner.Update(msg)
		return m, cmd

	case tea.QuitMsg:
		return m, tea.Quit
	}

	return m, nil
}

// Bubbletea lifecycle implementation: View.
func (m ProcessStatus) View() tea.View {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("%s%s\n", labelStyle.Render("Action: "), actionStyle.Render(opx.Ternary(m.action == "", "-", m.action))))
	sb.WriteString(fmt.Sprintf("%s%s\n", labelStyle.Render("Item:   "), itemStyle.Render(opx.Ternary(m.item == "", "-", m.item))))
	sb.WriteString(fmt.Sprintf("%s%s\n", labelStyle.Render(""), m.itemProgress.ViewAs(m.itemPercent)))
	if m.totalInfinited {
		if m.totalPercent >= 1 {
			sb.WriteString(fmt.Sprintf("%s\n", labelStyle.Render("Done.")))
		} else {
			sb.WriteString(fmt.Sprintf("%s%s\n", "Please wait", m.itemSpinner.View()))
		}
	} else {
		sb.WriteString(fmt.Sprintf("%s%s\n", labelStyle.Render(""), m.totalProgress.ViewAs(m.totalPercent)))
	}

	// Errors
	shown := m.errors
	if len(shown) > maxErrorLines {
		shown = shown[len(shown)-maxErrorLines:]
	}
	for _, e := range shown {
		if strings.HasPrefix(e, "[WARN]") {
			sb.WriteString(warningStyle.Render(e))
		} else {
			sb.WriteString(errorStyle.Render(e))
		}
		sb.WriteString("\n")
	}
	if m.allowInterrupt {
		sb.WriteString(fmt.Sprintf("\n%s interrupt  \n",
			shortcutStyle.Render(" esc "),
		))
	}

	return tea.NewView(sb.String())
}
