import os
import shutil
import logging
import traceback
import pandas as pd
from PyQt6.QtWidgets import (QWidget, QVBoxLayout, QHBoxLayout, QLabel,
                             QScrollArea, QDoubleSpinBox, QTextEdit, QPushButton,
                             QFormLayout, QFileDialog, QComboBox, QLineEdit, QCompleter, QSplitter)
from PyQt6.QtCore import Qt, QTimer, QStringListModel, QObject, QEvent
from core.pdf_viewer import PDFViewer

logging.basicConfig(filename='error_log.txt', level=logging.ERROR, 
                    format='%(asctime)s - %(levelname)s - %(message)s')

class WheelBlocker(QObject):
    def eventFilter(self, obj, event):
        if event.type() == QEvent.Type.Wheel:
            return True
        return False

class GradingApp(QWidget):
    def __init__(self, data_manager):
        super().__init__()
        self.dm = data_manager
        self.current_idx = 0
        self.spinboxes = {}
        self.lock_buttons = {}
        self.lock_states = {}
        self.is_dark_mode = True
        self.is_loading = False
        self.is_desc_locked = False
        self.session_edits = {}
        self.wheel_blocker = WheelBlocker()
        self.setup_ui()
        self.apply_theme()
        self.pdf_viewer.change_file_callback = self.request_change_file
        self.populate_student_dropdown()
        self.load_student()

    def setup_ui(self):
        self.setWindowTitle("Assignment Grading System")
        self.resize(1200, 800)
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

        self.btn_search = QPushButton("🔍")
        self.btn_search.setFixedSize(35, 35)
        self.btn_search.setStyleSheet("font-size: 16px; border-radius: 6px;")
        self.btn_search.setCursor(Qt.CursorShape.PointingHandCursor)
        self.btn_search.clicked.connect(self.toggle_search)

        self.search_input = QLineEdit()
        self.search_input.setPlaceholderText("Search Name or ID...")
        self.search_input.setFixedHeight(35)
        self.search_input.setMinimumWidth(200)
        self.search_input.setMaximumWidth(300)
        self.search_input.setVisible(False)
        
        self.search_completer = QCompleter()
        self.search_completer.setCaseSensitivity(Qt.CaseSensitivity.CaseInsensitive)
        self.search_completer.setFilterMode(Qt.MatchFlag.MatchContains)
        self.search_input.setCompleter(self.search_completer)
        
        self.search_completer.activated.connect(self.on_search_selected)
        self.search_input.returnPressed.connect(self.execute_search)

        self.btn_toggle_stats = QPushButton("📊")
        self.btn_toggle_stats.setFixedSize(35, 35)
        self.btn_toggle_stats.setStyleSheet("font-size: 18px; border-radius: 6px;")
        self.btn_toggle_stats.setCursor(Qt.CursorShape.PointingHandCursor)
        self.btn_toggle_stats.clicked.connect(self.toggle_stats_sidebar)

        self.btn_theme = QPushButton("☀️")
        self.btn_theme.setFixedSize(35, 35)
        self.btn_theme.setStyleSheet("font-size: 18px; border-radius: 17px;")
        self.btn_theme.setCursor(Qt.CursorShape.PointingHandCursor)
        self.btn_theme.clicked.connect(self.toggle_theme)

        top_bar.addWidget(self.btn_toggle_sidebar)
        top_bar.addWidget(self.btn_toggle_pdf_toolbar)
        top_bar.addWidget(self.btn_search)
        top_bar.addWidget(self.search_input)
        top_bar.addStretch()
        top_bar.addWidget(self.btn_toggle_stats)
        top_bar.addWidget(self.btn_theme)
        
        main_app_layout.addLayout(top_bar)

        self.main_splitter = QSplitter(Qt.Orientation.Horizontal)
        self.main_splitter.setHandleWidth(8)

        self.sidebar = QWidget()
        sidebar_layout = QVBoxLayout(self.sidebar)
        sidebar_layout.setSpacing(10)
        sidebar_layout.setContentsMargins(0, 0, 0, 0)

        self.student_combo = QComboBox()
        self.student_combo.setFixedHeight(35)
        self.student_combo.currentIndexChanged.connect(self.on_dropdown_changed)
        sidebar_layout.addWidget(self.student_combo)

        self.info_label = QLabel()
        self.info_label.setWordWrap(True)
        sidebar_layout.addWidget(self.info_label)
        
        status_score_layout = QHBoxLayout()
        self.status_label = QLabel("Status: Not Submitted Yet")
        self.total_score_label = QLabel("Total: 0.00")
        status_score_layout.addWidget(self.status_label)
        status_score_layout.addWidget(self.total_score_label)
        sidebar_layout.addLayout(status_score_layout)

        vertical_splitter = QSplitter(Qt.Orientation.Vertical)
        vertical_splitter.setHandleWidth(6)

        form_layout = QFormLayout()
        form_layout.setSpacing(12)
        for idx, q in enumerate(self.dm.questions):
            max_g = self.dm.max_grades[idx]
            
            row_widget = QWidget()
            row_container = QHBoxLayout(row_widget)
            row_container.setContentsMargins(0, 0, 0, 0)
            row_container.setSpacing(5)
            
            sb = QDoubleSpinBox()
            sb.setRange(0, float(max_g))
            sb.setDecimals(2)
            sb.setFixedHeight(35)
            sb.installEventFilter(self.wheel_blocker)
            sb.valueChanged.connect(self.on_input_changed)
            self.spinboxes[q] = sb
            
            btn_q_lock = QPushButton("🔓")
            btn_q_lock.setFixedSize(30, 35)
            btn_q_lock.setCursor(Qt.CursorShape.PointingHandCursor)
            btn_q_lock.setStyleSheet("background: transparent; border: none; font-size: 16px;")
            btn_q_lock.clicked.connect(lambda checked, question=q: self.toggle_question_lock(question))
            self.lock_buttons[q] = btn_q_lock
            self.lock_states[q] = False
            
            row_container.addWidget(sb)
            row_container.addWidget(btn_q_lock)
            
            form_layout.addRow(QLabel(f"{q} \u200E(out of {max_g}):"), row_widget)

        scroll_form = QScrollArea()
        scroll_form_widget = QWidget()
        scroll_form_widget.setLayout(form_layout)
        scroll_form.setWidget(scroll_form_widget)
        scroll_form.setWidgetResizable(True)
        vertical_splitter.addWidget(scroll_form)

        desc_widget = QWidget()
        desc_layout = QVBoxLayout(desc_widget)
        desc_layout.setContentsMargins(0, 0, 0, 0)
        desc_layout.setSpacing(5)
        
        desc_header = QWidget()
        desc_header_layout = QHBoxLayout(desc_header)
        desc_header_layout.setContentsMargins(0, 0, 0, 0)
        desc_header_layout.addWidget(QLabel("Description:"))
        desc_header_layout.addStretch()
        
        self.btn_lock_desc = QPushButton("🔓")
        self.btn_lock_desc.setFixedSize(30, 25)
        self.btn_lock_desc.setCursor(Qt.CursorShape.PointingHandCursor)
        self.btn_lock_desc.setStyleSheet("background: transparent; border: none; font-size: 16px;")
        self.btn_lock_desc.clicked.connect(self.toggle_desc_lock)
        desc_header_layout.addWidget(self.btn_lock_desc)
        
        desc_layout.addWidget(desc_header)
        
        self.comments_edit = QTextEdit()
        self.comments_edit.textChanged.connect(self.on_input_changed)
        desc_layout.addWidget(self.comments_edit)
        vertical_splitter.addWidget(desc_widget)

        vertical_splitter.setStretchFactor(0, 2)
        vertical_splitter.setStretchFactor(1, 1)
        sidebar_layout.addWidget(vertical_splitter)

        submit_layout = QHBoxLayout()
        self.btn_clear = QPushButton("Clear")
        self.btn_clear.setFixedHeight(45)
        self.btn_clear.clicked.connect(self.clear_inputs)
        
        self.btn_submit = QPushButton("Submit")
        self.btn_submit.setFixedHeight(45)
        self.btn_submit.clicked.connect(self.submit_grade)
        
        submit_layout.addWidget(self.btn_clear)
        submit_layout.addWidget(self.btn_submit)
        sidebar_layout.addLayout(submit_layout)

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

        self.main_splitter.addWidget(self.sidebar)

        self.pdf_viewer = PDFViewer()
        self.main_splitter.addWidget(self.pdf_viewer)

        self.stats_sidebar = QWidget()
        self.stats_sidebar.setVisible(False)
        stats_layout = QVBoxLayout(self.stats_sidebar)
        stats_layout.setContentsMargins(5, 0, 0, 0)
        stats_title = QLabel("📊 Live Statistics")
        stats_title.setStyleSheet("font-size: 16px; font-weight: bold; padding: 5px;")
        stats_layout.addWidget(stats_title)
        self.stats_text = QTextEdit()
        self.stats_text.setReadOnly(True)
        stats_layout.addWidget(self.stats_text)
        self.main_splitter.addWidget(self.stats_sidebar)

        self.main_splitter.setStretchFactor(0, 20)
        self.main_splitter.setStretchFactor(1, 60)
        self.main_splitter.setStretchFactor(2, 20)

        main_app_layout.addWidget(self.main_splitter)

    def toggle_question_lock(self, q):
        self.lock_states[q] = not self.lock_states[q]
        self.lock_buttons[q].setText("🔒" if self.lock_states[q] else "🔓")
        student = self.dm.students[self.current_idx]
        if student['pdf'] is not None:
            self.spinboxes[q].setReadOnly(self.lock_states[q])

    def toggle_desc_lock(self):
        self.is_desc_locked = not self.is_desc_locked
        self.btn_lock_desc.setText("🔒" if self.is_desc_locked else "🔓")
        self.comments_edit.setReadOnly(self.is_desc_locked)

    def toggle_stats_sidebar(self):
        is_visible = self.stats_sidebar.isVisible()
        self.stats_sidebar.setVisible(not is_visible)
        if not is_visible:
            self.update_stats_sidebar()
            self.main_splitter.setSizes([int(self.width() * 0.20), int(self.width() * 0.50), int(self.width() * 0.30)])
        else:
            self.main_splitter.setSizes([int(self.width() * 0.25), int(self.width() * 0.75), 0])
        QTimer.singleShot(50, self.refresh_pdf_size)

    def update_stats_sidebar(self):
        if not self.stats_sidebar.isVisible():
            return
            
        total = len(self.dm.students)
        with_pdf = sum(1 for s in self.dm.students if s['pdf'])
        without_pdf = total - with_pdf
        
        submitted = 0
        unsaved = 0
        not_sub = 0
        
        for s in self.dm.students:
            sid = s['id']
            is_sub = self.dm.is_submitted(sid)
            saved_grades, saved_desc = self.dm.get_saved_data(sid)
            
            changed = False
            if sid in self.session_edits:
                cg = self.session_edits[sid]['grades']
                cc = self.session_edits[sid]['comments']
                if cg != saved_grades or cc != saved_desc:
                    changed = True
                    
            if changed:
                unsaved += 1
            elif is_sub:
                submitted += 1
            else:
                not_sub += 1

        text_color = "#e6edf3" if self.is_dark_mode else "#1f2328"
        border_color = "#30363d" if self.is_dark_mode else "#d0d7de"
        header_bg = "#21262d" if self.is_dark_mode else "#f6f8fa"
        
        html = f"""
        <div style="color: {text_color}; font-size: 12px; font-family: 'Inter', 'Vazirmatn';">
            <p><b>Total Students:</b> {total}</p>
            <p><b>Files:</b> {with_pdf} PDFs / {without_pdf} Missing</p>
            <p><b>Submissions:</b> {submitted} Sub / {unsaved} Unsav / {not_sub} Not</p>
            <br>
            <table style="border-collapse: collapse; width: 100%; border: 1px solid {border_color};">
                <thead>
                    <tr style="background-color: {header_bg}; font-weight: bold; text-align: center;">
                        <th style="border: 1px solid {border_color}; padding: 4px; text-align: left;">Sec</th>
                        <th style="border: 1px solid {border_color}; padding: 4px;">Avg</th>
                        <th style="border: 1px solid {border_color}; padding: 4px;">Med</th>
                        <th style="border: 1px solid {border_color}; padding: 4px;">Max</th>
                        <th style="border: 1px solid {border_color}; padding: 4px;">Min</th>
                    </tr>
                </thead>
                <tbody>
        """
        
        df = self.dm.grades_df[self.dm.grades_df['_Submitted'] == True]
        cols_to_check = self.dm.questions + ['Total Score']
        
        for col in cols_to_check:
            if col in df.columns:
                scores = pd.to_numeric(df[col], errors='coerce').dropna()
                mean_v = round(scores.mean(), 2) if not scores.empty else 0
                median_v = round(scores.median(), 2) if not scores.empty else 0
                max_v = round(scores.max(), 2) if not scores.empty else 0
                min_v = round(scores.min(), 2) if not scores.empty else 0
                
                title_col = col if col != 'Total Score' else '<b>TOTAL</b>'
                html += f"""
                    <tr style="text-align: center;">
                        <td style="border: 1px solid {border_color}; padding: 4px; text-align: left;">{title_col}</td>
                        <td style="border: 1px solid {border_color}; padding: 4px;">{mean_v}</td>
                        <td style="border: 1px solid {border_color}; padding: 4px;">{median_v}</td>
                        <td style="border: 1px solid {border_color}; padding: 4px;">{max_v}</td>
                        <td style="border: 1px solid {border_color}; padding: 4px;">{min_v}</td>
                    </tr>
                """
                
        html += "</tbody></table></div>"
        self.stats_text.setHtml(html)

    def populate_student_dropdown(self):
        current_index = self.student_combo.currentIndex()
        self.student_combo.blockSignals(True)
        self.student_combo.clear()
        search_terms = []
        for idx, s in enumerate(self.dm.students):
            file_icon = "✅" if s['pdf'] else "❌"
            
            is_sub = self.dm.is_submitted(s['id'])
            saved_grades, saved_desc = self.dm.get_saved_data(s['id'])
            
            changed = False
            if s['id'] in self.session_edits:
                cg = self.session_edits[s['id']]['grades']
                cc = self.session_edits[s['id']]['comments']
                if cg != saved_grades or cc != saved_desc:
                    changed = True
                    
            if changed:
                status_icon = "🟡 Unsaved"
            elif is_sub:
                status_icon = "🟢 Submitted"
            else:
                status_icon = "🔴 Not Submitted"

            display_text = f"{file_icon} | {status_icon} | {s['name']} {s['surname']}"
            self.student_combo.addItem(display_text, userData=idx)
            search_terms.append(f"{s['name']} {s['surname']} - {s['id']}")
            
        self.search_model = QStringListModel(search_terms)
        self.search_completer.setModel(self.search_model)
        
        if 0 <= current_index < self.student_combo.count():
            self.student_combo.setCurrentIndex(current_index)
            
        self.student_combo.blockSignals(False)
        self.update_stats_sidebar()

    def on_dropdown_changed(self, index):
        if not self.is_loading and index >= 0:
            self.current_idx = index
            self.load_student()

    def toggle_sidebar(self):
        is_visible = self.sidebar.isVisible()
        self.sidebar.setVisible(not is_visible)
        QTimer.singleShot(50, self.refresh_pdf_size)

    def toggle_search(self):
        is_visible = self.search_input.isVisible()
        self.search_input.setVisible(not is_visible)
        if not is_visible:
            self.search_input.setFocus()
            self.search_input.selectAll()

    def on_search_selected(self, text):
        if "-" in text:
            sid = text.split("-")[-1].strip()
            for idx, s in enumerate(self.dm.students):
                if s['id'] == sid:
                    self.student_combo.setCurrentIndex(idx)
                    self.search_input.clear()
                    self.search_input.setVisible(False)
                    return

    def execute_search(self):
        query = self.search_input.text().strip().lower()
        if not query:
            return
            
        for idx, s in enumerate(self.dm.students):
            if query in s['id'].lower() or query in s['name'].lower() or query in s['surname'].lower():
                self.student_combo.setCurrentIndex(idx)
                self.search_input.clear()
                self.search_input.setVisible(False)
                return

    def clear_inputs(self):
        if not self.dm.students:
            return
        
        self.is_loading = True
        for q, sb in self.spinboxes.items():
            if sb.isEnabled() and not self.lock_states.get(q, False):
                sb.setValue(0.0)
        if not self.is_desc_locked:
            self.comments_edit.clear()
        self.is_loading = False
        self.on_input_changed()

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
        self.populate_student_dropdown()

    def update_status_label(self):
        if not self.dm.students:
            return
            
        student_id = self.dm.students[self.current_idx]['id']
        is_sub = self.dm.is_submitted(student_id)
        
        current_grades = {q: sb.value() for q, sb in self.spinboxes.items()}
        current_desc = self.comments_edit.toPlainText()
        saved_grades, saved_desc = self.dm.get_saved_data(student_id)
        
        changed = (current_grades != saved_grades) or (current_desc != saved_desc)
        
        if changed:
            status = "Unsaved Changes"
            color = "#dbab09" if self.is_dark_mode else "#b08800"
        elif is_sub:
            status = "Submitted"
            color = "#3fb950" if self.is_dark_mode else "#28a745"
        else:
            status = "Not Submitted Yet"
            color = "#ff7b72" if self.is_dark_mode else "#d73a49"
            
        self.status_label.setText(f"Status: {status}")
        self.status_label.setStyleSheet(f"font-size: 14px; font-weight: bold; color: {color};")
        
        total_score = sum(current_grades.values())
        self.total_score_label.setText(f"Total: {total_score:.2f}")

    def toggle_pdf_toolbar(self):
        self.pdf_viewer.is_toolbar_expanded = not self.pdf_viewer.is_toolbar_expanded
        self.pdf_viewer.toolbar_widget.setVisible(self.pdf_viewer.is_toolbar_expanded)
        self.btn_toggle_pdf_toolbar.setText("⌃" if self.pdf_viewer.is_toolbar_expanded else "⌄")

    def toggle_theme(self):
        self.is_dark_mode = not self.is_dark_mode
        self.btn_theme.setText("☀️" if self.is_dark_mode else "🌙")
        self.apply_theme()
        self.update_status_label()
        self.update_stats_sidebar()

    def refresh_pdf_size(self):
        if self.pdf_viewer.current_fit_mode == 'width':
            self.pdf_viewer.fit_width()
        elif self.pdf_viewer.current_fit_mode == 'screen':
            self.pdf_viewer.fit_screen()

    def apply_theme(self):
        font_css = "font-family: 'Inter', 'Vazirmatn';"
        if self.is_dark_mode:
            self.setStyleSheet(f"""
                QWidget {{ {font_css} background-color: #0d1117; color: #e6edf3; }}
                QScrollArea {{ border: 1px solid #30363d; border-radius: 8px; background-color: #161b22; }}
                QSplitter::handle {{ background-color: #30363d; margin: 2px; }}
                QLineEdit {{ background-color: #161b22; border: 1px solid #30363d; border-radius: 6px; padding: 5px; color: #e6edf3; font-size: 13px; }}
                QLineEdit:focus {{ border: 2px solid #58a6ff; }}
                QDoubleSpinBox, QTextEdit, QComboBox, QSpinBox {{ background-color: #161b22; border: 1px solid #30363d; border-radius: 6px; padding: 5px; color: #e6edf3; }}
                QDoubleSpinBox:focus, QTextEdit:focus, QComboBox:focus, QSpinBox:focus {{ border: 2px solid #58a6ff; }}
                QPushButton {{ background-color: #21262d; color: #c9d1d9; border: 1px solid #30363d; border-radius: 6px; font-weight: bold; font-size: 13px; }}
                QPushButton:hover {{ background-color: #30363d; }}
                QPushButton#btn_submit {{ background-color: #1f6feb; color: white; border: none; }}
                QPushButton#btn_submit:hover {{ background-color: #388bfd; }}
                QPushButton#btn_clear {{ background-color: #21262d; color: #c9d1d9; border: 1px solid #30363d; }}
                QPushButton#btn_clear:hover {{ background-color: #da3633; color: white; border: none; }}
                QPushButton:disabled {{ background-color: #161b22; color: #484f58; border: 1px solid #21262d; }}
                QLabel#info_label {{ font-size: 15px; font-weight: bold; color: #58a6ff; padding: 12px; background-color: #161b22; border-radius: 8px; border: 1px solid #30363d; }}
                QLabel#total_score_label {{ font-size: 14px; font-weight: bold; color: #58a6ff; }}
                QLabel[class="no-file-lbl"] {{ font-size: 18px; font-weight: bold; color: #8b949e; }}
                QPushButton[class="toolbar-btn"] {{ background-color: #1f6feb; color: white; border: none; padding: 5px 15px; }}
                QPushButton[class="toolbar-btn"]:hover {{ background-color: #388bfd; }}
                QPushButton[class="toolbar-btn"]:disabled {{ background-color: #21262d; color: #484f58; border: 1px solid #30363d; }}
            """)
        else:
            self.setStyleSheet(f"""
                QWidget {{ {font_css} background-color: #f4f5f7; color: #1f2328; }}
                QScrollArea {{ border: 1px solid #d0d7de; border-radius: 8px; background-color: #ffffff; }}
                QSplitter::handle {{ background-color: #d0d7de; margin: 2px; }}
                QLineEdit {{ background-color: #ffffff; border: 1px solid #d0d7de; border-radius: 6px; padding: 5px; color: #1f2328; font-size: 13px; }}
                QLineEdit:focus {{ border: 2px solid #0969da; }}
                QDoubleSpinBox, QTextEdit, QComboBox, QSpinBox {{ background-color: #ffffff; border: 1px solid #d0d7de; border-radius: 6px; padding: 5px; }}
                QDoubleSpinBox:focus, QTextEdit:focus, QComboBox:focus, QSpinBox:focus {{ border: 2px solid #0969da; }}
                QPushButton {{ background-color: #f6f8fa; color: #24292f; border: 1px solid #d0d7de; border-radius: 6px; font-weight: bold; font-size: 13px; }}
                QPushButton:hover {{ background-color: #f3f4f6; }}
                QPushButton#btn_submit {{ background-color: #0969da; color: white; border: none; }}
                QPushButton#btn_submit:hover {{ background-color: #0353a4; }}
                QPushButton#btn_clear {{ background-color: #f6f8fa; color: #24292f; border: 1px solid #d0d7de; }}
                QPushButton#btn_clear:hover {{ background-color: #cf222e; color: white; border: none; }}
                QPushButton:disabled {{ background-color: #eaeef2; color: #8c959f; border: 1px solid #d0d7de; }}
                QLabel#info_label {{ font-size: 15px; font-weight: bold; color: #0969da; padding: 12px; background-color: #ffffff; border-radius: 8px; border: 1px solid #d0d7de; }}
                QLabel#total_score_label {{ font-size: 14px; font-weight: bold; color: #0969da; }}
                QLabel[class="no-file-lbl"] {{ font-size: 18px; font-weight: bold; color: #57606a; }}
                QPushButton[class="toolbar-btn"] {{ background-color: #0969da; color: white; border: none; padding: 5px 15px; }}
                QPushButton[class="toolbar-btn"]:hover {{ background-color: #0353a4; }}
                QPushButton[class="toolbar-btn"]:disabled {{ background-color: #eaeef2; color: #8c959f; border: 1px solid #d0d7de; }}
            """)
        self.info_label.setObjectName("info_label")
        self.btn_submit.setObjectName("btn_submit")
        self.btn_clear.setObjectName("btn_clear")
        self.total_score_label.setObjectName("total_score_label")

    def load_student(self):
        if not self.dm.students:
            self.info_label.setText("No valid student files found.")
            self.btn_submit.setEnabled(False)
            self.btn_prev.setEnabled(False)
            self.btn_next.setEnabled(False)
            self.btn_clear.setEnabled(False)
            return

        self.is_loading = True
        student = self.dm.students[self.current_idx]
        total_students = len(self.dm.students)
        
        self.student_combo.blockSignals(True)
        self.student_combo.setCurrentIndex(self.current_idx)
        self.student_combo.blockSignals(False)
        
        self.info_label.setText(f"Student {self.current_idx + 1}/{total_students}\n{student['name']} {student['surname']}\nID: {student['id']}")

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
                sb.setReadOnly(self.lock_states.get(q, False))
                
        self.comments_edit.setPlainText(comments)
        self.comments_edit.setReadOnly(self.is_desc_locked)
        self.pdf_viewer.load_pdf(student['pdf'])

        self.btn_prev.setEnabled(self.current_idx > 0)
        self.btn_next.setEnabled(self.current_idx < len(self.dm.students) - 1)
        self.btn_clear.setEnabled(student['pdf'] is not None)
        
        self.is_loading = False
        self.update_status_label()

    def request_change_file(self):
        student = self.dm.students[self.current_idx]
        file_path, _ = QFileDialog.getOpenFileName(self, "Select PDF", "", "PDF Files (*.pdf)")
        if file_path:
            if not os.path.exists(self.dm.pdf_dir):
                os.makedirs(self.dm.pdf_dir)
                
            moodle_id = "0000000"
            for filename in os.listdir(self.dm.pdf_dir):
                if filename.endswith(".pdf"):
                    parts = filename.split('_')
                    if len(parts) > 2 and parts[1].isdigit():
                        moodle_id = parts[1]
                        break

            new_name = f"{student['name']} {student['surname']}_{moodle_id}_assignsubmission_file_{student['id']}.pdf"
            dest_path = os.path.join(self.dm.pdf_dir, new_name)
            
            shutil.copy(file_path, dest_path)
            student['pdf'] = dest_path
            
            self.populate_student_dropdown()
            self.load_student()

    def submit_grade(self):
        student = self.dm.students[self.current_idx]
        grades = {q: sb.value() for q, sb in self.spinboxes.items()}
        total = sum(grades.values())
        comments = self.comments_edit.toPlainText()

        try:
            self.dm.save_grade(student['id'], grades, total, comments)
            self.session_edits.pop(student['id'], None)
            self.populate_student_dropdown()
            self.update_status_label()
            self.update_stats_sidebar()
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