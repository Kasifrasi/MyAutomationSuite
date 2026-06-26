#![windows_subsystem = "windows"]

slint::include_modules!();

// 1. Die drei Haupt-Ordner als Module anmelden
mod features;
mod shared;
mod shell;

// 2. Deine MainWindow Instanz
use slint::Model;

const APP_NAME: &str = "automation-tool";

fn main() -> Result<(), slint::PlatformError> {
    let ui = MainWindow::new()?;

    // Fenster unter Windows maximiert starten
    #[cfg(target_os = "windows")]
    ui.window().set_maximized(true);

    // 3. Setup-Aufrufe (delegieren die Arbeit an die ui.rs Dateien)
    shell::ui::setup(&ui); // Theme & Dark Mode (falls deine Datei shell/ui.rs heißt)
    shell::updater::setup(&ui); // Updater Callbacks

    features::fb_generator::ui::setup(&ui);
    features::budget_to_fb::ui::setup(&ui);
    features::budget_to_vp::ui::setup(&ui);
    features::folder_generator::ui::setup(&ui);

    ui.run()
}
