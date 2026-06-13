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
        self.pdf_dir = "PDFs"

        self.setup_excel()
        self.students = self.get_matched_students()

    def setup_excel(self):
        if not os.path.exists(self.output_dir):
            os.makedirs(self.output_dir)
        if not os.path.exists(self.excel_path):
            cols = ['نام', 'نام خانوادگی', 'شماره دانشجویی'] + self.questions + ['مجموع', 'توضیحات']
            df = pd.DataFrame(columns=cols)
            df.to_excel(self.excel_path, index=False)
            self.prefill_excel_students()
        self.apply_excel_formatting()

    def prefill_excel_students(self):
        wb = load_workbook(self.excel_path)
        ws = wb.active
        for _, row in self.students_df.iterrows():
            if row['نام کاربری'] != 'nan' and bool(row['نام کاربری']):
                ws.cell(row=ws.max_row + 1, column=1, value=row['نام'])
                ws.cell(row=ws.max_row, column=2, value=row['نام خانوادگی'])
                ws.cell(row=ws.max_row, column=3, value=row['نام کاربری'])
        wb.save(self.excel_path)

    def apply_excel_formatting(self):
        wb = load_workbook(self.excel_path)
        ws = wb.active
        vazir_font = Font(name='Vazirmatn')
        align_right = Alignment(horizontal='right', vertical='center')
        for row in ws.iter_rows():
            for cell in row:
                cell.font = vazir_font
                cell.alignment = align_right
        ws.sheet_view.rightToLeft = True
        wb.save(self.excel_path)

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

    def save_grade(self, student_id, grades_dict, total, comments):
        wb = load_workbook(self.excel_path)
        ws = wb.active
        row_to_update = None
        for row in range(2, ws.max_row + 2):
            if str(ws.cell(row=row, column=3).value) == str(student_id):
                row_to_update = row
                break

        if row_to_update is not None:
            for idx, q in enumerate(self.questions):
                ws.cell(row=row_to_update, column=4 + idx, value=grades_dict.get(q, 0))

            ws.cell(row=row_to_update, column=4 + len(self.questions), value=total)
            ws.cell(row=row_to_update, column=5 + len(self.questions), value=comments)
            self.apply_excel_formatting()