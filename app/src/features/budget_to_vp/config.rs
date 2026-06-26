use crate::shared::models::SheetProtectionOptions;
use crate::{MainWindow, VorpruefungState, APP_NAME};

use slint::ComponentHandle;

#[derive(serde::Serialize, serde::Deserialize)]
struct VpSettings {
    src_folder: String,
    out_folder: String,
    name: String,
    protect_sheet: bool,
    protect_workbook: bool,
    sheet_password: String,
    workbook_password: String,
    protection: SheetProtectionOptions,
}

impl Default for VpSettings {
    fn default() -> Self {
        Self {
            src_folder: String::new(),
            out_folder: String::new(),
            name: "Pruefvorlage_{pn}_{la}.xlsx".to_string(),
            protect_sheet: true,
            protect_workbook: true,
            sheet_password: String::new(),
            workbook_password: String::new(),
            protection: SheetProtectionOptions {
                select_locked_cells: true,
                select_unlocked_cells: true,
                format_cells: true,
                format_columns: true,
                format_rows: true,
                insert_columns: false,
                insert_rows: false,
                insert_hyperlinks: true,
                delete_columns: true,
                delete_rows: true,
                sort: true,
                auto_filter: true,
                pivot_tables: true,
                edit_objects: false,
                edit_scenarios: true,
            },
        }
    }
}

pub fn apply_vp_defaults(ui: &MainWindow) {
    let vp = ui.global::<VorpruefungState>();
    vp.set_src_folder("".into());
    vp.set_out_folder("".into());
    vp.set_name("Pruefvorlage_{pn}_{la}.xlsx".into());
    vp.set_protect_sheet(true);
    vp.set_protect_workbook(true);
    vp.set_sheet_password("".into());
    vp.set_workbook_password("".into());
    vp.set_show_settings(true);
    vp.set_status_type("idle".into());
    vp.set_status_message("".into());
    vp.set_table_data(slint::ModelRc::default());
    vp.set_table_columns(slint::ModelRc::default());

    vp.set_sheet_permissions(
        SheetProtectionOptions {
            select_locked_cells: true,
            select_unlocked_cells: true,
            format_cells: true,
            format_columns: true,
            format_rows: true,
            insert_columns: false,
            insert_rows: false,
            insert_hyperlinks: true,
            delete_columns: true,
            delete_rows: true,
            sort: true,
            auto_filter: true,
            pivot_tables: true,
            edit_objects: false,
            edit_scenarios: true,
        }
        .into(),
    );
}

pub fn load_vp_settings(ui: &MainWindow) {
    let s: VpSettings = confy::load(APP_NAME, "vorpruefung").unwrap_or_default();
    let vp = ui.global::<VorpruefungState>();
    vp.set_name(s.name.into());
    vp.set_protect_sheet(s.protect_sheet);
    vp.set_protect_workbook(s.protect_workbook);
    vp.set_sheet_password(s.sheet_password.into());
    vp.set_workbook_password(s.workbook_password.into());
    if !s.src_folder.is_empty() {
        vp.set_src_folder(s.src_folder.into());
    }
    if !s.out_folder.is_empty() {
        vp.set_out_folder(s.out_folder.into());
    }
    vp.set_sheet_permissions(s.protection.into());
}

pub fn save_vp_settings(ui: &MainWindow) {
    let vp = ui.global::<VorpruefungState>();
    let s = VpSettings {
        src_folder: vp.get_src_folder().to_string(),
        out_folder: vp.get_out_folder().to_string(),
        name: vp.get_name().to_string(),
        protect_sheet: vp.get_protect_sheet(),
        protect_workbook: vp.get_protect_workbook(),
        sheet_password: vp.get_sheet_password().to_string(),
        workbook_password: vp.get_workbook_password().to_string(),
        protection: vp.get_sheet_permissions().into(),
    };
    let _ = confy::store(APP_NAME, "vorpruefung", &s);
}
