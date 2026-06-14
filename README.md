# Homework Grading System

A comprehensive desktop application designed to streamline and accelerate the grading process for assignments and exams. This system is specifically tailored to process, manage, and grade files exported and downloaded from Moodle systems, eliminating the need to manually open files and record grades in separate spreadsheets.

## Key Features

* **Automated File Matching:** Automatically detects and links PDF files to students based on their First and Last names, without relying on strict file naming conventions.
* **Built-in PDF Viewer:** View submissions directly within the app with features like zoom, page navigation, and dynamic resizing (Fit Width / Fit Screen).
* **Grade State Management:** Smartly tracks the status of each student, categorizing them into "Not Submitted Yet", "Unsaved Changes", and "Submitted".
* **Quick Search & Navigation:** Jump directly to any student using the autocomplete search bar by typing their Name, Surname, or Student ID.
* **Advanced Excel Export:** Automatically generates a fully formatted Excel file featuring customized fonts, cell styling, and calculated statistical metrics (Mean, Median, Max, Min).
* **Theme Support:** Includes native Dark Mode and Light Mode for comfortable viewing during extended grading sessions.

## Directory and File Structure

For the application to run correctly, your project must strictly follow this structure:

```text
ProjectFolder/
├── PDFs/
├── Data/
│   ├── students.csv
│   └── rubric.csv
├── fonts/
│   ├── Inter-Regular.ttf
│   ├── Vazirmatn.ttf
│   └── ...
├── core/
│   ├── __init__.py
│   ├── data_manager.py
│   ├── pdf_viewer.py
│   └── gui.py
├── main.py
└── requirements.txt
```

## Input Files Guide

To ensure data is processed accurately, your input files must follow these structures:

### 1. Assignment Files (PDFs)
Place all PDF files downloaded from Moodle into the `PDFs` folder. The application searches for the student's First and Last name within the file name.
* **Valid File Name Example:**
`امیرعلی دهقانی_1995188_assignsubmission_file_HW4-Spring2026.pdf`

### 2. Student List (students.csv)
This file must be placed in the `Data` folder and saved with UTF-8 encoding. You do not need to modify the standard Moodle export columns.
* **File Structure Example:**
```csv
نام,"نام خانوادگی","نام کاربری","آدرس پست الکترونیک",گروه‌ها
امیرعلی,دهقانی,810102443,amiralidehqani@ut.ac.ir,
```

### 3. Grading Rubric (rubric.csv)
This file must also be placed in the `Data` folder. The application dynamically reads the first column as the Question Name and the second column as the Maximum Score. The header names do not matter.
* **File Structure Example:**
```csv
بخش,نمره
سوال 1,20
سوال 2,40
سوال 3 الف,20
سوال 3 ب,20
```

## Installation

Ensure Python is installed on your system. Open your terminal in the root directory of the project and run the following command to install the required dependencies:

```bash
pip install -r requirements.txt
```

## Usage

Once your files and directories are set up, run the application using the following command:

```bash
python main.py
```

## Outputs

Upon submitting your first grade, the application will automatically create an `output` folder in the root directory. The following files are generated and updated with every submission:

* **grades.csv:** An internal database used to persist the submission states (e.g., "Not Submitted Yet") across different application sessions.
* **grades.xlsx:** The final, cleanly formatted export file containing the itemized grades, descriptions, and class statistics, ready for archiving or distribution.