import sys
import os
from PyQt6.QtWidgets import QApplication
from PyQt6.QtGui import QFontDatabase, QFont
from core.data_manager import DataManager
from core.gui import GradingApp

def main():
    app = QApplication(sys.argv)

    fonts_to_load = [
        "Inter-Regular.ttf", "Inter-Bold.ttf", 
        "Inter-Italic.ttf", "Inter-BoldItalic.ttf", 
        "Vazirmatn.ttf", "Vazirmatn-Bold.ttf"
    ]
    
    inter_family = None
    for font_file in fonts_to_load:
        font_path = os.path.join("fonts", font_file)
        if os.path.exists(font_path):
            font_id = QFontDatabase.addApplicationFont(font_path)
            if font_id != -1 and "Inter" in font_file and not inter_family:
                inter_family = QFontDatabase.applicationFontFamilies(font_id)[0]
    
    if inter_family:
        app.setFont(QFont(inter_family, 10))

    dm = DataManager()
    window = GradingApp(dm)
    window.show()
    sys.exit(app.exec())

if __name__ == "__main__":
    main()