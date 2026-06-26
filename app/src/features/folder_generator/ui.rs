use crate::{FolderState, MainWindow, Model};

use std::path::PathBuf;

use slint::ComponentHandle;

use super::config::{apply_folder_defaults, load_folder_settings, save_folder_settings};
use super::utils::{get_subfolders_vec, sort_subfolders, validate_project_name};

pub fn setup(ui: &MainWindow) {
    apply_folder_defaults(ui);
    load_folder_settings(ui);

    // ==========================================
    // Folder-Creation Callbacks
    // ==========================================

    ui.global::<FolderState>().on_validate_project_name({
        let ui_handle = ui.as_weak();
        move || {
            if let Some(ui) = ui_handle.upgrade() {
                validate_project_name(&ui);
            }
        }
    });

    ui.global::<FolderState>().on_select_folder({
        let ui_handle = ui.as_weak();
        move || {
            if let Some(ui) = ui_handle.upgrade() {
                if let Some(path) = rfd::FileDialog::new().pick_folder() {
                    ui.global::<FolderState>()
                        .set_target_folder(path.to_string_lossy().to_string().into());
                    save_folder_settings(&ui);
                    validate_project_name(&ui);
                }
            }
        }
    });

    ui.global::<FolderState>().on_select_template({
        let ui_handle = ui.as_weak();
        move || {
            if let Some(ui) = ui_handle.upgrade() {
                if let Some(path) = rfd::FileDialog::new()
                    .add_filter("Excel", &["xlsm", "xlsx"])
                    .pick_file()
                {
                    ui.global::<FolderState>()
                        .set_template_file(path.to_string_lossy().to_string().into());
                    save_folder_settings(&ui);
                }
            }
        }
    });

    ui.global::<FolderState>().on_reset({
        let ui_handle = ui.as_weak();
        move || {
            if let Some(ui) = ui_handle.upgrade() {
                apply_folder_defaults(&ui);
                save_folder_settings(&ui);
            }
        }
    });

    ui.global::<FolderState>().on_dismiss_status({
        let ui_handle = ui.as_weak();
        move || {
            if let Some(ui) = ui_handle.upgrade() {
                let fs = ui.global::<FolderState>();
                fs.set_status_type("idle".into());
                fs.set_status_message("".into());
            }
        }
    });

    ui.global::<FolderState>().on_select_csv_file({
        let ui_handle = ui.as_weak();
        move || {
            if let Some(ui) = ui_handle.upgrade() {
                if let Some(path) = rfd::FileDialog::new()
                    .add_filter("CSV", &["csv"])
                    .pick_file()
                {
                    ui.global::<FolderState>()
                        .set_csv_file(path.to_string_lossy().to_string().into());
                }
            }
        }
    });

    ui.global::<FolderState>().on_select_csv_target({
        let ui_handle = ui.as_weak();
        move || {
            if let Some(ui) = ui_handle.upgrade() {
                if let Some(path) = rfd::FileDialog::new().pick_folder() {
                    ui.global::<FolderState>()
                        .set_csv_target_folder(path.to_string_lossy().to_string().into());
                }
            }
        }
    });

    ui.global::<FolderState>().on_import_csv({
            let ui_handle = ui.as_weak();
            move || {
                if let Some(ui) = ui_handle.upgrade() {
                    let fs = ui.global::<FolderState>();
                    let csv_path = PathBuf::from(fs.get_csv_file().to_string());
                    let target = PathBuf::from(fs.get_csv_target_folder().to_string());
                    let template = PathBuf::from(fs.get_template_file().to_string());

                    if !target.exists() {
                        fs.set_status_type("error".into());
                        fs.set_status_message("CSV-Zielverzeichnis existiert nicht.".into());
                        return;
                    }
                    if !template.exists() {
                        fs.set_status_type("error".into());
                        fs.set_status_message(
                            "Excel-Vorlage nicht gefunden (oben im Einzelordner-Bereich auswählen)."
                                .into(),
                        );
                        return;
                    }

                    fs.set_status_type("pending".into());
                    fs.set_status_message("Erstelle Ordner...".into());

                    let subs = get_subfolders_vec(&ui);
                    let ui_handle_clone = ui_handle.clone();

                    std::thread::spawn(move || {
                        let subs_refs: Vec<&str> = subs.iter().map(|s| s.as_str()).collect();
                        match folder_generator::import_csv(&csv_path, &target, &template, &subs_refs) {
                            Ok(result) => {
                                let _ = ui_handle_clone.upgrade_in_event_loop(move |ui| {
                                    let fs = ui.global::<FolderState>();
                                    if result.errors.is_empty() {
                                        fs.set_status_type("success".into());
                                        fs.set_status_message(
                                            format!(
                                                "{} Ordner erstellt, {} übersprungen (existierten bereits).",
                                                result.created, result.skipped
                                            )
                                            .into(),
                                        );
                                    } else {
                                        fs.set_status_type("error".into());
                                        fs.set_status_message(
                                            format!(
                                                "{} erstellt, {} übersprungen, {} Fehler: {}",
                                                result.created,
                                                result.skipped,
                                                result.errors.len(),
                                                result.errors.join("; ")
                                            )
                                            .into(),
                                        );
                                    }
                                });
                            }
                            Err(e) => {
                                let err_str = e.to_string();
                                let _ = ui_handle_clone.upgrade_in_event_loop(move |ui| {
                                    let fs = ui.global::<FolderState>();
                                    fs.set_status_type("error".into());
                                    fs.set_status_message(err_str.into());
                                });
                            }
                        }
                    });
                }
            }
        });

    // ─── Subfolder-Liste: Callbacks ───
    ui.global::<FolderState>().on_add_subfolder({
        let ui_handle = ui.as_weak();
        move |name| {
            if let Some(ui) = ui_handle.upgrade() {
                let fs = ui.global::<FolderState>();
                let model = fs.get_subfolders();
                let mut items: Vec<slint::SharedString> = model.iter().collect();
                items.push(name);
                sort_subfolders(&mut items);
                fs.set_subfolders(std::rc::Rc::new(slint::VecModel::from(items)).into());
                save_folder_settings(&ui);
            }
        }
    });

    ui.global::<FolderState>().on_remove_subfolder({
        let ui_handle = ui.as_weak();
        move |idx| {
            if let Some(ui) = ui_handle.upgrade() {
                let fs = ui.global::<FolderState>();
                let model = fs.get_subfolders();
                let mut items: Vec<slint::SharedString> = model.iter().collect();
                if (idx as usize) < items.len() {
                    items.remove(idx as usize);
                    fs.set_subfolders(std::rc::Rc::new(slint::VecModel::from(items)).into());
                    save_folder_settings(&ui);
                }
            }
        }
    });

    ui.global::<FolderState>().on_rename_subfolder({
        let ui_handle = ui.as_weak();
        move |idx, new_name| {
            if let Some(ui) = ui_handle.upgrade() {
                let fs = ui.global::<FolderState>();
                let model = fs.get_subfolders();
                let mut items: Vec<slint::SharedString> = model.iter().collect();
                if (idx as usize) < items.len() {
                    items[idx as usize] = new_name;
                    sort_subfolders(&mut items);
                    fs.set_subfolders(std::rc::Rc::new(slint::VecModel::from(items)).into());
                    save_folder_settings(&ui);
                }
            }
        }
    });

    ui.global::<FolderState>().on_reset_subfolders({
        let ui_handle = ui.as_weak();
        move || {
            if let Some(ui) = ui_handle.upgrade() {
                let mut defaults: Vec<slint::SharedString> = folder_generator::SUBFOLDERS
                    .iter()
                    .map(|s| (*s).into())
                    .collect();
                sort_subfolders(&mut defaults);
                ui.global::<FolderState>()
                    .set_subfolders(std::rc::Rc::new(slint::VecModel::from(defaults)).into());
                save_folder_settings(&ui);
            }
        }
    });

    ui.global::<FolderState>().on_create_folders({
        let ui_handle = ui.as_weak();
        move || {
            if let Some(ui) = ui_handle.upgrade() {
                let fs = ui.global::<FolderState>();
                let project_name = fs.get_project_name().to_string();
                let target = PathBuf::from(fs.get_target_folder().to_string());
                let template = PathBuf::from(fs.get_template_file().to_string());

                if project_name.is_empty() {
                    fs.set_status_type("error".into());
                    fs.set_status_message("Bitte Projektnamen angeben.".into());
                    return;
                }
                if !target.exists() {
                    fs.set_status_type("error".into());
                    fs.set_status_message("Zielverzeichnis existiert nicht.".into());
                    return;
                }
                if !template.exists() {
                    fs.set_status_type("error".into());
                    fs.set_status_message("Excel-Vorlage nicht gefunden.".into());
                    return;
                }

                fs.set_status_type("pending".into());
                fs.set_status_message("Erstelle Projektordner...".into());

                let subs = get_subfolders_vec(&ui);
                let ui_handle_clone = ui_handle.clone();

                std::thread::spawn(move || {
                    let subs_refs: Vec<&str> = subs.iter().map(|s| s.as_str()).collect();
                    match folder_generator::create_project_folder(
                        &project_name,
                        &target,
                        &template,
                        &subs_refs,
                    ) {
                        Ok(root) => {
                            let root_display = root.display().to_string();
                            let _ = ui_handle_clone.upgrade_in_event_loop(move |ui| {
                                let fs = ui.global::<FolderState>();
                                fs.set_project_name("".into());
                                fs.set_project_name_valid(false);
                                fs.set_status_type("success".into());
                                fs.set_status_message(
                                    format!("Projektordner erstellt: {}", root_display).into(),
                                );
                            });
                        }
                        Err(e) => {
                            let err_str = e.to_string();
                            let _ = ui_handle_clone.upgrade_in_event_loop(move |ui| {
                                let fs = ui.global::<FolderState>();
                                fs.set_status_type("error".into());
                                fs.set_status_message(err_str.into());
                            });
                        }
                    }
                });
            }
        }
    });
}
