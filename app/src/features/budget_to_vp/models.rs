
use serde::Serialize;

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