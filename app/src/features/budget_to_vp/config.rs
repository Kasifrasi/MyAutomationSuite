use crate::{MainWindow, VorpruefungState, APP_NAME};

use slint::ComponentHandle;

#[derive(serde::Serialize, serde::Deserialize)]
struct VpSettings {
    src_folder: String,
    out_folder: String,
    name: String,
    protect_workbook: bool,
    workbook_password: String,
}

impl Default for VpSettings {
    fn default() -> Self {
        Self {
            src_folder: String::new(),
            out_folder: String::new(),
            name: "Pruefvorlage_{pn}_{la}.xlsx".to_string(),
            protect_workbook: true,
            workbook_password: String::new(),
        }
    }
}

pub fn apply_vp_defaults(ui: &MainWindow) {
    let vp = ui.global::<VorpruefungState>();
    vp.set_src_folder("".into());
    vp.set_out_folder("".into());
    vp.set_name("Pruefvorlage_{pn}_{la}.xlsx".into());
    vp.set_protect_workbook(true);
    vp.set_workbook_password("".into());
    vp.set_show_settings(true);
    vp.set_status_type("idle".into());
    vp.set_status_message("".into());
    vp.set_table_data(slint::ModelRc::default());
    vp.set_table_columns(slint::ModelRc::default());
}

pub fn load_vp_settings(ui: &MainWindow) {
    let s: VpSettings = confy::load(APP_NAME, "vorpruefung").unwrap_or_default();
    let vp = ui.global::<VorpruefungState>();
    vp.set_name(s.name.into());
    vp.set_protect_workbook(s.protect_workbook);
    vp.set_workbook_password(s.workbook_password.into());
    if !s.src_folder.is_empty() {
        vp.set_src_folder(s.src_folder.into());
    }
    if !s.out_folder.is_empty() {
        vp.set_out_folder(s.out_folder.into());
    }
}

pub fn save_vp_settings(ui: &MainWindow) {
    let vp = ui.global::<VorpruefungState>();
    let s = VpSettings {
        src_folder: vp.get_src_folder().to_string(),
        out_folder: vp.get_out_folder().to_string(),
        name: vp.get_name().to_string(),
        protect_workbook: vp.get_protect_workbook(),
        workbook_password: vp.get_workbook_password().to_string(),
    };
    let _ = confy::store(APP_NAME, "vorpruefung", &s);
}
