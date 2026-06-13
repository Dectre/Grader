import fitz
from PyQt6.QtWidgets import QScrollArea, QLabel, QWidget, QVBoxLayout, QHBoxLayout, QPushButton
from PyQt6.QtGui import QImage, QPixmap, QPainter
from PyQt6.QtCore import Qt

class PDFViewer(QWidget):
    def __init__(self):
        super().__init__()
        self.current_pdf_path = None
        self.scale_factor = 1.0

        self.layout = QVBoxLayout(self)
        self.layout.setContentsMargins(0, 0, 0, 0)
        self.layout.setSpacing(10)

        self.toolbar = QHBoxLayout()
        self.btn_zoom_in = QPushButton("Zoom In")
        self.btn_zoom_out = QPushButton("Zoom Out")
        self.btn_reset = QPushButton("Original Size")
        self.btn_fit_screen = QPushButton("Fit to Screen")

        for btn in [self.btn_zoom_in, self.btn_zoom_out, self.btn_reset, self.btn_fit_screen]:
            btn.setFixedHeight(35)
            btn.setCursor(Qt.CursorShape.PointingHandCursor)
            self.toolbar.addWidget(btn)

        self.toolbar.addStretch()
        self.layout.addLayout(self.toolbar)

        self.scroll_area = QScrollArea()
        self.pdf_label = QLabel()
        self.pdf_label.setAlignment(Qt.AlignmentFlag.AlignCenter)
        self.scroll_area.setWidget(self.pdf_label)
        self.scroll_area.setWidgetResizable(True)

        self.layout.addWidget(self.scroll_area)

    def load_pdf(self, path):
        self.current_pdf_path = path
        self.scale_factor = 1.0
        self.render_pdf()

    def render_pdf(self):
        if not self.current_pdf_path:
            return

        doc = fitz.open(self.current_pdf_path)
        images = []
        mat = fitz.Matrix(2 * self.scale_factor, 2 * self.scale_factor)

        for page in doc:
            pix = page.get_pixmap(matrix=mat)
            img = QImage(pix.samples, pix.width, pix.height, pix.stride, QImage.Format.Format_RGB888)
            images.append(QPixmap.fromImage(img))

        if not images:
            return

        total_height = sum(img.height() for img in images)
        max_width = max(img.width() for img in images)
        combined = QPixmap(max_width, total_height)
        combined.fill(Qt.GlobalColor.white)

        painter = QPainter(combined)
        y_offset = 0
        for img in images:
            painter.drawPixmap(0, y_offset, img)
            y_offset += img.height()
        painter.end()

        self.pdf_label.setPixmap(combined)

    def zoom_in(self):
        self.scale_factor *= 1.2
        self.render_pdf()

    def zoom_out(self):
        self.scale_factor /= 1.2
        self.render_pdf()

    def reset_zoom(self):
        self.scale_factor = 1.0
        self.render_pdf()

    def fit_to_screen(self):
        if not self.current_pdf_path:
            return
        doc = fitz.open(self.current_pdf_path)
        if len(doc) > 0:
            base_width = doc[0].rect.width
            view_width = self.scroll_area.viewport().width()
            if base_width > 0:
                self.scale_factor = (view_width - 25) / (base_width * 2)
            self.render_pdf()

    def clear(self):
        self.current_pdf_path = None
        self.pdf_label.clear()