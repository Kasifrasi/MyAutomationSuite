#![windows_subsystem = "windows"]

slint::include_modules!();

mod updater;

use slint::Model;
use std::path::PathBuf;

const APP_NAME: &str = "automation-tool";

// Theme: Einstellungen

#[derive(serde::Serialize, serde::Deserialize)]
struct ThemeSettings {
    dark_mode: bool,
}

impl Default for ThemeSettings {
    fn default() -> Self {
        Self {
            dark_mode: matches!(dark_light::detect(), Ok(dark_light::Mode::Dark)),
        }
    }
}

fn load_theme_settings(ui: &MainWindow) {
    let s: ThemeSettings = confy::load(APP_NAME, "theme").unwrap_or_default();
    ui.set_dark_mode(s.dark_mode);
    let scheme = if s.dark_mode {
        slint::language::ColorScheme::Dark
    } else {
        slint::language::ColorScheme::Light
    };
    ui.global::<Palette>().set_color_scheme(scheme);
}

fn save_theme_settings(dark_mode: bool) {
    let _ = confy::store(APP_NAME, "theme", &ThemeSettings { dark_mode });
}

// Budget-Scanner & FB-Generator: Einstellungen

#[derive(serde::Serialize, serde::Deserialize)]
struct B2fSettings {
    protect_sheet: bool,
    protect_workbook: bool,
    sheet_password: String,
    workbook_password: String,
    hide_columns: bool,
    hide_lang_sheet: bool,
    version: String,
    select_locked: bool,
    select_unlocked: bool,
    format_cells: bool,
    format_columns: bool,
    format_rows: bool,
    insert_columns: bool,
    insert_rows: bool,
    insert_hyperlinks: bool,
    delete_columns: bool,
    delete_rows: bool,
    sort: bool,
    autofilter: bool,
    pivot_tables: bool,
    edit_objects: bool,
    edit_scenarios: bool,
    contents: bool,
}

impl Default for B2fSettings {
    fn default() -> Self {
        Self {
            protect_sheet: true,
            protect_workbook: true,
            sheet_password: String::new(),
            workbook_password: String::new(),
            hide_columns: true,
            hide_lang_sheet: true,
            version: String::new(),
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
        }
    }
}

#[derive(serde::Serialize, serde::Deserialize)]
struct FbSettings {
    langs: [bool; 5], // de, en, fr, es, pt
    categories: [i32; 8],
    protect_sheet: bool,
    protect_workbook: bool,
    sheet_password: String,
    workbook_password: String,
    hide_columns: bool,
    hide_lang_sheet: bool,
    version: String,
    select_locked: bool,
    select_unlocked: bool,
    format_cells: bool,
    format_columns: bool,
    format_rows: bool,
    insert_columns: bool,
    insert_rows: bool,
    insert_hyperlinks: bool,
    delete_columns: bool,
    delete_rows: bool,
    sort: bool,
    autofilter: bool,
    pivot_tables: bool,
    edit_objects: bool,
    edit_scenarios: bool,
    contents: bool,
}

impl Default for FbSettings {
    fn default() -> Self {
        Self {
            langs: [true, false, false, false, false],
            categories: [20, 20, 30, 30, 20, 0, 0, 0],
            protect_sheet: true,
            protect_workbook: true,
            sheet_password: String::new(),
            workbook_password: String::new(),
            hide_columns: true,
            hide_lang_sheet: true,
            version: String::new(),
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
        }
    }
}

// Budget-Scanner & FB-Generator: Laden / Speichern

#[allow(clippy::too_many_arguments)]
fn permissions_from_settings(
    select_locked: bool,
    select_unlocked: bool,
    format_cells: bool,
    format_columns: bool,
    format_rows: bool,
    insert_columns: bool,
    insert_rows: bool,
    insert_hyperlinks: bool,
    delete_columns: bool,
    delete_rows: bool,
    sort: bool,
    autofilter: bool,
    pivot_tables: bool,
    edit_objects: bool,
    edit_scenarios: bool,
    contents: bool,
) -> SheetPermissions {
    SheetPermissions {
        select_locked,
        select_unlocked,
        format_cells,
        format_columns,
        format_rows,
        insert_columns,
        insert_rows,
        insert_hyperlinks,
        delete_columns,
        delete_rows,
        sort,
        autofilter,
        pivot_tables,
        edit_objects,
        edit_scenarios,
        contents,
    }
}

fn load_b2f_settings(ui: &MainWindow) {
    let s: B2fSettings = confy::load(APP_NAME, "b2f").unwrap_or_default();
    let b2f = ui.global::<BudgetState>();
    b2f.set_version(s.version.into());
    b2f.set_protect_sheet(s.protect_sheet);
    b2f.set_protect_workbook(s.protect_workbook);
    b2f.set_sheet_password(s.sheet_password.into());
    b2f.set_workbook_password(s.workbook_password.into());
    b2f.set_hide_columns(s.hide_columns);
    b2f.set_hide_lang_sheet(s.hide_lang_sheet);
    b2f.set_sheet_permissions(permissions_from_settings(
        s.select_locked,
        s.select_unlocked,
        s.format_cells,
        s.format_columns,
        s.format_rows,
        s.insert_columns,
        s.insert_rows,
        s.insert_hyperlinks,
        s.delete_columns,
        s.delete_rows,
        s.sort,
        s.autofilter,
        s.pivot_tables,
        s.edit_objects,
        s.edit_scenarios,
        s.contents,
    ));
}

fn save_b2f_settings(ui: &MainWindow) {
    let b2f = ui.global::<BudgetState>();
    let sp = b2f.get_sheet_permissions();
    let s = B2fSettings {
        version: b2f.get_version().to_string(),
        protect_sheet: b2f.get_protect_sheet(),
        protect_workbook: b2f.get_protect_workbook(),
        sheet_password: b2f.get_sheet_password().to_string(),
        workbook_password: b2f.get_workbook_password().to_string(),
        hide_columns: b2f.get_hide_columns(),
        hide_lang_sheet: b2f.get_hide_lang_sheet(),
        select_locked: sp.select_locked,
        select_unlocked: sp.select_unlocked,
        format_cells: sp.format_cells,
        format_columns: sp.format_columns,
        format_rows: sp.format_rows,
        insert_columns: sp.insert_columns,
        insert_rows: sp.insert_rows,
        insert_hyperlinks: sp.insert_hyperlinks,
        delete_columns: sp.delete_columns,
        delete_rows: sp.delete_rows,
        sort: sp.sort,
        autofilter: sp.autofilter,
        pivot_tables: sp.pivot_tables,
        edit_objects: sp.edit_objects,
        edit_scenarios: sp.edit_scenarios,
        contents: sp.contents,
    };
    let _ = confy::store(APP_NAME, "b2f", &s);
}

fn load_fb_settings(ui: &MainWindow) {
    let s: FbSettings = confy::load(APP_NAME, "fb").unwrap_or_default();
    let fb = ui.global::<FBState>();
    fb.set_langs(Languages {
        de: s.langs[0],
        en: s.langs[1],
        fr: s.langs[2],
        es: s.langs[3],
        pt: s.langs[4],
    });
    fb.set_categories(Categories {
        cat1: s.categories[0],
        cat2: s.categories[1],
        cat3: s.categories[2],
        cat4: s.categories[3],
        cat5: s.categories[4],
        cat6: s.categories[5],
        cat7: s.categories[6],
        cat8: s.categories[7],
    });
    fb.set_version(s.version.into());
    fb.set_protect_sheet(s.protect_sheet);
    fb.set_protect_workbook(s.protect_workbook);
    fb.set_sheet_password(s.sheet_password.into());
    fb.set_workbook_password(s.workbook_password.into());
    fb.set_hide_columns(s.hide_columns);
    fb.set_hide_lang_sheet(s.hide_lang_sheet);
    fb.set_sheet_permissions(permissions_from_settings(
        s.select_locked,
        s.select_unlocked,
        s.format_cells,
        s.format_columns,
        s.format_rows,
        s.insert_columns,
        s.insert_rows,
        s.insert_hyperlinks,
        s.delete_columns,
        s.delete_rows,
        s.sort,
        s.autofilter,
        s.pivot_tables,
        s.edit_objects,
        s.edit_scenarios,
        s.contents,
    ));
}

fn save_fb_settings(ui: &MainWindow) {
    let fb = ui.global::<FBState>();
    let sp = fb.get_sheet_permissions();
    let langs = fb.get_langs();
    let cats = fb.get_categories();
    let s = FbSettings {
        langs: [langs.de, langs.en, langs.fr, langs.es, langs.pt],
        categories: [
            cats.cat1, cats.cat2, cats.cat3, cats.cat4, cats.cat5, cats.cat6, cats.cat7, cats.cat8,
        ],
        version: fb.get_version().to_string(),
        protect_sheet: fb.get_protect_sheet(),
        protect_workbook: fb.get_protect_workbook(),
        sheet_password: fb.get_sheet_password().to_string(),
        workbook_password: fb.get_workbook_password().to_string(),
        hide_columns: fb.get_hide_columns(),
        hide_lang_sheet: fb.get_hide_lang_sheet(),
        select_locked: sp.select_locked,
        select_unlocked: sp.select_unlocked,
        format_cells: sp.format_cells,
        format_columns: sp.format_columns,
        format_rows: sp.format_rows,
        insert_columns: sp.insert_columns,
        insert_rows: sp.insert_rows,
        insert_hyperlinks: sp.insert_hyperlinks,
        delete_columns: sp.delete_columns,
        delete_rows: sp.delete_rows,
        sort: sp.sort,
        autofilter: sp.autofilter,
        pivot_tables: sp.pivot_tables,
        edit_objects: sp.edit_objects,
        edit_scenarios: sp.edit_scenarios,
        contents: sp.contents,
    };
    let _ = confy::store(APP_NAME, "fb", &s);
}

// Budget-Scanner & FB-Generator: Standardwerte

fn apply_fb_defaults(ui: &MainWindow) {
    let fb = ui.global::<FBState>();

    fb.set_langs(Languages {
        de: true,
        en: true,
        fr: true,
        es: true,
        pt: true,
    });
    fb.set_all_langs(true);

    fb.set_version("".into());
    fb.set_folder("".into());

    fb.set_categories(Categories {
        cat1: 20,
        cat2: 20,
        cat3: 30,
        cat4: 30,
        cat5: 20,
        cat6: 0,
        cat7: 0,
        cat8: 0,
    });

    fb.set_protect_sheet(true);
    fb.set_protect_workbook(true);
    fb.set_sheet_password("".into());
    fb.set_workbook_password("".into());
    fb.set_hide_columns(true);
    fb.set_hide_lang_sheet(true);

    fb.set_sheet_permissions(SheetPermissions {
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

    fb.set_status_type("idle".into());
    fb.set_status_message("".into());
}

fn apply_b2f_defaults(ui: &MainWindow) {
    let b2f = ui.global::<BudgetState>();
    b2f.set_src_folder("".into());
    b2f.set_out_folder("".into());
    b2f.set_status_type("idle".into());
    b2f.set_status_message("".into());

    b2f.set_version("".into());
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

/// Sortiert wie Windows Explorer: Zahlen numerisch, dann alphabetisch (case-insensitive).
fn sort_subfolders(items: &mut [slint::SharedString]) {
    items.sort_by(|a, b| natural_cmp(&a.to_string(), &b.to_string()));
}

/// Natural sort: "2. Vertrag" < "10. Berichte" (nicht lexikographisch).
fn natural_cmp(a: &str, b: &str) -> std::cmp::Ordering {
    let mut ai = a.chars().peekable();
    let mut bi = b.chars().peekable();

    loop {
        match (ai.peek(), bi.peek()) {
            (None, None) => return std::cmp::Ordering::Equal,
            (None, Some(_)) => return std::cmp::Ordering::Less,
            (Some(_), None) => return std::cmp::Ordering::Greater,
            (Some(ac), Some(bc)) => {
                if ac.is_ascii_digit() && bc.is_ascii_digit() {
                    let na: String = ai.by_ref().take_while(|c| c.is_ascii_digit()).collect();
                    let nb: String = bi.by_ref().take_while(|c| c.is_ascii_digit()).collect();
                    let cmp = na.len().cmp(&nb.len()).then_with(|| na.cmp(&nb));
                    if cmp != std::cmp::Ordering::Equal {
                        return cmp;
                    }
                } else {
                    let ca = ai.next().unwrap().to_ascii_lowercase();
                    let cb = bi.next().unwrap().to_ascii_lowercase();
                    let cmp = ca.cmp(&cb);
                    if cmp != std::cmp::Ordering::Equal {
                        return cmp;
                    }
                }
            }
        }
    }
}

fn get_subfolders_vec(ui: &MainWindow) -> Vec<String> {
    ui.global::<FolderState>()
        .get_subfolders()
        .iter()
        .map(|s| s.to_string())
        .collect()
}

fn validate_project_name(ui: &MainWindow) {
    let fs = ui.global::<FolderState>();
    let raw = fs.get_project_name().to_string();
    let skip = fs.get_skip_validation();

    if !skip && !raw.is_empty() {
        let formatted = folder_generator::format_project_name(&raw);
        if formatted != raw {
            fs.set_project_name(formatted.into());
        }
    }

    let name = fs.get_project_name().to_string();
    let valid = if skip {
        !name.is_empty()
    } else {
        folder_generator::is_valid_project_number(&name)
    };
    fs.set_project_name_valid(valid);

    let target = fs.get_target_folder().to_string();
    if !target.is_empty() && !name.is_empty() {
        fs.set_folder_exists(PathBuf::from(&target).join(&name).exists());
    } else {
        fs.set_folder_exists(false);
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

// Einstiegspunkt

#[derive(serde::Serialize, Debug, Clone)]
struct ExportOptions {
    pub protect_sheet: bool,
    pub protect_workbook: bool,
    pub sheet_password: String,
    pub workbook_password: String,
    pub hide_columns: bool,
    pub hide_lang_sheet: bool,
    pub select_locked: bool,
    pub select_unlocked: bool,
    pub format_cells: bool,
    pub format_columns: bool,
    pub format_rows: bool,
    pub insert_columns: bool,
    pub insert_rows: bool,
    pub insert_hyperlinks: bool,
    pub delete_columns: bool,
    pub delete_rows: bool,
    pub sort: bool,
    pub autofilter: bool,
    pub pivot_tables: bool,
    pub edit_objects: bool,
    pub edit_scenarios: bool,
}

#[derive(serde::Deserialize, Debug)]
struct ProgressMessage {
    pub status: String,
    pub file: Option<String>,
    pub current: Option<u32>,
    pub total: Option<u32>,
    pub message: String,
}

fn get_sidecar_path() -> std::path::PathBuf {
    // Hier binden wir die kompilierte Go-Exe direkt in die Rust-Anwendung ein!
    let sidecar_bytes = include_bytes!("../../sidecars/Excelize/scanner.exe");

    // Wir entpacken sie in den Temp-Ordner
    let dir = std::env::temp_dir().join("MyAutomationSuite");
    let _ = std::fs::create_dir_all(&dir);

    let exe_name = if cfg!(windows) { "scanner.exe" } else { "scanner" };
    let exe_path = dir.join(exe_name);

    // Nur neu schreiben, wenn sie noch nicht existiert oder sich die Größe geändert hat (z.B. nach einem App-Update)
    // Das verhindert unnötige Schreibvorgänge und beruhigt Antivirenscanner.
    let needs_write = match std::fs::metadata(&exe_path) {
        Ok(meta) => meta.len() as usize != sidecar_bytes.len(),
        Err(_) => true,
    };

    if needs_write {
        let _ = std::fs::write(&exe_path, sidecar_bytes);

        // Auf Linux/macOS müssen wir die Datei ausführbar machen
        #[cfg(unix)]
        {
            use std::os::unix::fs::PermissionsExt;
            if let Ok(mut perms) = std::fs::metadata(&exe_path).map(|m| m.permissions()) {
                perms.set_mode(0o755);
                let _ = std::fs::set_permissions(&exe_path, perms);
            }
        }
    }

    exe_path
}

fn main() -> Result<(), slint::PlatformError> {
    let ui = MainWindow::new()?;

    // Defaults setzen, dann gespeicherte Settings laden
    apply_fb_defaults(&ui);
    load_fb_settings(&ui);
    apply_b2f_defaults(&ui);
    load_b2f_settings(&ui);
    apply_folder_defaults(&ui);
    load_folder_settings(&ui);

    // Theme: gespeicherte Einstellung laden (Fallback: System-Erkennung)
    load_theme_settings(&ui);

    // ==========================================
    // UpdateState: Versionsnummer + Callback
    // ==========================================
    ui.global::<UpdateState>()
        .set_app_version(env!("CARGO_PKG_VERSION").into());

    // Beim Start automatisch nach Updates suchen
    updater::spawn_check(ui.as_weak());

    ui.global::<UpdateState>().on_check_for_update({
        let ui_handle = ui.as_weak();
        move || {
            if let Some(ui) = ui_handle.upgrade() {
                let us = ui.global::<UpdateState>();
                if us.get_update_available() {
                    updater::spawn_install(ui_handle.clone());
                } else {
                    updater::spawn_check(ui_handle.clone());
                }
            }
        }
    });

    ui.global::<UpdateState>().on_restart({
        let ui_handle = ui.as_weak();
        move || {
            // Aktuellen Pfad der .exe ermitteln und neu starten
            if let Ok(exe) = std::env::current_exe() {
                let _ = std::process::Command::new(exe).spawn();
            }
            // Aktuelles Fenster schließen
            if let Some(ui) = ui_handle.upgrade() {
                let _ = ui.hide();
            }
            std::process::exit(0);
        }
    });

    // Dark Mode Toggle
    ui.on_toggle_dark_mode({
        let ui_handle = ui.as_weak();
        move |dark| {
            if let Some(ui) = ui_handle.upgrade() {
                let scheme = if dark {
                    slint::language::ColorScheme::Dark
                } else {
                    slint::language::ColorScheme::Light
                };
                ui.global::<Palette>().set_color_scheme(scheme);
                save_theme_settings(dark);
            }
        }
    });

    // ==========================================
    // FB-Generator Callbacks
    // ==========================================

    ui.global::<FBState>().on_select_folder({
        let ui_handle = ui.as_weak();
        move || {
            if let Some(ui) = ui_handle.upgrade() {
                if let Some(path) = rfd::FileDialog::new().pick_folder() {
                    let fb = ui.global::<FBState>();
                    fb.set_folder(path.to_string_lossy().to_string().into());
                    fb.set_status_type("idle".into());
                    fb.set_status_message("".into());
                }
            }
        }
    });

    ui.global::<FBState>().on_reset({
        let ui_handle = ui.as_weak();
        move || {
            if let Some(ui) = ui_handle.upgrade() {
                apply_fb_defaults(&ui);
                save_fb_settings(&ui);
            }
        }
    });

    ui.global::<FBState>().on_dismiss_status({
        let ui_handle = ui.as_weak();
        move || {
            if let Some(ui) = ui_handle.upgrade() {
                let fb = ui.global::<FBState>();
                fb.set_status_type("idle".into());
                fb.set_status_message("".into());
            }
        }
    });

    ui.global::<FBState>().on_save_settings({
        let ui_handle = ui.as_weak();
        move || {
            if let Some(ui) = ui_handle.upgrade() {
                save_fb_settings(&ui);
            }
        }
    });

    ui.global::<FBState>().on_toggle_settings({
        let ui_handle = ui.as_weak();
        move || {
            if let Some(ui) = ui_handle.upgrade() {
                let fb = ui.global::<FBState>();
                fb.set_show_settings(!fb.get_show_settings());
            }
        }
    });

    ui.global::<FBState>().on_generate_report({
        let ui_handle = ui.as_weak();
        move || {
            if let Some(ui) = ui_handle.upgrade() {
                let fb = ui.global::<FBState>();

                let folder = fb.get_folder().to_string();
                if folder.is_empty() {
                    fb.set_status_type("error".into());
                    fb.set_status_message("Bitte Ausgabeordner wählen.".into());
                    return;
                }

                let version = fb.get_version().to_string();
                if version.is_empty() {
                    fb.set_status_type("error".into());
                    fb.set_status_message("Bitte Version angeben.".into());
                    return;
                }

                let langs = fb.get_langs();
                let mut lang_list = Vec::new();
                if langs.de {
                    lang_list.push("de");
                }
                if langs.en {
                    lang_list.push("en");
                }
                if langs.fr {
                    lang_list.push("fr");
                }
                if langs.es {
                    lang_list.push("es");
                }
                if langs.pt {
                    lang_list.push("pt");
                }

                if lang_list.is_empty() {
                    fb.set_status_type("error".into());
                    fb.set_status_message("Bitte mindestens eine Sprache wählen.".into());
                    return;
                }

                fb.set_status_type("pending".into());
                fb.set_status_message("Export läuft...".into());

                let cats = fb.get_categories();
                let counts = [
                    cats.cat1 as u16,
                    cats.cat2 as u16,
                    cats.cat3 as u16,
                    cats.cat4 as u16,
                    cats.cat5 as u16,
                    cats.cat6 as u16,
                    cats.cat7 as u16,
                    cats.cat8 as u16,
                ];

                let version = fb.get_version().to_string();

                let sp = fb.get_sheet_permissions();
                let options = ExportOptions {
                    protect_sheet: fb.get_protect_sheet(),
                    protect_workbook: fb.get_protect_workbook(),
                    sheet_password: fb.get_sheet_password().to_string(),
                    workbook_password: fb.get_workbook_password().to_string(),
                    hide_columns: fb.get_hide_columns(),
                    hide_lang_sheet: fb.get_hide_lang_sheet(),
                    select_locked: sp.select_locked,
                    select_unlocked: sp.select_unlocked,
                    format_cells: sp.format_cells,
                    format_columns: sp.format_columns,
                    format_rows: sp.format_rows,
                    insert_columns: sp.insert_columns,
                    insert_rows: sp.insert_rows,
                    insert_hyperlinks: sp.insert_hyperlinks,
                    delete_columns: sp.delete_columns,
                    delete_rows: sp.delete_rows,
                    sort: sp.sort,
                    autofilter: sp.autofilter,
                    pivot_tables: sp.pivot_tables,
                    edit_objects: sp.edit_objects,
                    edit_scenarios: sp.edit_scenarios,
                };
                let wb_hash = if options.protect_workbook {
                    Some(excel_protection::precompute_hash(&options.workbook_password))
                } else { None };

                let sh_hash = if options.protect_sheet {
                    Some(excel_protection::precompute_hash(&options.sheet_password))
                } else { None };

                let sh_opts = if options.protect_sheet {
                    Some(excel_protection::SheetProtectionOptions {
                        select_locked_cells: options.select_locked,
                        select_unlocked_cells: options.select_unlocked,
                        format_cells: options.format_cells,
                        format_columns: options.format_columns,
                        format_rows: options.format_rows,
                        insert_columns: options.insert_columns,
                        insert_rows: options.insert_rows,
                        insert_hyperlinks: options.insert_hyperlinks,
                        delete_columns: options.delete_columns,
                        delete_rows: options.delete_rows,
                        sort: options.sort,
                        auto_filter: options.autofilter,
                        pivot_tables: options.pivot_tables,
                        objects: options.edit_objects,
                        scenarios: options.edit_scenarios,
                    })
                } else { None };

                let mut sidecar_options = options.clone();
                sidecar_options.protect_sheet = false;
                sidecar_options.protect_workbook = false;
                let options_json = serde_json::to_string(&sidecar_options).unwrap_or_default();

                let ui_handle_clone = ui_handle.clone();
                std::thread::spawn(move || {
                    let mut templates = Vec::new();

                    for lang in &lang_list {
                        let mut positions = Vec::new();
                        for (i, &pos_count) in counts.iter().enumerate() {
                            let category = i + 1;
                            for p in 0..pos_count {
                                positions.push(budget_scanner::BudgetPosition {
                                    number: format!("{category}.{}", p + 1),
                                    label: String::new(),
                                    cost_col1: String::new(),
                                    cost_col2: String::new(),
                                });
                            }
                        }

                        templates.push(budget_scanner::BudgetData {
                            file_path: std::path::PathBuf::from(format!("Vorlage_{lang}.xlsx")),
                            sheet_name: "Budget".into(),
                            version: version.clone(),
                            project_title: "".into(),
                            project_number: "Vorlage".into(),
                            language: lang.to_string(),
                            local_currency: "".into(),
                            cost_col1: 8,
                            cost_col2: Some(13),
                            eigenleistung: "0".into(),
                            drittmittel: "0".into(),
                            kmw_mittel: "0".into(),
                            positions,
                        });
                    }

                    let output_dir = std::path::PathBuf::from(&folder);
                    let _ = std::fs::create_dir_all(&output_dir);

                    let tmp_json_path = std::env::temp_dir().join(format!(
                        "template_{}.json",
                        std::time::SystemTime::now()
                            .duration_since(std::time::UNIX_EPOCH)
                            .unwrap()
                            .as_millis()
                    ));
                    if let Err(e) = std::fs::File::create(&tmp_json_path).and_then(|mut f| {
                        let json = serde_json::to_string(&templates)?;
                        std::io::Write::write_all(&mut f, json.as_bytes())?;
                        Ok(())
                    }) {
                        let _ = ui_handle_clone.upgrade_in_event_loop(move |ui| {
                            let fb = ui.global::<FBState>();
                            fb.set_status_type("error".into());
                            fb.set_status_message(
                                format!("Fehler beim Speichern der JSON: {e}").into(),
                            );
                        });
                        return;
                    }

                    let sidecar_exe = get_sidecar_path();

                    let mut cmd = std::process::Command::new(&sidecar_exe);
                    cmd.arg("-input")
                        .arg(&tmp_json_path)
                        .arg("-output")
                        .arg(&output_dir)
                        .arg("-options")
                        .arg(&options_json);

                    if !version.is_empty() {
                        cmd.arg("-filename")
                            .arg(format!("Vorlage_{{la}}_{version}_FB.xlsx"));
                    } else {
                        cmd.arg("-filename").arg("Vorlage_{la}_FB.xlsx");
                    }

                    cmd.stdout(std::process::Stdio::piped());

                    let mut child = match cmd.spawn() {
                        Ok(c) => c,
                        Err(e) => {
                            let _ = ui_handle_clone.upgrade_in_event_loop(move |ui| {
                                let fb = ui.global::<FBState>();
                                fb.set_status_type("error".into());
                                fb.set_status_message(
                                    format!("Fehler beim Starten von {}: {e}", sidecar_exe.display()).into(),
                                );
                            });
                            return;
                        }
                    };

                    let stdout = child.stdout.take().unwrap();
                    let reader = std::io::BufReader::new(stdout);

                    use std::io::BufRead;
                    for line in reader.lines().map_while(Result::ok) {
                        if let Ok(msg) = serde_json::from_str::<ProgressMessage>(&line) {
                            let _ = ui_handle_clone.upgrade_in_event_loop({
                                let msg_status = msg.status.clone();
                                let msg_text = msg.message.clone();
                                let current = msg.current.unwrap_or(0);
                                let total = msg.total.unwrap_or(0);

                                move |ui| {
                                    let fb = ui.global::<FBState>();

                                    if msg_status == "error" {
                                        fb.set_status_type("error".into());
                                    } else if msg_status == "done" {
                                        fb.set_status_type("success".into());
                                    } else {
                                        fb.set_status_type("pending".into());
                                    }

                                    if total > 0 {
                                        fb.set_status_message(
                                            format!("{current}/{total} - {msg_text}").into(),
                                        );
                                    } else {
                                        fb.set_status_message(msg_text.into());
                                    }
                                }
                            });
                        }
                    }

                    let _ = child.wait();
                    let _ = std::fs::remove_file(&tmp_json_path);

                    if wb_hash.is_some() || sh_hash.is_some() {
                        let _ = ui_handle_clone.upgrade_in_event_loop(move |ui| {
                            let fb = ui.global::<FBState>();
                            fb.set_status_message("Wende Verschlüsselung an...".into());
                        });

                        use rayon::prelude::*;
                        if let Ok(entries) = std::fs::read_dir(&output_dir) {
                            let paths: Vec<_> = entries.flatten().map(|e| e.path()).collect();
                            paths.into_par_iter().for_each(|p| {
                                if p.extension().map_or(false, |ext| ext == "xlsx") {
                                    let _ = excel_protection::apply_protection_in_place(
                                        &p,
                                        wb_hash.as_ref(),
                                        sh_hash.as_ref(),
                                        sh_opts.as_ref()
                                    );
                                }
                            });
                        }

                        let _ = ui_handle_clone.upgrade_in_event_loop(move |ui| {
                            let fb = ui.global::<FBState>();
                            fb.set_status_message("Export und Verschlüsselung erfolgreich abgeschlossen!".into());
                        });
                    }
                });
            }
        }
    });

    // ==========================================
    // Budget-to-FB Callbacks
    // ==========================================

    ui.global::<BudgetState>().on_select_src({
        let ui_handle = ui.as_weak();
        move || {
            if let Some(ui) = ui_handle.upgrade() {
                if let Some(path) = rfd::FileDialog::new().pick_folder() {
                    ui.global::<BudgetState>()
                        .set_src_folder(path.to_string_lossy().to_string().into());
                }
            }
        }
    });

    ui.global::<BudgetState>().on_select_out({
        let ui_handle = ui.as_weak();
        move || {
            if let Some(ui) = ui_handle.upgrade() {
                if let Some(path) = rfd::FileDialog::new().pick_folder() {
                    ui.global::<BudgetState>()
                        .set_out_folder(path.to_string_lossy().to_string().into());
                }
            }
        }
    });

    ui.global::<BudgetState>().on_scan({
        let ui_handle = ui.as_weak();
        move || {
            if let Some(ui) = ui_handle.upgrade() {
                let b2f = ui.global::<BudgetState>();

                let src = b2f.get_src_folder().to_string();
                let out_base = b2f.get_out_folder().to_string();

                if src.is_empty() {
                    b2f.set_status_type("error".into());
                    b2f.set_status_message("Bitte Quellordner wählen.".into());
                    return;
                }
                if out_base.is_empty() {
                    b2f.set_status_type("error".into());
                    b2f.set_status_message("Bitte Ausgabeordner wählen.".into());
                    return;
                }

                b2f.set_status_type("pending".into());
                b2f.set_status_message("Scannt...".into());

                let version = b2f.get_version().to_string();

                let sp = b2f.get_sheet_permissions();
                let options = ExportOptions {
                    protect_sheet: b2f.get_protect_sheet(),
                    protect_workbook: b2f.get_protect_workbook(),
                    sheet_password: b2f.get_sheet_password().to_string(),
                    workbook_password: b2f.get_workbook_password().to_string(),
                    hide_columns: b2f.get_hide_columns(),
                    hide_lang_sheet: b2f.get_hide_lang_sheet(),
                    select_locked: sp.select_locked,
                    select_unlocked: sp.select_unlocked,
                    format_cells: sp.format_cells,
                    format_columns: sp.format_columns,
                    format_rows: sp.format_rows,
                    insert_columns: sp.insert_columns,
                    insert_rows: sp.insert_rows,
                    insert_hyperlinks: sp.insert_hyperlinks,
                    delete_columns: sp.delete_columns,
                    delete_rows: sp.delete_rows,
                    sort: sp.sort,
                    autofilter: sp.autofilter,
                    pivot_tables: sp.pivot_tables,
                    edit_objects: sp.edit_objects,
                    edit_scenarios: sp.edit_scenarios,
                };

                let wb_hash = if options.protect_workbook {
                    Some(excel_protection::precompute_hash(&options.workbook_password))
                } else { None };

                let sh_hash = if options.protect_sheet {
                    Some(excel_protection::precompute_hash(&options.sheet_password))
                } else { None };

                let sh_opts = if options.protect_sheet {
                    Some(excel_protection::SheetProtectionOptions {
                        select_locked_cells: options.select_locked,
                        select_unlocked_cells: options.select_unlocked,
                        format_cells: options.format_cells,
                        format_columns: options.format_columns,
                        format_rows: options.format_rows,
                        insert_columns: options.insert_columns,
                        insert_rows: options.insert_rows,
                        insert_hyperlinks: options.insert_hyperlinks,
                        delete_columns: options.delete_columns,
                        delete_rows: options.delete_rows,
                        sort: options.sort,
                        auto_filter: options.autofilter,
                        pivot_tables: options.pivot_tables,
                        objects: options.edit_objects,
                        scenarios: options.edit_scenarios,
                    })
                } else { None };

                // Dem Sidecar geben wir protect=false mit, damit es das XML nicht verschlüsselt
                let mut sidecar_options = options.clone();
                sidecar_options.protect_sheet = false;
                sidecar_options.protect_workbook = false;
                let options_json = serde_json::to_string(&sidecar_options).unwrap_or_default();

                let ui_handle_clone = ui_handle.clone();
                std::thread::spawn(move || {
                    let src_path = std::path::PathBuf::from(&src);
                    let out_base_path = std::path::PathBuf::from(&out_base);

                    // 1. Budget-Dateien scannen
                    let mut result = budget_scanner::scan_directory(&src_path);

                    // 2. Override version
                    if !version.is_empty() {
                        for data in &mut result.successes {
                            data.version = version.clone();
                        }
                    }

                    // 3. Output-Ordner bestimmen
                    let output_dir = budget_scanner::resolve_output_dir(&out_base_path);
                    let _ = std::fs::create_dir_all(&output_dir);

                    // 4. Temporäres JSON erstellen
                    let tmp_json_path = std::env::temp_dir().join(format!(
                        "scan_{}.json",
                        std::time::SystemTime::now()
                            .duration_since(std::time::UNIX_EPOCH)
                            .unwrap()
                            .as_millis()
                    ));
                    if let Err(e) = std::fs::File::create(&tmp_json_path).and_then(|mut f| {
                        let json = serde_json::to_string(&result.successes)?;
                        std::io::Write::write_all(&mut f, json.as_bytes())?;
                        Ok(())
                    }) {
                        let _ = ui_handle_clone.upgrade_in_event_loop(move |ui| {
                            let b2f = ui.global::<BudgetState>();
                            b2f.set_status_type("error".into());
                            b2f.set_status_message(
                                format!("Fehler beim Speichern der JSON: {e}").into(),
                            );
                        });
                        return;
                    }

                    // 5. Go Sidecar aufrufen
                    let sidecar_exe = get_sidecar_path();

                    let mut cmd = std::process::Command::new(&sidecar_exe);
                    cmd.arg("-input")
                        .arg(&tmp_json_path)
                        .arg("-output")
                        .arg(&output_dir)
                        .arg("-options")
                        .arg(&options_json);

                    // Wenn version gesetzt ist, fügen wir es ins Namensmuster ein
                    if !version.is_empty() {
                        cmd.arg("-filename")
                            .arg(format!("{{pn}}_{{la}}_{version}_FB.xlsx"));
                    } else {
                        cmd.arg("-filename").arg("{pn}_{la}_FB.xlsx");
                    }

                    cmd.stdout(std::process::Stdio::piped());

                    let mut child = match cmd.spawn() {
                        Ok(c) => c,
                        Err(e) => {
                            let _ = ui_handle_clone.upgrade_in_event_loop(move |ui| {
                                let b2f = ui.global::<BudgetState>();
                                b2f.set_status_type("error".into());
                                b2f.set_status_message(
                                    format!("Fehler beim Starten von {}: {e}", sidecar_exe.display()).into(),
                                );
                            });
                            return;
                        }
                    };

                    let stdout = child.stdout.take().unwrap();
                    let reader = std::io::BufReader::new(stdout);

                    use std::io::BufRead;
                    for line in reader.lines().map_while(Result::ok) {
                        if let Ok(msg) = serde_json::from_str::<ProgressMessage>(&line) {
                            let _ = ui_handle_clone.upgrade_in_event_loop({
                                let msg_status = msg.status.clone();
                                let msg_text = msg.message.clone();
                                let current = msg.current.unwrap_or(0);
                                let total = msg.total.unwrap_or(0);

                                move |ui| {
                                    let b2f = ui.global::<BudgetState>();

                                    if msg_status == "error" {
                                        b2f.set_status_type("error".into());
                                    } else if msg_status == "done" {
                                        b2f.set_status_type("success".into());
                                    } else {
                                        b2f.set_status_type("pending".into());
                                    }

                                    if total > 0 {
                                        b2f.set_status_message(
                                            format!("{current}/{total} - {msg_text}").into(),
                                        );
                                    } else {
                                        b2f.set_status_message(msg_text.into());
                                    }
                                }
                            });
                        }
                    }

                    let _ = child.wait();
                    let _ = std::fs::remove_file(&tmp_json_path);

                    // --- RUST EXCEL PROTECTION ---
                    // Wir durchlaufen alle generierten XLSX Dateien und wenden den schnellen XML-Schutz an
                    if wb_hash.is_some() || sh_hash.is_some() {
                        let _ = ui_handle_clone.upgrade_in_event_loop(move |ui| {
                            let b2f = ui.global::<BudgetState>();
                            b2f.set_status_message("Wende Verschlüsselung an...".into());
                        });

                        use rayon::prelude::*;
                        if let Ok(entries) = std::fs::read_dir(&output_dir) {
                            let paths: Vec<_> = entries.flatten().map(|e| e.path()).collect();
                            paths.into_par_iter().for_each(|p| {
                                if p.extension().map_or(false, |ext| ext == "xlsx") {
                                    let _ = excel_protection::apply_protection_in_place(
                                        &p,
                                        wb_hash.as_ref(),
                                        sh_hash.as_ref(),
                                        sh_opts.as_ref()
                                    );
                                }
                            });
                        }

                        let _ = ui_handle_clone.upgrade_in_event_loop(move |ui| {
                            let b2f = ui.global::<BudgetState>();
                            b2f.set_status_message("Scan und Verschlüsselung erfolgreich abgeschlossen!".into());
                        });
                    }

                    // 6. Fehler-CSV schreiben
                    if !result.failures.is_empty() {
                        let csv_path = output_dir.join("scan_fehler.csv");
                        let _ = budget_scanner::write_failure_report(&result.failures, &csv_path);
                    }

                    // 7. Tabelle aktualisieren
                    let _ = ui_handle_clone.upgrade_in_event_loop(move |ui| {
                        let b2f = ui.global::<BudgetState>();

                        let mk_col = |t: &str| {
                            let mut c = slint::TableColumn::default();
                            c.title = t.into();
                            c
                        };
                        let columns = slint::ModelRc::new(slint::VecModel::from(vec![
                            mk_col("Dateiname"),
                            mk_col("Status"),
                            mk_col("Details"),
                        ]));
                        b2f.set_table_columns(columns);

                        let mut rows: Vec<slint::ModelRc<slint::StandardListViewItem>> = Vec::new();

                        for data in &result.successes {
                            let fname = data
                                .file_path
                                .file_name()
                                .map(|n| n.to_string_lossy().to_string())
                                .unwrap_or_default();

                            rows.push(slint::ModelRc::new(slint::VecModel::from(vec![
                                slint::StandardListViewItem::from(slint::SharedString::from(
                                    &fname,
                                )),
                                slint::StandardListViewItem::from(slint::SharedString::from("OK")),
                                slint::StandardListViewItem::from(slint::SharedString::from(
                                    "Generiert",
                                )),
                            ])));
                        }

                        for f in &result.failures {
                            rows.push(slint::ModelRc::new(slint::VecModel::from(vec![
                                slint::StandardListViewItem::from(slint::SharedString::from(
                                    &f.file_name,
                                )),
                                slint::StandardListViewItem::from(slint::SharedString::from(
                                    "Fehler",
                                )),
                                slint::StandardListViewItem::from(slint::SharedString::from(
                                    f.reason.to_string(),
                                )),
                            ])));
                        }

                        let table_data = slint::ModelRc::new(slint::VecModel::from(rows));
                        b2f.set_table_data(table_data);
                    });
                });
            }
        }
    });

    ui.global::<BudgetState>().on_do_export_txt({
        let ui_handle = ui.as_weak();
        move || {
            if let Some(ui) = ui_handle.upgrade() {
                let b2f = ui.global::<BudgetState>();
                let table_data = b2f.get_table_data();
                let columns = b2f.get_table_columns();

                if let Some(path) = rfd::FileDialog::new()
                    .set_file_name("scan_ergebnis.csv")
                    .add_filter("CSV", &["csv"])
                    .save_file()
                {
                    let mut out = String::new();
                    let col_count = columns.row_count();
                    for c in 0..col_count {
                        if c > 0 {
                            out.push(';');
                        }
                        out.push_str(
                            &columns
                                .row_data(c)
                                .map(|col| col.title.to_string())
                                .unwrap_or_default(),
                        );
                    }
                    out.push('\n');
                    for r in 0..table_data.row_count() {
                        if let Some(row) = table_data.row_data(r) {
                            for c in 0..col_count {
                                if c > 0 {
                                    out.push(';');
                                }
                                out.push_str(
                                    &row.row_data(c)
                                        .map(|item| item.text.to_string())
                                        .unwrap_or_default(),
                                );
                            }
                            out.push('\n');
                        }
                    }
                    match std::fs::write(&path, &out) {
                        Ok(()) => {
                            b2f.set_status_type("success".into());
                            b2f.set_status_message(
                                format!("CSV exportiert: {}", path.display()).into(),
                            );
                        }
                        Err(e) => {
                            b2f.set_status_type("error".into());
                            b2f.set_status_message(format!("CSV-Export Fehler: {e}").into());
                        }
                    }
                }
            }
        }
    });

    ui.global::<BudgetState>().on_do_export_excel({
        let ui_handle = ui.as_weak();
        move || {
            if let Some(ui) = ui_handle.upgrade() {
                let b2f = ui.global::<BudgetState>();
                let table_data = b2f.get_table_data();
                let columns = b2f.get_table_columns();

                if let Some(path) = rfd::FileDialog::new()
                    .set_file_name("scan_ergebnis.xlsx")
                    .add_filter("Excel", &["xlsx"])
                    .save_file()
                {
                    let col_count = columns.row_count();
                    let mut workbook = rust_xlsxwriter::Workbook::new();
                    let sheet = workbook.add_worksheet();

                    // Header
                    for c in 0..col_count {
                        let title = columns
                            .row_data(c)
                            .map(|col| col.title.to_string())
                            .unwrap_or_default();
                        let _ = sheet.write_string(0, c as u16, &title);
                    }

                    // Rows
                    for r in 0..table_data.row_count() {
                        if let Some(row) = table_data.row_data(r) {
                            for c in 0..col_count {
                                let text = row
                                    .row_data(c)
                                    .map(|item| item.text.to_string())
                                    .unwrap_or_default();
                                let _ = sheet.write_string((r + 1) as u32, c as u16, &text);
                            }
                        }
                    }

                    match workbook.save(&path) {
                        Ok(()) => {
                            b2f.set_status_type("success".into());
                            b2f.set_status_message(
                                format!("Excel exportiert: {}", path.display()).into(),
                            );
                        }
                        Err(e) => {
                            b2f.set_status_type("error".into());
                            b2f.set_status_message(format!("Excel-Export Fehler: {e}").into());
                        }
                    }
                }
            }
        }
    });

    ui.global::<BudgetState>().on_dismiss_status({
        let ui_handle = ui.as_weak();
        move || {
            if let Some(ui) = ui_handle.upgrade() {
                let b2f = ui.global::<BudgetState>();
                b2f.set_status_type("idle".into());
                b2f.set_status_message("".into());
            }
        }
    });

    ui.global::<BudgetState>().on_do_reset({
        let ui_handle = ui.as_weak();
        move || {
            if let Some(ui) = ui_handle.upgrade() {
                apply_b2f_defaults(&ui);
                let b2f = ui.global::<BudgetState>();
                b2f.set_table_data(slint::ModelRc::default());
                b2f.set_table_columns(slint::ModelRc::default());
                save_b2f_settings(&ui);
            }
        }
    });

    ui.global::<BudgetState>().on_toggle_settings({
        let ui_handle = ui.as_weak();
        move || {
            if let Some(ui) = ui_handle.upgrade() {
                let b2f = ui.global::<BudgetState>();
                b2f.set_show_settings(!b2f.get_show_settings());
            }
        }
    });

    ui.global::<BudgetState>().on_save_settings({
        let ui_handle = ui.as_weak();
        move || {
            if let Some(ui) = ui_handle.upgrade() {
                save_b2f_settings(&ui);
            }
        }
    });

    // ==========================================
    // Folder-Creation Callbacks
    // ==========================================

    ui.global::<FolderState>().on_validate_project_name({
        let ui_handle = ui.as_weak();
        move || {
            if let Some(ui) = ui_handle.upgrade() {
                validate_project_name(&ui);
            }
        }
    });

    ui.global::<FolderState>().on_select_folder({
        let ui_handle = ui.as_weak();
        move || {
            if let Some(ui) = ui_handle.upgrade() {
                if let Some(path) = rfd::FileDialog::new().pick_folder() {
                    ui.global::<FolderState>()
                        .set_target_folder(path.to_string_lossy().to_string().into());
                    save_folder_settings(&ui);
                    validate_project_name(&ui);
                }
            }
        }
    });

    ui.global::<FolderState>().on_select_template({
        let ui_handle = ui.as_weak();
        move || {
            if let Some(ui) = ui_handle.upgrade() {
                if let Some(path) = rfd::FileDialog::new()
                    .add_filter("Excel", &["xlsm", "xlsx"])
                    .pick_file()
                {
                    ui.global::<FolderState>()
                        .set_template_file(path.to_string_lossy().to_string().into());
                    save_folder_settings(&ui);
                }
            }
        }
    });

    ui.global::<FolderState>().on_reset({
        let ui_handle = ui.as_weak();
        move || {
            if let Some(ui) = ui_handle.upgrade() {
                apply_folder_defaults(&ui);
                save_folder_settings(&ui);
            }
        }
    });

    ui.global::<FolderState>().on_dismiss_status({
        let ui_handle = ui.as_weak();
        move || {
            if let Some(ui) = ui_handle.upgrade() {
                let fs = ui.global::<FolderState>();
                fs.set_status_type("idle".into());
                fs.set_status_message("".into());
            }
        }
    });

    ui.global::<FolderState>().on_select_csv_file({
        let ui_handle = ui.as_weak();
        move || {
            if let Some(ui) = ui_handle.upgrade() {
                if let Some(path) = rfd::FileDialog::new()
                    .add_filter("CSV", &["csv"])
                    .pick_file()
                {
                    ui.global::<FolderState>()
                        .set_csv_file(path.to_string_lossy().to_string().into());
                }
            }
        }
    });

    ui.global::<FolderState>().on_select_csv_target({
        let ui_handle = ui.as_weak();
        move || {
            if let Some(ui) = ui_handle.upgrade() {
                if let Some(path) = rfd::FileDialog::new().pick_folder() {
                    ui.global::<FolderState>()
                        .set_csv_target_folder(path.to_string_lossy().to_string().into());
                }
            }
        }
    });

    ui.global::<FolderState>().on_import_csv({
        let ui_handle = ui.as_weak();
        move || {
            if let Some(ui) = ui_handle.upgrade() {
                let fs = ui.global::<FolderState>();
                let csv_path = PathBuf::from(fs.get_csv_file().to_string());
                let target = PathBuf::from(fs.get_csv_target_folder().to_string());
                let template = PathBuf::from(fs.get_template_file().to_string());

                if !target.exists() {
                    fs.set_status_type("error".into());
                    fs.set_status_message("CSV-Zielverzeichnis existiert nicht.".into());
                    return;
                }
                if !template.exists() {
                    fs.set_status_type("error".into());
                    fs.set_status_message(
                        "Excel-Vorlage nicht gefunden (oben im Einzelordner-Bereich auswählen)."
                            .into(),
                    );
                    return;
                }

                let subs = get_subfolders_vec(&ui);
                let subs_refs: Vec<&str> = subs.iter().map(|s| s.as_str()).collect();
                match folder_generator::import_csv(&csv_path, &target, &template, &subs_refs) {
                    Ok(result) => {
                        if result.errors.is_empty() {
                            fs.set_status_type("success".into());
                            fs.set_status_message(
                                format!(
                                    "{} Ordner erstellt, {} übersprungen (existierten bereits).",
                                    result.created, result.skipped
                                )
                                .into(),
                            );
                        } else {
                            fs.set_status_type("error".into());
                            fs.set_status_message(
                                format!(
                                    "{} erstellt, {} übersprungen, {} Fehler: {}",
                                    result.created,
                                    result.skipped,
                                    result.errors.len(),
                                    result.errors.join("; ")
                                )
                                .into(),
                            );
                        }
                    }
                    Err(e) => {
                        fs.set_status_type("error".into());
                        fs.set_status_message(e.to_string().into());
                    }
                }
            }
        }
    });

    // ─── Subfolder-Liste: Callbacks ───
    ui.global::<FolderState>().on_add_subfolder({
        let ui_handle = ui.as_weak();
        move |name| {
            if let Some(ui) = ui_handle.upgrade() {
                let fs = ui.global::<FolderState>();
                let model = fs.get_subfolders();
                let mut items: Vec<slint::SharedString> = model.iter().collect();
                items.push(name);
                sort_subfolders(&mut items);
                fs.set_subfolders(std::rc::Rc::new(slint::VecModel::from(items)).into());
                save_folder_settings(&ui);
            }
        }
    });

    ui.global::<FolderState>().on_remove_subfolder({
        let ui_handle = ui.as_weak();
        move |idx| {
            if let Some(ui) = ui_handle.upgrade() {
                let fs = ui.global::<FolderState>();
                let model = fs.get_subfolders();
                let mut items: Vec<slint::SharedString> = model.iter().collect();
                if (idx as usize) < items.len() {
                    items.remove(idx as usize);
                    fs.set_subfolders(std::rc::Rc::new(slint::VecModel::from(items)).into());
                    save_folder_settings(&ui);
                }
            }
        }
    });

    ui.global::<FolderState>().on_rename_subfolder({
        let ui_handle = ui.as_weak();
        move |idx, new_name| {
            if let Some(ui) = ui_handle.upgrade() {
                let fs = ui.global::<FolderState>();
                let model = fs.get_subfolders();
                let mut items: Vec<slint::SharedString> = model.iter().collect();
                if (idx as usize) < items.len() {
                    items[idx as usize] = new_name;
                    sort_subfolders(&mut items);
                    fs.set_subfolders(std::rc::Rc::new(slint::VecModel::from(items)).into());
                    save_folder_settings(&ui);
                }
            }
        }
    });

    ui.global::<FolderState>().on_reset_subfolders({
        let ui_handle = ui.as_weak();
        move || {
            if let Some(ui) = ui_handle.upgrade() {
                let mut defaults: Vec<slint::SharedString> = folder_generator::SUBFOLDERS
                    .iter()
                    .map(|s| (*s).into())
                    .collect();
                sort_subfolders(&mut defaults);
                ui.global::<FolderState>()
                    .set_subfolders(std::rc::Rc::new(slint::VecModel::from(defaults)).into());
                save_folder_settings(&ui);
            }
        }
    });

    ui.global::<FolderState>().on_create_folders({
        let ui_handle = ui.as_weak();
        move || {
            if let Some(ui) = ui_handle.upgrade() {
                let fs = ui.global::<FolderState>();
                let project_name = fs.get_project_name().to_string();
                let target = PathBuf::from(fs.get_target_folder().to_string());
                let template = PathBuf::from(fs.get_template_file().to_string());

                if project_name.is_empty() {
                    fs.set_status_type("error".into());
                    fs.set_status_message("Bitte Projektnamen angeben.".into());
                    return;
                }
                if !target.exists() {
                    fs.set_status_type("error".into());
                    fs.set_status_message("Zielverzeichnis existiert nicht.".into());
                    return;
                }
                if !template.exists() {
                    fs.set_status_type("error".into());
                    fs.set_status_message("Excel-Vorlage nicht gefunden.".into());
                    return;
                }

                let subs = get_subfolders_vec(&ui);
                let subs_refs: Vec<&str> = subs.iter().map(|s| s.as_str()).collect();
                match folder_generator::create_project_folder(
                    &project_name,
                    &target,
                    &template,
                    &subs_refs,
                ) {
                    Ok(root) => {
                        fs.set_project_name("".into());
                        fs.set_project_name_valid(false);
                        fs.set_status_type("success".into());
                        fs.set_status_message(
                            format!("Projektordner erstellt: {}", root.display()).into(),
                        );
                    }
                    Err(e) => {
                        fs.set_status_type("error".into());
                        fs.set_status_message(e.to_string().into());
                    }
                }
            }
        }
    });

    ui.run()
}
