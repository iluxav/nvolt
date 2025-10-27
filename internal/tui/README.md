# nvolt TUI (Terminal User Interface)

An interactive Terminal User Interface for managing environment variables and user permissions in nvolt.

## Features

- **Dual Panel Layout**: Browse environment variables and users side-by-side
- **Multi-Environment Support**: Switch between environments (default, staging, production) with arrow keys
- **Interactive Navigation**: Full keyboard-driven interface
- **Value Reveal/Hide**: Toggle between encrypted and decrypted values
- **Delete Confirmations**: Modal dialogs for safe deletion
- **Real-time Updates**: Changes reflected immediately in the UI

## Usage

```bash
# Launch TUI with auto-detected project
nvolt tui

# Launch TUI for a specific project
nvolt tui -p my-project

# Launch TUI for a specific project and environment
nvolt tui -p my-project -e production
```

## Keyboard Shortcuts

### Navigation
- **Tab**: Switch between Variables panel (left) and Users panel (right)
- **←/→ (Arrow Left/Right)**: Navigate between environments
- **↑/↓ (Arrow Up/Down)**: Navigate through table rows
- **j/k**: Vim-style navigation (down/up)

### Actions
- **Enter/Space**: Reveal/hide encrypted value (in Variables panel)
- **d/Delete**: Delete selected item (shows confirmation modal)
- **q/Ctrl+C**: Quit TUI

### Modal
- **y/Enter**: Confirm action
- **n/Esc**: Cancel action

## Architecture

The TUI is built using:
- **[Bubble Tea](https://github.com/charmbracelet/bubbletea)**: The Elm Architecture for Go
- **[Lip Gloss](https://github.com/charmbracelet/lipgloss)**: Style definitions and layout
- **[Bubbles](https://github.com/charmbracelet/bubbles)**: Pre-built components

### File Structure

```
internal/tui/
├── model.go           # Main TUI model and update logic
├── types.go           # Type definitions
├── styles.go          # Lip Gloss styling
├── variables_panel.go # Left panel: environment variables
├── users_panel.go     # Right panel: users and permissions
└── README.md          # This file
```

## Extending the TUI

### Adding New Panels

1. Create a new file (e.g., `my_panel.go`)
2. Add rendering method to `Model`:
   ```go
   func (m Model) renderMyPanel(width int) string {
       // Your rendering logic
   }
   ```
3. Update `View()` method to include your panel

### Adding New Modals

1. Add modal type to `types.go`:
   ```go
   const (
       MyNewModal ModalType = iota
   )
   ```
2. Update `renderModal()` in `model.go`
3. Add keyboard handler in `handleModalKeys()`

## Future Enhancements

- [ ] Fetch live data from server (currently using mock data)
- [ ] Add/Edit variables directly from TUI
- [ ] Add/Edit users directly from TUI
- [ ] Search/filter functionality
- [ ] Copy values to clipboard
- [ ] Syntax highlighting for values
- [ ] Export to file functionality
- [ ] Audit log viewer
