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
    src_folder: String,
    out_folder: String,
    protect_sheet: bool,
    protect_workbook: bool,
    sheet_password: String,
    workbook_password: String,
    hide_columns: bool,
    name: String,
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
    pub empty_rows: i32,
}

impl Default for B2fSettings {
    fn default() -> Self {
        Self {
            src_folder: String::new(),
            out_folder: String::new(),
            protect_sheet: true,
            protect_workbook: true,
            sheet_password: String::new(),
            workbook_password: String::new(),
            hide_columns: true,
            name: "{pn}_{la}_{version}_FB.xlsx".to_string(),
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
            empty_rows: 3,
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
    name: String,
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
    pub empty_rows: i32,
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
            name: "Vorlage_{la}_{version}_FB.xlsx".to_string(),
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
            empty_rows: 3,
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
        src_folder: b2f.get_src_folder().to_string(),
        out_folder: b2f.get_out_folder().to_string(),
        name: b2f.get_name().to_string(),
        protect_sheet: b2f.get_protect_sheet(),
        protect_workbook: b2f.get_protect_workbook(),
        sheet_password: b2f.get_sheet_password().to_string(),
        workbook_password: b2f.get_workbook_password().to_string(),
        hide_columns: b2f.get_hide_columns(),
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
        empty_rows: b2f.get_empty_rows(),
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
    fb.set_name(s.name.into());
    fb.set_protect_sheet(s.protect_sheet);
    fb.set_protect_workbook(s.protect_workbook);
    fb.set_sheet_password(s.sheet_password.into());
    fb.set_workbook_password(s.workbook_password.into());
    fb.set_hide_columns(s.hide_columns);
    fb.set_empty_rows(s.empty_rows);
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
        name: fb.get_name().to_string(),
        protect_sheet: fb.get_protect_sheet(),
        protect_workbook: fb.get_protect_workbook(),
        sheet_password: fb.get_sheet_password().to_string(),
        workbook_password: fb.get_workbook_password().to_string(),
        hide_columns: fb.get_hide_columns(),
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
        empty_rows: fb.get_empty_rows(),
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

    fb.set_name("Vorlage_{la}_{version}_FB.xlsx".into());
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

// Budget-zu-Prüfvorlage (Vorpruefung): Einstellungen

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

fn load_vp_settings(ui: &MainWindow) {
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

fn save_vp_settings(ui: &MainWindow) {
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

fn apply_vp_defaults(ui: &MainWindow) {
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
    pub empty_rows: i32,
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
    let sidecar_bytes = include_bytes!("../../sidecars/FB/generator.exe");

    // Wir entpacken sie in den Temp-Ordner
    let dir = std::env::temp_dir().join("MyAutomationSuite");
    let _ = std::fs::create_dir_all(&dir);

    let exe_name = if cfg!(windows) {
        "generator.exe"
    } else {
        "generator"
    };
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

fn get_vorpruefung_path() -> std::path::PathBuf {
    // Vorpruefung-Sidecar (Go) wird wie der FB-Generator direkt eingebettet.
    // Vor `cargo build` muss er via `build-go` als sidecars/Vorpruefung/vorpruefung.exe
    // erzeugt worden sein.
    let sidecar_bytes = include_bytes!("../../sidecars/Vorpruefung/vorpruefung.exe");

    let dir = std::env::temp_dir().join("MyAutomationSuite");
    let _ = std::fs::create_dir_all(&dir);

    let exe_name = if cfg!(windows) {
        "vorpruefung.exe"
    } else {
        "vorpruefung"
    };
    let exe_path = dir.join(exe_name);

    let needs_write = match std::fs::metadata(&exe_path) {
        Ok(meta) => meta.len() as usize != sidecar_bytes.len(),
        Err(_) => true,
    };

    if needs_write {
        let _ = std::fs::write(&exe_path, sidecar_bytes);
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

// ==========================================================================
// Budget → Vorpruefung-Prüfvorlage: Mapping & Hilfsfunktionen
// ==========================================================================

// Kostenkategorien in der Reihenfolge, die der Vorpruefung-Generator (BG_CATEGORIES)
// erwartet. Die Positionsnummer "n.m" aus dem Budget bestimmt über n (1..8) die Kategorie.
const VP_CATEGORIES: [&str; 8] = [
    "Bauausgaben",
    "Investitionen",
    "Personalkosten",
    "Projektaktivitaeten",
    "Projektverwaltung",
    "Evaluierung",
    "Audit",
    "Reserve",
];

// Die folgenden Structs spiegeln exakt das BudgetConfig-Schema von
// sidecars/Vorpruefung/config.go (das den Decoder mit DisallowUnknownFields nutzt).
#[derive(serde::Serialize)]
struct VpBudget {
    #[serde(skip_serializing_if = "Option::is_none")]
    kurs: Option<f64>,
    eigenmittel: VpIncome,
    drittmittel: VpDritt,
    #[serde(rename = "kmwMittel")]
    kmw_mittel: VpIncome,
    ausgaben: Vec<VpAusgabe>,
    #[serde(rename = "reserveFreigabe")]
    reserve_freigabe: bool,
}

#[derive(serde::Serialize, Default)]
struct VpIncome {
    #[serde(skip_serializing_if = "Option::is_none")]
    lc: Option<f64>,
    #[serde(skip_serializing_if = "Option::is_none")]
    y1: Option<f64>,
    #[serde(skip_serializing_if = "Option::is_none")]
    y2: Option<f64>,
    #[serde(skip_serializing_if = "Option::is_none")]
    y3: Option<f64>,
    #[serde(skip_serializing_if = "Option::is_none")]
    eur: Option<f64>,
}

#[derive(serde::Serialize)]
struct VpDritt {
    #[serde(skip_serializing_if = "Option::is_none")]
    y1: Option<f64>,
    #[serde(skip_serializing_if = "Option::is_none")]
    y2: Option<f64>,
    #[serde(skip_serializing_if = "Option::is_none")]
    y3: Option<f64>,
    geber: Vec<VpGeber>,
    #[serde(skip_serializing_if = "Option::is_none")]
    sonstiges: Option<VpSonstiges>,
}

#[derive(serde::Serialize)]
struct VpGeber {
    geber: String,
    #[serde(skip_serializing_if = "Option::is_none")]
    lc: Option<f64>,
    #[serde(skip_serializing_if = "Option::is_none")]
    eur: Option<f64>,
}

#[derive(serde::Serialize)]
struct VpSonstiges {
    #[serde(skip_serializing_if = "Option::is_none")]
    lc: Option<f64>,
    #[serde(skip_serializing_if = "Option::is_none")]
    eur: Option<f64>,
}

#[derive(serde::Serialize)]
struct VpAusgabe {
    kategorie: String,
    id: String,
    position: String,
    #[serde(skip_serializing_if = "Option::is_none")]
    lc: Option<f64>,
    #[serde(skip_serializing_if = "Option::is_none")]
    y1: Option<f64>,
    #[serde(skip_serializing_if = "Option::is_none")]
    y2: Option<f64>,
    #[serde(skip_serializing_if = "Option::is_none")]
    y3: Option<f64>,
    #[serde(skip_serializing_if = "Option::is_none")]
    eur: Option<f64>,
}

/// Parst einen Geldbetrag aus dem Budget (z.B. "10,000", "5,000.00 €", "1.234,56").
/// Erkennt Tausender-/Dezimaltrenner heuristisch. Leer/unparsbar ⇒ None (leeres
/// Eingabefeld in der Prüfvorlage).
fn parse_amount(raw: &str) -> Option<f64> {
    let mut s: String = raw
        .chars()
        .filter(|c| c.is_ascii_digit() || *c == ',' || *c == '.' || *c == '-')
        .collect();
    if s.is_empty() || s == "-" {
        return None;
    }

    let has_dot = s.contains('.');
    let has_comma = s.contains(',');

    if has_dot && has_comma {
        // Der zuletzt auftretende Trenner ist der Dezimaltrenner.
        let last_dot = s.rfind('.').unwrap();
        let last_comma = s.rfind(',').unwrap();
        if last_comma > last_dot {
            s = s.replace('.', "").replace(',', ".");
        } else {
            s = s.replace(',', "");
        }
    } else if has_comma {
        let after = s.rsplit(',').next().map(|p| p.len()).unwrap_or(0);
        let commas = s.matches(',').count();
        if commas == 1 && after != 3 {
            // Einzelnes Komma, nicht im Tausenderformat ⇒ Dezimaltrenner.
            s = s.replace(',', ".");
        } else {
            s = s.replace(',', "");
        }
    }

    s.parse::<f64>().ok()
}

fn vp_income_from(row: &budget_scanner::FinancingRow) -> VpIncome {
    VpIncome {
        lc: parse_amount(&row.lc),
        y1: parse_amount(&row.year1),
        y2: parse_amount(&row.year2),
        y3: parse_amount(&row.year3),
        eur: parse_amount(&row.eur),
    }
}

/// Bildet die gescannten Budgetdaten auf das Vorpruefung-Budget-Schema ab.
/// Nicht spezifizierte Drittmittelgeber landen gesammelt unter "Sonstige".
fn budget_to_vp_budget(d: &budget_scanner::BudgetData) -> VpBudget {
    let mut ausgaben = Vec::with_capacity(d.positions.len());
    for p in &d.positions {
        let mut parts = p.number.splitn(2, '.');
        let cat = parts.next().unwrap_or("").trim();
        let sub = parts.next().unwrap_or("").trim();
        let idx = match cat.parse::<usize>() {
            Ok(n) if (1..=8).contains(&n) => n - 1,
            _ => continue,
        };

        let lc = parse_amount(&p.cost_col1);
        let y1 = parse_amount(&p.cost_year1);
        let y2 = parse_amount(&p.cost_year2);
        let y3 = parse_amount(&p.cost_year3);
        let eur = parse_amount(&p.cost_col2);

        // Wertlose Positionen (kein/0-Wert in allen Spalten) weglassen, wenn sie entweder
        // eine reine Kategorie-Kopfzeile sind (keine Unterposition, label = Kategoriename)
        // ODER ein namenloser Platzhalter. Ihre ID wird gar nicht erst an das Prüftool
        // übergeben. Da die ID direkt aus dem Budget übernommen wird (id = p.number),
        // behalten die übrigen Positionen ihre Original-Nummern – es entsteht höchstens eine
        // Lücke (z. B. 1.1, 1.3), aber keine Verschiebung. Die Sonderkategorien
        // (Evaluierung/Audit/Reserve) tragen ihren Wert direkt auf der Kategoriezeile und
        // bleiben dadurch erhalten; benannte Positionen mit 0 ebenfalls (bewusste Eingabe).
        let is_zero = |v: Option<f64>| v.is_none_or(|x| x == 0.0);
        let valueless = is_zero(lc) && is_zero(y1) && is_zero(y2) && is_zero(y3) && is_zero(eur);
        let label_empty = p.label.trim().is_empty();
        let is_header = sub.is_empty();
        if valueless && (is_header || label_empty) {
            continue;
        }

        let position = if label_empty {
            VP_CATEGORIES[idx].to_string()
        } else {
            p.label.clone()
        };

        ausgaben.push(VpAusgabe {
            kategorie: VP_CATEGORIES[idx].to_string(),
            id: p.number.clone(),
            position,
            lc,
            y1,
            y2,
            y3,
            eur,
        });
    }

    let fin = &d.financing;
    VpBudget {
        kurs: None,
        eigenmittel: vp_income_from(&fin.eigenleistung),
        drittmittel: VpDritt {
            y1: parse_amount(&fin.drittmittel.year1),
            y2: parse_amount(&fin.drittmittel.year2),
            y3: parse_amount(&fin.drittmittel.year3),
            geber: Vec::new(),
            sonstiges: Some(VpSonstiges {
                lc: parse_amount(&fin.drittmittel.lc),
                eur: parse_amount(&fin.drittmittel.eur),
            }),
        },
        kmw_mittel: vp_income_from(&fin.kmw_mittel),
        ausgaben,
        reserve_freigabe: false,
    }
}

/// Ersetzt Platzhalter im Dateinamen und erzwingt Eindeutigkeit (.xlsx).
/// {pn}=Projektnummer, {la}=Sprache, {version}=Version, {i}=Duplikate-Zähler.
fn vp_output_name(
    pattern: &str,
    data: &budget_scanner::BudgetData,
    used: &mut std::collections::HashSet<String>,
) -> String {
    let sanitize = |s: &str| -> String {
        s.chars()
            .map(|c| if "\\/:*?\"<>|".contains(c) { '_' } else { c })
            .collect()
    };

    let mut name = pattern
        .replace("{pn}", &sanitize(&data.project_number))
        .replace("{la}", &sanitize(&data.language))
        .replace("{version}", &sanitize(&data.version));
    if !name.to_lowercase().ends_with(".xlsx") {
        name.push_str(".xlsx");
    }

    let (stem, ext) = match name.to_lowercase().rfind(".xlsx") {
        Some(pos) => (name[..pos].to_string(), name[pos..].to_string()),
        None => (name.clone(), String::new()),
    };
    let has_counter = stem.contains("{i}");

    let mut n = 1u32;
    loop {
        let candidate = if has_counter {
            format!("{}{}", stem.replace("{i}", &n.to_string()), ext)
        } else if n == 1 {
            format!("{stem}{ext}")
        } else {
            format!("{stem}_{n}{ext}")
        };
        if used.insert(candidate.clone()) {
            return candidate;
        }
        n += 1;
    }
}

fn main() -> Result<(), slint::PlatformError> {
    let ui = MainWindow::new()?;

    // Defaults setzen, dann gespeicherte Settings laden
    apply_fb_defaults(&ui);
    load_fb_settings(&ui);
    apply_b2f_defaults(&ui);
    load_b2f_settings(&ui);
    apply_vp_defaults(&ui);
    load_vp_settings(&ui);
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

                let name = fb.get_name().to_string();
                if name.is_empty() {
                    fb.set_status_type("error".into());
                    fb.set_status_message("Bitte Dateinamens-Muster angeben.".into());
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

                let name_clone = name.clone();

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
                    empty_rows: fb.get_empty_rows(),
                };
                let wb_hash = if options.protect_workbook {
                    Some(excel_protection::precompute_hash(
                        &options.workbook_password,
                    ))
                } else {
                    None
                };

                let sh_hash = if options.protect_sheet {
                    Some(excel_protection::precompute_hash(&options.sheet_password))
                } else {
                    None
                };

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
                } else {
                    None
                };

                let mut sidecar_options = options.clone();
                sidecar_options.protect_sheet = false;
                sidecar_options.protect_workbook = false;
                let options_json = serde_json::to_string(&sidecar_options).unwrap_or_default();

                let ui_handle_clone = ui_handle.clone();
                std::thread::spawn(move || {
                    let start_time = std::time::Instant::now();
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
                                    cost_year1: String::new(),
                                    cost_year2: String::new(),
                                    cost_year3: String::new(),
                                });
                            }
                        }

                        templates.push(budget_scanner::BudgetData {
                            file_path: std::path::PathBuf::from(format!("Vorlage_{lang}.xlsx")),
                            sheet_name: "Budget".into(),
                            version: String::new(),
                            project_title: "".into(),
                            project_number: "Vorlage".into(),
                            language: lang.to_string(),
                            local_currency: "".into(),
                            cost_col1: 8,
                            cost_col2: Some(13),
                            eigenleistung: "0".into(),
                            drittmittel: "0".into(),
                            kmw_mittel: "0".into(),
                            financing: budget_scanner::FinancingDetail::default(),
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

                    // 4. Go Sidecar aufrufen
                    let sidecar_exe = get_sidecar_path();

                    let mut cmd = std::process::Command::new(&sidecar_exe);

                    #[cfg(target_os = "windows")]
                    {
                        use std::os::windows::process::CommandExt;
                        const CREATE_NO_WINDOW: u32 = 0x08000000;
                        cmd.creation_flags(CREATE_NO_WINDOW);
                    }

                    cmd.arg("-input")
                        .arg(&tmp_json_path)
                        .arg("-output")
                        .arg(&output_dir)
                        .arg("-options")
                        .arg(&options_json)
                        .arg("-filename")
                        .arg(&name_clone);

                    cmd.stdout(std::process::Stdio::piped());

                    let mut child = match cmd.spawn() {
                        Ok(c) => c,
                        Err(e) => {
                            let _ = ui_handle_clone.upgrade_in_event_loop(move |ui| {
                                let fb = ui.global::<FBState>();
                                fb.set_status_type("error".into());
                                fb.set_status_message(
                                    format!(
                                        "Fehler beim Starten von {}: {e}",
                                        sidecar_exe.display()
                                    )
                                    .into(),
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
                            fb.set_status_message("Wende Schutz an...".into());
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
                                        sh_opts.as_ref(),
                                    );
                                }
                            });
                        }
                    }

                    let success_count = templates.len();
                    let _ = ui_handle_clone.upgrade_in_event_loop(move |ui| {
                        let fb = ui.global::<FBState>();
                        let elapsed_sec = start_time.elapsed().as_secs_f64();
                        fb.set_status_type("success".into());
                        fb.set_status_message(
                            format!(
                                "Erfolgreich abgeschlossen! {} Datei(en) in {:.2}s erstellt.",
                                success_count, elapsed_sec
                            )
                            .into(),
                        );
                    });
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
                    save_b2f_settings(&ui);
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
                    save_b2f_settings(&ui);
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

                let name = b2f.get_name().to_string();
                let name_clone = name.clone();

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
                    empty_rows: b2f.get_empty_rows(),
                };

                let wb_hash = if options.protect_workbook {
                    Some(excel_protection::precompute_hash(
                        &options.workbook_password,
                    ))
                } else {
                    None
                };

                let sh_hash = if options.protect_sheet {
                    Some(excel_protection::precompute_hash(&options.sheet_password))
                } else {
                    None
                };

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
                } else {
                    None
                };

                // Dem Sidecar geben wir protect=false mit, damit es das XML nicht verschlüsselt
                let mut sidecar_options = options.clone();
                sidecar_options.protect_sheet = false;
                sidecar_options.protect_workbook = false;
                let options_json = serde_json::to_string(&sidecar_options).unwrap_or_default();

                let ui_handle_clone = ui_handle.clone();
                std::thread::spawn(move || {
                    let start_time = std::time::Instant::now();
                    let src_path = std::path::PathBuf::from(&src);
                    let out_base_path = std::path::PathBuf::from(&out_base);

                    // 1. Budget-Dateien scannen
                    let result = budget_scanner::scan_directory(&src_path);

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

                    #[cfg(target_os = "windows")]
                    {
                        use std::os::windows::process::CommandExt;
                        const CREATE_NO_WINDOW: u32 = 0x08000000;
                        cmd.creation_flags(CREATE_NO_WINDOW);
                    }

                    cmd.arg("-input")
                        .arg(&tmp_json_path)
                        .arg("-output")
                        .arg(&output_dir)
                        .arg("-options")
                        .arg(&options_json)
                        .arg("-filename")
                        .arg(&name_clone);

                    cmd.stdout(std::process::Stdio::piped());

                    let mut child = match cmd.spawn() {
                        Ok(c) => c,
                        Err(e) => {
                            let _ = ui_handle_clone.upgrade_in_event_loop(move |ui| {
                                let b2f = ui.global::<BudgetState>();
                                b2f.set_status_type("error".into());
                                b2f.set_status_message(
                                    format!(
                                        "Fehler beim Starten von {}: {e}",
                                        sidecar_exe.display()
                                    )
                                    .into(),
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
                            b2f.set_status_message("Wende Schutz an...".into());
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
                                        sh_opts.as_ref(),
                                    );
                                }
                            });
                        }
                    }

                    // 6. Fehler-CSV schreiben
                    if !result.failures.is_empty() {
                        let csv_path = output_dir.join("scan_fehler.csv");
                        let _ = budget_scanner::write_failure_report(&result.failures, &csv_path);
                    }

                    // 7. Tabelle aktualisieren
                    let success_count = result.successes.len();
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

                        let elapsed_sec = start_time.elapsed().as_secs_f64();
                        b2f.set_status_type("success".into());
                        b2f.set_status_message(
                            format!(
                                "Erfolgreich abgeschlossen! {} Datei(en) in {:.2}s erstellt.",
                                success_count, elapsed_sec
                            )
                            .into(),
                        );
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
    // Budget-zu-Prüfvorlage (Vorpruefung) Callbacks
    // ==========================================

    ui.global::<VorpruefungState>().on_select_src({
        let ui_handle = ui.as_weak();
        move || {
            if let Some(ui) = ui_handle.upgrade() {
                if let Some(path) = rfd::FileDialog::new().pick_folder() {
                    ui.global::<VorpruefungState>()
                        .set_src_folder(path.to_string_lossy().to_string().into());
                    save_vp_settings(&ui);
                }
            }
        }
    });

    ui.global::<VorpruefungState>().on_select_out({
        let ui_handle = ui.as_weak();
        move || {
            if let Some(ui) = ui_handle.upgrade() {
                if let Some(path) = rfd::FileDialog::new().pick_folder() {
                    ui.global::<VorpruefungState>()
                        .set_out_folder(path.to_string_lossy().to_string().into());
                    save_vp_settings(&ui);
                }
            }
        }
    });

    ui.global::<VorpruefungState>().on_dismiss_status({
        let ui_handle = ui.as_weak();
        move || {
            if let Some(ui) = ui_handle.upgrade() {
                let vp = ui.global::<VorpruefungState>();
                vp.set_status_type("idle".into());
                vp.set_status_message("".into());
            }
        }
    });

    ui.global::<VorpruefungState>().on_toggle_settings({
        let ui_handle = ui.as_weak();
        move || {
            if let Some(ui) = ui_handle.upgrade() {
                let vp = ui.global::<VorpruefungState>();
                vp.set_show_settings(!vp.get_show_settings());
            }
        }
    });

    ui.global::<VorpruefungState>().on_save_settings({
        let ui_handle = ui.as_weak();
        move || {
            if let Some(ui) = ui_handle.upgrade() {
                save_vp_settings(&ui);
            }
        }
    });

    ui.global::<VorpruefungState>().on_do_reset({
        let ui_handle = ui.as_weak();
        move || {
            if let Some(ui) = ui_handle.upgrade() {
                apply_vp_defaults(&ui);
                save_vp_settings(&ui);
            }
        }
    });

    ui.global::<VorpruefungState>().on_generate({
        let ui_handle = ui.as_weak();
        move || {
            if let Some(ui) = ui_handle.upgrade() {
                let vp = ui.global::<VorpruefungState>();

                let src = vp.get_src_folder().to_string();
                let out_base = vp.get_out_folder().to_string();
                let name = vp.get_name().to_string();

                if src.is_empty() {
                    vp.set_status_type("error".into());
                    vp.set_status_message("Bitte Budget-Ordner wählen.".into());
                    return;
                }
                if out_base.is_empty() {
                    vp.set_status_type("error".into());
                    vp.set_status_message("Bitte Ausgabeordner wählen.".into());
                    return;
                }
                if name.is_empty() {
                    vp.set_status_type("error".into());
                    vp.set_status_message("Bitte Dateinamens-Muster angeben.".into());
                    return;
                }

                vp.set_status_type("pending".into());
                vp.set_status_message("Scannt Budgets...".into());

                let protect_workbook = vp.get_protect_workbook();
                let wb_password = vp.get_workbook_password().to_string();

                let ui_handle_clone = ui_handle.clone();
                std::thread::spawn(move || {
                    let start_time = std::time::Instant::now();
                    let src_path = std::path::PathBuf::from(&src);
                    let out_base_path = std::path::PathBuf::from(&out_base);

                    // 1. Budgets scannen
                    let result = budget_scanner::scan_directory(&src_path);

                    // 2. Output-Ordner
                    let output_dir = budget_scanner::resolve_output_dir(&out_base_path);
                    let _ = std::fs::create_dir_all(&output_dir);

                    let wb_hash = if protect_workbook {
                        Some(excel_protection::precompute_hash(&wb_password))
                    } else {
                        None
                    };

                    let sidecar_exe = get_vorpruefung_path();
                    let total = result.successes.len() as u32;

                    // (Quelldateiname, Status, Detail)
                    let mut rows_info: Vec<(String, String, String)> = Vec::new();
                    let mut used_names: std::collections::HashSet<String> =
                        std::collections::HashSet::new();
                    let mut ok_count = 0u32;

                    for (i, data) in result.successes.iter().enumerate() {
                        let current = (i + 1) as u32;
                        let src_name = data
                            .file_path
                            .file_name()
                            .map(|n| n.to_string_lossy().to_string())
                            .unwrap_or_default();

                        let _ = ui_handle_clone.upgrade_in_event_loop({
                            let label = src_name.clone();
                            move |ui| {
                                let vp = ui.global::<VorpruefungState>();
                                vp.set_status_type("pending".into());
                                vp.set_status_message(
                                    format!("{current}/{total} – {label}").into(),
                                );
                            }
                        });

                        // 3. Budget → Vorpruefung-JSON
                        let vp_budget = budget_to_vp_budget(data);
                        let json = match serde_json::to_string(&vp_budget) {
                            Ok(j) => j,
                            Err(e) => {
                                rows_info.push((
                                    src_name,
                                    "Fehler".into(),
                                    format!("JSON-Fehler: {e}"),
                                ));
                                continue;
                            }
                        };

                        let tmp_json_path = std::env::temp_dir().join(format!(
                            "vp_budget_{}_{}.json",
                            std::time::SystemTime::now()
                                .duration_since(std::time::UNIX_EPOCH)
                                .map(|d| d.as_millis())
                                .unwrap_or(0),
                            i
                        ));
                        if let Err(e) = std::fs::write(&tmp_json_path, json.as_bytes()) {
                            rows_info.push((src_name, "Fehler".into(), format!("Temp-JSON: {e}")));
                            continue;
                        }

                        // 4. Zieldateiname + Sidecar-Aufruf
                        let out_name = vp_output_name(&name, data, &mut used_names);
                        let out_path = output_dir.join(&out_name);

                        let mut cmd = std::process::Command::new(&sidecar_exe);
                        #[cfg(target_os = "windows")]
                        {
                            use std::os::windows::process::CommandExt;
                            const CREATE_NO_WINDOW: u32 = 0x08000000;
                            cmd.creation_flags(CREATE_NO_WINDOW);
                        }
                        cmd.arg("-budget")
                            .arg(&tmp_json_path)
                            .arg("-o")
                            .arg(&out_path);

                        let run = cmd.output();
                        let _ = std::fs::remove_file(&tmp_json_path);

                        match run {
                            Ok(o) if o.status.success() => {
                                // 5. Optionaler Mappenschutz (sperrt keine Eingabezellen)
                                if let Some(h) = wb_hash.as_ref() {
                                    let _ = excel_protection::apply_protection_in_place(
                                        &out_path,
                                        Some(h),
                                        None,
                                        None,
                                    );
                                }
                                ok_count += 1;
                                rows_info.push((src_name, "OK".into(), out_name));
                            }
                            Ok(o) => {
                                let err = String::from_utf8_lossy(&o.stderr);
                                let detail = err
                                    .lines()
                                    .last()
                                    .map(|s| s.to_string())
                                    .filter(|s| !s.is_empty())
                                    .unwrap_or_else(|| "Generierung fehlgeschlagen".into());
                                rows_info.push((src_name, "Fehler".into(), detail));
                            }
                            Err(e) => {
                                rows_info.push((
                                    src_name,
                                    "Fehler".into(),
                                    format!("Sidecar-Start: {e}"),
                                ));
                            }
                        }
                    }

                    // 6. Scan-Fehler ergänzen + CSV
                    for f in &result.failures {
                        rows_info.push((
                            f.file_name.clone(),
                            "Fehler".into(),
                            f.reason.to_string(),
                        ));
                    }
                    if !result.failures.is_empty() {
                        let csv_path = output_dir.join("scan_fehler.csv");
                        let _ = budget_scanner::write_failure_report(&result.failures, &csv_path);
                    }

                    // 7. Tabelle + Status aktualisieren
                    let _ = ui_handle_clone.upgrade_in_event_loop(move |ui| {
                        let vp = ui.global::<VorpruefungState>();

                        let mk_col = |t: &str| {
                            let mut c = slint::TableColumn::default();
                            c.title = t.into();
                            c
                        };
                        vp.set_table_columns(slint::ModelRc::new(slint::VecModel::from(vec![
                            mk_col("Budget-Datei"),
                            mk_col("Status"),
                            mk_col("Details"),
                        ])));

                        let rows: Vec<slint::ModelRc<slint::StandardListViewItem>> = rows_info
                            .iter()
                            .map(|(file, status, detail)| {
                                slint::ModelRc::new(slint::VecModel::from(vec![
                                    slint::StandardListViewItem::from(slint::SharedString::from(
                                        file.as_str(),
                                    )),
                                    slint::StandardListViewItem::from(slint::SharedString::from(
                                        status.as_str(),
                                    )),
                                    slint::StandardListViewItem::from(slint::SharedString::from(
                                        detail.as_str(),
                                    )),
                                ]))
                            })
                            .collect();
                        vp.set_table_data(slint::ModelRc::new(slint::VecModel::from(rows)));

                        let elapsed = start_time.elapsed().as_secs_f64();
                        let fail_count = result.failures.len() as u32 + (total - ok_count);
                        if ok_count == 0 {
                            vp.set_status_type("error".into());
                            vp.set_status_message(
                                format!("Keine Prüfvorlage erstellt ({fail_count} Fehler).").into(),
                            );
                        } else {
                            vp.set_status_type("success".into());
                            vp.set_status_message(
                                format!(
                                    "{ok_count} Prüfvorlage(n) in {elapsed:.2}s erstellt{}.",
                                    if fail_count > 0 {
                                        format!(", {fail_count} Fehler")
                                    } else {
                                        String::new()
                                    }
                                )
                                .into(),
                            );
                        }
                    });
                });
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

#[cfg(test)]
mod vp_tests {
    use super::*;

    fn pos(number: &str, lc: &str, y1: &str, eur: &str) -> budget_scanner::BudgetPosition {
        posn(number, "", lc, y1, eur)
    }

    fn posn(
        number: &str,
        label: &str,
        lc: &str,
        y1: &str,
        eur: &str,
    ) -> budget_scanner::BudgetPosition {
        budget_scanner::BudgetPosition {
            number: number.into(),
            label: label.into(),
            cost_col1: lc.into(),
            cost_col2: eur.into(),
            cost_year1: y1.into(),
            cost_year2: String::new(),
            cost_year3: String::new(),
        }
    }

    fn data_with(positions: Vec<budget_scanner::BudgetPosition>) -> budget_scanner::BudgetData {
        budget_scanner::BudgetData {
            file_path: std::path::PathBuf::from("test.xlsx"),
            sheet_name: "Budget".into(),
            version: "V2".into(),
            project_title: String::new(),
            project_number: "P1".into(),
            language: "deutsch".into(),
            local_currency: "USD".into(),
            cost_col1: 8,
            cost_col2: Some(13),
            eigenleistung: String::new(),
            drittmittel: String::new(),
            kmw_mittel: String::new(),
            financing: budget_scanner::FinancingDetail::default(),
            positions,
        }
    }

    #[test]
    fn parse_amount_formats() {
        assert_eq!(parse_amount("10,000"), Some(10000.0));
        assert_eq!(parse_amount("5,000.00 €"), Some(5000.0));
        assert_eq!(parse_amount("1.234,56"), Some(1234.56));
        assert_eq!(parse_amount("0"), Some(0.0));
        assert_eq!(parse_amount(""), None);
        assert_eq!(parse_amount("1589000"), Some(1589000.0));
    }

    #[test]
    fn skips_empty_category_header_but_keeps_special_categories() {
        let d = data_with(vec![
            // Kopfzeile trägt (wie im echten Scanner) den Kategorienamen als Label, hat
            // aber keine Werte -> muss trotzdem gefiltert werden.
            posn("1.", "Bauausgaben", "", "", ""),
            pos("1.1", "10000", "10000", "5000"), // normale Position
            posn("6.", "Evaluierung", "10000", "10000", "5000"), // Wert auf Kategoriezeile
            posn("7.", "Audit", "10000", "10000", "5000"),       // Audit
            posn("8.", "Reserve", "79000", "79000", "39500"),    // Reserve
        ]);
        let vp = budget_to_vp_budget(&d);
        let ids: Vec<&str> = vp.ausgaben.iter().map(|a| a.id.as_str()).collect();
        assert_eq!(ids, vec!["1.1", "6.", "7.", "8."], "Kopfzeile raus, Sonderkat. drin");

        let eval = vp.ausgaben.iter().find(|a| a.id == "6.").unwrap();
        assert_eq!(eval.kategorie, "Evaluierung");
        assert_eq!(eval.position, "Evaluierung"); // leeres Label -> Kategoriename
        assert_eq!(eval.lc, Some(10000.0));
        let reserve = vp.ausgaben.iter().find(|a| a.id == "8.").unwrap();
        assert_eq!(reserve.kategorie, "Reserve");
        assert_eq!(reserve.lc, Some(79000.0));
    }

    #[test]
    fn filters_empty_placeholders_without_renumbering() {
        let d = data_with(vec![
            pos("1.", "", "", ""),                  // Kopfzeile -> raus
            pos("1.1", "10000", "10000", "5000"),   // Wert -> bleibt
            pos("1.2", "0", "", "0"),               // leer + 0 -> raus (ID 1.2 entfällt)
            pos("1.3", "11000", "11000", "5500"),   // Wert -> bleibt
            posn("1.4", "Büromaterial", "0", "", "0"), // Name vorhanden -> bleibt (auch bei 0)
        ]);
        let vp = budget_to_vp_budget(&d);
        let ids: Vec<&str> = vp.ausgaben.iter().map(|a| a.id.as_str()).collect();
        // 1.2 fehlt (Lücke), aber 1.3 behält seine Original-ID – keine Verschiebung.
        assert_eq!(ids, vec!["1.1", "1.3", "1.4"]);
        let bm = vp.ausgaben.iter().find(|a| a.id == "1.4").unwrap();
        assert_eq!(bm.position, "Büromaterial");
        assert_eq!(bm.lc, Some(0.0));
    }

    // Voller Pfad mit echtem Budget (nur wenn Testdatei vorhanden, sonst übersprungen).
    #[test]
    fn maps_real_budget_and_writes_json() {
        let budget = std::path::Path::new(env!("CARGO_MANIFEST_DIR"))
            .join("../sidecars/FB/data/budgets/de.xlsx");
        if !budget.exists() {
            eprintln!("de.xlsx fehlt – Test übersprungen");
            return;
        }
        let d = budget_scanner::scan_file(&budget).expect("scan ok");
        let vp = budget_to_vp_budget(&d);

        // Sonderkategorien müssen jetzt enthalten sein
        for cat in ["Evaluierung", "Audit", "Reserve"] {
            assert!(
                vp.ausgaben.iter().any(|a| a.kategorie == cat),
                "Kategorie {cat} fehlt im Mapping"
            );
        }
        let json = serde_json::to_string_pretty(&vp).unwrap();
        std::fs::write("/tmp/vp_real.json", json).unwrap();
        eprintln!(
            "vp_real.json geschrieben: {} Ausgaben",
            vp.ausgaben.len()
        );
    }
}
