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

pub fn run_sidecar_batch<T: serde::Serialize>(
    sidecar_exe: &std::path::Path,
    data: &[T],
    out_dir: &std::path::Path,
    filename_pattern: &str,
    options_json: Option<&str>,
    wb_hash: Option<excel_protection::PrecomputedHash>,
    sh_hash: Option<excel_protection::PrecomputedHash>,
    sh_opts: Option<excel_protection::SheetProtectionOptions>,
    mut on_progress: impl FnMut(crate::shared::models::ProgressMessage),
) -> Result<u32, String> {
    // 1. Temp-JSON schreiben
    let mut tmp_json_file = tempfile::Builder::new()
        .prefix("sidecar_batch_")
        .suffix(".json")
        .tempfile()
        .map_err(|e| format!("Temp-JSON Fehler: {e}"))?;

    let json = serde_json::to_string(data).map_err(|e| format!("JSON-Serialize Fehler: {e}"))?;

    std::io::Write::write_all(&mut tmp_json_file, json.as_bytes())
        .map_err(|e| format!("Temp-JSON Schreibfehler: {e}"))?;
    let _ = std::io::Write::flush(&mut tmp_json_file);
    let tmp_json_path = tmp_json_file.into_temp_path();

    // 2. Prozess starten
    let mut cmd = std::process::Command::new(sidecar_exe);
    #[cfg(target_os = "windows")]
    {
        use std::os::windows::process::CommandExt;
        const CREATE_NO_WINDOW: u32 = 0x08000000;
        cmd.creation_flags(CREATE_NO_WINDOW);
    }

    cmd.arg("-input")
        .arg(&tmp_json_path)
        .arg("-output")
        .arg(out_dir)
        .arg("-filename")
        .arg(filename_pattern);

    if let Some(opt) = options_json {
        cmd.arg("-options").arg(opt);
    }

    cmd.stdout(std::process::Stdio::piped());

    let mut child = cmd
        .spawn()
        .map_err(|e| format!("Fehler beim Starten von {}: {e}", sidecar_exe.display()))?;

    let stdout = child.stdout.take().ok_or("Konnte Ausgabe nicht lesen")?;
    let reader = std::io::BufReader::new(stdout);
    let mut ok_count = 0u32;

    use std::io::BufRead;
    for line in reader.lines().map_while(Result::ok) {
        if let Ok(msg) = serde_json::from_str::<crate::shared::models::ProgressMessage>(&line) {
            if msg.status == "success" || msg.status == "progress" {
                ok_count += 1;
            }
            on_progress(msg);
        }
    }

    let _ = child.wait();

    // 3. Optionaler Excel-Schutz anwenden
    if wb_hash.is_some() || sh_hash.is_some() {
        on_progress(crate::shared::models::ProgressMessage {
            status: "pending".into(),
            message: "Wende Schutz an...".into(),
            current: None,
            total: None,
            file: None,
        });

        use rayon::prelude::*;
        if let Ok(entries) = std::fs::read_dir(out_dir) {
            let paths: Vec<_> = entries.flatten().map(|e| e.path()).collect();
            paths.into_par_iter().for_each(|p| {
                if p.extension().is_some_and(|ext| ext == "xlsx") {
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

    Ok(ok_count)
}
