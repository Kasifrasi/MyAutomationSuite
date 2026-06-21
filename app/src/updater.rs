//! Auto-Update via `self_update` (GitHub Releases Backend).
//! Läuft komplett in einem Hintergrundthread; UI-Updates gehen via
//! `slint::invoke_from_event_loop` zurück an die Slint-Event-Loop.

use crate::{MainWindow, UpdateState};
use slint::{ComponentHandle, Weak};

const GITHUB_OWNER: &str = "Kasifrasi";
const GITHUB_REPO: &str = "Automation-Tool";
const ASSET_NAME: &str = "gui";

// ── Öffentliche API ──────────────────────────────────────────────────────────

/// Phase 1: Prüft ob eine neuere Version auf GitHub verfügbar ist.
/// Setzt `update-available = true` und zeigt den Tag, wenn ein Update gefunden wird.
pub fn spawn_check(handle: Weak<MainWindow>) {
    let h = handle.clone();
    std::thread::spawn(move || {
        set_ui(&h, "Prüfe...", true, false);

        match do_check() {
            Ok(None) => set_ui(&h, "Bereits aktuell", false, false),
            Ok(Some(tag)) => {
                let msg = format!("Update verfügbar: {}", tag);
                let h2 = h.clone();
                let _ = slint::invoke_from_event_loop(move || {
                    if let Some(ui) = h2.upgrade() {
                        let us = ui.global::<UpdateState>();
                        us.set_checking(false);
                        us.set_update_available(true);
                        us.set_status(msg.into());
                    }
                });
            }
            Err(e) => set_ui(&h, &format!("Fehler: {}", e), false, false),
        }
    });
}

/// Phase 2: Lädt die neue `.exe` herunter und ersetzt die laufende Datei.
/// Auf Windows: Neustart erforderlich (self_update verwendet einen Temp-File-Trick).
pub fn spawn_install(handle: Weak<MainWindow>) {
    let h = handle.clone();
    std::thread::spawn(move || {
        set_ui(&h, "Installiere...", true, true);

        match do_install() {
            Ok(tag) => {
                let msg = format!("Installiert – bitte neu starten ({})", tag);
                let h2 = h.clone();
                let _ = slint::invoke_from_event_loop(move || {
                    if let Some(ui) = h2.upgrade() {
                        let us = ui.global::<UpdateState>();
                        us.set_checking(false);
                        us.set_update_available(false);
                        us.set_installed(true);
                        us.set_status(msg.into());
                    }
                });
            }
            Err(e) => set_ui(&h, &format!("Fehler: {}", e), false, false),
        }
    });
}

// ── Interne Hilfsfunktionen ──────────────────────────────────────────────────

fn set_ui(handle: &Weak<MainWindow>, status: &str, checking: bool, update_available: bool) {
    let status = status.to_string();
    let h = handle.clone();
    let _ = slint::invoke_from_event_loop(move || {
        if let Some(ui) = h.upgrade() {
            let us = ui.global::<UpdateState>();
            us.set_checking(checking);
            us.set_update_available(update_available);
            us.set_status(status.into());
        }
    });
}

fn build_updater() -> Result<self_update::backends::github::UpdateBuilder, Box<dyn std::error::Error + Send + Sync>> {
    let mut builder = self_update::backends::github::Update::configure();
    builder
        .repo_owner(GITHUB_OWNER)
        .repo_name(GITHUB_REPO)
        .bin_name(ASSET_NAME)
        .current_version(env!("CARGO_PKG_VERSION"))
        .no_confirm(true);

    // Für private Repos: GITHUB_TOKEN aus Umgebungsvariable lesen
    if let Ok(token) = std::env::var("GITHUB_TOKEN") {
        builder.auth_token(&token);
    }

    Ok(builder)
}

/// Gibt `Some(tag)` zurück wenn ein neueres Release existiert, sonst `None`.
fn do_check() -> Result<Option<String>, Box<dyn std::error::Error + Send + Sync>> {
    let updater = build_updater()?.build()?;
    let latest = updater.get_latest_release()?;

    // Ungültige Tags (z.B. "v_test") ignorieren → "Bereits aktuell"
    let Some(latest_ver) = semver::Version::parse(latest.version.trim_start_matches('v')).ok() else {
        return Ok(None);
    };
    let current_ver = semver::Version::parse(env!("CARGO_PKG_VERSION"))?;

    if latest_ver > current_ver {
        Ok(Some(latest.version))
    } else {
        Ok(None)
    }
}

/// Führt den Download und das Ersetzen der laufenden `.exe` durch.
fn do_install() -> Result<String, Box<dyn std::error::Error + Send + Sync>> {
    let updater = build_updater()?.build()?;

    let latest = updater.get_latest_release()?;
    let tag = latest.version.clone();
    updater.update()?;
    Ok(tag)
}
