use std::path::Path;

use serde::{Deserialize, Serialize};
use crate::shared::models::SheetProtectionOptions;

/// Flaches Export-Modell für das FB-Go-Sidecar (`sidecars/FB`).
///
/// Der `budget_scanner` liefert mit `BudgetData` eine generische, volle
/// Repräsentation (Finanzierung inkl. Jahre/EUR, Positionen mit Jahreswerten).
/// Das Go-Sidecar erwartet jedoch ein flaches Schema mit String-Feldern
/// `eigenleistung`/`drittmittel`/`kmw_mittel` und Positionen ohne
/// Jahresaufschlüsselung. Diese Übersetzung lebt bewusst hier in der
/// FB-Feature-Schicht, damit der Scanner frei von seitenspezifischer Logik
/// bleibt.
#[derive(Serialize)]
pub struct FbExportModel<'a> {
    pub file_path: &'a Path,
    pub sheet_name: &'a str,
    pub version: &'a str,
    pub project_title: &'a str,
    pub project_number: &'a str,
    pub language: &'a str,
    pub local_currency: &'a str,
    pub cost_col1: usize,
    pub cost_col2: Option<usize>,
    pub eigenleistung: &'a str,
    pub drittmittel: &'a str,
    pub kmw_mittel: &'a str,
    pub positions: Vec<FbPosition<'a>>,
}

/// Flache Kostenposition für das FB-Sidecar (nur die vom Go-Code gelesenen Felder).
#[derive(Serialize)]
pub struct FbPosition<'a> {
    pub number: &'a str,
    pub label: &'a str,
    pub cost_col1: &'a str,
    pub cost_col2: &'a str,
}

impl<'a> FbExportModel<'a> {
    /// Bildet eine generische `BudgetData` auf das flache FB-Schema ab. Die flachen
    /// Finanzierungs-Strings werden aus dem LC-Gesamtwert (`financing.*.lc`) abgeleitet.
    pub fn from_budget(d: &'a budget_scanner::BudgetData) -> Self {
        FbExportModel {
            file_path: &d.file_path,
            sheet_name: &d.sheet_name,
            version: &d.version,
            project_title: &d.project_title,
            project_number: &d.project_number,
            language: &d.language,
            local_currency: &d.local_currency,
            cost_col1: d.cost_col1,
            cost_col2: d.cost_col2,
            eigenleistung: &d.financing.eigenleistung.lc,
            drittmittel: &d.financing.drittmittel.lc,
            kmw_mittel: &d.financing.kmw_mittel.lc,
            positions: d
                .positions
                .iter()
                .map(|p| FbPosition {
                    number: &p.number,
                    label: &p.label,
                    cost_col1: &p.cost_col1,
                    cost_col2: &p.cost_col2,
                })
                .collect(),
        }
    }
}

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