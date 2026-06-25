
fn apply_b2f_defaults(ui: &MainWindow) {
    let b2f = ui.global::<BudgetState>();
    b2f.set_src_folder("".into());
    b2f.set_out_folder("".into());
    b2f.set_status_type("idle".into());
    b2f.set_status_message("".into());

    b2f.set_name("{pn}_{la}_{version}_FB.xlsx".into());
    b2f.set_protect_sheet(true);
    b2f.set_protect_workbook(true);
    b2f.set_sheet_password("".into());
    b2f.set_workbook_password("".into());
    b2f.set_hide_columns(true);
    b2f.set_hide_lang_sheet(true);
    b2f.set_show_settings(true);

    b2f.set_sheet_permissions(SheetPermissions {
        select_locked: true,
        select_unlocked: true,
        format_cells: true,
        format_columns: true,
        format_rows: true,
        insert_columns: false,
        insert_rows: false,
        insert_hyperlinks: true,
        delete_columns: true,
        delete_rows: true,
        sort: true,
        autofilter: true,
        pivot_tables: true,
        edit_objects: false,
        edit_scenarios: true,
        contents: false,
    });
}

#[allow(clippy::too_many_arguments)]
fb.set_sheet_permissions(s.protection.into());

fn load_b2f_settings(ui: &MainWindow) {
    let s: B2fSettings = confy::load(APP_NAME, "b2f").unwrap_or_default();
    let b2f = ui.global::<BudgetState>();
    b2f.set_name(s.name.into());
    b2f.set_protect_sheet(s.protect_sheet);
    b2f.set_protect_workbook(s.protect_workbook);
    b2f.set_sheet_password(s.sheet_password.into());
    b2f.set_workbook_password(s.workbook_password.into());
    b2f.set_hide_columns(s.hide_columns);
    b2f.set_empty_rows(s.empty_rows);
    if !s.src_folder.is_empty() {
        b2f.set_src_folder(s.src_folder.into());
    }
    if !s.out_folder.is_empty() {
        b2f.set_out_folder(s.out_folder.into());
    }
    b2f.set_sheet_permissions(s.protection.into());
}

fn save_b2f_settings(ui: &MainWindow) {
    let b2f = ui.global::<BudgetState>();
    let sp = b2f.get_sheet_permissions();
    let s = B2fSettings {
        src_folder: b2f.get_src_folder().to_string(),
        out_folder: b2f.get_out_folder().to_string(),
        name: b2f.get_name().to_string(),
        protect_sheet: b2f.get_protect_sheet(),
        protect_workbook: b2f.get_protect_workbook(),
        sheet_password: b2f.get_sheet_password().to_string(),
        workbook_password: b2f.get_workbook_password().to_string(),
        hide_columns: b2f.get_hide_columns(),
        empty_rows: b2f.get_empty_rows(),

        protection: b2f.get_sheet_permissions().into(), 
    };
    let _ = confy::store(APP_NAME, "b2f", &s);
}