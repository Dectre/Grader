import os
import shutil
import pandas as pd
from fastapi import FastAPI, Request, UploadFile, File
from fastapi.responses import FileResponse, HTMLResponse, JSONResponse
from fastapi.staticfiles import StaticFiles
import uvicorn
from core.data_manager import DataManager

app = FastAPI()
dm = DataManager()

if os.path.exists("fonts"):
    app.mount("/fonts", StaticFiles(directory="fonts"), name="fonts")

@app.get("/")
def index():
    with open("index.html", "r", encoding="utf-8") as f:
        return HTMLResponse(f.read())

@app.get("/api/init")
def get_init_data():
    students = []
    for idx, s in enumerate(dm.students):
        is_sub = dm.is_submitted(s['id'])
        grades, desc, not_sub = dm.get_saved_data(s['id'])
        students.append({
            "index": idx,
            "id": s['id'],
            "name": s['name'],
            "surname": s['surname'],
            "has_pdf": bool(s['pdf']),
            "is_submitted": is_sub,
            "not_submitted": not_sub
        })
    
    return {
        "questions": dm.questions,
        "max_grades": dm.max_grades,
        "students": students,
        "stats": get_stats()
    }

@app.get("/api/student/{student_id}")
def get_student(student_id: str):
    grades, desc, not_sub = dm.get_saved_data(student_id)
    is_sub = dm.is_submitted(student_id)
    return {
        "grades": grades,
        "description": desc,
        "not_submitted": not_sub,
        "is_submitted": is_sub
    }

@app.post("/api/student/{student_id}/submit")
async def submit_student(student_id: str, request: Request):
    data = await request.json()
    grades = data.get("grades", {})
    total = sum(float(v) for v in grades.values())
    desc = data.get("comments", "")
    not_sub = data.get("not_submitted", False)
    
    dm.save_grade(student_id, grades, total, desc, not_sub)
    return {"status": "success", "stats": get_stats()}

@app.post("/api/student/{student_id}/upload")
async def upload_pdf(student_id: str, file: UploadFile = File(...)):
    student = next((s for s in dm.students if str(s['id']) == student_id), None)
    if not student:
        return JSONResponse(status_code=404, content={"message": "Not Found"})
        
    if not os.path.exists(dm.pdf_dir):
        os.makedirs(dm.pdf_dir)
        
    moodle_id = "0000000"
    for filename in os.listdir(dm.pdf_dir):
        if filename.endswith(".pdf"):
            parts = filename.split('_')
            if len(parts) > 2 and parts[1].isdigit():
                moodle_id = parts[1]
                break

    new_name = f"{student['name']} {student['surname']}_{moodle_id}_assignsubmission_file_{student['id']}.pdf"
    file_path = os.path.join(dm.pdf_dir, new_name)
    
    with open(file_path, "wb") as buffer:
        shutil.copyfileobj(file.file, buffer)
        
    student['pdf'] = file_path
    return {"status": "success"}

@app.get("/api/pdf/{student_id}")
def get_pdf(student_id: str):
    for s in dm.students:
        if str(s['id']) == student_id and s['pdf'] and os.path.exists(s['pdf']):
            return FileResponse(
                s['pdf'], 
                media_type="application/pdf",
                headers={"Content-Disposition": "inline"}
            )
    return JSONResponse(status_code=404, content={"message": "Not Found"})

@app.get("/api/download/grades")
def download_grades():
    return FileResponse(
        dm.excel_path, 
        filename="grades.xlsx",
        media_type="application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
    )

@app.get("/api/download/pdf_view")
def download_pdf_view():
    df = dm.grades_df.copy()
    if '_Submitted' in df.columns:
        df = df.drop(columns=['_Submitted'])
    
    valid_df = df[df['Not Submitted'] != True]
    cols_for_stats = dm.questions + ['Total Score']
    
    stats_html = ""
    labels = ["Mean", "Median", "Max", "Min"]
    for label in labels:
        stats_html += f"<tr class='stat-row'><th colspan='3'>{label}</th>"
        for col in df.columns[3:]:
            if col in cols_for_stats:
                scores = pd.to_numeric(valid_df[col], errors='coerce').dropna()
                if label == "Mean": val = round(scores.mean(), 2) if not scores.empty else 0
                elif label == "Median": val = round(scores.median(), 2) if not scores.empty else 0
                elif label == "Max": val = round(scores.max(), 2) if not scores.empty else 0
                elif label == "Min": val = round(scores.min(), 2) if not scores.empty else 0
                stats_html += f"<td>{val}</td>"
            else:
                stats_html += "<td></td>"
        stats_html += "</tr>"
    
    header_html = "<tr class='header-row'>"
    for col in df.columns:
        header_html += f"<th>{col}</th>"
    header_html += "</tr>"
    
    body_html = ""
    for _, row in df.iterrows():
        body_html += "<tr>"
        for col in df.columns:
            val = row[col]
            if pd.isna(val):
                val = ""
            
            if col == 'Description':
                val = str(val).replace('\\n', '<br>').replace('\n', '<br>')
                body_html += f"<td class='desc-col'>{val}</td>"
            else:
                body_html += f"<td>{val}</td>"
        body_html += "</tr>"
        
    html_content = f"""
    <!DOCTYPE html>
    <html lang="fa" dir="rtl">
    <head>
        <meta charset="UTF-8">
        <meta name="viewport" content="width=device-width, initial-scale=1.0">
        <title>Grades Report</title>
        <style>
            @font-face {{ font-family: 'Vazirmatn'; src: url('/fonts/Vazirmatn.ttf'); }}
            body {{ font-family: 'Vazirmatn', Tahoma, Arial, sans-serif; padding: 20px; background: #fff; color: #000; direction: rtl; }}
            h2 {{ text-align: center; color: #333; margin-bottom: 20px; }}
            .table {{ width: 100%; border-collapse: collapse; font-size: 13px; text-align: center; border: 2px solid #000; }}
            .table th, .table td {{ border: 1px solid #000; padding: 6px; vertical-align: middle; }}
            .table th {{ background-color: #f4f4f4; color: #000; font-weight: bold; }}
            .stat-row th {{ text-align: left; padding-left: 15px; border-bottom: 1px solid #000; }}
            .header-row th {{ border-top: 2px solid #000; border-bottom: 2px solid #000; }}
            .desc-col {{ text-align: right; max-width: 400px; line-height: 1.6; }}
            .table tbody tr:last-child td {{ border-bottom: 2px solid #000; }}
            @media print {{
                body {{ padding: 0; }}
                .no-print {{ display: none; }}
            }}
        </style>
    </head>
    <body>
        <div class="no-print" style="text-align: center; margin-bottom: 20px;">
            <button onclick="window.print()" style="padding: 10px 20px; font-size: 16px; cursor: pointer; background: #1f6feb; color: white; border: none; border-radius: 5px; font-family: 'Vazirmatn';">Save as PDF / Print</button>
        </div>
        <h2>گزارش نمرات دانشجویان</h2>
        <table class="table">
            <thead>
                {stats_html}
                {header_html}
            </thead>
            <tbody>
                {body_html}
            </tbody>
        </table>
    </body>
    </html>
    """
    return HTMLResponse(content=html_content)

def get_stats():
    df = dm.grades_df[(dm.grades_df['_Submitted'] == True) & (dm.grades_df['Not Submitted'] != True)]
    stats_data = {}
    for col in dm.questions + ['Total Score']:
        if col in df.columns:
            scores = df[col].dropna()
            stats_data[col] = {
                "avg": round(scores.mean(), 2) if not scores.empty else 0,
                "med": round(scores.median(), 2) if not scores.empty else 0,
                "max": round(scores.max(), 2) if not scores.empty else 0,
                "min": round(scores.min(), 2) if not scores.empty else 0
            }
    return {"total_students": len(dm.students), "data": stats_data}

if __name__ == "__main__":
    uvicorn.run(app, host="0.0.0.0", port=8000)