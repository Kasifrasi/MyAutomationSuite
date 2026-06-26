pub fn get_fb_path() -> std::path::PathBuf {
    // Hier binden wir die kompilierte Go-Exe direkt in die Rust-Anwendung ein!
    let sidecar_bytes = include_bytes!("../../../sidecars/FB/fb_generator.exe");

    // Wir entpacken sie in den Temp-Ordner
    let dir = std::env::temp_dir().join("MyAutomationSuite");
    let _ = std::fs::create_dir_all(&dir);

    let exe_name = if cfg!(windows) {
        "fb_generator.exe"
    } else {
        "fb_generator"
    };
    let exe_path = dir.join(exe_name);

    // Nur neu schreiben, wenn sie noch nicht existiert oder sich die Größe geändert hat (z.B. nach einem App-Update)
    // Das verhindert unnötige Schreibvorgänge und beruhigt Antivirenscanner.
    let needs_write = match std::fs::metadata(&exe_path) {
        Ok(meta) => meta.len() as usize != sidecar_bytes.len(),
        Err(_) => true,
    };

    if needs_write {
        let _ = std::fs::write(&exe_path, sidecar_bytes);

        // Auf Linux/macOS müssen wir die Datei ausführbar machen
        #[cfg(unix)]
        {
            use std::os::unix::fs::PermissionsExt;
            if let Ok(mut perms) = std::fs::metadata(&exe_path).map(|m| m.permissions()) {
                perms.set_mode(0o755);
                let _ = std::fs::set_permissions(&exe_path, perms);
            }
        }
    }

    exe_path
}

pub fn get_vorpruefung_path() -> std::path::PathBuf {
    // Vorpruefung-Sidecar (Go) wird wie der FB-Generator direkt eingebettet.
    // Vor `cargo build` muss er via `build-go` als sidecars/Vorpruefung/vp_generator.exe
    // erzeugt worden sein.
    let sidecar_bytes = include_bytes!("../../../sidecars/Vorpruefung/vp_generator.exe");

    let dir = std::env::temp_dir().join("MyAutomationSuite");
    let _ = std::fs::create_dir_all(&dir);

    let exe_name = if cfg!(windows) {
        "vp_generator.exe"
    } else {
        "vp_generator"
    };
    let exe_path = dir.join(exe_name);

    let needs_write = match std::fs::metadata(&exe_path) {
        Ok(meta) => meta.len() as usize != sidecar_bytes.len(),
        Err(_) => true,
    };

    if needs_write {
        let _ = std::fs::write(&exe_path, sidecar_bytes);
        #[cfg(unix)]
        {
            use std::os::unix::fs::PermissionsExt;
            if let Ok(mut perms) = std::fs::metadata(&exe_path).map(|m| m.permissions()) {
                perms.set_mode(0o755);
                let _ = std::fs::set_permissions(&exe_path, perms);
            }
        }
    }

    exe_path
}