
/// Sanitisiert einen String für die Verwendung in Dateinamen
pub fn sanitize_filename(s: &str) -> String {
    s.chars()
        .map(|c| if "\\/:*?\"<>|".contains(c) { '_' } else { c })
        .collect()
}

/// Ersetzt Platzhalter in einem Template-String und stellt Eindeutigkeit sicher.
/// `replacements`: Map von Platzhalter → Wert (z.B. `{"{pn}" -> "12345"}`)
pub fn render_unique_filename(
    pattern: &str,
    replacements: &std::collections::HashMap<String, String>,
    used: &mut std::collections::HashSet<String>,
) -> String {
    let mut name = pattern.to_string();
    for (placeholder, value) in replacements {
        name = name.replace(placeholder, &sanitize_filename(value));
    }
    
    if !name.to_lowercase().ends_with(".xlsx") {
        name.push_str(".xlsx");
    }

    let (stem, ext) = match name.to_lowercase().rfind(".xlsx") {
        Some(pos) => (name[..pos].to_string(), name[pos..].to_string()),
        None => (name.clone(), String::new()),
    };
    let has_counter = stem.contains("{i}");

    let mut n = 1u32;
    loop {
        let candidate = if has_counter {
            format!("{}{}", stem.replace("{i}", &n.to_string()), ext)
        } else if n == 1 {
            format!("{stem}{ext}")
        } else {
            format!("{stem}_{n}{ext}")
        };
        if used.insert(candidate.clone()) {
            return candidate;
        }
        n += 1;
    }
}

/// Parst einen Geldbetrag aus dem Budget (z.B. "10,000", "5,000.00 €", "1.234,56").
/// Erkennt Tausender-/Dezimaltrenner heuristisch. Leer/unparsbar ⇒ None (leeres
/// Eingabefeld in der Prüfvorlage).
fn parse_amount(raw: &str) -> Option<f64> {
    let mut s: String = raw
        .chars()
        .filter(|c| c.is_ascii_digit() || *c == ',' || *c == '.' || *c == '-')
        .collect();
    if s.is_empty() || s == "-" {
        return None;
    }

    let has_dot = s.contains('.');
    let has_comma = s.contains(',');

    if has_dot && has_comma {
        // Der zuletzt auftretende Trenner ist der Dezimaltrenner.
        let last_dot = s.rfind('.').unwrap();
        let last_comma = s.rfind(',').unwrap();
        if last_comma > last_dot {
            s = s.replace('.', "").replace(',', ".");
        } else {
            s = s.replace(',', "");
        }
    } else if has_comma {
        let after = s.rsplit(',').next().map(|p| p.len()).unwrap_or(0);
        let commas = s.matches(',').count();
        if commas == 1 && after != 3 {
            // Einzelnes Komma, nicht im Tausenderformat ⇒ Dezimaltrenner.
            s = s.replace(',', ".");
        } else {
            s = s.replace(',', "");
        }
    }

    s.parse::<f64>().ok()
}