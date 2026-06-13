import sys
import os
from PyQt6.QtWidgets import QApplication
from PyQt6.QtGui import QFontDatabase, QFont
from core.data_manager import DataManager
from core.gui import GradingApp

def main():
    app = QApplication(sys.argv)

    font_path = os.path.join("fonts", "Vazirmatn.ttf")
    font_id = QFontDatabase.addApplicationFont(font_path)
    if font_id != -1:
        font_family = QFontDatabase.applicationFontFamilies(font_id)[0]
        app.setFont(QFont(font_family, 10))

    dm = DataManager()
    window = GradingApp(dm)
    window.show()
    sys.exit(app.exec())

if __name__ == "__main__":
    main()