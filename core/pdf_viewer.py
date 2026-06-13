import fitz
from PyQt6.QtWidgets import (QScrollArea, QLabel, QWidget, QVBoxLayout, 
                             QHBoxLayout, QPushButton, QComboBox, QSpinBox, QSizePolicy)
from PyQt6.QtGui import QImage, QPixmap, QPainter
from PyQt6.QtCore import Qt

class PDFViewer(QWidget):
    def __init__(self):
        super().__init__()
        self.doc = None
        self.scale_factor = 1.0
        self.total_pages = 0
        self.current_page = 0
        self.current_fit_mode = None
        self.is_toolbar_expanded = True

        self.layout = QVBoxLayout(self)
        self.layout.setContentsMargins(0, 0, 0, 0)
        self.layout.setSpacing(5)

        self.toolbar_widget = QWidget()
        self.toolbar_widget.setSizePolicy(QSizePolicy.Policy.Expanding, QSizePolicy.Policy.Maximum)
        toolbar_layout = QHBoxLayout(self.toolbar_widget)
        toolbar_layout.setContentsMargins(0, 0, 0, 0)
        toolbar_layout.setSpacing(10)
        
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
        self.btn_fit_width = QPushButton("↔️ Fit Width")
        self.btn_fit_screen = QPushButton("🖵 Fit Screen")

        self.btn_zoom_in.clicked.connect(self.zoom_in)
        self.btn_zoom_out.clicked.connect(self.zoom_out)
        self.btn_fit_width.clicked.connect(self.fit_width)
        self.btn_fit_screen.clicked.connect(self.fit_screen)

        controls = [self.btn_prev_page, self.btn_next_page, self.btn_zoom_in, 
                    self.btn_zoom_out, self.btn_fit_width, self.btn_fit_screen]
        
        toolbar_layout.addWidget(self.view_mode)
        for widget in controls:
            if widget in [self.btn_zoom_in, self.btn_zoom_out, self.btn_fit_width, self.btn_fit_screen]:
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
        self.scroll_area.setWidget(self.pdf_label)
        self.scroll_area.setWidgetResizable(True)

        self.layout.addWidget(self.scroll_area)

    def resizeEvent(self, event):
        super().resizeEvent(event)
        if self.current_fit_mode == 'width':
            self.fit_width()
        elif self.current_fit_mode == 'screen':
            self.fit_screen()

    def load_pdf(self, path):
        if self.doc:
            self.doc.close()
        self.doc = fitz.open(path)
        self.total_pages = len(self.doc)
        self.current_page = 0
        self.scale_factor = 1.0
        self.current_fit_mode = None
        self.page_spinbox.setMaximum(self.total_pages)
        self.page_label.setText(f"/ {self.total_pages}")
        self.render_pdf()

    def update_ui_states(self):
        is_single = self.view_mode.currentText() == "Single Page"
        self.btn_prev_page.setVisible(is_single)
        self.btn_next_page.setVisible(is_single)
        self.page_spinbox.setVisible(is_single)
        self.page_label.setVisible(is_single)
        
        if is_single:
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

    def render_pdf(self):
        if not self.doc:
            return

        self.update_ui_states()

        images = []
        mat = fitz.Matrix(2 * self.scale_factor, 2 * self.scale_factor)

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
        base_width = page.rect.width
        view_width = self.scroll_area.viewport().width()
        if base_width > 0:
            self.scale_factor = (view_width - 30) / (base_width * 2)
        self.render_pdf()

    def fit_screen(self):
        if not self.doc: return
        self.current_fit_mode = 'screen'
        page = self.doc[self.current_page if self.view_mode.currentText() == "Single Page" else 0]
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