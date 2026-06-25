
pub fn setup(ui: &MainWindow) {
    apply_fb_defaults(ui);
    load_fb_settings(ui);
}

// ==========================================
// FB-Generator Callbacks
// ==========================================

ui.global::<FBState>().on_select_folder({
    let ui_handle = ui.as_weak();
    move || {
        if let Some(ui) = ui_handle.upgrade() {
            if let Some(path) = rfd::FileDialog::new().pick_folder() {
                let fb = ui.global::<FBState>();
                fb.set_folder(path.to_string_lossy().to_string().into());
                fb.set_status_type("idle".into());
                fb.set_status_message("".into());
            }
        }
    }
});

ui.global::<FBState>().on_reset({
    let ui_handle = ui.as_weak();
    move || {
        if let Some(ui) = ui_handle.upgrade() {
            apply_fb_defaults(&ui);
            save_fb_settings(&ui);
        }
    }
});

ui.global::<FBState>().on_dismiss_status({
    let ui_handle = ui.as_weak();
    move || {
        if let Some(ui) = ui_handle.upgrade() {
            let fb = ui.global::<FBState>();
            fb.set_status_type("idle".into());
            fb.set_status_message("".into());
        }
    }
});

ui.global::<FBState>().on_save_settings({
    let ui_handle = ui.as_weak();
    move || {
        if let Some(ui) = ui_handle.upgrade() {
            save_fb_settings(&ui);
        }
    }
});

ui.global::<FBState>().on_toggle_settings({
    let ui_handle = ui.as_weak();
    move || {
        if let Some(ui) = ui_handle.upgrade() {
            let fb = ui.global::<FBState>();
            fb.set_show_settings(!fb.get_show_settings());
        }
    }
});

ui.global::<FBState>().on_generate_report({
    let ui_handle = ui.as_weak();

    move || {
        if let Some(ui) = ui_handle.upgrade() {
            let fb = ui.global::<FBState>();

            let folder = fb.get_folder().to_string();
            if folder.is_empty() {
                fb.set_status_type("error".into());
                fb.set_status_message("Bitte Ausgabeordner wählen.".into());
                return;
            }

            let name = fb.get_name().to_string();
            if name.is_empty() {
                fb.set_status_type("error".into());
                fb.set_status_message("Bitte Dateinamens-Muster angeben.".into());
                return;
            }

            let langs = fb.get_langs();
            let mut lang_list = Vec::new();
            if langs.de {
                lang_list.push("de");
            }
            if langs.en {
                lang_list.push("en");
            }
            if langs.fr {
                lang_list.push("fr");
            }
            if langs.es {
                lang_list.push("es");
            }
            if langs.pt {
                lang_list.push("pt");
            }

            if lang_list.is_empty() {
                fb.set_status_type("error".into());
                fb.set_status_message("Bitte mindestens eine Sprache wählen.".into());
                return;
            }

            fb.set_status_type("pending".into());
            fb.set_status_message("Export läuft...".into());

            let cats = fb.get_categories();
            let counts = [
                cats.cat1 as u16,
                cats.cat2 as u16,
                cats.cat3 as u16,
                cats.cat4 as u16,
                cats.cat5 as u16,
                cats.cat6 as u16,
                cats.cat7 as u16,
                cats.cat8 as u16,
            ];

            let name_clone = name.clone();

            let sp = fb.get_sheet_permissions();
            let options = ExportOptions {
                protect_sheet: fb.get_protect_sheet(),
                protect_workbook: fb.get_protect_workbook(),
                sheet_password: fb.get_sheet_password().to_string(),
                workbook_password: fb.get_workbook_password().to_string(),
                hide_columns: fb.get_hide_columns(),
                hide_lang_sheet: fb.get_hide_lang_sheet(),
                select_locked: sp.select_locked,
                select_unlocked: sp.select_unlocked,
                format_cells: sp.format_cells,
                format_columns: sp.format_columns,
                format_rows: sp.format_rows,
                insert_columns: sp.insert_columns,
                insert_rows: sp.insert_rows,
                insert_hyperlinks: sp.insert_hyperlinks,
                delete_columns: sp.delete_columns,
                delete_rows: sp.delete_rows,
                sort: sp.sort,
                autofilter: sp.autofilter,
                pivot_tables: sp.pivot_tables,
                edit_objects: sp.edit_objects,
                edit_scenarios: sp.edit_scenarios,
                empty_rows: fb.get_empty_rows(),
            };
            let wb_hash = if options.protect_workbook {
                Some(excel_protection::precompute_hash(
                    &options.workbook_password,
                ))
            } else {
                None
            };

            let sh_hash = if options.protect_sheet {
                Some(excel_protection::precompute_hash(&options.sheet_password))
            } else {
                None
            };

            let sh_opts = if options.protect_sheet {
                Some(excel_protection::SheetProtectionOptions {
                    select_locked_cells: options.select_locked,
                    select_unlocked_cells: options.select_unlocked,
                    format_cells: options.format_cells,
                    format_columns: options.format_columns,
                    format_rows: options.format_rows,
                    insert_columns: options.insert_columns,
                    insert_rows: options.insert_rows,
                    insert_hyperlinks: options.insert_hyperlinks,
                    delete_columns: options.delete_columns,
                    delete_rows: options.delete_rows,
                    sort: options.sort,
                    auto_filter: options.autofilter,
                    pivot_tables: options.pivot_tables,
                    objects: options.edit_objects,
                    scenarios: options.edit_scenarios,
                })
            } else {
                None
            };

            let mut sidecar_options = options.clone();
            sidecar_options.protect_sheet = false;
            sidecar_options.protect_workbook = false;
            let options_json = serde_json::to_string(&sidecar_options).unwrap_or_default();

            let ui_handle_clone = ui_handle.clone();
            std::thread::spawn(move || {
                let start_time = std::time::Instant::now();
                let mut templates = Vec::new();

                for lang in &lang_list {
                    let mut positions = Vec::new();
                    for (i, &pos_count) in counts.iter().enumerate() {
                        let category = i + 1;
                        for p in 0..pos_count {
                            positions.push(budget_scanner::BudgetPosition {
                                number: format!("{category}.{}", p + 1),
                                label: String::new(),
                                cost_col1: String::new(),
                                cost_col2: String::new(),
                                cost_year1: String::new(),
                                cost_year2: String::new(),
                                cost_year3: String::new(),
                            });
                        }
                    }

                    templates.push(budget_scanner::BudgetData {
                        file_path: std::path::PathBuf::from(format!("Vorlage_{lang}.xlsx")),
                        sheet_name: "Budget".into(),
                        version: String::new(),
                        project_title: "".into(),
                        project_number: "Vorlage".into(),
                        language: lang.to_string(),
                        local_currency: "".into(),
                        cost_col1: 8,
                        cost_col2: Some(13),
                        eigenleistung: "0".into(),
                        drittmittel: "0".into(),
                        kmw_mittel: "0".into(),
                        financing: budget_scanner::FinancingDetail::default(),
                        positions,
                    });
                }

                let output_dir = std::path::PathBuf::from(&folder);
                let _ = std::fs::create_dir_all(&output_dir);

                let tmp_json_path = std::env::temp_dir().join(format!(
                    "template_{}.json",
                    std::time::SystemTime::now()
                        .duration_since(std::time::UNIX_EPOCH)
                        .unwrap()
                        .as_millis()
                ));
                if let Err(e) = std::fs::File::create(&tmp_json_path).and_then(|mut f| {
                    let json = serde_json::to_string(&templates)?;
                    std::io::Write::write_all(&mut f, json.as_bytes())?;
                    Ok(())
                }) {
                    let _ = ui_handle_clone.upgrade_in_event_loop(move |ui| {
                        let fb = ui.global::<FBState>();
                        fb.set_status_type("error".into());
                        fb.set_status_message(
                            format!("Fehler beim Speichern der JSON: {e}").into(),
                        );
                    });
                    return;
                }

                // 4. Go Sidecar aufrufen
                let sidecar_exe = get_fb_path();

                let mut cmd = std::process::Command::new(&sidecar_exe);

                #[cfg(target_os = "windows")]
                {
                    use std::os::windows::process::CommandExt;
                    const CREATE_NO_WINDOW: u32 = 0x08000000;
                    cmd.creation_flags(CREATE_NO_WINDOW);
                }

                cmd.arg("-input")
                    .arg(&tmp_json_path)
                    .arg("-output")
                    .arg(&output_dir)
                    .arg("-options")
                    .arg(&options_json)
                    .arg("-filename")
                    .arg(&name_clone);

                cmd.stdout(std::process::Stdio::piped());

                let mut child = match cmd.spawn() {
                    Ok(c) => c,
                    Err(e) => {
                        let _ = ui_handle_clone.upgrade_in_event_loop(move |ui| {
                            let fb = ui.global::<FBState>();
                            fb.set_status_type("error".into());
                            fb.set_status_message(
                                format!(
                                    "Fehler beim Starten von {}: {e}",
                                    sidecar_exe.display()
                                )
                                .into(),
                            );
                        });
                        return;
                    }
                };

                let stdout = child.stdout.take().unwrap();
                let reader = std::io::BufReader::new(stdout);

                use std::io::BufRead;
                for line in reader.lines().map_while(Result::ok) {
                    if let Ok(msg) = serde_json::from_str::<ProgressMessage>(&line) {
                        let _ = ui_handle_clone.upgrade_in_event_loop({
                            let msg_status = msg.status.clone();
                            let msg_text = msg.message.clone();
                            let current = msg.current.unwrap_or(0);
                            let total = msg.total.unwrap_or(0);

                            move |ui| {
                                let fb = ui.global::<FBState>();

                                if msg_status == "error" {
                                    fb.set_status_type("error".into());
                                } else if msg_status == "done" {
                                    fb.set_status_type("success".into());
                                } else {
                                    fb.set_status_type("pending".into());
                                }

                                if total > 0 {
                                    fb.set_status_message(
                                        format!("{current}/{total} - {msg_text}").into(),
                                    );
                                } else {
                                    fb.set_status_message(msg_text.into());
                                }
                            }
                        });
                    }
                }

                let _ = child.wait();
                let _ = std::fs::remove_file(&tmp_json_path);

                if wb_hash.is_some() || sh_hash.is_some() {
                    let _ = ui_handle_clone.upgrade_in_event_loop(move |ui| {
                        let fb = ui.global::<FBState>();
                        fb.set_status_message("Wende Schutz an...".into());
                    });

                    use rayon::prelude::*;
                    if let Ok(entries) = std::fs::read_dir(&output_dir) {
                        let paths: Vec<_> = entries.flatten().map(|e| e.path()).collect();
                        paths.into_par_iter().for_each(|p| {
                            if p.extension().map_or(false, |ext| ext == "xlsx") {
                                let _ = excel_protection::apply_protection_in_place(
                                    &p,
                                    wb_hash.as_ref(),
                                    sh_hash.as_ref(),
                                    sh_opts.as_ref(),
                                );
                            }
                        });
                    }
                }

                let success_count = templates.len();
                let _ = ui_handle_clone.upgrade_in_event_loop(move |ui| {
                    let fb = ui.global::<FBState>();
                    let elapsed_sec = start_time.elapsed().as_secs_f64();
                    fb.set_status_type("success".into());
                    fb.set_status_message(
                        format!(
                            "Erfolgreich abgeschlossen! {} Datei(en) in {:.2}s erstellt.",
                            success_count, elapsed_sec
                        )
                        .into(),
                    );
                });
            });
        }
    }
});