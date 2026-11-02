package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Focus area constants
const (
	focusTables = iota
	focusSQL
	focusContent
)

// Model is the main application model.
type Model struct {
	db *DB

	// Window dimensions
	width  int
	height int

	// Focus
	focus int

	// Table panel
	tables       []string
	tableCursor  int
	tableErr     error
	selectedTable string

	// SQL panel
	sqlInput  textarea.Model
	sqlResult *QueryResult
	sqlErr    error

	// Content panel
	content        *QueryResult
	contentScrollX int // horizontal scroll offset in columns
	contentScrollY int // vertical scroll offset in rows
	totalRows      int // actual total row count (may be > len(content.Rows) if truncated)
}

// NewModel creates a new application model.
func NewModel(db *DB) Model {
	tables, err := db.TableNames()
	if err != nil {
		// Store error to show later
		tables = nil
	}

	ta := textarea.New()
	ta.Placeholder = "Enter SQL query..."
	ta.SetWidth(60)
	ta.SetHeight(4)
	ta.ShowLineNumbers = false
	ta.CharLimit = 0
	
	// Style the textarea
	ta.FocusedStyle.CursorLine = lipgloss.NewStyle().Foreground(colorText)
	ta.BlurredStyle.CursorLine = lipgloss.NewStyle().Foreground(colorDim)
	ta.FocusedStyle.Placeholder = lipgloss.NewStyle().Foreground(colorDim)
	ta.BlurredStyle.Placeholder = lipgloss.NewStyle().Foreground(colorDim)
	ta.FocusedStyle.Text = lipgloss.NewStyle().Foreground(colorText)
	ta.BlurredStyle.Text = lipgloss.NewStyle().Foreground(colorDim)
	ta.Cursor.Style = lipgloss.NewStyle().Foreground(colorAccent)
	
	// Remove default key bindings that conflict
	ta.KeyMap.InsertNewline.SetEnabled(false)

	m := Model{
		db:          db,
		tables:      tables,
		tableErr:    err,
		tableCursor: 0,
		sqlInput:    ta,
		focus:       focusTables,
	}
	return m
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		// Resize SQL textarea to fit
		sqlWidth := m.width - 28 // account for table panel width + borders
		if sqlWidth < 20 {
			sqlWidth = 20
		}
		m.sqlInput.SetWidth(sqlWidth - 4) // borders/padding

	case tea.KeyMsg:
		key := msg.String()

		// Global keys
		if key == "ctrl+c" {
			return m, tea.Quit
		}

		// Tab to switch focus (forward: Tables → SQL → Content → Tables)
		if key == "tab" {
			switch m.focus {
			case focusTables:
				m.focus = focusSQL
				m.sqlInput.Focus()
				return m, m.sqlInput.Focus()
			case focusSQL:
				m.focus = focusContent
				m.sqlInput.Blur()
				return m, nil
			default:
				m.focus = focusTables
				m.sqlInput.Blur()
				return m, nil
			}
		}

		// Ctrl+Tab to switch focus (reverse: Tables → Content → SQL → Tables)
		if key == "ctrl+tab" || key == "shift+tab" {
			switch m.focus {
			case focusTables:
				m.focus = focusContent
				m.sqlInput.Blur()
				return m, nil
			case focusContent:
				m.focus = focusSQL
				m.sqlInput.Focus()
				return m, m.sqlInput.Focus()
			default:
				m.focus = focusTables
				m.sqlInput.Blur()
				return m, nil
			}
		}

		if m.focus == focusTables {
			return m.updateTables(msg)
		}

		if m.focus == focusContent {
			return m.updateContent(msg)
		}

		// SQL focus: intercept keys before textarea sees them
		if m.focus == focusSQL {
			// Execute query: ctrl+r, ctrl+enter, or F5
			if key == "ctrl+r" || key == "ctrl+enter" || key == "F5" {
				m.executeSQL()
				return m, nil
			}

			// Enter inserts newline (don't pass to textarea)
			if key == "enter" || key == "shift+enter" {
				m.sqlInput.InsertString("\n")
				return m, nil
			}

			// Let textarea handle everything else
			var cmd tea.Cmd
			m.sqlInput, cmd = m.sqlInput.Update(msg)
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
		}
	}

	return m, tea.Batch(cmds...)
}

func (m Model) updateTables(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.tableCursor > 0 {
			m.tableCursor--
		}
	case "down", "j":
		if m.tableCursor < len(m.tables)-1 {
			m.tableCursor++
		}
	case "left", "h":
		// Horizontal scroll in content panel
		if m.contentScrollX > 0 {
			m.contentScrollX--
		}
	case "right", "l":
		// Horizontal scroll in content panel
		if m.content != nil && m.contentScrollX < len(m.content.Columns)-1 {
			m.contentScrollX++
		}
	case "enter":
		if len(m.tables) > 0 {
			tableName := m.tables[m.tableCursor]
			m.selectedTable = tableName
			// Get total row count
			countResult, countErr := m.db.ExecQuery(fmt.Sprintf("SELECT COUNT(*) FROM \"%s\"", tableName))
			if countErr == nil && len(countResult.Rows) == 1 && len(countResult.Rows[0]) == 1 {
				fmt.Sscanf(countResult.Rows[0][0], "%d", &m.totalRows)
			} else {
				m.totalRows = 0
			}
			result, err := m.db.ExecQuery(fmt.Sprintf("SELECT * FROM \"%s\" LIMIT 10000", tableName))
			if err != nil {
				m.sqlErr = err
			} else {
				m.sqlErr = nil
				m.content = result
				m.contentScrollX = 0
				m.contentScrollY = 0
			}
		}
	}
	return m, nil
}

func (m Model) updateContent(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Calculate how many rows are visible on screen for paging
	pageSize := m.contentPageSize()
	if pageSize < 1 {
		pageSize = 1
	}
	maxRow := 0
	if m.content != nil {
		maxRow = len(m.content.Rows) - 1
	}

	switch msg.String() {
	case "up", "k":
		if m.contentScrollY > 0 {
			m.contentScrollY--
		}
	case "down", "j":
		if m.content != nil && m.contentScrollY < maxRow {
			m.contentScrollY++
		}
	case "pgup":
		m.contentScrollY -= pageSize
		if m.contentScrollY < 0 {
			m.contentScrollY = 0
		}
	case "pgdown":
		m.contentScrollY += pageSize
		if m.contentScrollY > maxRow {
			m.contentScrollY = maxRow
		}
	case "home":
		m.contentScrollY = 0
	case "end":
		m.contentScrollY = maxRow
	case "left", "h":
		if m.contentScrollX > 0 {
			m.contentScrollX--
		}
	case "right", "l":
		if m.content != nil && m.contentScrollX < len(m.content.Columns)-1 {
			m.contentScrollX++
		}
	}
	return m, nil
}

// contentPageSize returns the number of visible data rows in the content panel.
func (m Model) contentPageSize() int {
	footerHeight := 3
	sqlPanelHeight := 8
	if m.sqlErr != nil {
		sqlPanelHeight += 2
	}
	contentHeight := m.height - sqlPanelHeight - footerHeight - 2
	if contentHeight < 5 {
		contentHeight = 5
	}
	// renderGrid gets height-4 (panel borders+padding), then maxRows = that - 3 (header, separator, status line)
	gridHeight := contentHeight - 4
	maxRows := gridHeight - 3
	if maxRows < 1 {
		maxRows = 1
	}
	return maxRows
}

func (m *Model) executeSQL() {
	query := m.sqlInput.Value()
	if strings.TrimSpace(query) != "" {
		result, err := m.db.ExecQuery(query)
		if err != nil {
			m.sqlErr = err
			m.content = nil
		} else {
			m.sqlErr = nil
			m.content = result
			m.contentScrollX = 0
			m.contentScrollY = 0
			m.selectedTable = "" // clear so title shows "Results"
		}
	}
}

func (m Model) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	// Layout dimensions
	tablePanelWidth := 26
	footerHeight := 3
	sqlPanelHeight := 8
	if m.sqlErr != nil {
		sqlPanelHeight += 2 // extra space for error
	}
	contentHeight := m.height - sqlPanelHeight - footerHeight - 2
	if contentHeight < 5 {
		contentHeight = 5
	}

	// Render each panel
	tablePanel := m.renderTablePanel(tablePanelWidth-2, m.height-footerHeight-2)
	sqlPanel := m.renderSQLPanel(m.width-tablePanelWidth-4, sqlPanelHeight-2)
	contentPanel := m.renderContentPanel(m.width-tablePanelWidth-4, contentHeight)
	footer := m.renderFooter(m.width, footerHeight)

	// Right side: SQL on top, content below
	rightSide := lipgloss.JoinVertical(lipgloss.Left, sqlPanel, contentPanel)

	// Main layout: table panel left, right side right
	mainArea := lipgloss.JoinHorizontal(lipgloss.Top, tablePanel, rightSide)

	// Full layout: main area + footer
	return lipgloss.JoinVertical(lipgloss.Left, mainArea, footer)
}

func (m Model) renderTablePanel(width, height int) string {
	title := panelTitleStyle.Render("📋 Tables")

	var items []string
	if m.tableErr != nil {
		items = append(items, sqlErrorStyle.Render("Error: "+m.tableErr.Error()))
	} else if len(m.tables) == 0 {
		items = append(items, statusStyle.Render("No tables found"))
	} else {
		for i, table := range m.tables {
			label := fmt.Sprintf("  %s", table)
			if i == m.tableCursor {
				label = fmt.Sprintf("▸ %s", table)
				items = append(items, selectedItemStyle.Render(label))
			} else {
				items = append(items, normalItemStyle.Render(label))
			}
		}
	}

	listContent := strings.Join(items, "\n")

	// Scroll to show selected item
	maxVisible := height - 2
	if maxVisible < 1 {
		maxVisible = 1
	}
	if len(items) > maxVisible {
		start := m.tableCursor - maxVisible/2
		if start < 0 {
			start = 0
		}
		end := start + maxVisible
		if end > len(items) {
			end = len(items)
			start = end - maxVisible
			if start < 0 {
				start = 0
			}
		}
		listContent = strings.Join(items[start:end], "\n")
	}

	inner := lipgloss.JoinVertical(lipgloss.Left, title, "", listContent)

	styled := panelStyle.
		Width(width).
		Height(height)

	if m.focus == focusTables {
		styled = panelFocusedStyle.
			Width(width).
			Height(height)
	}

	return styled.Render(inner)
}

func (m Model) renderSQLPanel(width, height int) string {
	title := panelTitleStyle.Render("🔍 SQL Query")
	hint := sqlHintStyle.Render("Ctrl+R: execute · Enter: newline · Tab: switch focus")

	textareaView := m.sqlInput.View()

	var inner string
	if m.focus == focusSQL {
		title = panelTitleStyle.Copy().Foreground(colorAccent).Render("🔍 SQL Query")
	}
	inner = lipgloss.JoinVertical(lipgloss.Left, title, hint, textareaView)

	// Show SQL error if any
	if m.sqlErr != nil {
		errMsg := sqlErrorStyle.Render("✗ " + m.sqlErr.Error())
		inner = lipgloss.JoinVertical(lipgloss.Left, inner, errMsg)
	}

	styled := panelStyle.
		Width(width).
		Height(height)

	if m.focus == focusSQL {
		styled = panelFocusedStyle.
			Width(width).
			Height(height)
	}

	return styled.Render(inner)
}

func (m Model) renderContentPanel(width, height int) string {
	var title string
	if m.selectedTable != "" {
		// Add row count to title
		titleText := "📊 Table: " + m.selectedTable
		if m.content != nil {
			totalCount := m.totalRows
			if totalCount == 0 {
				totalCount = len(m.content.Rows)
			}
			if totalCount > len(m.content.Rows) {
				titleText += fmt.Sprintf(" (%d rows, showing first %d)", totalCount, len(m.content.Rows))
			} else {
				titleText += fmt.Sprintf(" (%d rows)", totalCount)
			}
		}
		title = panelTitleStyle.Render(titleText)
	} else {
		// Show row count for query results too
		titleText := "📊 Results"
		if m.content != nil {
			titleText += fmt.Sprintf(" (%d rows)", len(m.content.Rows))
		}
		title = panelTitleStyle.Render(titleText)
	}

	var tableContent string
	if m.content == nil {
		tableContent = statusStyle.Render("Select a table or run a query to see results")
	} else {
		tableContent = m.renderGrid(m.content, width-4, height-4)
	}

	inner := lipgloss.JoinVertical(lipgloss.Left, title, "", tableContent)

	styled := panelStyle.
		Width(width).
		Height(height)

	if m.focus == focusContent {
		styled = panelFocusedStyle.
			Width(width).
			Height(height)
	}

	return styled.Render(inner)
}

func (m Model) renderGrid(result *QueryResult, maxWidth, maxHeight int) string {
	if len(result.Columns) == 0 {
		return statusStyle.Render("No results")
	}

	// Calculate optimal column widths (max 20 chars, but use smallest that fits)
	colWidths := make([]int, len(result.Columns))
	const maxTextWidth = 20

	// Start with column name widths
	for i, col := range result.Columns {
		colWidths[i] = len(col)
	}

	// Scan data to find actual max width needed per column
	for _, row := range result.Rows {
		for i, cell := range row {
			if i < len(colWidths) && len(cell) > colWidths[i] {
				colWidths[i] = len(cell)
			}
		}
	}

	// Cap at maxTextWidth
	for i := range colWidths {
		if colWidths[i] > maxTextWidth {
			colWidths[i] = maxTextWidth
		}
	}

	// padRight pads a string to exactly n characters wide, right-padding with spaces.
	padRight := func(s string, n int) string {
		if len(s) >= n {
			return s[:n]
		}
		return s + strings.Repeat(" ", n-len(s))
	}

	var lines []string

	// Determine visible columns based on scroll offset and available width
	startCol := m.contentScrollX
	if startCol >= len(result.Columns) {
		startCol = len(result.Columns) - 1
	}
	
	// Calculate how many columns can fit and adjust their widths
	const separatorWidth = 3 // " │ " between columns
	const minWidthPerCol = 8
	
	remainingCols := len(result.Columns) - startCol
	
	// Calculate how many columns can fit at minimum width
	// Formula: n * minWidth + (n-1) * separatorWidth <= maxWidth
	// n * (minWidth + separatorWidth) <= maxWidth + separatorWidth
	maxColsAtMin := (maxWidth + separatorWidth) / (minWidthPerCol + separatorWidth)
	if maxColsAtMin < 1 && remainingCols > 0 {
		maxColsAtMin = 1
	}
	
	numVisible := remainingCols
	if numVisible > maxColsAtMin {
		numVisible = maxColsAtMin
	}
	if numVisible < 1 && remainingCols > 0 {
		numVisible = 1
	}
	
	// Calculate uniform width that lets all visible columns fit
	// totalWidth = numVisible * colWidth + (numVisible-1) * separatorWidth <= maxWidth
	// colWidth = (maxWidth - (numVisible-1) * separatorWidth) / numVisible
	if numVisible > 0 {
		availableForData := maxWidth - (numVisible-1)*separatorWidth
		uniformWidth := availableForData / numVisible
		
		// Use uniform width if it's smaller than natural width, or if natural would overflow
		for i := 0; i < numVisible; i++ {
			colIdx := startCol + i
			if colIdx < len(colWidths) {
				if colWidths[colIdx] > uniformWidth {
					colWidths[colIdx] = uniformWidth
				}
			}
		}
	}
	
	visibleColumns := result.Columns[startCol : startCol+numVisible]

	// Build header row
	headerCells := make([]string, len(visibleColumns))
	for i, col := range visibleColumns {
		colIdx := startCol + i
		truncated := truncate(col, colWidths[colIdx])
		headerCells[i] = headerCellStyle.Render(padRight(truncated, colWidths[colIdx]))
	}
	lines = append(lines, strings.Join(headerCells, " │ "))

	// Build separator row
	sepParts := make([]string, len(visibleColumns))
	for i := range visibleColumns {
		colIdx := startCol + i
		sepParts[i] = strings.Repeat("─", colWidths[colIdx])
	}
	lines = append(lines, strings.Join(sepParts, "───"))

	// Build data rows
	maxRows := maxHeight - 3 // account for header, separator, and row count
	if maxRows < 1 {
		maxRows = 1
	}

	// Apply vertical scroll offset
	startRow := m.contentScrollY
	if startRow >= len(result.Rows) {
		startRow = len(result.Rows) - 1
	}
	if startRow < 0 {
		startRow = 0
	}

	endRow := startRow + maxRows
	if endRow > len(result.Rows) {
		endRow = len(result.Rows)
	}

	for ri := startRow; ri < endRow; ri++ {
		row := result.Rows[ri]
		cells := make([]string, len(visibleColumns))
		for i := range visibleColumns {
			colIdx := startCol + i
			val := ""
			if colIdx < len(row) {
				val = row[colIdx]
			}
			truncated := truncate(val, colWidths[colIdx])
			cells[i] = cellStyle.Render(padRight(truncated, colWidths[colIdx]))
		}
		lines = append(lines, strings.Join(cells, " │ "))
	}

	// Add scroll indicator if needed
	scrollInfo := ""
	if m.contentScrollX > 0 || startCol+len(visibleColumns) < len(result.Columns) {
		scrollInfo = fmt.Sprintf(" [columns %d-%d of %d, use ←→ to scroll]",
			startCol+1, startCol+len(visibleColumns), len(result.Columns))
	}
	
	rowScrollInfo := ""
	if len(result.Rows) > maxRows {
		rowScrollInfo = fmt.Sprintf(" [rows %d-%d of %d, use ↑↓ to scroll]",
			startRow+1, endRow, len(result.Rows))
	}

	// Combine scroll info
	combinedInfo := scrollInfo + rowScrollInfo
	if combinedInfo != "" {
		rowInfo := statusStyle.Render(combinedInfo)
		lines = append(lines, rowInfo)
	}

	return strings.Join(lines, "\n")
}

func truncate(s string, maxLen int) string {
	if maxLen <= 0 {
		return ""
	}
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-2] + ".."
}

func (m Model) renderFooter(width, height int) string {
	exitBtn := exitButtonStyle.Render(" Exit (Ctrl+C) ")

	focusIndicator := ""
	switch m.focus {
	case focusTables:
		focusIndicator = statusStyle.Render("Focus: Tables │ Tab: switch")
	case focusSQL:
		focusIndicator = statusStyle.Render("Focus: SQL │ Tab: switch")
	case focusContent:
		focusIndicator = statusStyle.Render("Focus: Content │ Tab: switch")
	}

	tableInfo := statusStyle.Render(fmt.Sprintf("%d tables loaded", len(m.tables)))

	left := lipgloss.JoinHorizontal(lipgloss.Center, exitBtn, "  ", focusIndicator, "  ", tableInfo)

	return footerStyle.
		Width(width).
		Height(height).
		Render(left)
}
