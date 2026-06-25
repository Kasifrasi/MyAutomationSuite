use super::models::{VpBudget, VpIncome, VpAusgabe, VpDritt, VpSonstiges, VP_CATEGORIES};
use crate::shared::utils::parse_amount;


pub fn vp_income_from(row: &budget_scanner::FinancingRow) -> VpIncome {
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
pub fn budget_to_vp_budget(d: &budget_scanner::BudgetData) -> VpBudget {
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

pub fn vp_output_name(
    pattern: &str,
    data: &budget_scanner::BudgetData,
    used: &mut std::collections::HashSet<String>,
) -> String {
    let mut replacements = std::collections::HashMap::new();

    // Originalen Dateinamen ohne Endung ermitteln (z.B. "de" statt "de.xlsx")
    let file_stem = data.file_path
        .file_stem()
        .map(|s| s.to_string_lossy().to_string())
        .unwrap_or_default();

    // Platzhalter-Ersetzungen definieren
    replacements.insert("{pn}".to_string(), data.project_number.clone());
    replacements.insert("{la}".to_string(), data.language.clone());
    replacements.insert("{pt}".to_string(), data.project_title.clone());
    replacements.insert("{fn}".to_string(), file_stem);
    replacements.insert("{vs}".to_string(), data.version.clone());

    // Aufruf der Shared-Util-Funktion zur sicheren Namensgenerierung
    crate::shared::utils::render_unique_filename(pattern, &replacements, used)
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