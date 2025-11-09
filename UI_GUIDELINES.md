# nvolt UI Guidelines

## Color Scheme

This document defines the consistent color scheme used across all nvolt CLI commands.

### Color Usage

| Element | Color | Function | Example |
|---------|-------|----------|---------|
| Success messages | Bright Green | `ui.Success()` | `✓ Vault initialized` |
| Progress/Steps | Cyan | `ui.Step()` | `→ Pushing secrets to vault` |
| Values (IDs, names) | Cyan | `ui.Cyan()` | Project name, machine ID, environment |
| Paths & metadata | Gray | `ui.Gray()` | File paths, fingerprints, timestamps |
| Warnings | Yellow | `ui.Warning()` | Warning messages |
| Errors | Red | `ui.Error()` | Error messages |
| Section headers | Cyan | `ui.Section()` | "Next steps:", "Secrets:" |
| Bullet points | Gray | `ui.Substep()` | `  • item` |

### Helper Functions

#### Core Functions

- **`ui.Success(format, args...)`** - Success message with ✓ checkmark (bright green)
- **`ui.Step(format, args...)`** - Progress indicator with → arrow (cyan)
- **`ui.Warning(format, args...)`** - Warning message with "WARNING:" prefix (yellow)
- **`ui.Error(format, args...)`** - Error message with "ERROR:" prefix (red)
- **`ui.Info(format, args...)`** - Standard informational message

#### Formatting Functions

- **`ui.Section(message)`** - Section header (cyan, with newline)
- **`ui.Substep(message)`** - Indented bullet point (`  • message` in gray)
- **`ui.PrintKeyValue(key, value)`** - Key-value pair display
- **`ui.PrintDetected(label, value)`** - Shows detected values (e.g., "Project: myproject")
- **`ui.PrintModeInfo(mode)`** - Shows mode information (e.g., "Mode: Local")

#### Color Functions

- **`ui.Green(text)`** - Green text
- **`ui.BrightGreen(text)`** - Bright green text
- **`ui.Cyan(text)`** - Cyan text (for important values)
- **`ui.Yellow(text)`** - Yellow text (for warnings/notes)
- **`ui.Red(text)`** - Red text (for errors)
- **`ui.Gray(text)`** - Gray text (for metadata)
- **`ui.Bold(text)`** - Bold text

## Usage Examples

### Success Flow
```go
ui.Step("Pulling secrets from vault")
// ... do work ...
ui.Success("Repository up to date")
```

### Showing Information
```go
ui.PrintDetected("Project", project)
ui.PrintKeyValue("  Environment", ui.Cyan(environment))
ui.PrintKeyValue("  Vault", ui.Gray(vaultPath))
```

### Section with Items
```go
ui.Section("Secrets encrypted:")
for key := range secrets {
    ui.Substep(key)
}
```

### Notes and Warnings
```go
ui.Info(ui.Yellow("Note: ") + "All machines now have access to decrypt secrets.")
ui.Warning("In local mode, you manage Git operations manually")
```

## Commands Updated

All commands now follow this consistent scheme:

- ✅ **init** - Vault initialization
- ✅ **push** - Push secrets to vault
- ✅ **pull** - Pull secrets from vault
- ✅ **machine add/rm/list** - Machine management
- ✅ **run** - Run commands with secrets
- ✅ **sync** - Key rotation and re-wrapping
- ✅ **vault show/verify** - Vault inspection
- ✅ **status** - Already uses lipgloss (no changes needed)

## Visual Hierarchy

```
Section Header (Cyan)
  → Step in progress (Cyan arrow)
  ✓ Success message (Bright green)
    Key: value (normal: cyan value)
    • Substep (gray bullet)

⚠ Warning message (Yellow)
✗ Error message (Red)
```

## Consistency Rules

1. **Always use helper functions** - Never use raw fmt.Printf with colors
2. **Cyan for important values** - Project names, machine IDs, environments
3. **Gray for metadata** - Paths, fingerprints, timestamps
4. **Step before action** - Always show what you're about to do
5. **Success after completion** - Confirm what was done
6. **Notes in yellow** - Important information that isn't an error

## Example Output

```
→ Checking machine keypair
✓ Machine keypair ready
  Machine ID: m-abc123
  Fingerprint: SHA256:xyz...

Mode: Local

✓ Vault initialized
  • /path/to/vault

Next steps:

  • nvolt push -f .env
  • nvolt pull

  Note: In local mode, you manage Git operations manually
```
