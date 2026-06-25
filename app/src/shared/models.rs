use serde::{Deserialize, Serialize};

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