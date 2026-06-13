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

        self.toolbar = QHBoxLayout()
        self.btn_zoom_in = QPushButton("زوم +")
        self.btn_zoom_out = QPushButton("زوم -")
        self.btn_reset = QPushButton("اندازه اصلی")

        self.btn_zoom_in.clicked.connect(self.zoom_in)
        self.btn_zoom_out.clicked.connect(self.zoom_out)
        self.btn_reset.clicked.connect(self.reset_zoom)

        self.toolbar.addWidget(self.btn_zoom_in)
        self.toolbar.addWidget(self.btn_zoom_out)
        self.toolbar.addWidget(self.btn_reset)
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

    def clear(self):
        self.current_pdf_path = None
        self.pdf_label.clear()