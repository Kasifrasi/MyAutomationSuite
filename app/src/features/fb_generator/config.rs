use super::models::FbSettings;
use crate::shared::models::SheetProtectionOptions;
use crate::{Categories, FBState, Languages, MainWindow, APP_NAME};

use slint::ComponentHandle;

pub fn apply_fb_defaults(ui: &MainWindow) {
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
        cat1: 30,
        cat2: 20,
        cat3: 30,
        cat4: 30,
        cat5: 20,
        cat6: 1, // Fest auf 1
        cat7: 1, // Fest auf 1
        cat8: 1, // Fest auf 1
    });

    fb.set_protect_sheet(true);
    fb.set_protect_workbook(true);
    fb.set_sheet_password("".into());
    fb.set_workbook_password("".into());
    fb.set_hide_columns(true);
    fb.set_hide_lang_sheet(true);

    fb.set_sheet_permissions(
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

    fb.set_status_type("idle".into());
    fb.set_status_message("".into());
}

pub fn load_fb_settings(ui: &MainWindow) {
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
        cat6: 1, // Fest auf 1 setzen
        cat7: 1, // Fest auf 1 setzen
        cat8: 1, // Fest auf 1 setzen
    });
    fb.set_name(s.name.into());
    fb.set_protect_sheet(s.protect_sheet);
    fb.set_protect_workbook(s.protect_workbook);
    fb.set_sheet_password(s.sheet_password.into());
    fb.set_workbook_password(s.workbook_password.into());
    fb.set_hide_columns(s.hide_columns);
    fb.set_empty_rows(0); // Fest auf 0 setzen
    fb.set_sheet_permissions(s.protection.into());
}

pub fn save_fb_settings(ui: &MainWindow) {
    let fb = ui.global::<FBState>();
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
        protection: fb.get_sheet_permissions().into(),
        empty_rows: fb.get_empty_rows(),
    };
    let _ = confy::store(APP_NAME, "fb", &s);
}
