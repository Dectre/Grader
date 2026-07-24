import streamlit as st
import streamlit.components.v1 as components
import base64
import os
import pandas as pd
from core.data_manager import DataManager

st.set_page_config(
    page_title="Assignment Grading System",
    page_icon="📝",
    layout="wide",
    initial_sidebar_state="expanded"
)

if 'dm' not in st.session_state:
    st.session_state.dm = DataManager()
if 'dark_mode' not in st.session_state:
    st.session_state.dark_mode = True
if 'show_stats' not in st.session_state:
    st.session_state.show_stats = False
if 'show_pdf' not in st.session_state:
    st.session_state.show_pdf = True

dm = st.session_state.dm

theme_bg = "#0d1117" if st.session_state.dark_mode else "#f4f5f7"
theme_card = "#161b22" if st.session_state.dark_mode else "#ffffff"
theme_text = "#e6edf3" if st.session_state.dark_mode else "#1f2328"
theme_border = "#30363d" if st.session_state.dark_mode else "#d0d7de"

st.markdown(f"""
    <style>
    .block-container {{
        padding-top: 1.5rem !important;
        padding-bottom: 1rem !important;
        max-width: 98% !important;
    }}
    header {{
        visibility: hidden;
    }}
    .stApp {{
        background-color: {theme_bg};
        color: {theme_text};
    }}
    .student-card {{
        background-color: {theme_card};
        border: 1px solid {theme_border};
        border-radius: 8px;
        padding: 12px;
        margin-bottom: 12px;
    }}
    iframe, object {{
        border: 1px solid {theme_border};
        border-radius: 8px;
        background-color: #ffffff;
    }}
    </style>
""", unsafe_allow_html=True)

def show_pdf(file_path):
    if file_path and os.path.exists(file_path):
        with open(file_path, "rb") as f:
            base64_pdf = base64.b64encode(f.read()).decode('utf-8')
        
        pdf_display = f'<object data="data:application/pdf;base64,{base64_pdf}" type="application/pdf" width="100%" height="800px"><p>Your browser does not support inline PDFs.</p></object>'
        st.markdown(pdf_display, unsafe_allow_html=True)
    else:
        st.warning("📄 No PDF file uploaded for this student.")

def open_pdf_new_tab(base64_string, dark_mode):
    bg_color = "#0d1117" if dark_mode else "#f4f5f7"
    btn_bg = "transparent" if dark_mode else "#f6f8fa"
    btn_color = "#e6edf3" if dark_mode else "#24292f"
    btn_border = "#30363d" if dark_mode else "#d0d7de"
    btn_hover = "#30363d" if dark_mode else "#f3f4f6"
    
    html_code = f"""
    <!DOCTYPE html>
    <html>
    <head>
    <style>
        body {{
            margin: 0;
            padding: 0;
            background-color: {bg_color};
            overflow: hidden;
        }}
        .btn {{
            display: flex;
            justify-content: center;
            align-items: center;
            width: 100%;
            height: 35px;
            background-color: {btn_bg};
            color: {btn_color};
            font-family: sans-serif;
            font-weight: 600;
            border-radius: 6px;
            text-decoration: none;
            font-size: 14px;
            border: 1px solid {btn_border};
            box-sizing: border-box;
            cursor: pointer;
            transition: 0.2s;
        }}
        .btn:hover {{
            background-color: {btn_hover};
        }}
    </style>
    </head>
    <body>
        <a id="pdf-link" class="btn" target="_blank">↗️ Open in New Tab</a>
        <script>
            const b64 = "{base64_string}";
            const byteCharacters = atob(b64);
            const byteNumbers = new Array(byteCharacters.length);
            for (let i = 0; i < byteCharacters.length; i++) {{
                byteNumbers[i] = byteCharacters.charCodeAt(i);
            }}
            const byteArray = new Uint8Array(byteNumbers);
            const blob = new Blob([byteArray], {{type: 'application/pdf'}});
            const url = URL.createObjectURL(blob);
            document.getElementById('pdf-link').href = url;
        </script>
    </body>
    </html>
    """
    components.html(html_code, height=40)

top_col1, top_col2, top_col3, top_col4 = st.columns([2, 1, 1, 1])

with top_col1:
    search_query = st.text_input("🔍 Search Name/ID", placeholder="Search...", label_visibility="collapsed")

with top_col2:
    st.session_state.show_pdf = st.toggle("👁️ Show PDF", value=st.session_state.show_pdf)

with top_col3:
    if st.button("📊 Stats", use_container_width=True):
        st.session_state.show_stats = not st.session_state.show_stats

with top_col4:
    theme_icon = "☀️ Light" if st.session_state.dark_mode else "🌙 Dark"
    if st.button(theme_icon, use_container_width=True):
        st.session_state.dark_mode = not st.session_state.dark_mode
        st.rerun()

st.divider()

student_options = []
for idx, s in enumerate(dm.students):
    is_sub = dm.is_submitted(s['id'])
    saved_grades, saved_desc, saved_not_sub = dm.get_saved_data(s['id'])
    
    if saved_not_sub:
        status = "⛔ Absent"
    elif is_sub:
        status = "✅ Submitted"
    else:
        status = "⏳ Not Submitted"
        
    has_pdf = "📄" if s['pdf'] else "❌"
    label = f"{has_pdf} | {status} | {s['name']} {s['surname']} ({s['id']})"
    student_options.append((label, idx, s))

if search_query:
    filtered_options = [opt for opt in student_options if search_query.lower() in opt[2]['name'].lower() or search_query.lower() in opt[2]['surname'].lower() or search_query.lower() in str(opt[2]['id']).lower()]
    if filtered_options:
        student_options = filtered_options

st.sidebar.title("👨‍🎓 Grading System")

selected_label = st.sidebar.selectbox(
    "Select Student:",
    options=[opt[0] for opt in student_options],
    index=st.session_state.get('current_student_idx', 0) if st.session_state.get('current_student_idx', 0) < len(student_options) else 0
)

current_idx = [opt[1] for opt in student_options if opt[0] == selected_label][0]
st.session_state.current_student_idx = current_idx
student = dm.students[current_idx]

st.sidebar.divider()

col_prev, col_next = st.sidebar.columns(2)
if col_prev.button("⬅️ Previous", use_container_width=True, disabled=(current_idx == 0)):
    st.session_state.current_student_idx -= 1
    st.rerun()

if col_next.button("Next ➡️", use_container_width=True, disabled=(current_idx == len(dm.students) - 1)):
    st.session_state.current_student_idx += 1
    st.rerun()

st.sidebar.divider()

if os.path.exists(dm.excel_path):
    with open(dm.excel_path, "rb") as f:
        excel_bytes = f.read()
    st.sidebar.download_button(
        label="📥 Download Grades Excel",
        data=excel_bytes,
        file_name="grades.xlsx",
        mime="application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
        use_container_width=True
    )

if st.session_state.show_stats:
    with st.expander("📈 Live Statistics Table", expanded=True):
        df_sub = dm.grades_df[(dm.grades_df['_Submitted'] == True) & (dm.grades_df['Not Submitted'] != True)]
        
        stats_data = []
        cols_to_check = dm.questions + ['Total Score']
        
        for col in cols_to_check:
            if col in df_sub.columns:
                scores = pd.to_numeric(df_sub[col], errors='coerce').dropna()
                stats_data.append({
                    "Section": col,
                    "Avg": round(float(scores.mean()), 2) if not scores.empty else 0.0,
                    "Median": round(float(scores.median()), 2) if not scores.empty else 0.0,
                    "Max": round(float(scores.max()), 2) if not scores.empty else 0.0,
                    "Min": round(float(scores.min()), 2) if not scores.empty else 0.0
                })
        
        if stats_data:
            st.dataframe(pd.DataFrame(stats_data), use_container_width=True, hide_index=True)

if st.session_state.show_pdf:
    col_pdf, col_grading = st.columns([1.3, 1])
else:
    col_pdf = None
    col_grading = st.container()

saved_grades, saved_desc, saved_not_sub = dm.get_saved_data(student['id'])
is_sub = dm.is_submitted(student['id'])

if col_pdf:
    with col_pdf:
        pdf_top_left, pdf_top_right = st.columns([2, 1.2])
        with pdf_top_left:
            st.markdown(f"### 📄 {student['name']} {student['surname']}")
        with pdf_top_right:
            if student['pdf'] and os.path.exists(student['pdf']):
                with open(student['pdf'], "rb") as pdf_file:
                    b64_pdf = base64.b64encode(pdf_file.read()).decode('utf-8')
                open_pdf_new_tab(b64_pdf, st.session_state.dark_mode)
                
        show_pdf(student['pdf'])

with col_grading:
    grading_panel = st.container(height=850)
    with grading_panel:
        total_students = len(dm.students)
        
        st.markdown(f"""
            <div class="student-card">
                <h4 style="margin:0; color:#58a6ff;">Student {current_idx + 1}/{total_students}</h4>
                <h3 style="margin:4px 0;">{student['name']} {student['surname']}</h3>
                <p style="margin:0; color:#8b949e;">ID: <b>{student['id']}</b></p>
            </div>
        """, unsafe_allow_html=True)
        
        if saved_not_sub:
            status_label = "Did Not Submit"
            status_color = "#ff7b72"
        elif is_sub:
            status_label = "Submitted"
            status_color = "#3fb950"
        else:
            status_label = "Not Submitted Yet"
            status_color = "#ff7b72"
            
        st.markdown(f"**Status:** <span style='color:{status_color}; font-weight:bold;'>{status_label}</span>", unsafe_allow_html=True)

        with st.form(key=f"grading_form_{student['id']}"):
            not_submitted = st.checkbox("Did Not Submit", value=saved_not_sub)
            
            input_grades = {}
            st.write("---")
            
            for q_idx, q_name in enumerate(dm.questions):
                max_g = float(dm.max_grades[q_idx])
                default_val = float(saved_grades.get(q_name, 0.0))
                
                input_grades[q_name] = st.number_input(
                    label=f"{q_name} (out of {max_g}):",
                    min_value=0.0,
                    max_value=max_g,
                    value=default_val,
                    step=0.25,
                    disabled=not_submitted
                )

            comments = st.text_area("Description:", value=saved_desc, height=120)
            
            if not not_submitted:
                total_calc = round(sum(float(v) for v in input_grades.values()), 2)
            else:
                total_calc = 0.0
                
            st.markdown(f"### Total Score: <span style='color:#58a6ff;'>{total_calc:.2f}</span>", unsafe_allow_html=True)
            
            col_btn_clear, col_btn_sub = st.columns([1, 2])
            with col_btn_sub:
                btn_submit = st.form_submit_button("💾 Submit Grade", use_container_width=True)

        if btn_submit:
            dm.save_grade(
                student_id=student['id'],
                grades_dict=input_grades,
                total=total_calc,
                comments=comments,
                not_submitted=not_submitted
            )
            st.success("✅ Successfully submitted!")
            st.rerun()