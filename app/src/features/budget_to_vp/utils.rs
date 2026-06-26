/// Erzeugt einen eindeutigen Ausgabedateinamen aus dem Muster und den Metadaten des
/// gescannten Budgets. Die fachliche Abbildung der Budgetdaten auf das Prüfvorlagen-
/// Schema erledigt das Vorpruefung-Sidecar selbst — Rust gibt die kanonische
/// `budget_scanner::BudgetData` unverändert weiter.
pub fn vp_output_name(
    pattern: &str,
    data: &budget_scanner::BudgetData,
    used: &mut std::collections::HashSet<String>,
) -> String {
    let mut replacements = std::collections::HashMap::new();

    // Originalen Dateinamen ohne Endung ermitteln (z.B. "de" statt "de.xlsx")
    let file_stem = data
        .file_path
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
