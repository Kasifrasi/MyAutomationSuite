#![windows_subsystem = "windows"]

slint::include_modules!();

mod updater;

use fb_generator::{
    Language, PositionEntry, ReportBody, ReportConfig, ReportHeader, ReportOptions, SheetProtection,
};
use slint::Model;
use std::path::PathBuf;

const APP_NAME: &str = "automation-tool";

// ==========================================
// Theme: Einstellungen
// ==========================================

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

// ==========================================
// Budget-Scanner & FB-Generator: Einstellungen
// ==========================================

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

// ==========================================
// Budget-Scanner & FB-Generator: Laden / Speichern
// ==========================================

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

// ==========================================
// Budget-Scanner & FB-Generator: Standardwerte
// ==========================================

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

// ==========================================
// Ordner-Generator: Einstellungen & Hilfsfunktionen
// ==========================================

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
            subfolders: folder_generator::SUBFOLDERS.iter().map(|s| s.to_string()).collect(),
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
    ui.global::<FolderState>().get_subfolders().iter().map(|s| s.to_string()).collect()
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
    let defaults: Vec<slint::SharedString> = folder_generator::SUBFOLDERS.iter().map(|s| (*s).into()).collect();
    fs.set_subfolders(std::rc::Rc::new(slint::VecModel::from(defaults)).into());
}

// ==========================================
// Einstiegspunkt
// ==========================================

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
    ui.global::<UpdateState>().set_app_version(env!("CARGO_PKG_VERSION").into());

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
                    lang_list.push(Language::Deutsch);
                }
                if langs.en {
                    lang_list.push(Language::English);
                }
                if langs.fr {
                    lang_list.push(Language::Francais);
                }
                if langs.es {
                    lang_list.push(Language::Espanol);
                }
                if langs.pt {
                    lang_list.push(Language::Portugues);
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

                let sheet_prot = if fb.get_protect_sheet() {
                    let sp = fb.get_sheet_permissions();
                    Some(
                        SheetProtection::new()
                            .with_password(fb.get_sheet_password().to_string())
                            .allow_select_locked_cells(sp.select_locked)
                            .allow_select_unlocked_cells(sp.select_unlocked)
                            .allow_format_cells(sp.format_cells)
                            .allow_format_columns(sp.format_columns)
                            .allow_format_rows(sp.format_rows)
                            .allow_insert_columns(sp.insert_columns)
                            .allow_insert_rows(sp.insert_rows)
                            .allow_insert_hyperlinks(sp.insert_hyperlinks)
                            .allow_delete_columns(sp.delete_columns)
                            .allow_delete_rows(sp.delete_rows)
                            .allow_sort(sp.sort)
                            .allow_autofilter(sp.autofilter)
                            .allow_pivot_tables(sp.pivot_tables)
                            .allow_edit_objects(sp.edit_objects)
                            .allow_edit_scenarios(sp.edit_scenarios)
                            .allow_contents(sp.contents),
                    )
                } else {
                    None
                };

                let workbook_pw = if fb.get_protect_workbook() {
                    Some(fb.get_workbook_password().to_string())
                } else {
                    None
                };

                match generate_excel(
                    lang_list,
                    counts,
                    sheet_prot,
                    workbook_pw.as_deref(),
                    fb.get_hide_columns(),
                    fb.get_hide_lang_sheet(),
                    &folder,
                    &version,
                ) {
                    Ok(count) => {
                        fb.set_status_type("success".into());
                        fb.set_status_message(
                            format!("{count} Datei(en) erfolgreich erstellt!").into(),
                        );
                    }
                    Err(e) => {
                        fb.set_status_type("error".into());
                        fb.set_status_message(format!("Fehler: {e}").into());
                    }
                }
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

                let src_path = std::path::Path::new(&src);
                let out_base_path = std::path::Path::new(&out_base);

                // 1. Budget-Dateien scannen
                let result = budget_scanner::scan_directory(src_path);

                // 2. Output-Ordner bestimmen
                let output_dir = budget_scanner::resolve_output_dir(out_base_path);

                // 3. ReportOptions aus Settings bauen
                let sheet_prot = if b2f.get_protect_sheet() {
                    let sp = b2f.get_sheet_permissions();
                    Some(
                        SheetProtection::new()
                            .with_password(b2f.get_sheet_password().to_string())
                            .allow_select_locked_cells(sp.select_locked)
                            .allow_select_unlocked_cells(sp.select_unlocked)
                            .allow_format_cells(sp.format_cells)
                            .allow_format_columns(sp.format_columns)
                            .allow_format_rows(sp.format_rows)
                            .allow_insert_columns(sp.insert_columns)
                            .allow_insert_rows(sp.insert_rows)
                            .allow_insert_hyperlinks(sp.insert_hyperlinks)
                            .allow_delete_columns(sp.delete_columns)
                            .allow_delete_rows(sp.delete_rows)
                            .allow_sort(sp.sort)
                            .allow_autofilter(sp.autofilter)
                            .allow_pivot_tables(sp.pivot_tables)
                            .allow_edit_objects(sp.edit_objects)
                            .allow_edit_scenarios(sp.edit_scenarios)
                            .allow_contents(sp.contents),
                    )
                } else {
                    None
                };

                let mut options_builder = ReportOptions::builder();
                if let Some(prot) = sheet_prot {
                    options_builder = options_builder.sheet_protection(prot);
                }
                if b2f.get_protect_workbook() {
                    let pw = b2f.get_workbook_password().to_string();
                    options_builder = options_builder.workbook_password(pw);
                }
                options_builder = options_builder
                    .hide_columns_qv(b2f.get_hide_columns())
                    .hide_language_sheet(b2f.get_hide_lang_sheet());
                let report_options = options_builder.build();
                let version = b2f.get_version().to_string();

                // 4. Finanzberichte generieren
                let mut generated = 0u32;
                let mut gen_errors: Vec<(String, String)> = Vec::new();

                for data in &result.successes {
                    let relative = data
                        .file_path
                        .strip_prefix(src_path)
                        .unwrap_or(&data.file_path);

                    // _FB Suffix im Dateinamen
                    let stem = relative
                        .file_stem()
                        .unwrap_or_default()
                        .to_string_lossy();
                    let fb_name = if version.is_empty() {
                        format!("{stem}_FB.xlsx")
                    } else {
                        format!("{stem}_{version}_FB.xlsx")
                    };
                    let out_path = output_dir.join(
                        relative.with_file_name(&fb_name),
                    );

                    // Verzeichnisse erstellen
                    if let Some(parent) = out_path.parent() {
                        if let Err(e) = std::fs::create_dir_all(parent) {
                            gen_errors.push((
                                data.file_path.display().to_string(),
                                format!("Ordner erstellen fehlgeschlagen: {e}"),
                            ));
                            continue;
                        }
                    }

                    let config =
                        budget_scanner::budget_to_report_config(data, report_options.clone(), &version);
                    match config.write_to(&out_path) {
                        Ok(()) => generated += 1,
                        Err(e) => gen_errors.push((
                            data.file_path.display().to_string(),
                            format!("FB-Generierung fehlgeschlagen: {e}"),
                        )),
                    }
                }

                // 5. Fehler-CSV schreiben
                if !result.failures.is_empty() {
                    let csv_path = output_dir.join("scan_fehler.csv");
                    let _ = std::fs::create_dir_all(&output_dir);
                    let _ = budget_scanner::write_failure_report(&result.failures, &csv_path);
                }

                // 6. Tabelle aktualisieren
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
                    let fname = data.file_path.file_name()
                        .map(|n| n.to_string_lossy().to_string())
                        .unwrap_or_default();
                    let status = if gen_errors.iter().any(|(p, _)| *p == data.file_path.display().to_string()) {
                        "Fehler"
                    } else {
                        "OK"
                    };
                    let detail = gen_errors.iter()
                        .find(|(p, _)| *p == data.file_path.display().to_string())
                        .map(|(_, e)| e.as_str())
                        .unwrap_or("FB erstellt");

                    rows.push(slint::ModelRc::new(slint::VecModel::from(vec![
                        slint::StandardListViewItem::from(slint::SharedString::from(&fname)),
                        slint::StandardListViewItem::from(slint::SharedString::from(status)),
                        slint::StandardListViewItem::from(slint::SharedString::from(detail)),
                    ])));
                }

                for f in &result.failures {
                    rows.push(slint::ModelRc::new(slint::VecModel::from(vec![
                        slint::StandardListViewItem::from(slint::SharedString::from(&f.file_name)),
                        slint::StandardListViewItem::from(slint::SharedString::from("Fehler")),
                        slint::StandardListViewItem::from(slint::SharedString::from(f.reason.to_string())),
                    ])));
                }

                let table_data = slint::ModelRc::new(slint::VecModel::from(rows));
                b2f.set_table_data(table_data);

                // 7. Status
                let scan_fail = result.failures.len();
                let gen_fail = gen_errors.len();
                let total = result.successes.len() + scan_fail;

                if scan_fail == 0 && gen_fail == 0 {
                    b2f.set_status_type("success".into());
                    b2f.set_status_message(
                        format!("{generated}/{total} Finanzberichte erstellt → {}", output_dir.display()).into(),
                    );
                } else {
                    b2f.set_status_type("error".into());
                    b2f.set_status_message(
                        format!(
                            "{generated} FB erstellt, {scan_fail} Scan-Fehler, {gen_fail} Generierungs-Fehler → {}",
                            output_dir.display()
                        ).into(),
                    );
                }
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
                        if c > 0 { out.push(';'); }
                        out.push_str(&columns.row_data(c).map(|col| col.title.to_string()).unwrap_or_default());
                    }
                    out.push('\n');
                    for r in 0..table_data.row_count() {
                        if let Some(row) = table_data.row_data(r) {
                            for c in 0..col_count {
                                if c > 0 { out.push(';'); }
                                out.push_str(&row.row_data(c).map(|item| item.text.to_string()).unwrap_or_default());
                            }
                            out.push('\n');
                        }
                    }
                    match std::fs::write(&path, &out) {
                        Ok(()) => {
                            b2f.set_status_type("success".into());
                            b2f.set_status_message(format!("CSV exportiert: {}", path.display()).into());
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
                        let title = columns.row_data(c).map(|col| col.title.to_string()).unwrap_or_default();
                        let _ = sheet.write_string(0, c as u16, &title);
                    }

                    // Rows
                    for r in 0..table_data.row_count() {
                        if let Some(row) = table_data.row_data(r) {
                            for c in 0..col_count {
                                let text = row.row_data(c).map(|item| item.text.to_string()).unwrap_or_default();
                                let _ = sheet.write_string((r + 1) as u32, c as u16, &text);
                            }
                        }
                    }

                    match workbook.save(&path) {
                        Ok(()) => {
                            b2f.set_status_type("success".into());
                            b2f.set_status_message(format!("Excel exportiert: {}", path.display()).into());
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
                    fs.set_status_message("Excel-Vorlage nicht gefunden (oben im Einzelordner-Bereich auswählen).".into());
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
                let mut defaults: Vec<slint::SharedString> = folder_generator::SUBFOLDERS.iter().map(|s| (*s).into()).collect();
                sort_subfolders(&mut defaults);
                ui.global::<FolderState>().set_subfolders(std::rc::Rc::new(slint::VecModel::from(defaults)).into());
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
                match folder_generator::create_project_folder(&project_name, &target, &template, &subs_refs) {
                    Ok(root) => {
                        fs.set_project_name("".into());
                        fs.set_project_name_valid(false);
                        fs.set_status_type("success".into());
                        fs.set_status_message(format!("Projektordner erstellt: {}", root.display()).into());
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

// ==========================================
// FB-Generator: Excel-Ausgabe
// ==========================================

#[allow(clippy::too_many_arguments)]
fn generate_excel(
    langs: Vec<Language>,
    counts: [u16; 8],
    sheet_prot: Option<SheetProtection>,
    workbook_pw: Option<&str>,
    hide_columns: bool,
    hide_lang_sheet: bool,
    folder: &str,
    version: &str,
) -> Result<usize, fb_generator::ReportError> {
    let folder_path = PathBuf::from(folder);
    if !folder_path.exists() {
        std::fs::create_dir_all(&folder_path)?;
    }

    // Mappenschutz-Hash vorab berechnen (~25ms Ersparnis pro Datei)
    let precomputed_hash = workbook_pw.map(fb_generator::precompute_hash);

    let mut count = 0;

    for lang in langs {
        let header = ReportHeader::builder()
            .language(lang)
            .version(version)
            .build();

        let mut body_builder = ReportBody::builder();
        for (i, &pos_count) in counts.iter().enumerate() {
            let category = (i + 1) as u8;
            if pos_count > 0 {
                let positions = (0..pos_count).map(|_| PositionEntry::builder().build());
                body_builder = body_builder.add_positions(category, positions);
            } else {
                body_builder =
                    body_builder.set_header_input(category, PositionEntry::builder().build());
            }
        }

        let mut options_builder = ReportOptions::builder();
        if let Some(ref prot) = sheet_prot {
            options_builder = options_builder.sheet_protection(prot.clone());
        }
        if let Some(pw) = workbook_pw {
            options_builder = options_builder.workbook_password(pw);
        }
        if hide_columns {
            options_builder = options_builder.hide_columns_qv(true);
        }
        if hide_lang_sheet {
            options_builder = options_builder.hide_language_sheet(true);
        }

        let config = ReportConfig::builder()
            .header(header)
            .body(body_builder.build())
            .options(options_builder.build())
            .build();

        let lang_code = match lang {
            Language::Deutsch => "de",
            Language::English => "en",
            Language::Francais => "fr",
            Language::Espanol => "es",
            Language::Portugues => "po",
        };

        let path = folder_path.join(format!("{version}_{lang_code}.xlsx"));

        if let Some(ref hash) = precomputed_hash {
            config.write_to_precomputed(&path, hash)?;
        } else {
            config.write_to(&path)?;
        }

        count += 1;
    }

    Ok(count)
}
