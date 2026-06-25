use crate::{MainWindow, FBState, Languages, Categories, SheetPermissions, APP_NAME};
use super::models::FbSettings;

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
        contents: false,
    });

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
        s.select_locked_cells,
        s.select_unlocked_cells,
        s.format_cells,
        s.format_columns,
        s.format_rows,
        s.insert_columns,
        s.insert_rows,
        s.insert_hyperlinks,
        s.delete_columns,
        s.delete_rows,
        s.sort,
        s.auto_filter,
        s.pivot_tables,
        s.edit_objects,
        s.edit_scenarios,
        s.contents,
    ));
}

pub fn save_fb_settings(ui: &MainWindow) {
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
        select_locked_cells: sp.select_locked_cells,
        select_unlocked_cells: sp.select_unlocked_cells,
        format_cells: sp.format_cells,
        format_columns: sp.format_columns,
        format_rows: sp.format_rows,
        insert_columns: sp.insert_columns,
        insert_rows: sp.insert_rows,
        insert_hyperlinks: sp.insert_hyperlinks,
        delete_columns: sp.delete_columns,
        delete_rows: sp.delete_rows,
        sort: sp.sort,
        auto_filter: sp.auto_filter,
        pivot_tables: sp.pivot_tables,
        edit_objects: sp.edit_objects,
        edit_scenarios: sp.edit_scenarios,
        contents: sp.contents,
        empty_rows: fb.get_empty_rows(),
    };
    let _ = confy::store(APP_NAME, "fb", &s);
}

pub fn permissions_from_settings(
    select_locked_cells: bool,
    select_unlocked_cells: bool,
    format_cells: bool,
    format_columns: bool,
    format_rows: bool,
    insert_columns: bool,
    insert_rows: bool,
    insert_hyperlinks: bool,
    delete_columns: bool,
    delete_rows: bool,
    sort: bool,
    auto_filter: bool,
    pivot_tables: bool,
    edit_objects: bool,
    edit_scenarios: bool,
    contents: bool,
) -> SheetPermissions {
    SheetPermissions {
        select_locked_cells,
        select_unlocked_cells,
        format_cells,
        format_columns,
        format_rows,
        insert_columns,
        insert_rows,
        insert_hyperlinks,
        delete_columns,
        delete_rows,
        sort,
        auto_filter,
        pivot_tables,
        edit_objects,
        edit_scenarios,
        contents,
    }
}