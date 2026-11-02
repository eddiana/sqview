# sqview

A terminal-based SQLite table viewer built in Go using [Bubble Tea](https://github.com/charmbracelet/bubbletea). Tested in Linux and Windows 11.

Browse tables, run arbitrary SQL queries, and scroll through results — all from the command line.

Coded using Qwen 3.7 Max. 

## Features

- Three-panel layout: table list, SQL editor, and results grid
- Navigate tables and query results with keyboard shortcuts
- Horizontal and vertical scrolling with fixed column headers
- Page Up/Down and Home/End support for large datasets
- Table queries capped at 10,000 rows (SQL queries are unrestricted)

## Usage

```bash
./sqview <database.db>
```

### Controls

| Key | Context | Action |
|-----|---------|--------|
| `Tab` | Any | Cycle focus forward (Tables → SQL → Content) |
| `Shift+Tab` | Any | Cycle focus backward (Tables → Content → SQL) |
| `↑/↓` or `j/k` | Tables | Navigate table list |
| `↑/↓` or `j/k` | Content | Scroll rows vertically (headers stay fixed) |
| `PgUp/PgDn` | Content | Page through rows |
| `Home/End` | Content | Jump to first/last row |
| `←/→` or `h/l` | Tables/Content | Scroll columns horizontally |
| `Enter` | Tables | Load selected table (first 10,000 rows) |
| `Enter` | SQL | Insert newline |
| `Ctrl+R` | SQL | Execute query |
| `Ctrl+C` | Any | Quit |

## Build

Requires [Go](https://go.dev/dl/) 1.23+.

```bash
git clone https://github.com/eddiana/sqview.git
cd sqview
go build -o sqview .
```

The binary will be built as `sqview` in the project directory.
