// Copyright (C) 2025 T-Force I/O
// This file is part of TFunifiler
//
// TFunifiler is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// TFunifiler is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with TFunifiler. If not, see <https://www.gnu.org/licenses/>.

package tui

import "charm.land/lipgloss/v2"

var (
	styleAction   = lipgloss.NewStyle().Bold(true)
	styleItem     = lipgloss.NewStyle().Bold(true)
	styleLabel    = lipgloss.NewStyle().Bold(true)
	styleShortcut = lipgloss.NewStyle().Bold(true).Reverse(true)
	styleError    = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
	styleWarning  = lipgloss.NewStyle().Foreground(lipgloss.Color("11"))
)
