<p align="center">
  <img src="https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat-square&logo=go&logoColor=white"/>
  <img src="https://img.shields.io/badge/Gin-Web_Framework-blue?style=flat-square"/>
  <img src="https://img.shields.io/github/v/release/Dectre/Grader?style=flat-square&color=green"/>
  <img src="https://img.shields.io/github/license/Dectre/Grader?style=flat-square"/>
  <img src="https://img.shields.io/badge/platform-Windows%20%7C%20macOS%20%7C%20Linux-lightgrey?style=flat-square"/>
</p>

<h1 align="center">📚 Grader</h1>
<p align="center"><strong>A fast, browser-based homework grading assistant built in Go</strong></p>

---

## 🎯 What is Grader?

Grader is a **zero-dependency, cross-platform desktop app** that turns the painful workflow of grading PDF submissions into a fast, keyboard-driven, one-screen experience.

It's built for teaching assistants who receive zip exports from **Moodle**, **eLearn**, or similar LMSs and need to:

- View each student's PDF submission next to the rubric,
- Type grades, descriptions, and flags,
- Export a polished Excel report with live statistics,
- Do it all — **without** juggling 30 PDF windows and a separate spreadsheet.

Written in **Go + Gin**, the UI (HTML/CSS/JS) is embedded into a **single executable** (~15 MB). No Python, no Node, no installer. Just double-click.

---

## ✨ Features

### 🗂️ Multi-Assignment System
- Each assignment lives in its own folder: `Assignments/<name>/{Data, PDFs, output}`.
- A persistent `assignments.json` registry prevents stray folders from appearing as assignments.
- Switch between assignments at any time — pick, create, rename, or delete from a single screen.

### 🌐 Web-Based Setup
- Upload `rubric.csv` and `students.csv` directly from the browser — no manual folder editing.
- Inline editor for quick paste/edit without leaving the app.
- Bulk PDF import: drop a **zip** (eLearn export) or multiple PDFs; zip entries are renamed after their parent folder automatically.

### 🖥️ Split-Screen Grading
- PDF on the right, controls on the left — always in view.
- Resizable student panel (drag the edge, touch-friendly) with auto-collapse when too small.
- Collapsible accordion sections (Info / Questions / Description).
- Collapsible top bar (`Alt+H`) for maximum PDF real-estate.

### ⌨️ Keyboard-First Workflow
A full set of typing-safe shortcuts — never reach for the mouse:

| Action | Shortcut |
|---|---|
| Next / Prev student | `PageDown` / `PageUp` |
| Next / Prev grade field | `↓`/`Enter` — `↑`/`Shift+Enter` |
| Submit | `Ctrl+Enter` |
| Submit + next student | `Ctrl+Shift+Enter` |
| Clear form | `Alt+C` |
| Did Not Submit | `Alt+X` |
| Flag for review | `F` |
| Max (focused field) | `Alt+M` |
| Lock (focused field) | `Alt+L` |
| Lock / Unlock all | `Ctrl+Shift+L` |
| Search | `/` |
| Toggle Sidebar / PDF / Stats / Theme | `Alt+B` / `Alt+P` / `Alt+S` / `Alt+T` |
| Collapse / expand top bar | `Alt+H` |
| Open shortcuts help | `?` |
| Leave field / close dialogs | `Esc` |

Shortcuts **never fire while typing** — only field navigation is active inside inputs.

### 🎨 Modern UI
- **Dark / Light theme** persisted across all screens (select page + grader).
- Smooth transitions, soft shadows, custom scrollbars, gradient accents.
- Natural alphanumeric sorting.
- Sort assignments by **Name**, **Created date**, or **Modified date** (asc/desc, remembered).
- Last-selected assignment highlighted ⭐.

### 📊 Live Statistics
- Per-question `Mean`, `Median`, `Max`, `Min`, `Count` — updated on every submission.
- Shown in a toggleable sidebar, **and** baked into the exported Excel as live formulas.

### 📥 Advanced Excel Export
- Generated `grades.xlsx` with the **Vazirmatn** Persian font, frozen panes, thick/thin borders.
- Dynamic formulas for `Total Score` (ignores "Not Submitted" students).
- Auto-stat rows (Graded Count, Mean, Median, Max, Min).
- Flagged students are highlighted yellow via conditional formatting.
- Separate `grades.csv` acts as an internal state database for persistence.

### 🔒 Per-Student Locks
- Lock individual grade fields or the description — **per student**, so locking one doesn't affect the next.
- `Max` button respects locks (won't overwrite a graded field).
- `Lock All` / `Unlock All` toggle.

### 🚩 Flags & Absent Markers
- Flag a student for review — visible in the dropdown with a 🚩 icon.
- Mark a student as "Did Not Submit" — zeroed out in totals, excluded from stats.

### 🔎 Smart Search
- Search by name, surname, or student ID.
- Inline results with RTL support for Persian names.

---

## 📁 Directory Structure

```text
GraderApp/
├── Grader.exe                    # The compiled binary
└── Assignments/
    ├── assignments.json          # Registry of assignments (auto-managed)
    ├── HW1/
    │   ├── Data/
    │   │   ├── students.csv      # Student list (Moodle export format)
    │   │   └── rubric.csv        # Questions + max scores
    │   ├── PDFs/                 # Student PDF submissions
    │   └── output/
    │       ├── grades.csv        # Internal state DB
    │       └── grades.xlsx       # Final Excel report
    └── CA2/
        └── ...                   # Another assignment
```

You don't create this structure manually — the app **prompts** you to create a new assignment via the web UI on first launch.

---

## 🚀 Quick Start

### For End Users
1. Download the latest release from [Releases](https://github.com/Dectre/Grader/releases).
2. Place `Grader.exe` (or `grader` on macOS/Linux) in any folder.
3. Double-click it. Your default browser opens at **http://localhost:8080**.
4. The web UI appears. Click **Create** to make your first assignment, or pick an existing one.
5. Use the ⚙️ button to upload your `students.csv`, `rubric.csv`, and PDFs.
6. Select the assignment → start grading.

### For Developers

Requires **Go 1.21+**.

```bash
git clone https://github.com/Dectre/Grader.git
cd Grader
go mod tidy

# Build
go build -o Grader.exe          # Windows
go build -o grader              # Linux/macOS

# Cross-compile for Windows from Linux/macOS
GOOS=windows GOARCH=amd64 go build -o Grader.exe
```

Run it:
```bash
./Grader.exe   # or ./grader
```

---

## 🗃️ Input Files

### `students.csv`
Moodle's student export format. Must have these columns (UTF-8 with or without BOM):

```csv
نام,نام خانوادگی,نام کاربری
علی,احمدی,810101234
زهرا,محمدی,810105678
```

### `rubric.csv`
Two columns: `Question,Max Grade`.

```csv
Question,Max Grade
سوال ۱,20
سوال ۲,30
سوال ۳,50
```

### PDFs
Two ways to import:
- **Bulk zip upload** (recommended): upload the Moodle zip; each PDF is renamed after its parent folder.
- **Multiple PDFs**: drop several PDFs at once.

The app matches each PDF to a student by searching the filename for the student's first and last name (ignoring spaces, underscores, and ZWNJ characters).

---

## 🏗️ Architecture

```text
grader/
├── main.go                          # Entry point + browser launch
├── index.html                       # Grader UI (embedded)
├── select.html                      # Assignment selector UI (embedded)
├── fonts/                           # Vazirmatn + Inter (embedded)
└── internal/
    ├── api/
    │   ├── server.go                # Gin router + auth middleware
    │   └── select_handlers.go       # Assignment CRUD + file handlers
    ├── manager/
    │   ├── manager.go               # DataManager: load/save students, grades
    │   ├── exporter.go              # Excel + CSV report builder
    │   ├── assignments.go           # Registry (assignments.json)
    │   └── pdfimport.go             # Zip extraction + PDF import
    ├── models/                      # Student struct
    └── utils/                       # String cleaning helpers
```

The `index.html`, `select.html`, and `fonts/` are **embedded** into the binary via `go:embed`, so the compiled executable is fully self-contained.

---

## 🧪 Testing the Workflow

1. Run the app → browser opens.
2. Click **Create**, name it `HW1`.
3. Click ⚙️ → upload a zip of PDFs, or edit `rubric.csv` and `students.csv` inline.
4. Click the `HW1` item to open the grader.
5. Use `PageDown` to cycle students, `Enter` to jump between grade fields, `Ctrl+Enter` to submit.
6. Press `?` anytime to see the shortcut cheat-sheet.
7. When done, click ⬇️ **Excel** to download the final report.

---

## ⚠️ Troubleshooting

| Problem | Solution |
|---|---|
| **"Port 8080 already in use"** | Another service is using 8080. Stop it, or edit `server.Run(":8080")` in `main.go`. |
| **PDF not matching a student** | Ensure the PDF filename contains both the first and last name. Spaces and underscores are ignored. |
| **Excel file locked** | Close `grades.xlsx` in Excel before submitting new grades — the app writes to it on every save. |
| **Theme or settings lost** | Settings are stored in `localStorage`. Clearing browser data resets them. |
| **Assignment disappeared** | Only directories with a `Data/` subfolder are auto-imported. Check `Assignments/assignments.json`. |

---

## 🤝 Contributing

Pull requests are welcome! For major changes, please open an issue first.

```bash
# Create a feature branch
git checkout -b feature/my-thing

# Keep commits focused
git commit -m "feat(ui): my new feature"

# Push and open a PR
git push origin feature/my-thing
```

Commit message conventions: `feat:`, `fix:`, `refactor:`, `style:`, `docs:`.

---

## 📜 License

This project is open source. See [LICENSE](LICENSE) for details.

---

## 💡 Why Go?

- **Single binary, zero dependencies** — ship a `.exe` to TAs who have never heard of Python.
- **Blazing fast** — Excel generation and PDF matching happen instantly, even with 200+ students.
- **Cross-platform** — same source compiles to Windows, macOS, and Linux.
- **Self-hosted** — no cloud, no telemetry, no data leaves the machine.

---

<div align="center">

**⭐ If you found this repository helpful, please give it a star!**

</div>