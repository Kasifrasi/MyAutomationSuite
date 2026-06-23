import zipfile

with zipfile.ZipFile('/tmp/xlsx_good.xlsx', 'r') as z:
    try:
        print(z.read('xl/worksheets/sheet7.xml').decode('utf-8'))
    except KeyError:
        print("sheet7 not found")
