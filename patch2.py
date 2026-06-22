import re

with open("app/src/main.rs", "r") as f:
    content = f.read()

# remove fields from structs
content = re.sub(r'\s+pub unhide_all_columns: bool,', '', content)
content = re.sub(r'\s+pub unhide_all_rows: bool,', '', content)

# remove defaults
content = re.sub(r'\s+unhide_all_columns: false,', '', content)
content = re.sub(r'\s+unhide_all_rows: false,', '', content)

# remove setters
content = re.sub(r'\s+b2f.set_unhide_all_columns\(s\.unhide_all_columns\);', '', content)
content = re.sub(r'\s+b2f.set_unhide_all_rows\(s\.unhide_all_rows\);', '', content)
content = re.sub(r'\s+fb.set_unhide_all_columns\(s\.unhide_all_columns\);', '', content)
content = re.sub(r'\s+fb.set_unhide_all_rows\(s\.unhide_all_rows\);', '', content)

# remove getters
content = re.sub(r'\s+unhide_all_columns: b2f.get_unhide_all_columns\(\),', '', content)
content = re.sub(r'\s+unhide_all_rows: b2f.get_unhide_all_rows\(\),', '', content)
content = re.sub(r'\s+unhide_all_columns: fb.get_unhide_all_columns\(\),', '', content)
content = re.sub(r'\s+unhide_all_rows: fb.get_unhide_all_rows\(\),', '', content)

# remove resets
content = re.sub(r'\s+b2f.set_unhide_all_columns\(false\);', '', content)
content = re.sub(r'\s+b2f.set_unhide_all_rows\(false\);', '', content)
content = re.sub(r'\s+fb.set_unhide_all_columns\(false\);', '', content)
content = re.sub(r'\s+fb.set_unhide_all_rows\(false\);', '', content)


with open("app/src/main.rs", "w") as f:
    f.write(content)

