use crate::MainWindow;
use slint::ComponentHandle; // Für .as_weak()
use slint::Global; // Für ui.global()

use super::config::{load_theme_settings, save_theme_settings};

pub fn setup(ui: &MainWindow) {
    // Theme laden
    load_theme_settings(ui);

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
                // "Palette" wird von Slint automatisch in den Crate-Root importiert
                // wenn du slint::include_modules!() nutzt
                ui.global::<crate::Palette>().set_color_scheme(scheme);
                save_theme_settings(dark);
            }
        }
    });
}