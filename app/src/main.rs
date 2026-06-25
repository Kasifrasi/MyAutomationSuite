#![windows_subsystem = "windows"]

slint::include_modules!();

mod shell/updater;
mod shell/models;

use slint::Model;
use std::path::PathBuf;

const APP_NAME: &str = "automation-tool";

fn main() -> Result<(), slint::PlatformError> {
    let ui = MainWindow::new()?;

    // Defaults setzen, dann gespeicherte Settings laden
    apply_fb_defaults(&ui);
    load_fb_settings(&ui);
    apply_b2f_defaults(&ui);
    load_b2f_settings(&ui);
    apply_vp_defaults(&ui);
    load_vp_settings(&ui);
    apply_folder_defaults(&ui);
    load_folder_settings(&ui);

    // Theme: gespeicherte Einstellung laden (Fallback: System-Erkennung)
    load_theme_settings(&ui);

    // ==========================================
    // UpdateState: Versionsnummer + Callback
    // ==========================================
    ui.global::<UpdateState>()
        .set_app_version(env!("CARGO_PKG_VERSION").into());

    // Beim Start automatisch nach Updates suchen
    updater::spawn_check(ui.as_weak());

    ui.global::<UpdateState>().on_check_for_update({
        let ui_handle = ui.as_weak();
        move || {
            if let Some(ui) = ui_handle.upgrade() {
                let us = ui.global::<UpdateState>();
                if us.get_update_available() {
                    updater::spawn_install(ui_handle.clone());
                } else {
                    updater::spawn_check(ui_handle.clone());
                }
            }
        }
    });

    ui.global::<UpdateState>().on_restart({
        let ui_handle = ui.as_weak();
        move || {
            // Aktuellen Pfad der .exe ermitteln und neu starten
            if let Ok(exe) = std::env::current_exe() {
                let _ = std::process::Command::new(exe).spawn();
            }
            // Aktuelles Fenster schließen
            if let Some(ui) = ui_handle.upgrade() {
                let _ = ui.hide();
            }
            std::process::exit(0);
        }
    });

    // Dark Mode Toggle
    ui.on_toggle_dark_mode({
        let ui_handle = ui.as_weak();
        move |dark| {
            if let Some(ui) = ui_handle.upgrade() {
                let scheme = if dark {
                    slint::language::ColorScheme::Dark
                } else {
                    slint::language::ColorScheme::Light
                };
                ui.global::<Palette>().set_color_scheme(scheme);
                save_theme_settings(dark);
            }
        }
    });

    ui.run()
}

