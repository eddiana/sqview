# sqview - SQLite Table Viewer (TUI)

A terminal-based SQLite database viewer built in Go using the [Bubble Tea](https://github.com/charmbracelet/bubbletea) framework.

**Coded using Qwen 3.7 Max**

## Usage

```
./sqview <database.db>
```

## Architecture

The application is structured into four source files:

| File | Purpose |
|------|---------|
| `main.go` | Entry point, argument parsing, launches the TUI, defines Release constant |
| `model.go` | Bubble Tea model with Update/View logic, layout, grid rendering, and scrolling |
| `db.go` | SQLite database wrapper (table listing, query execution) |
| `styles.go` | Lipgloss style definitions (colors, borders, formatting) |

### Dependencies

- **charmbracelet/bubbletea** — TUI framework
- **charmbracelet/bubbles** — UI components (textarea)
- **charmbracelet/lipgloss** — Terminal styling
- **modernc.org/sqlite** — Pure-Go SQLite driver (no CGO)

## Layout

The interface has four panels:

```
┌──────────┐┌─────────────────────────────────┐
│  Tables   ││  SQL Query                       │
│           ││  (multi-line textarea)           │
│           │└─────────────────────────────────┘
│           │┌─────────────────────────────────┐
│           ││  Results / Table: <name> (X rows)│
│           ││  ┌─────┬────────┬────────┐      │
│           ││  │ id  │ name   │ email  │ ...  │
│           ││  ├─────┼────────┼────────┤      │
│           ││  │ 1   │ Alice  │ a@..   │ ...  │
│           ││  │ 2   │ Bob    │ b@..   │ ...  │
│           ││  └─────┴────────┴────────┘      │
│           ││  [rows 1-50, cols 1-9, ←→ ↑↓]   │
├──────────┴┴─────────────────────────────────┤
│  Exit (Ctrl+C)  Focus: Content  Tab: switch │
└─────────────────────────────────────────────┘
```

### Panel Descriptions

- **Tables panel** (left) — Lists all user tables in the database. Scrollable with up/down arrows.
- **SQL panel** (top-right) — Multi-line SQL query editor with placeholder text and keybinding hints.
- **Content panel** (bottom-right) — Displays query/table results in a formatted grid with fixed headers and comprehensive scrolling.
- **Footer** — Exit button, current focus indicator, and table count.

## Controls

| Key | Context | Action |
|-----|---------|--------|
| `Tab` | Any | Cycle focus forward: Tables → SQL → Content |
| `Shift+Tab` or `Ctrl+Tab` | Any | Cycle focus backward: Tables → Content → SQL |
| `↑/↓` or `j/k` | Tables | Navigate table list |
| `↑/↓` or `j/k` | Content | Scroll rows vertically (headers stay fixed) |
| `Page Up/Down` | Content | Page through rows (one screen at a time) |
| `Home/End` | Content | Jump to first/last row |
| `←/→` or `h/l` | Tables/Content | Scroll columns horizontally |
| `Enter` | Tables | Load selected table (first 10,000 rows) |
| `Enter` | SQL | Insert newline |
| `Ctrl+R` | SQL | Execute query |
| `Ctrl+C` | Any | Exit application |

## Key Features

### Three-Panel Focus System
- Tab cycles forward through all three interactive panels (Tables → SQL → Content)
- Shift+Tab or Ctrl+Tab cycles backward (Tables → Content → SQL)
- Focused panel gets a distinctive **double-line border** for clear visual feedback
- Footer shows which panel is currently active

### Grid Rendering
- Columns are sized intelligently: smallest width that fits the data, capped at 20 characters
- All cells are padded to exact widths for perfect column alignment
- Columns that don't fit on screen are hidden and accessible via horizontal scrolling
- Header row stays fixed at the top while scrolling data rows vertically

### Comprehensive Vertical Scrolling
- **Up/down arrows** (or j/k) — scroll one row at a time
- **Page Up/Down** — scroll one full page of visible rows (calculated dynamically)
- **Home/End** — jump to first/last row
- Scroll indicator shows current row range (e.g., "rows 1-50 of 10000")
- Scroll position resets when loading a new table or executing a new query

### Horizontal Scrolling
- Only shows as many columns as can fit at a readable minimum width (8 characters)
- Scroll indicator shows current column range (e.g., "columns 1-9 of 19")
- Left/right arrows (or h/l) scroll columns horizontally

### Smart Row Limits
- Table queries (via Enter) limited to 10,000 rows for performance
- SQL queries (via Ctrl+R) have no limit — user controls what they query
- Title bar shows total row count: "📊 Table: employees (107 rows)"
- For truncated tables: "📊 Table: bodies (21166 rows, showing first 10000)"
- Query results show count: "📊 Results (523 rows)"

### Release Information
- Version constant `Release = 1` defined in main.go
- Usage message shows: "sqview rel 1 - SQLite table viewer"

## Build

### Linux
```bash
go build -o sqview .
```

### Windows
```bash
GOOS=windows GOARCH=amd64 go build -o sqview.exe .
```

Binaries have been tested on both Linux and Windows (Windows Terminal, PowerShell).

## Test Database

The project includes `test.db` with 20 tables of realistic sample data, including tables with 20+ columns and 100+ rows to exercise scrolling features.

## Future Enhancement Ideas

**Quick wins:**
- Query history — Save and re-run previous queries
- Export to CSV/JSON — Save query results to files
- Refresh — Reload data with a keypress if database changes

**Medium effort:**
- Cell expansion — Press Enter on a cell to see full content (great for long text/JSON)
- Performance metrics — Show query execution time
- Search/filter — Quick search within current results
- Column sorting — Sort results by clicking columns

**Bigger projects:**
- Schema panel — Show table structure (column types, keys, indexes)
- SQL autocomplete — Suggest table/column names while typing
- Foreign key navigation — Jump to related records
- Data type formatting — Better display for dates, BLOBs, NULLs

## Notes

- Colors are currently set to white across all styles — planned for future customization
- The application uses the alternate screen buffer, so the terminal content is restored on exit
- Tested with databases up to 21,000+ rows with smooth scrolling performance
