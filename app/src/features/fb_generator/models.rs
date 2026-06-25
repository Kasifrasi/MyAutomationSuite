use serde::{Deserialize, Serialize};

#[derive(Serialize, Deserialize, Clone, Debug)]
pub struct FbSettings {
    pub langs: [bool; 5],
    pub categories: [i32; 8],
    pub name: String,
    pub protect_sheet: bool,
    pub protect_workbook: bool,
    pub sheet_password: String,
    pub workbook_password: String,
    pub hide_columns: bool,
    pub select_locked_cells: bool,
    pub select_unlocked_cells: bool,
    pub format_cells: bool,
    pub format_columns: bool,
    pub format_rows: bool,
    pub insert_columns: bool,
    pub insert_rows: bool,
    pub insert_hyperlinks: bool,
    pub delete_columns: bool,
    pub delete_rows: bool,
    pub sort: bool,
    pub auto_filter: bool,
    pub pivot_tables: bool,
    pub edit_objects: bool,
    pub edit_scenarios: bool,
    pub contents: bool,
    pub empty_rows: i32,
}

impl Default for FbSettings {
    fn default() -> Self {
        Self {
            langs: [true, true, true, true, true],
            categories: [20, 20, 30, 30, 20, 0, 0, 0],
            name: "Vorlage_{la}_{version}_FB.xlsx".to_string(),
            protect_sheet: true,
            protect_workbook: true,
            sheet_password: String::new(),
            workbook_password: String::new(),
            hide_columns: true,
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
            empty_rows: 3,
        }
    }
}
