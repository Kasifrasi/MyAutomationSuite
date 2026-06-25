
#[derive(serde::Serialize, serde::Deserialize)]
struct ThemeSettings {
    dark_mode: bool,
}

impl Default for ThemeSettings {
    fn default() -> Self {
        Self {
            dark_mode: matches!(dark_light::detect(), Ok(dark_light::Mode::Dark)),
        }
    }
}


fn load_theme_settings(ui: &MainWindow) {
    let s: ThemeSettings = confy::load(APP_NAME, "theme").unwrap_or_default();
    ui.set_dark_mode(s.dark_mode);
    let scheme = if s.dark_mode {
        slint::language::ColorScheme::Dark
    } else {
        slint::language::ColorScheme::Light
    };
    ui.global::<Palette>().set_color_scheme(scheme);
}

fn save_theme_settings(dark_mode: bool) {
    let _ = confy::store(APP_NAME, "theme", &ThemeSettings { dark_mode });
}