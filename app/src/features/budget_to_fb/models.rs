use serde::{Deserialize, Serialize};
use crate::shared::models::SheetProtectionOptions;

#[derive(Serialize, Deserialize, Default)]
pub struct B2fSettings {
    pub src_folder: String,
    pub out_folder: String,
    pub name: String,
    pub protect_sheet: bool,
    pub protect_workbook: bool,
    pub sheet_password: String,
    pub workbook_password: String,
    pub hide_columns: bool,
    pub hide_lang_sheet: bool,
    pub empty_rows: i32,
    pub protection: SheetProtectionOptions,
}