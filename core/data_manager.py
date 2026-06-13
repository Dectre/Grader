import os
import pandas as pd
from openpyxl import load_workbook
from openpyxl.styles import Font, Alignment

class DataManager:
    def __init__(self):
        self.students_df = pd.read_csv(os.path.join("Data", "students.csv"), encoding='utf-8-sig')
        self.students_df['نام کاربری'] = self.students_df['نام کاربری'].astype(str).str.replace(r'\.0$', '', regex=True).str.strip()
        self.students_df = self.students_df.sort_values(by='نام کاربری')

        self.rubric_df = pd.read_csv(os.path.join("Data", "rubric.csv"), encoding='utf-8-sig')
        self.questions = self.rubric_df['بخش'].tolist()
        self.max_grades = self.rubric_df['نمره'].tolist()

        self.output_dir = "output"
        self.excel_path = os.path.join(self.output_dir, "grades.xlsx")
        self.csv_path = os.path.join(self.output_dir, "grades.csv")
        self.pdf_dir = "PDFs"

        self.setup_dataframes()
        self.students = self.get_matched_students()

    def setup_dataframes(self):
        if not os.path.exists(self.output_dir):
            os.makedirs(self.output_dir)

        cols = ['First Name', 'Last Name', 'Student ID'] + self.questions + ['Total Score', 'Description']

        if os.path.exists(self.excel_path):
            self.grades_df = pd.read_excel(self.excel_path)
            self.grades_df['Student ID'] = self.grades_df['Student ID'].astype(str).str.replace(r'\.0$', '', regex=True).str.strip()
        else:
            self.grades_df = pd.DataFrame(columns=cols)
            new_rows = []
            for _, row in self.students_df.iterrows():
                if row['نام کاربری'] != 'nan' and bool(row['نام کاربری']):
                    new_rows.append({
                        'First Name': row['نام'],
                        'Last Name': row['نام خانوادگی'],
                        'Student ID': row['نام کاربری']
                    })
            if new_rows:
                self.grades_df = pd.concat([self.grades_df, pd.DataFrame(new_rows)], ignore_index=True)
            self.export_files()

    def export_files(self):
        self.grades_df.to_excel(self.excel_path, index=False)
        self.grades_df.to_csv(self.csv_path, index=False, encoding='utf-8-sig')
        self.apply_excel_formatting()

    def apply_excel_formatting(self):
        try:
            wb = load_workbook(self.excel_path)
            ws = wb.active
            vazir_font = Font(name='Vazirmatn')
            align_left = Alignment(horizontal='left', vertical='center')
            for row in ws.iter_rows():
                for cell in row:
                    cell.font = vazir_font
                    cell.alignment = align_left
            ws.sheet_view.rightToLeft = False
            wb.save(self.excel_path)
        except Exception:
            pass

    def get_matched_students(self):
        students_list = []
        for _, row in self.students_df.iterrows():
            student_id = row['نام کاربری']
            if student_id == 'nan' or not student_id:
                continue
            pdf_path = self.find_pdf(student_id)
            if pdf_path:
                students_list.append({
                    'name': row['نام'],
                    'surname': row['نام خانوادگی'],
                    'id': student_id,
                    'pdf': pdf_path
                })
        return students_list

    def find_pdf(self, student_id):
        if not os.path.exists(self.pdf_dir):
            return None
        for filename in os.listdir(self.pdf_dir):
            if str(student_id) in filename and filename.lower().endswith('.pdf'):
                return os.path.join(self.pdf_dir, filename)
        return None

    def get_saved_data(self, student_id):
        idx = self.grades_df.index[self.grades_df['Student ID'] == str(student_id)].tolist()
        if idx:
            row = self.grades_df.iloc[idx[0]]
            grades = {}
            for q in self.questions:
                val = row.get(q, 0.0)
                grades[q] = 0.0 if pd.isna(val) else float(val)
            desc = row.get('Description', '')
            desc = "" if pd.isna(desc) else str(desc)
            return grades, desc
        return {q: 0.0 for q in self.questions}, ""

    def save_grade(self, student_id, grades_dict, total, comments):
        idx = self.grades_df.index[self.grades_df['Student ID'] == str(student_id)].tolist()
        if idx:
            row_idx = idx[0]
            for q in self.questions:
                self.grades_df.at[row_idx, q] = grades_dict.get(q, 0)
            self.grades_df.at[row_idx, 'Total Score'] = total
            self.grades_df.at[row_idx, 'Description'] = comments

            self.export_files()