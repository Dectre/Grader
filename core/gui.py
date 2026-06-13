from PyQt6.QtWidgets import (QWidget, QVBoxLayout, QHBoxLayout, QLabel,
                             QScrollArea, QDoubleSpinBox, QTextEdit, QPushButton,
                             QFormLayout, QComboBox)
from PyQt6.QtCore import Qt
from core.pdf_viewer import PDFViewer

class GradingApp(QWidget):
    def __init__(self, data_manager):
        super().__init__()
        self.dm = data_manager
        self.current_idx = 0
        self.spinboxes = {}
        self.setup_ui()
        self.apply_dark_theme()
        self.load_student()

    def setup_ui(self):
        self.setWindowTitle("Assignment Grading System")
        self.resize(1300, 850)
        self.setLayoutDirection(Qt.LayoutDirection.LeftToRight)
        main_layout = QHBoxLayout(self)
        main_layout.setSpacing(15)
        main_layout.setContentsMargins(20, 20, 20, 20)

        sidebar = QWidget()
        sidebar.setFixedWidth(380)
        sidebar_layout = QVBoxLayout(sidebar)
        sidebar_layout.setSpacing(15)
        sidebar_layout.setContentsMargins(0, 0, 0, 0)

        top_bar = QHBoxLayout()
        self.theme_combo = QComboBox()
        self.theme_combo.addItems(["Dark Mode", "Light Mode"])
        self.theme_combo.currentTextChanged.connect(self.change_theme)
        self.theme_combo.setFixedHeight(35)
        top_bar.addWidget(QLabel("Theme:"))
        top_bar.addWidget(self.theme_combo)
        sidebar_layout.addLayout(top_bar)

        self.info_label = QLabel()
        self.info_label.setWordWrap(True)
        sidebar_layout.addWidget(self.info_label)

        form_layout = QFormLayout()
        form_layout.setSpacing(12)
        for idx, q in enumerate(self.dm.questions):
            max_g = self.dm.max_grades[idx]
            sb = QDoubleSpinBox()
            sb.setRange(0, float(max_g))
            sb.setDecimals(2)
            sb.setFixedHeight(35)
            self.spinboxes[q] = sb
            form_layout.addRow(QLabel(f"{q} (out of {max_g}):"), sb)

        scroll_form = QScrollArea()
        scroll_form_widget = QWidget()
        scroll_form_widget.setLayout(form_layout)
        scroll_form.setWidget(scroll_form_widget)
        scroll_form.setWidgetResizable(True)
        sidebar_layout.addWidget(scroll_form)

        sidebar_layout.addWidget(QLabel("Description:"))
        self.comments_edit = QTextEdit()
        sidebar_layout.addWidget(self.comments_edit)

        self.submit_btn = QPushButton("Submit & Next")
        self.submit_btn.setFixedHeight(45)
        self.submit_btn.clicked.connect(self.submit_grade)
        sidebar_layout.addWidget(self.submit_btn)

        main_layout.addWidget(sidebar)

        self.pdf_viewer = PDFViewer()
        main_layout.addWidget(self.pdf_viewer)

    def change_theme(self, theme_name):
        if theme_name == "Dark Mode":
            self.apply_dark_theme()
        else:
            self.apply_light_theme()

    def apply_light_theme(self):
        self.setStyleSheet("""
            QWidget {
                background-color: #f4f5f7;
                color: #1f2328;
            }
            QScrollArea {
                border: 1px solid #d0d7de;
                border-radius: 8px;
                background-color: #ffffff;
            }
            QDoubleSpinBox, QTextEdit, QComboBox {
                background-color: #ffffff;
                border: 1px solid #d0d7de;
                border-radius: 6px;
                padding: 5px;
            }
            QDoubleSpinBox:focus, QTextEdit:focus, QComboBox:focus {
                border: 2px solid #0969da;
            }
            QPushButton {
                background-color: #0969da;
                color: white;
                border: none;
                border-radius: 6px;
                font-weight: bold;
                font-size: 13px;
            }
            QPushButton:hover {
                background-color: #0353a4;
            }
            QPushButton:disabled {
                background-color: #8c959f;
            }
            QLabel#info_label {
                font-size: 15px;
                font-weight: bold;
                color: #0969da;
                padding: 12px;
                background-color: #ffffff;
                border-radius: 8px;
                border: 1px solid #d0d7de;
            }
        """)
        self.info_label.setObjectName("info_label")

    def apply_dark_theme(self):
        self.setStyleSheet("""
            QWidget {
                background-color: #0d1117;
                color: #e6edf3;
            }
            QScrollArea {
                border: 1px solid #30363d;
                border-radius: 8px;
                background-color: #161b22;
            }
            QDoubleSpinBox, QTextEdit, QComboBox {
                background-color: #161b22;
                border: 1px solid #30363d;
                border-radius: 6px;
                padding: 5px;
                color: #e6edf3;
            }
            QDoubleSpinBox:focus, QTextEdit:focus, QComboBox:focus {
                border: 2px solid #58a6ff;
            }
            QPushButton {
                background-color: #238636;
                color: white;
                border: none;
                border-radius: 6px;
                font-weight: bold;
                font-size: 13px;
            }
            QPushButton:hover {
                background-color: #2ea043;
            }
            QPushButton:disabled {
                background-color: #21262d;
                color: #8b949e;
            }
            QLabel#info_label {
                font-size: 15px;
                font-weight: bold;
                color: #58a6ff;
                padding: 12px;
                background-color: #161b22;
                border-radius: 8px;
                border: 1px solid #30363d;
            }
        """)
        self.info_label.setObjectName("info_label")

    def load_student(self):
        if self.current_idx >= len(self.dm.students):
            self.info_label.setText("All files have been reviewed.")
            self.pdf_viewer.clear()
            self.submit_btn.setEnabled(False)
            return

        student = self.dm.students[self.current_idx]
        self.info_label.setText(f"Student: {student['name']} {student['surname']}\nID: {student['id']}")

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