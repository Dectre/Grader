import os
import shutil
import logging
import traceback
from PyQt6.QtWidgets import (QWidget, QVBoxLayout, QHBoxLayout, QLabel,
                             QScrollArea, QDoubleSpinBox, QTextEdit, QPushButton,
                             QFormLayout, QFileDialog)
from PyQt6.QtCore import Qt, QTimer
from core.pdf_viewer import PDFViewer

logging.basicConfig(filename='error_log.txt', level=logging.ERROR, 
                    format='%(asctime)s - %(levelname)s - %(message)s')

class GradingApp(QWidget):
    def __init__(self, data_manager):
        super().__init__()
        self.dm = data_manager
        self.current_idx = 0
        self.spinboxes = {}
        self.is_dark_mode = True
        self.is_loading = False
        self.session_edits = {}
        self.setup_ui()
        self.apply_theme()
        self.pdf_viewer.change_file_callback = self.request_change_file
        self.load_student()

    def setup_ui(self):
        self.setWindowTitle("Assignment Grading System")
        self.resize(1400, 900)
        self.setLayoutDirection(Qt.LayoutDirection.LeftToRight)
        
        main_app_layout = QVBoxLayout(self)
        main_app_layout.setContentsMargins(10, 5, 10, 10)
        main_app_layout.setSpacing(5)

        top_bar = QHBoxLayout()
        top_bar.setContentsMargins(0, 0, 0, 0)
        
        self.btn_toggle_sidebar = QPushButton("☰")
        self.btn_toggle_sidebar.setFixedSize(35, 35)
        self.btn_toggle_sidebar.setStyleSheet("font-size: 18px; border-radius: 6px;")
        self.btn_toggle_sidebar.setCursor(Qt.CursorShape.PointingHandCursor)
        self.btn_toggle_sidebar.clicked.connect(self.toggle_sidebar)

        self.btn_toggle_pdf_toolbar = QPushButton("⌃")
        self.btn_toggle_pdf_toolbar.setFixedSize(35, 35)
        self.btn_toggle_pdf_toolbar.setStyleSheet("font-size: 18px; border-radius: 6px;")
        self.btn_toggle_pdf_toolbar.setCursor(Qt.CursorShape.PointingHandCursor)
        self.btn_toggle_pdf_toolbar.clicked.connect(self.toggle_pdf_toolbar)

        self.btn_theme = QPushButton("☀️")
        self.btn_theme.setFixedSize(35, 35)
        self.btn_theme.setStyleSheet("font-size: 18px; border-radius: 17px;")
        self.btn_theme.setCursor(Qt.CursorShape.PointingHandCursor)
        self.btn_theme.clicked.connect(self.toggle_theme)

        top_bar.addWidget(self.btn_toggle_sidebar)
        top_bar.addWidget(self.btn_toggle_pdf_toolbar)
        top_bar.addStretch()
        top_bar.addWidget(self.btn_theme)
        
        main_app_layout.addLayout(top_bar)

        content_layout = QHBoxLayout()
        content_layout.setSpacing(10)

        self.sidebar = QWidget()
        self.sidebar.setFixedWidth(380)
        sidebar_layout = QVBoxLayout(self.sidebar)
        sidebar_layout.setSpacing(15)
        sidebar_layout.setContentsMargins(0, 0, 0, 0)

        self.info_label = QLabel()
        self.info_label.setWordWrap(True)
        sidebar_layout.addWidget(self.info_label)
        
        self.status_label = QLabel("Status: Not Submitted")
        sidebar_layout.addWidget(self.status_label)

        form_layout = QFormLayout()
        form_layout.setSpacing(12)
        for idx, q in enumerate(self.dm.questions):
            max_g = self.dm.max_grades[idx]
            sb = QDoubleSpinBox()
            sb.setRange(0, float(max_g))
            sb.setDecimals(2)
            sb.setFixedHeight(35)
            sb.valueChanged.connect(self.on_input_changed)
            self.spinboxes[q] = sb
            form_layout.addRow(QLabel(f"{q} \u200E(out of {max_g}):"), sb)

        scroll_form = QScrollArea()
        scroll_form_widget = QWidget()
        scroll_form_widget.setLayout(form_layout)
        scroll_form.setWidget(scroll_form_widget)
        scroll_form.setWidgetResizable(True)
        sidebar_layout.addWidget(scroll_form)

        sidebar_layout.addWidget(QLabel("Description:"))
        self.comments_edit = QTextEdit()
        self.comments_edit.textChanged.connect(self.on_input_changed)
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

    def on_input_changed(self):
        if self.is_loading or not self.dm.students:
            return
            
        student_id = self.dm.students[self.current_idx]['id']
        current_grades = {q: sb.value() for q, sb in self.spinboxes.items()}
        current_comments = self.comments_edit.toPlainText()
        
        self.session_edits[student_id] = {
            'grades': current_grades,
            'comments': current_comments
        }
        self.update_status_label()

    def update_status_label(self):
        if not self.dm.students:
            return
            
        student_id = self.dm.students[self.current_idx]['id']
        is_sub = self.dm.is_submitted(student_id)
        
        current_grades = {q: sb.value() for q, sb in self.spinboxes.items()}
        current_desc = self.comments_edit.toPlainText()
        saved_grades, saved_desc = self.dm.get_saved_data(student_id)
        
        changed = (current_grades != saved_grades) or (current_desc != saved_desc)
        
        if not is_sub:
            status = "Not Submitted (Editing...)" if changed else "Not Submitted"
            color = "#d73a49" if not self.is_dark_mode else "#ff7b72"
        else:
            status = "Unsaved Changes" if changed else "Submitted"
            color = "#dbab09" if changed else ("#28a745" if not self.is_dark_mode else "#3fb950")
            
        self.status_label.setText(f"Status: {status}")
        self.status_label.setStyleSheet(f"font-size: 14px; font-weight: bold; color: {color};")

    def toggle_sidebar(self):
        self.sidebar.setVisible(not self.sidebar.isVisible())
        if self.pdf_viewer.current_fit_mode == 'width':
            self.pdf_viewer.fit_width()
        elif self.pdf_viewer.current_fit_mode == 'screen':
            self.pdf_viewer.fit_screen()

    def toggle_pdf_toolbar(self):
        self.pdf_viewer.is_toolbar_expanded = not self.pdf_viewer.is_toolbar_expanded
        self.pdf_viewer.toolbar_widget.setVisible(self.pdf_viewer.is_toolbar_expanded)
        self.btn_toggle_pdf_toolbar.setText("⌃" if self.pdf_viewer.is_toolbar_expanded else "⌄")

    def toggle_theme(self):
        self.is_dark_mode = not self.is_dark_mode
        self.btn_theme.setText("☀️" if self.is_dark_mode else "🌙")
        self.apply_theme()
        self.update_status_label()

    def apply_theme(self):
        font_css = "font-family: 'Inter', 'Vazirmatn';"
        if self.is_dark_mode:
            self.setStyleSheet(f"""
                QWidget {{ {font_css} background-color: #0d1117; color: #e6edf3; }}
                QScrollArea {{ border: 1px solid #30363d; border-radius: 8px; background-color: #161b22; }}
                QDoubleSpinBox, QTextEdit, QComboBox, QSpinBox {{ background-color: #161b22; border: 1px solid #30363d; border-radius: 6px; padding: 5px; color: #e6edf3; }}
                QDoubleSpinBox:focus, QTextEdit:focus, QComboBox:focus, QSpinBox:focus {{ border: 2px solid #58a6ff; }}
                QPushButton {{ background-color: #21262d; color: #c9d1d9; border: 1px solid #30363d; border-radius: 6px; font-weight: bold; font-size: 13px; }}
                QPushButton:hover {{ background-color: #30363d; }}
                QPushButton#btn_submit {{ background-color: #1f6feb; color: white; border: none; }}
                QPushButton#btn_submit:hover {{ background-color: #388bfd; }}
                QPushButton:disabled {{ background-color: #161b22; color: #484f58; border: 1px solid #21262d; }}
                QLabel#info_label {{ font-size: 15px; font-weight: bold; color: #58a6ff; padding: 12px; background-color: #161b22; border-radius: 8px; border: 1px solid #30363d; }}
                QLabel[class="no-file-lbl"] {{ font-size: 18px; font-weight: bold; color: #8b949e; }}
                QPushButton[class="toolbar-btn"] {{ background-color: #1f6feb; color: white; border: none; padding: 5px 15px; }}
                QPushButton[class="toolbar-btn"]:hover {{ background-color: #388bfd; }}
                QPushButton[class="toolbar-btn"]:disabled {{ background-color: #21262d; color: #484f58; border: 1px solid #30363d; }}
            """)
        else:
            self.setStyleSheet(f"""
                QWidget {{ {font_css} background-color: #f4f5f7; color: #1f2328; }}
                QScrollArea {{ border: 1px solid #d0d7de; border-radius: 8px; background-color: #ffffff; }}
                QDoubleSpinBox, QTextEdit, QComboBox, QSpinBox {{ background-color: #ffffff; border: 1px solid #d0d7de; border-radius: 6px; padding: 5px; }}
                QDoubleSpinBox:focus, QTextEdit:focus, QComboBox:focus, QSpinBox:focus {{ border: 2px solid #0969da; }}
                QPushButton {{ background-color: #f6f8fa; color: #24292f; border: 1px solid #d0d7de; border-radius: 6px; font-weight: bold; font-size: 13px; }}
                QPushButton:hover {{ background-color: #f3f4f6; }}
                QPushButton#btn_submit {{ background-color: #0969da; color: white; border: none; }}
                QPushButton#btn_submit:hover {{ background-color: #0353a4; }}
                QPushButton:disabled {{ background-color: #eaeef2; color: #8c959f; border: 1px solid #d0d7de; }}
                QLabel#info_label {{ font-size: 15px; font-weight: bold; color: #0969da; padding: 12px; background-color: #ffffff; border-radius: 8px; border: 1px solid #d0d7de; }}
                QLabel[class="no-file-lbl"] {{ font-size: 18px; font-weight: bold; color: #57606a; }}
                QPushButton[class="toolbar-btn"] {{ background-color: #0969da; color: white; border: none; padding: 5px 15px; }}
                QPushButton[class="toolbar-btn"]:hover {{ background-color: #0353a4; }}
                QPushButton[class="toolbar-btn"]:disabled {{ background-color: #eaeef2; color: #8c959f; border: 1px solid #d0d7de; }}
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

        self.is_loading = True
        student = self.dm.students[self.current_idx]
        self.info_label.setText(f"Student: {student['name']} {student['surname']}\nID: {student['id']}")

        if student['id'] in self.session_edits:
            grades = self.session_edits[student['id']]['grades']
            comments = self.session_edits[student['id']]['comments']
        else:
            grades, comments = self.dm.get_saved_data(student['id'])

        for q, sb in self.spinboxes.items():
            sb.setValue(grades.get(q, 0.0))
            if student['pdf'] is None:
                sb.setValue(0.0)
                sb.setEnabled(False)
            else:
                sb.setEnabled(True)
                
        self.comments_edit.setPlainText(comments)
        self.pdf_viewer.load_pdf(student['pdf'])

        self.btn_prev.setEnabled(self.current_idx > 0)
        self.btn_next.setEnabled(self.current_idx < len(self.dm.students) - 1)
        
        self.is_loading = False
        self.update_status_label()

    def request_change_file(self):
        student = self.dm.students[self.current_idx]
        file_path, _ = QFileDialog.getOpenFileName(self, "Select PDF", "", "PDF Files (*.pdf)")
        if file_path:
            if not os.path.exists(self.dm.pdf_dir):
                os.makedirs(self.dm.pdf_dir)
                
            if student['pdf']:
                new_name = os.path.basename(student['pdf'])
            else:
                new_name = f"{student['name']}_{student['surname']}_{student['id']}.pdf"
                
            dest_path = os.path.join(self.dm.pdf_dir, new_name)
            shutil.copy(file_path, dest_path)
            student['pdf'] = dest_path
            self.load_student()

    def submit_grade(self):
        student = self.dm.students[self.current_idx]
        grades = {q: sb.value() for q, sb in self.spinboxes.items()}
        total = sum(grades.values())
        comments = self.comments_edit.toPlainText()

        try:
            self.dm.save_grade(student['id'], grades, total, comments)
            self.update_status_label()
            self.btn_submit.setText("Successfully submitted!")
            self.btn_submit.setStyleSheet("background-color: #238636; color: white; border: none; font-weight: bold; border-radius: 6px;")
            QTimer.singleShot(1500, self.reset_submit_btn)
        except Exception as e:
            logging.error(f"Error saving grade for student {student['id']}: {str(e)}\n{traceback.format_exc()}")
            self.btn_submit.setText("Error! Check logs.")
            self.btn_submit.setStyleSheet("background-color: #da3633; color: white; border: none; font-weight: bold; border-radius: 6px;")
            QTimer.singleShot(2000, self.reset_submit_btn)

    def reset_submit_btn(self):
        self.btn_submit.setText("Submit")
        self.btn_submit.setStyleSheet("")

    def go_previous(self):
        if self.current_idx > 0:
            self.current_idx -= 1
            self.load_student()

    def go_next(self):
        if self.current_idx < len(self.dm.students) - 1:
            self.current_idx += 1
            self.load_student()