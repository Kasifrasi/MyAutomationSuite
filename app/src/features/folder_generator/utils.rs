/// Sortiert wie Windows Explorer: Zahlen numerisch, dann alphabetisch (case-insensitive).
fn sort_subfolders(items: &mut [slint::SharedString]) {
    items.sort_by(|a, b| natural_cmp(&a.to_string(), &b.to_string()));
}

/// Natural sort: "2. Vertrag" < "10. Berichte" (nicht lexikographisch).
fn natural_cmp(a: &str, b: &str) -> std::cmp::Ordering {
    let mut ai = a.chars().peekable();
    let mut bi = b.chars().peekable();

    loop {
        match (ai.peek(), bi.peek()) {
            (None, None) => return std::cmp::Ordering::Equal,
            (None, Some(_)) => return std::cmp::Ordering::Less,
            (Some(_), None) => return std::cmp::Ordering::Greater,
            (Some(ac), Some(bc)) => {
                if ac.is_ascii_digit() && bc.is_ascii_digit() {
                    let na: String = ai.by_ref().take_while(|c| c.is_ascii_digit()).collect();
                    let nb: String = bi.by_ref().take_while(|c| c.is_ascii_digit()).collect();
                    let cmp = na.len().cmp(&nb.len()).then_with(|| na.cmp(&nb));
                    if cmp != std::cmp::Ordering::Equal {
                        return cmp;
                    }
                } else {
                    let ca = ai.next().unwrap().to_ascii_lowercase();
                    let cb = bi.next().unwrap().to_ascii_lowercase();
                    let cmp = ca.cmp(&cb);
                    if cmp != std::cmp::Ordering::Equal {
                        return cmp;
                    }
                }
            }
        }
    }
}

fn get_subfolders_vec(ui: &MainWindow) -> Vec<String> {
    ui.global::<FolderState>()
        .get_subfolders()
        .iter()
        .map(|s| s.to_string())
        .collect()
}

fn validate_project_name(ui: &MainWindow) {
    let fs = ui.global::<FolderState>();
    let raw = fs.get_project_name().to_string();
    let skip = fs.get_skip_validation();

    if !skip && !raw.is_empty() {
        let formatted = folder_generator::format_project_name(&raw);
        if formatted != raw {
            fs.set_project_name(formatted.into());
        }
    }

    let name = fs.get_project_name().to_string();
    let valid = if skip {
        !name.is_empty()
    } else {
        folder_generator::is_valid_project_number(&name)
    };
    fs.set_project_name_valid(valid);

    let target = fs.get_target_folder().to_string();
    if !target.is_empty() && !name.is_empty() {
        fs.set_folder_exists(PathBuf::from(&target).join(&name).exists());
    } else {
        fs.set_folder_exists(false);
    }
}