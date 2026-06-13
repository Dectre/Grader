import fitz
from PyQt6.QtWidgets import QScrollArea, QLabel
from PyQt6.QtGui import QImage, QPixmap, QPainter
from PyQt6.QtCore import Qt

class PDFViewer(QScrollArea):
    def __init__(self):
        super().__init__()
        self.pdf_label = QLabel()
        self.pdf_label.setAlignment(Qt.AlignmentFlag.AlignCenter)
        self.setWidget(self.pdf_label)
        self.setWidgetResizable(True)

    def load_pdf(self, path):
        doc = fitz.open(path)
        images = []
        for page in doc:
            pix = page.get_pixmap(matrix=fitz.Matrix(2, 2))
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

    def clear(self):
        self.pdf_label.clear()