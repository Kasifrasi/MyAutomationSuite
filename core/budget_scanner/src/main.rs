use budget_scanner::{col_to_letter, scan_directory, scan_file, write_failure_report, BudgetData};
use comfy_table::{Attribute, Cell, ContentArrangement, Table};
use std::env;
use std::path::Path;

fn print_budget(data: &BudgetData) {
    println!("\n=== {} ===", data.file_path.display());
    println!("  Sheet:           {}", data.sheet_name);
    println!("  Version (A2):    {}", data.version);
    println!("  Projekttitel:    {}", data.project_title);
    println!("  Projektnummer:   {}", data.project_number);
    println!("  Sprache:         {}", data.language);
    println!("  Lokalwährung:    {}", data.local_currency);

    let c1 = col_to_letter(data.cost_col1).to_string();
    let c2 = data
        .cost_col2
        .map(|c| col_to_letter(c).to_string())
        .unwrap_or("-".into());
    println!("  Kostenspalten:   {} / {}", c1, c2);
    println!("  Eigenleistung:   {}", data.eigenleistung);
    println!("  Drittmittel:     {}", data.drittmittel);
    println!("  KMW-Mittel:      {}", data.kmw_mittel);

    let mut table = Table::new();
    table.set_content_arrangement(ContentArrangement::Dynamic);
    table.set_header(vec![
        Cell::new("Nr.").add_attribute(Attribute::Bold),
        Cell::new("Bezeichnung").add_attribute(Attribute::Bold),
        Cell::new(&c1).add_attribute(Attribute::Bold),
        Cell::new(&c2).add_attribute(Attribute::Bold),
    ]);

    for pos in &data.positions {
        table.add_row(vec![
            Cell::new(&pos.number),
            Cell::new(&pos.label),
            Cell::new(&pos.cost_col1),
            Cell::new(&pos.cost_col2),
        ]);
    }

    println!("{table}");
}

fn main() {
    let args: Vec<String> = env::args().collect();

    // Prüfe auf JSON-Modus
    let json_mode = args.iter().any(|a| a == "--json");
    let input_arg = args.iter().find(|a| *a != "--json" && *a != &args[0]);

    if input_arg.is_none() {
        eprintln!("Usage: {} [--json] <file_or_directory>", args[0]);
        std::process::exit(1);
    }

    let input = Path::new(input_arg.unwrap());

    if input.is_file() {
        match scan_file(input) {
            Ok(data) => {
                if json_mode {
                    println!("{}", serde_json::to_string(&data).unwrap());
                } else {
                    print_budget(&data);
                }
            }
            Err(failure) => {
                if json_mode {
                    eprintln!("{}", serde_json::to_string(&failure).unwrap());
                } else {
                    eprintln!("Fehler: {} — {}", failure.file_name, failure.reason);
                }
                std::process::exit(1);
            }
        }
    } else if input.is_dir() {
        let result = scan_directory(input);

        if json_mode {
            println!("{}", serde_json::to_string(&result).unwrap());
        } else {
            println!(
                "{} erfolgreich, {} fehlgeschlagen.",
                result.successes.len(),
                result.failures.len()
            );

            for data in &result.successes {
                print_budget(data);
            }

            if !result.failures.is_empty() {
                println!("\n--- Fehlgeschlagene Dateien ---");
                for f in &result.failures {
                    println!("  {} — {}", f.file_name, f.reason);
                }

                let report_path = input.join("scan_fehler.csv");
                match write_failure_report(&result.failures, &report_path) {
                    Ok(()) => println!("\nFehler-Report: {}", report_path.display()),
                    Err(e) => eprintln!("Fehler beim Schreiben des Reports: {}", e),
                }
            }
        }
    } else {
        eprintln!(
            "'{}' ist weder eine Datei noch ein Verzeichnis.",
            input.display()
        );
        std::process::exit(1);
    }
}
