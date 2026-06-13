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
        self.is_dark_mode = True
        self.setup_ui()
        self.apply_theme()
        self.load_student()

    def setup_ui(self):
        self.setWindowTitle("Assignment Grading System")
        self.resize(1400, 900)
        self.setLayoutDirection(Qt.LayoutDirection.LeftToRight)
        
        main_app_layout = QVBoxLayout(self)
        main_app_layout.setContentsMargins(10, 10, 10, 10)
        main_app_layout.setSpacing(10)

        top_bar = QHBoxLayout()
        self.btn_theme = QPushButton("☀️")
        self.btn_theme.setFixedSize(40, 40)
        self.btn_theme.setStyleSheet("font-size: 20px; border-radius: 20px;")
        self.btn_theme.setCursor(Qt.CursorShape.PointingHandCursor)
        self.btn_theme.clicked.connect(self.toggle_theme)

        self.btn_toggle_sidebar = QPushButton("☰")
        self.btn_toggle_sidebar.setFixedSize(40, 40)
        self.btn_toggle_sidebar.setStyleSheet("font-size: 20px; border-radius: 6px;")
        self.btn_toggle_sidebar.setCursor(Qt.CursorShape.PointingHandCursor)
        self.btn_toggle_sidebar.clicked.connect(self.toggle_sidebar)

        top_bar.addWidget(self.btn_theme)
        top_bar.addStretch()
        top_bar.addWidget(self.btn_toggle_sidebar)
        main_app_layout.addLayout(top_bar)

        content_layout = QHBoxLayout()
        content_layout.setSpacing(15)

        self.sidebar = QWidget()
        self.sidebar.setFixedWidth(380)
        sidebar_layout = QVBoxLayout(self.sidebar)
        sidebar_layout.setSpacing(15)
        sidebar_layout.setContentsMargins(0, 0, 0, 0)

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

        self.btn_submit = QPushButton("Submit")
        self.btn_submit.setFixedHeight(45)
        self.btn_submit.clicked.connect(self.submit_grade)
        sidebar_layout.addWidget(self.btn_submit)

        nav_layout = QHBoxLayout()
        self.btn_prev = QPushButton("Previous")
        self.btn_next = QPushButton("Next")
        self.btn_prev.setFixedHeight(40)
        self.btn_next.setFixedHeight(40)
        self.btn_prev.clicked.connect(self.go_previous)
        self.btn_next.clicked.connect(self.go_next)
        
        nav_layout.addWidget(self.btn_prev)
        nav_layout.addWidget(self.btn_next)
        sidebar_layout.addLayout(nav_layout)

        content_layout.addWidget(self.sidebar)

        self.pdf_viewer = PDFViewer()
        content_layout.addWidget(self.pdf_viewer)

        main_app_layout.addLayout(content_layout)

    def toggle_sidebar(self):
        self.sidebar.setVisible(not self.sidebar.isVisible())

    def toggle_theme(self):
        self.is_dark_mode = not self.is_dark_mode
        self.btn_theme.setText("☀️" if self.is_dark_mode else "🌙")
        self.apply_theme()

    def apply_theme(self):
        if self.is_dark_mode:
            self.setStyleSheet("""
                QWidget { background-color: #0d1117; color: #e6edf3; }
                QScrollArea { border: 1px solid #30363d; border-radius: 8px; background-color: #161b22; }
                QDoubleSpinBox, QTextEdit, QComboBox, QSpinBox { background-color: #161b22; border: 1px solid #30363d; border-radius: 6px; padding: 5px; color: #e6edf3; }
                QDoubleSpinBox:focus, QTextEdit:focus, QComboBox:focus, QSpinBox:focus { border: 2px solid #58a6ff; }
                QPushButton { background-color: #21262d; color: #c9d1d9; border: 1px solid #30363d; border-radius: 6px; font-weight: bold; font-size: 13px; }
                QPushButton:hover { background-color: #30363d; }
                QPushButton#btn_submit { background-color: #238636; color: white; border: none; }
                QPushButton#btn_submit:hover { background-color: #2ea043; }
                QPushButton:disabled { background-color: #161b22; color: #484f58; border: 1px solid #21262d; }
                QLabel#info_label { font-size: 15px; font-weight: bold; color: #58a6ff; padding: 12px; background-color: #161b22; border-radius: 8px; border: 1px solid #30363d; }
            """)
        else:
            self.setStyleSheet("""
                QWidget { background-color: #f4f5f7; color: #1f2328; }
                QScrollArea { border: 1px solid #d0d7de; border-radius: 8px; background-color: #ffffff; }
                QDoubleSpinBox, QTextEdit, QComboBox, QSpinBox { background-color: #ffffff; border: 1px solid #d0d7de; border-radius: 6px; padding: 5px; }
                QDoubleSpinBox:focus, QTextEdit:focus, QComboBox:focus, QSpinBox:focus { border: 2px solid #0969da; }
                QPushButton { background-color: #f6f8fa; color: #24292f; border: 1px solid #d0d7de; border-radius: 6px; font-weight: bold; font-size: 13px; }
                QPushButton:hover { background-color: #f3f4f6; }
                QPushButton#btn_submit { background-color: #0969da; color: white; border: none; }
                QPushButton#btn_submit:hover { background-color: #0353a4; }
                QPushButton:disabled { background-color: #eaeef2; color: #8c959f; border: 1px solid #d0d7de; }
                QLabel#info_label { font-size: 15px; font-weight: bold; color: #0969da; padding: 12px; background-color: #ffffff; border-radius: 8px; border: 1px solid #d0d7de; }
            """)
        self.info_label.setObjectName("info_label")
        self.btn_submit.setObjectName("btn_submit")

    def load_student(self):
        if not self.dm.students:
            self.info_label.setText("No valid student files found.")
            self.btn_submit.setEnabled(False)
            self.btn_prev.setEnabled(False)
            self.btn_next.setEnabled(False)
            return

        student = self.dm.students[self.current_idx]
        self.info_label.setText(f"Student: {student['name']} {student['surname']}\nID: {student['id']}")

        saved_grades, saved_desc = self.dm.get_saved_data(student['id'])
        for q, sb in self.spinboxes.items():
            sb.setValue(saved_grades.get(q, 0.0))
        self.comments_edit.setPlainText(saved_desc)

        self.pdf_viewer.load_pdf(student['pdf'])

        self.btn_prev.setEnabled(self.current_idx > 0)
        self.btn_next.setEnabled(self.current_idx < len(self.dm.students) - 1)

    def submit_grade(self):
        student = self.dm.students[self.current_idx]
        grades = {q: sb.value() for q, sb in self.spinboxes.items()}
        total = sum(grades.values())
        comments = self.comments_edit.toPlainText()

        self.dm.save_grade(student['id'], grades, total, comments)

    def go_previous(self):
        if self.current_idx > 0:
            self.current_idx -= 1
            self.load_student()

    def go_next(self):
        if self.current_idx < len(self.dm.students) - 1:
            self.current_idx += 1
            self.load_student()