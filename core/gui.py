from PyQt6.QtWidgets import (QWidget, QVBoxLayout, QHBoxLayout, QLabel,
                             QScrollArea, QDoubleSpinBox, QTextEdit, QPushButton,
                             QFormLayout)
from PyQt6.QtCore import Qt
from core.pdf_viewer import PDFViewer

class GradingApp(QWidget):
    def __init__(self, data_manager):
        super().__init__()
        self.dm = data_manager
        self.current_idx = 0
        self.spinboxes = {}
        self.setup_ui()
        self.load_student()

    def setup_ui(self):
        self.setWindowTitle("سیستم نمره‌دهی تمرینات")
        self.resize(1200, 800)
        self.setLayoutDirection(Qt.LayoutDirection.RightToLeft)
        main_layout = QHBoxLayout(self)

        sidebar = QWidget()
        sidebar.setFixedWidth(350)
        sidebar_layout = QVBoxLayout(sidebar)

        self.info_label = QLabel()
        self.info_label.setStyleSheet("font-size: 16px; font-weight: bold; margin-bottom: 10px;")
        sidebar_layout.addWidget(self.info_label)

        form_layout = QFormLayout()
        for idx, q in enumerate(self.dm.questions):
            max_g = self.dm.max_grades[idx]
            sb = QDoubleSpinBox()
            sb.setRange(0, float(max_g))
            sb.setDecimals(2)
            self.spinboxes[q] = sb
            form_layout.addRow(QLabel(f"{q} (از {max_g}):"), sb)

        scroll_form = QScrollArea()
        scroll_form_widget = QWidget()
        scroll_form_widget.setLayout(form_layout)
        scroll_form.setWidget(scroll_form_widget)
        scroll_form.setWidgetResizable(True)
        sidebar_layout.addWidget(scroll_form)

        sidebar_layout.addWidget(QLabel("توضیحات:"))
        self.comments_edit = QTextEdit()
        sidebar_layout.addWidget(self.comments_edit)

        self.submit_btn = QPushButton("ثبت و بعدی")
        self.submit_btn.setStyleSheet("padding: 10px; font-weight: bold;")
        self.submit_btn.clicked.connect(self.submit_grade)
        sidebar_layout.addWidget(self.submit_btn)

        main_layout.addWidget(sidebar)

        self.pdf_viewer = PDFViewer()
        main_layout.addWidget(self.pdf_viewer)

    def load_student(self):
        if self.current_idx >= len(self.dm.students):
            self.info_label.setText("تمام فایل‌ها بررسی شدند.")
            self.pdf_viewer.clear()
            self.submit_btn.setEnabled(False)
            return

        student = self.dm.students[self.current_idx]
        self.info_label.setText(f"دانشجو: {student['name']} {student['surname']}\nشماره دانشجویی: {student['id']}")

        for sb in self.spinboxes.values():
            sb.setValue(0.0)
        self.comments_edit.clear()

        self.pdf_viewer.load_pdf(student['pdf'])

    def submit_grade(self):
        student = self.dm.students[self.current_idx]
        grades = {q: sb.value() for q, sb in self.spinboxes.items()}
        total = sum(grades.values())
        comments = self.comments_edit.toPlainText()

        self.dm.save_grade(student['id'], grades, total, comments)

        self.current_idx += 1
        self.load_student()