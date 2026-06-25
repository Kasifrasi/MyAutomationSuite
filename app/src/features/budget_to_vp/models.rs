use serde::{Deserialize, Serialize};
use crate::shared::models::SheetProtectionOptions;

// Kostenkategorien in der Reihenfolge, die der Vorpruefung-Generator (BG_CATEGORIES)
// erwartet. Die Positionsnummer "n.m" aus dem Budget bestimmt über n (1..8) die Kategorie.
pub const VP_CATEGORIES: [&str; 8] = [
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
pub struct VpBudget {
    #[serde(skip_serializing_if = "Option::is_none")]
    pub kurs: Option<f64>,
    pub eigenmittel: VpIncome,
    pub drittmittel: VpDritt,
    #[serde(rename = "kmwMittel")]
    pub kmw_mittel: VpIncome,
    pub ausgaben: Vec<VpAusgabe>,
    #[serde(rename = "reserveFreigabe")]
    pub reserve_freigabe: bool,
}

#[derive(serde::Serialize, Default)]
pub struct VpIncome {
    #[serde(skip_serializing_if = "Option::is_none")]
    pub lc: Option<f64>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub y1: Option<f64>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub y2: Option<f64>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub y3: Option<f64>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub eur: Option<f64>,
}

#[derive(serde::Serialize)]
pub struct VpDritt {
    #[serde(skip_serializing_if = "Option::is_none")]
    pub y1: Option<f64>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub y2: Option<f64>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub y3: Option<f64>,
    pub geber: Vec<VpGeber>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub sonstiges: Option<VpSonstiges>,
}

#[derive(serde::Serialize)]
pub struct VpGeber {
    pub geber: String,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub lc: Option<f64>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub eur: Option<f64>,
}

#[derive(serde::Serialize)]
pub struct VpSonstiges {
    #[serde(skip_serializing_if = "Option::is_none")]
    pub lc: Option<f64>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub eur: Option<f64>,
}

#[derive(serde::Serialize)]
pub struct VpAusgabe {
    pub kategorie: String,
    pub id: String,
    pub position: String,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub lc: Option<f64>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub y1: Option<f64>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub y2: Option<f64>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub y3: Option<f64>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub eur: Option<f64>,
}