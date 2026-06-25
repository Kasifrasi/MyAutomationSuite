// Ordner-Generator: Einstellungen & Hilfsfunktionen

#[derive(serde::Serialize, serde::Deserialize)]
struct FolderSettings {
    target_folder: String,
    template_file: String,
    subfolders: Vec<String>,
}

impl Default for FolderSettings {
    fn default() -> Self {
        Self {
            target_folder: String::new(),
            template_file: String::new(),
            subfolders: folder_generator::SUBFOLDERS
                .iter()
                .map(|s| s.to_string())
                .collect(),
        }
    }
}

fn apply_folder_defaults(ui: &MainWindow) {
    let fs = ui.global::<FolderState>();
    fs.set_target_folder("".into());
    fs.set_template_file("".into());
    fs.set_project_name("".into());
    fs.set_skip_validation(false);
    fs.set_folder_exists(false);
    fs.set_project_name_valid(false);
    fs.set_csv_file("".into());
    fs.set_csv_target_folder("".into());
    fs.set_status_type("idle".into());
    fs.set_status_message("".into());
    let defaults: Vec<slint::SharedString> = folder_generator::SUBFOLDERS
        .iter()
        .map(|s| (*s).into())
        .collect();
    fs.set_subfolders(std::rc::Rc::new(slint::VecModel::from(defaults)).into());
}

fn load_folder_settings(ui: &MainWindow) {
    let s: FolderSettings = confy::load(APP_NAME, "folder").unwrap_or_default();
    let fs = ui.global::<FolderState>();
    if !s.target_folder.is_empty() {
        fs.set_target_folder(s.target_folder.into());
    }
    if !s.template_file.is_empty() {
        fs.set_template_file(s.template_file.into());
    }
    let model: Vec<slint::SharedString> = s.subfolders.iter().map(|s| s.into()).collect();
    fs.set_subfolders(std::rc::Rc::new(slint::VecModel::from(model)).into());
}

fn save_folder_settings(ui: &MainWindow) {
    let fs = ui.global::<FolderState>();
    let subfolders: Vec<String> = fs.get_subfolders().iter().map(|s| s.to_string()).collect();
    let s = FolderSettings {
        target_folder: fs.get_target_folder().to_string(),
        template_file: fs.get_template_file().to_string(),
        subfolders,
    };
    let _ = confy::store(APP_NAME, "folder", &s);
}