import fitz
from PyQt6.QtWidgets import (QScrollArea, QLabel, QWidget, QVBoxLayout, 
                             QHBoxLayout, QPushButton, QComboBox, QSpinBox, QSizePolicy)
from PyQt6.QtGui import QImage, QPixmap, QPainter
from PyQt6.QtCore import Qt, QEvent, QTimer

class PDFViewer(QWidget):
    def __init__(self):
        super().__init__()
        self.doc = None
        self.scale_factor = 1.0
        self.rotation_angle = 0
        self.total_pages = 0
        self.current_page = 0
        self.current_fit_mode = None
        self.is_toolbar_expanded = True
        self.change_file_callback = None
        self.last_mouse_pos = None

        self.resize_timer = QTimer()
        self.resize_timer.setSingleShot(True)
        self.resize_timer.timeout.connect(self.handle_resize_timeout)

        self.layout = QVBoxLayout(self)
        self.layout.setContentsMargins(0, 0, 0, 0)
        self.layout.setSpacing(5)

        self.toolbar_widget = QWidget()
        self.toolbar_widget.setSizePolicy(QSizePolicy.Policy.Expanding, QSizePolicy.Policy.Maximum)
        toolbar_layout = QHBoxLayout(self.toolbar_widget)
        toolbar_layout.setContentsMargins(0, 0, 0, 0)
        toolbar_layout.setSpacing(6)
        
        toolbar_layout.addStretch()
        
        self.view_mode = QComboBox()
        self.view_mode.addItems(["Continuous", "Single Page"])
        self.view_mode.currentTextChanged.connect(self.render_pdf)
        self.view_mode.setFixedHeight(35)
        
        self.btn_prev_page = QPushButton("◀")
        self.btn_next_page = QPushButton("▶")
        self.page_spinbox = QSpinBox()
        self.page_spinbox.setMinimum(1)
        self.page_label = QLabel("/ 0")

        self.btn_prev_page.clicked.connect(self.prev_page)
        self.btn_next_page.clicked.connect(self.next_page)
        self.page_spinbox.valueChanged.connect(self.go_to_page)

        self.btn_zoom_in = QPushButton("🔍+")
        self.btn_zoom_out = QPushButton("🔍-")
        self.btn_rotate = QPushButton("↻ Rotate")
        self.btn_fit_width = QPushButton("↔️ Width")
        self.btn_fit_screen = QPushButton("🖵 Screen")
        self.btn_change_file = QPushButton("📁 Change")

        self.btn_zoom_in.clicked.connect(self.zoom_in)
        self.btn_zoom_out.clicked.connect(self.zoom_out)
        self.btn_rotate.clicked.connect(self.rotate_pdf)
        self.btn_fit_width.clicked.connect(self.fit_width)
        self.btn_fit_screen.clicked.connect(self.fit_screen)
        self.btn_change_file.clicked.connect(self.trigger_change_file)

        self.controls = [self.btn_prev_page, self.btn_next_page, self.btn_zoom_in, 
                         self.btn_zoom_out, self.btn_rotate, self.btn_fit_width, self.btn_fit_screen, self.btn_change_file]
        
        toolbar_layout.addWidget(self.view_mode)
        for widget in self.controls:
            if widget in [self.btn_zoom_in, self.btn_zoom_out, self.btn_rotate, self.btn_fit_width, self.btn_fit_screen, self.btn_change_file]:
                widget.setFixedHeight(35)
                widget.setCursor(Qt.CursorShape.PointingHandCursor)
                widget.setProperty("class", "toolbar-btn")
            elif widget in [self.btn_prev_page, self.btn_next_page]:
                widget.setFixedHeight(35)
                widget.setCursor(Qt.CursorShape.PointingHandCursor)
                widget.setProperty("class", "toolbar-btn")
                
            if widget == self.btn_next_page:
                toolbar_layout.addWidget(self.btn_prev_page)
                toolbar_layout.addWidget(self.page_spinbox)
                toolbar_layout.addWidget(self.page_label)
                toolbar_layout.addWidget(self.btn_next_page)
            elif widget != self.btn_prev_page:
                toolbar_layout.addWidget(widget)

        toolbar_layout.addStretch()
        self.layout.addWidget(self.toolbar_widget)

        self.scroll_area = QScrollArea()
        self.pdf_label = QLabel()
        self.pdf_label.setAlignment(Qt.AlignmentFlag.AlignTop | Qt.AlignmentFlag.AlignHCenter)
        self.pdf_label.setCursor(Qt.CursorShape.OpenHandCursor)
        self.pdf_label.installEventFilter(self)
        
        self.scroll_area.setWidget(self.pdf_label)
        self.scroll_area.setWidgetResizable(True)

        self.empty_widget = QWidget()
        empty_layout = QVBoxLayout(self.empty_widget)
        self.lbl_no_file = QLabel("No file uploaded for this student.")
        self.lbl_no_file.setProperty("class", "no-file-lbl")
        self.lbl_no_file.setAlignment(Qt.AlignmentFlag.AlignCenter)
        self.btn_browse_center = QPushButton("Browse & Upload PDF")
        self.btn_browse_center.setFixedSize(200, 45)
        self.btn_browse_center.setProperty("class", "toolbar-btn")
        self.btn_browse_center.setCursor(Qt.CursorShape.PointingHandCursor)
        self.btn_browse_center.clicked.connect(self.trigger_change_file)
        
        empty_layout.addStretch()
        empty_layout.addWidget(self.lbl_no_file, alignment=Qt.AlignmentFlag.AlignCenter)
        empty_layout.addWidget(self.btn_browse_center, alignment=Qt.AlignmentFlag.AlignCenter)
        empty_layout.addStretch()

        self.layout.addWidget(self.scroll_area)
        self.layout.addWidget(self.empty_widget)
        self.empty_widget.hide()

    def eventFilter(self, source, event):
        if source == self.pdf_label:
            if event.type() == QEvent.Type.MouseButtonPress:
                if event.button() == Qt.MouseButton.LeftButton:
                    self.last_mouse_pos = event.globalPosition().toPoint()
                    self.pdf_label.setCursor(Qt.CursorShape.ClosedHandCursor)
                    return True
            elif event.type() == QEvent.Type.MouseMove:
                if self.last_mouse_pos is not None:
                    delta = event.globalPosition().toPoint() - self.last_mouse_pos
                    h_bar = self.scroll_area.horizontalScrollBar()
                    v_bar = self.scroll_area.verticalScrollBar()
                    h_bar.setValue(h_bar.value() - delta.x())
                    v_bar.setValue(v_bar.value() - delta.y())
                    self.last_mouse_pos = event.globalPosition().toPoint()
                    return True
            elif event.type() == QEvent.Type.MouseButtonRelease:
                if event.button() == Qt.MouseButton.LeftButton:
                    self.last_mouse_pos = None
                    self.pdf_label.setCursor(Qt.CursorShape.OpenHandCursor)
                    return True
        return super().eventFilter(source, event)

    def trigger_change_file(self):
        if self.change_file_callback:
            self.change_file_callback()

    def resizeEvent(self, event):
        super().resizeEvent(event)
        self.resize_timer.start(150)

    def handle_resize_timeout(self):
        if self.current_fit_mode == 'width':
            self.fit_width()
        elif self.current_fit_mode == 'screen':
            self.fit_screen()

    def load_pdf(self, path):
        if self.doc:
            self.doc.close()
            self.doc = None
            
        if not path:
            self.scroll_area.hide()
            self.empty_widget.show()
            for widget in self.controls:
                if widget != self.btn_change_file:
                    widget.setEnabled(False)
            self.view_mode.setEnabled(False)
            return
            
        self.empty_widget.hide()
        self.scroll_area.show()
        for widget in self.controls:
            widget.setEnabled(True)
        self.view_mode.setEnabled(True)

        self.doc = fitz.open(path)
        self.total_pages = len(self.doc)
        self.current_page = 0
        self.scale_factor = 1.0
        self.rotation_angle = 0
        self.current_fit_mode = 'width'
        self.page_spinbox.setMaximum(self.total_pages)
        self.page_label.setText(f"/ {self.total_pages}")
        self.fit_width()

    def update_ui_states(self):
        is_single = self.view_mode.currentText() == "Single Page"
        self.btn_prev_page.setVisible(is_single)
        self.btn_next_page.setVisible(is_single)
        self.page_spinbox.setVisible(is_single)
        self.page_label.setVisible(is_single)
        
        if is_single and self.doc:
            self.page_spinbox.blockSignals(True)
            self.page_spinbox.setValue(self.current_page + 1)
            self.page_spinbox.blockSignals(False)
            self.btn_prev_page.setEnabled(self.current_page > 0)
            self.btn_next_page.setEnabled(self.current_page < self.total_pages - 1)

    def go_to_page(self, page_num):
        self.current_page = page_num - 1
        self.render_pdf()

    def prev_page(self):
        if self.current_page > 0:
            self.current_page -= 1
            self.render_pdf()

    def next_page(self):
        if self.current_page < self.total_pages - 1:
            self.current_page += 1
            self.render_pdf()

    def rotate_pdf(self):
        self.rotation_angle = (self.rotation_angle + 90) % 360
        if self.current_fit_mode == 'width':
            self.fit_width()
        elif self.current_fit_mode == 'screen':
            self.fit_screen()
        else:
            self.render_pdf()

    def render_pdf(self):
        if not self.doc:
            return

        self.update_ui_states()

        images = []
        mat = fitz.Matrix(2 * self.scale_factor, 2 * self.scale_factor) * fitz.Matrix(self.rotation_angle)

        if self.view_mode.currentText() == "Continuous":
            pages = range(self.total_pages)
        else:
            pages = [self.current_page]

        for p_num in pages:
            page = self.doc[p_num]
            pix = page.get_pixmap(matrix=mat)
            img = QImage(pix.samples, pix.width, pix.height, pix.stride, QImage.Format.Format_RGB888)
            images.append(QPixmap.fromImage(img))

        if not images:
            return

        total_height = sum(img.height() for img in images)
        max_width = max(img.width() for img in images)
        combined = QPixmap(max_width, total_height)
        combined.fill(Qt.GlobalColor.transparent if self.view_mode.currentText() == "Single Page" else Qt.GlobalColor.white)

        painter = QPainter(combined)
        y_offset = 0
        for img in images:
            painter.drawPixmap(0, y_offset, img)
            y_offset += img.height()
        painter.end()

        self.pdf_label.setPixmap(combined)

    def zoom_in(self):
        self.current_fit_mode = None
        self.scale_factor *= 1.2
        self.render_pdf()

    def zoom_out(self):
        self.current_fit_mode = None
        self.scale_factor /= 1.2
        self.render_pdf()

    def fit_width(self):
        if not self.doc: return
        self.current_fit_mode = 'width'
        page = self.doc[self.current_page if self.view_mode.currentText() == "Single Page" else 0]
        
        if self.rotation_angle in [90, 270]:
            base_width = page.rect.height
        else:
            base_width = page.rect.width
            
        view_width = self.scroll_area.viewport().width()
        if base_width > 0:
            self.scale_factor = (view_width - 30) / (base_width * 2)
        self.render_pdf()

    def fit_screen(self):
        if not self.doc: return
        self.current_fit_mode = 'screen'
        page = self.doc[self.current_page if self.view_mode.currentText() == "Single Page" else 0]
        
        if self.rotation_angle in [90, 270]:
            base_width = page.rect.height
            base_height = page.rect.width
        else:
            base_width = page.rect.width
            base_height = page.rect.height
            
        view_width = self.scroll_area.viewport().width()
        view_height = self.scroll_area.viewport().height()
        if base_width > 0 and base_height > 0:
            scale_w = (view_width - 30) / (base_width * 2)
            scale_h = (view_height - 30) / (base_height * 2)
            self.scale_factor = min(scale_w, scale_h)
        self.render_pdf()

    def clear(self):
        if self.doc:
            self.doc.close()
            self.doc = None
        self.pdf_label.clear()