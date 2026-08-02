# Homework Grading System (Go Version)

A high-performance, cross-platform desktop application designed to streamline and accelerate the grading process for assignments and exams. This system is specifically tailored to process, manage, and grade files exported from Moodle, eliminating the need to manually open files and record grades in separate spreadsheets. 

Built with **Go (Golang)** and **Gin**, it offers blazing-fast performance and compiles into a single, easy-to-distribute executable file.

## Key Features

- **Automated File Matching**: Automatically detects and links PDF files to students based on their First and Last names, without relying on strict file naming conventions.
- **Built-in PDF Viewer**: View submissions directly within the browser with native inline rendering, zoom, and page navigation.
- **Grade State Management**: Smartly tracks the status of each student, categorizing them into "Not Submitted Yet", "Unsaved Changes", and "Submitted".
- **Quick Search & Navigation**: Jump directly to any student using the autocomplete search bar by typing their Name, Surname, or Student ID.
- **Advanced Excel Export**: Automatically generates a fully formatted Excel file (`grades.xlsx`) featuring customized fonts (Vazirmatn), cell styling, frozen panes, and dynamic statistical formulas (Mean, Median, Max, Min).
- **Theme Support**: Includes native Dark Mode and Light Mode for comfortable viewing during extended grading sessions.
- **Zero-Dependency Distribution**: The UI (`index.html`) and assets (`fonts/`) are embedded directly into the executable. No external runtime (like Python or Node.js) is required for the end-user.

## Directory and File Structure

For the application to run correctly, your project directory must strictly follow this structure. *(Note: `index.html` and `fonts/` are embedded in the compiled binary, so you only need to manage the folders below).*

```text
ProjectFolder/
 ├── Grader.exe          # The compiled executable (or 'grader' on Linux/macOS)
 ├── PDFs/               # Place all student PDF submissions here
 ├── Data/
 │   ├── students.csv    # List of students (Moodle export format)
 │   └── rubric.csv      # Grading criteria and max scores
 └── output/             # Automatically generated upon first run
     ├── grades.csv      # Internal state database
     └── grades.xlsx     # Final formatted report with statistics
```

## Input Files Guide

To ensure data is processed accurately, your input files must follow these structures:

### 1. Assignment Files (PDFs)
Place all PDF files downloaded from Moodle into the `PDFs` folder. The application searches for the student's First and Last name within the file name (ignoring spaces, underscores, and zero-width non-joiners).
- **Valid File Name Example**: `امیرعلی دهقانی_1995188_assignsubmission_file_HW4-Spring2026.pdf`

### 2. Student List (`students.csv`)
This file must be placed in the `Data` folder and saved with **UTF-8 encoding** (UTF-8 with BOM is fully supported and automatically handled). 
- **File Structure Example**:
  ```csv
  نام,نام خانوادگی,نام کاربری,آدرس پست الکترونیک,گروه‌ها
  امیرعلی,دهقانی,810102443,amiralidehqani@ut.ac.ir,
  ```

### 3. Grading Rubric (`rubric.csv`)
This file must also be placed in the `Data` folder. The application dynamically reads the first column as the Question Name and the second column as the Maximum Score.
- **File Structure Example**:
  ```csv
  بخش,نمره
  سوال 1,20
  سوال 2,40
  سوال 3 الف,20
  سوال 3 ب,20
  ```

## Installation & Building

### For End Users (Recommended)
1. Download the latest `Grader.exe` (or `grader` for macOS/Linux) from the [Releases](#) page.
2. Create the `Data` and `PDFs` folders next to the executable.
3. Add your `students.csv` and `rubric.csv` to the `Data` folder.
4. Double-click `Grader.exe` to run.

### For Developers
Ensure you have [Go 1.21+](https://go.dev/dl/) installed on your system.

1. Clone the repository:
   ```bash
   git clone <repository-url>
   cd <repository-folder>
   ```
2. Install dependencies:
   ```bash
   go mod tidy
   ```
3. Build the executable:
   ```bash
   # For Windows
   go build -o Grader.exe
   
   # For Linux/macOS
   go build -o grader
   
   # Cross-compile for Windows from Linux/macOS
   GOOS=windows GOARCH=amd64 go build -o Grader.exe
   ```

## Usage

1. Ensure your `Data/students.csv` and `Data/rubric.csv` are correctly populated.
2. Run the application:
   ```bash
   ./Grader.exe   # On Windows, just double-click or run 'Grader.exe'
   ```
3. The application will automatically open your default web browser at:
   **http://localhost:8080**
4. If the `Data` folder is missing or empty, the app will generate template files and prompt you to fill them before proceeding.

## Outputs

Upon submitting your first grade, the application will automatically create an `output` folder. The following files are generated and updated with every submission:

- **`grades.csv`**: An internal, lightweight database used to persist submission states (e.g., `_Submitted`, `Not Submitted`) across different application sessions.
- **`grades.xlsx`**: The final, cleanly formatted export file. It includes:
  - Itemized grades and descriptions.
  - Dynamic Excel formulas for `Total Score` (ignoring "Not Submitted" students).
  - Statistical rows (Mean, Median, Max, Min) calculated dynamically.
  - Professional styling with the "Vazirmatn" font, thick/thin borders, and frozen panes for easy scrolling.

## Troubleshooting

- **Port 8080 is already in use**: The application requires port `8080`. Ensure no other service (like another web server) is using this port.
- **PDF not loading**: Ensure the PDF file name contains both the student's first and last name exactly as they appear in `students.csv` (without spaces or special characters).
- **Excel formatting issues**: The `grades.xlsx` file is actively locked by the application while it runs. Do not keep the Excel file open in Microsoft Excel while submitting grades in the app, as this will cause a "file in use" error. Close Excel before saving new grades.