
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
