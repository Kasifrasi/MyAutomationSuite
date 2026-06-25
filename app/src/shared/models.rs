use serde::{Deserialize, Serialize};

pub use excel_protection::SheetProtectionOptions;

// 1. Umwandlung von Bibliothek-Optionen in Slint-UI Permissions
impl From<SheetProtectionOptions> for SheetPermissions {
    fn from(sp: SheetProtectionOptions) -> Self {
        Self {
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
            objects: sp.objects,
            scenarios: sp.scenarios,
            contents: false, // contents ist ein Slint-interner Standardwert
        }
    }
}

// 2. Umwandlung von Slint-UI Permissions zurück in Bibliothek-Optionen
impl From<SheetPermissions> for SheetProtectionOptions {
    fn from(sp: SheetPermissions) -> Self {
        Self {
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
            objects: sp.objects,
            scenarios: sp.scenarios,
        }
    }
}

#[derive(Serialize, Deserialize, Debug, Clone)]
pub struct ExportOptions {
    pub protect_sheet: bool,
    pub protect_workbook: bool,
    pub sheet_password: String,
    pub workbook_password: String,
    pub hide_columns: bool,
    pub hide_lang_sheet: bool,
    pub empty_rows: i32,
    #[serde(flatten)]
    pub protection: ExcelProtectionSettings,
}

#[derive(Deserialize, Debug)]
pub struct ProgressMessage {
    pub status: String,
    pub file: Option<String>,
    pub current: Option<u32>,
    pub total: Option<u32>,
    pub message: String,
}