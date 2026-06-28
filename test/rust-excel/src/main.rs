use std::fs::{self, File};
use std::io::{Read, Write, Cursor};
use std::time::Instant;
use zip::{ZipArchive, ZipWriter};
use zip::write::SimpleFileOptions;
use quick_xml::events::{Event, BytesStart, BytesEnd};
use quick_xml::reader::Reader;
use quick_xml::writer::Writer;
use std::collections::HashSet;
use rayon::prelude::*;

#[derive(Clone, PartialEq, Eq)]
struct ColData {
    style: String,
    attrs: Vec<(String, String)>,
}

fn find_fill_id_by_color(xml: &[u8], target_color: &str) -> Option<usize> {
    let mut reader = Reader::from_reader(xml);
    let mut fill_idx = 0;
    let mut in_fill = false;
    let mut in_fills = false;
    let target_upper = target_color.to_uppercase();
    let mut buf = Vec::new();

    loop {
        match reader.read_event_into(&mut buf) {
            Ok(Event::Start(ref e)) => {
                let name = e.name();
                if name.as_ref() == b"fills" {
                    in_fills = true;
                } else if in_fills && name.as_ref() == b"fill" {
                    in_fill = true;
                } else if in_fill && (name.as_ref() == b"fgColor" || name.as_ref() == b"bgColor") {
                    for attr in e.attributes().flatten() {
                        if attr.key.as_ref() == b"rgb" {
                            if let Ok(val) = std::str::from_utf8(&attr.value) {
                                if val.to_uppercase().ends_with(&target_upper) { return Some(fill_idx); }
                            }
                        }
                    }
                }
            }
            Ok(Event::Empty(ref e)) => {
                let name = e.name();
                if in_fills && name.as_ref() == b"fill" {
                    fill_idx += 1;
                } else if in_fill && (name.as_ref() == b"fgColor" || name.as_ref() == b"bgColor") {
                    for attr in e.attributes().flatten() {
                        if attr.key.as_ref() == b"rgb" {
                            if let Ok(val) = std::str::from_utf8(&attr.value) {
                                if val.to_uppercase().ends_with(&target_upper) { return Some(fill_idx); }
                            }
                        }
                    }
                }
            }
            Ok(Event::End(ref e)) => {
                let name = e.name();
                if name.as_ref() == b"fills" {
                    in_fills = false;
                } else if in_fills && name.as_ref() == b"fill" {
                    fill_idx += 1;
                    in_fill = false;
                }
            }
            Ok(Event::Eof) | Err(_) => break,
            _ => {}
        }
        buf.clear();
    }
    None
}

fn rewrite_styles_xml(xml: &[u8], target_fill_id: usize) -> Result<(Vec<u8>, usize, usize, HashSet<usize>, HashSet<usize>), Box<dyn std::error::Error + Send + Sync>> {
    let mut reader = Reader::from_reader(xml);
    let mut writer = Writer::new(Cursor::new(Vec::with_capacity(xml.len() + 1024)));

    let mut in_cell_xfs = false;
    let mut inside_xf = false;
    let mut current_xf_start = None;
    let mut current_xf_children = Vec::new();

    let mut count_val = 0;
    let mut current_xf_idx = 0;
    
    let mut yellow_ids = HashSet::new();
    let mut unlocked_ids = HashSet::new();
    let mut buf = Vec::new();

    loop {
        match reader.read_event_into(&mut buf) {
            Ok(Event::Start(e)) => {
                let name = e.name();
                if name.as_ref() == b"cellXfs" {
                    in_cell_xfs = true;
                    let mut new_e = BytesStart::new("cellXfs");
                    for attr in e.attributes().flatten() {
                        if attr.key.as_ref() == b"count" {
                            if let Ok(s) = std::str::from_utf8(&attr.value) {
                                if let Ok(count) = s.parse::<usize>() {
                                    count_val = count;
                                    new_e.push_attribute(("count", (count + 2).to_string().as_str()));
                                    continue;
                                }
                            }
                        }
                        new_e.push_attribute(attr);
                    }
                    writer.write_event(Event::Start(new_e))?;
                } else if in_cell_xfs && name.as_ref() == b"xf" {
                    inside_xf = true;
                    current_xf_start = Some(e.into_owned());
                    current_xf_children.clear();
                } else if inside_xf {
                    current_xf_children.push(Event::Start(e.into_owned()));
                } else {
                    writer.write_event(Event::Start(e.clone()))?;
                }
            }
            Ok(Event::Empty(e)) => {
                let name = e.name();
                if in_cell_xfs && name.as_ref() == b"xf" {
                    process_xf(
                        &mut writer, e.into_owned(), &[], target_fill_id, 
                        current_xf_idx, &mut yellow_ids, &mut unlocked_ids
                    )?;
                    current_xf_idx += 1;
                } else if inside_xf {
                    current_xf_children.push(Event::Empty(e.into_owned()));
                } else {
                    writer.write_event(Event::Empty(e.clone()))?;
                }
            }
            Ok(Event::End(e)) => {
                let name = e.name();
                if name.as_ref() == b"cellXfs" {
                    in_cell_xfs = false;
                    
                    let new_xf_unlocked = BytesStart::new("xf").with_attributes([
                        ("numFmtId", "0"), ("fontId", "0"), ("fillId", "0"), ("borderId", "0"), ("xfId", "0"),
                        ("applyAlignment", "false"), ("applyProtection", "true"),
                    ]);
                    writer.write_event(Event::Start(new_xf_unlocked))?;
                    writer.write_event(Event::Empty(BytesStart::new("alignment")))?;
                    writer.write_event(Event::Empty(
                        BytesStart::new("protection").with_attributes([("hidden", "false"), ("locked", "0")]),
                    ))?;
                    writer.write_event(Event::End(BytesEnd::new("xf")))?;

                    let new_xf_locked = BytesStart::new("xf").with_attributes([
                        ("numFmtId", "0"), ("fontId", "0"), ("fillId", "0"), ("borderId", "0"), ("xfId", "0"),
                        ("applyAlignment", "false"), ("applyProtection", "true"),
                    ]);
                    writer.write_event(Event::Start(new_xf_locked))?;
                    writer.write_event(Event::Empty(BytesStart::new("alignment")))?;
                    writer.write_event(Event::Empty(
                        BytesStart::new("protection").with_attributes([("hidden", "false"), ("locked", "1")]),
                    ))?;
                    writer.write_event(Event::End(BytesEnd::new("xf")))?;

                    writer.write_event(Event::End(e.clone()))?;
                } else if inside_xf && name.as_ref() == b"xf" {
                    process_xf(
                        &mut writer, current_xf_start.take().unwrap(), &current_xf_children, 
                        target_fill_id, current_xf_idx, &mut yellow_ids, &mut unlocked_ids
                    )?;
                    current_xf_idx += 1;
                    inside_xf = false;
                } else if inside_xf {
                    current_xf_children.push(Event::End(e.into_owned()));
                } else {
                    writer.write_event(Event::End(e.clone()))?;
                }
            }
            Ok(Event::Text(e)) => {
                if inside_xf {
                    current_xf_children.push(Event::Text(e.into_owned()));
                } else {
                    writer.write_event(Event::Text(e.clone()))?;
                }
            }
            Ok(Event::Eof) => break,
            Err(e) => return Err(e.into()),
            Ok(ev) => {
                if inside_xf {
                    current_xf_children.push(ev.into_owned());
                } else {
                    writer.write_event(ev)?;
                }
            }
        }
        buf.clear();
    }

    let unlocked_col_id = count_val;
    let explicit_lock_id = count_val + 1;
    let out = writer.into_inner().into_inner();
    Ok((out, unlocked_col_id, explicit_lock_id, yellow_ids, unlocked_ids))
}

fn process_xf(
    writer: &mut Writer<Cursor<Vec<u8>>>,
    start_event: BytesStart<'static>,
    children: &[Event<'static>],
    target_fill_id: usize,
    current_xf_idx: usize,
    yellow_ids: &mut HashSet<usize>,
    unlocked_ids: &mut HashSet<usize>,
) -> Result<(), Box<dyn std::error::Error + Send + Sync>> {
    let mut fill_id = None;
    let mut original_locked = None;
    let mut alignment_children = Vec::new();
    let mut other_children = Vec::new();

    for attr in start_event.attributes().flatten() {
        if attr.key.as_ref() == b"fillId" {
            fill_id = std::str::from_utf8(&attr.value).ok().and_then(|s| s.parse::<usize>().ok());
        }
    }

    let mut in_alignment = false;

    for child in children {
        match child {
            Event::Start(e) | Event::Empty(e) => {
                let name = e.name();
                if name.as_ref() == b"protection" {
                    for attr in e.attributes().flatten() {
                        if attr.key.as_ref() == b"locked" {
                            original_locked = Some(std::str::from_utf8(&attr.value).unwrap_or("").to_string());
                        }
                    }
                } else if name.as_ref() == b"alignment" {
                    in_alignment = matches!(child, Event::Start(_));
                    alignment_children.push(child.clone());
                } else {
                    if in_alignment { alignment_children.push(child.clone()); } 
                    else { other_children.push(child.clone()); }
                }
            }
            Event::End(e) => {
                let name = e.name();
                if name.as_ref() == b"alignment" {
                    alignment_children.push(child.clone());
                    in_alignment = false;
                } else if name.as_ref() != b"protection" {
                    if in_alignment { alignment_children.push(child.clone()); } 
                    else { other_children.push(child.clone()); }
                }
            }
            _ => {
                if in_alignment { alignment_children.push(child.clone()); } 
                else { other_children.push(child.clone()); }
            }
        }
    }

    let is_target_fill = fill_id == Some(target_fill_id);
    let mut target_locked = original_locked.unwrap_or_else(|| "1".to_string());

    if is_target_fill {
        target_locked = "0".to_string();
        yellow_ids.insert(current_xf_idx);
        unlocked_ids.insert(current_xf_idx);
    } else if target_locked == "0" || target_locked == "false" {
        unlocked_ids.insert(current_xf_idx);
    }

    let mut new_start = BytesStart::new("xf");
    for attr in start_event.attributes().flatten() {
        if attr.key.as_ref() != b"applyProtection" { new_start.push_attribute(attr); }
    }
    
    if is_target_fill {
        new_start.push_attribute(("applyProtection", "1"));
    } else {
        let mut had_apply = false;
        for attr in start_event.attributes().flatten() {
            if attr.key.as_ref() == b"applyProtection" {
                new_start.push_attribute(attr);
                had_apply = true;
            }
        }
        if !had_apply && target_locked == "0" {
             new_start.push_attribute(("applyProtection", "1"));
        }
    }

    if children.is_empty() && !is_target_fill && target_locked == "1" {
        writer.write_event(Event::Empty(new_start))?;
        return Ok(());
    }

    writer.write_event(Event::Start(new_start))?;
    for child in alignment_children { writer.write_event(child)?; }
    writer.write_event(Event::Empty(BytesStart::new("protection").with_attributes([("locked", target_locked.as_str())])))?;
    for child in other_children { writer.write_event(child)?; }
    writer.write_event(Event::End(BytesEnd::new("xf")))?;

    Ok(())
}

fn parse_col_event(e: &BytesStart, col_array: &mut Vec<Option<ColData>>) {
    let mut min = 0;
    let mut max = 0;
    let mut style = String::new();
    let mut attrs = Vec::new();

    for attr in e.attributes().flatten() {
        if let (Ok(k), Ok(v)) = (std::str::from_utf8(attr.key.as_ref()), std::str::from_utf8(attr.value.as_ref())) {
            if k == "min" { min = v.parse().unwrap_or(0); }
            else if k == "max" { max = v.parse().unwrap_or(0); }
            else if k == "style" { style = v.to_string(); }
            else { attrs.push((k.to_string(), v.to_string())); }
        }
    }

    if min > 0 && max >= min {
        let capped_max = std::cmp::min(max, 16384);
        for i in min..=capped_max {
            col_array[i] = Some(ColData { style: style.clone(), attrs: attrs.clone() });
        }
    }
}

fn generate_cols(
    writer: &mut Writer<Cursor<Vec<u8>>>,
    col_array: &mut Vec<Option<ColData>>,
    unlocked_col_id: usize,
) -> Result<(), Box<dyn std::error::Error + Send + Sync>> {
    for i in 1..=10 {
        if let Some(col) = &mut col_array[i] {
            col.style = unlocked_col_id.to_string();
        } else {
            col_array[i] = Some(ColData {
                style: unlocked_col_id.to_string(),
                attrs: vec![
                    ("customWidth".to_string(), "1".to_string()),
                    ("width".to_string(), "9.140625".to_string()),
                ],
            });
        }
    }

    let mut groups = Vec::new();
    let mut current_group: Option<(usize, usize, ColData)> = None;

    for i in 1..=16384 {
        match &col_array[i] {
            Some(data) => {
                match &mut current_group {
                    Some((_, current_max, current_data)) if current_data == data => {
                        *current_max = i;
                    }
                    _ => {
                        if let Some(g) = current_group.take() { groups.push(g); }
                        current_group = Some((i, i, data.clone()));
                    }
                }
            }
            None => {
                if let Some(g) = current_group.take() { groups.push(g); }
            }
        }
    }
    if let Some(g) = current_group.take() { groups.push(g); }

    if !groups.is_empty() {
        writer.write_event(Event::Start(BytesStart::new("cols")))?;
        for (min, max, data) in groups {
            let mut c = BytesStart::new("col");
            let min_str = min.to_string();
            let max_str = max.to_string();
            c.push_attribute(("min", min_str.as_str()));
            c.push_attribute(("max", max_str.as_str()));
            if !data.style.is_empty() { c.push_attribute(("style", data.style.as_str())); }
            for (k, v) in &data.attrs { c.push_attribute((k.as_str(), v.as_str())); }
            writer.write_event(Event::Empty(c))?;
        }
        writer.write_event(Event::End(BytesEnd::new("cols")))?;
    }
    Ok(())
}

fn rewrite_worksheet_xml(
    xml: &[u8],
    yellow_ids: &HashSet<usize>,
    unlocked_ids: &HashSet<usize>,
    explicit_lock_id: usize,
    unlocked_col_id: usize,
) -> Result<Vec<u8>, Box<dyn std::error::Error + Send + Sync>> {
    let mut reader = Reader::from_reader(xml);
    let mut writer = Writer::new(Cursor::new(Vec::with_capacity(xml.len() + 4096)));

    let mut in_c = false;
    let mut current_c_start: Option<BytesStart> = None;
    let mut c_style_idx = None;
    let mut c_has_content = false;
    let mut c_children = Vec::new();

    let mut in_cols = false;
    let mut saw_cols = false;
    let mut col_array: Vec<Option<ColData>> = vec![None; 16385];
    let mut buf = Vec::new();

    loop {
        match reader.read_event_into(&mut buf) {
            Ok(Event::Start(e)) => {
                let name = e.name();
                if name.as_ref() == b"cols" {
                    in_cols = true;
                    saw_cols = true;
                } else if name.as_ref() == b"sheetData" {
                    if !saw_cols {
                        generate_cols(&mut writer, &mut col_array, unlocked_col_id)?;
                        saw_cols = true;
                    }
                    writer.write_event(Event::Start(e.clone()))?;
                } else if name.as_ref() == b"c" {
                    in_c = true;
                    c_style_idx = None;
                    for attr in e.attributes().flatten() {
                        if attr.key.as_ref() == b"s" {
                            c_style_idx = std::str::from_utf8(&attr.value).ok().and_then(|s| s.parse::<usize>().ok());
                        }
                    }
                    c_has_content = false;
                    current_c_start = Some(e.into_owned());
                    c_children.clear();
                } else if in_c {
                    let name_ref = e.name();
                    if name_ref.as_ref() == b"v" || name_ref.as_ref() == b"f" || name_ref.as_ref() == b"is" {
                        c_has_content = true;
                    }
                    c_children.push(Event::Start(e.into_owned()));
                } else if !in_cols {
                    writer.write_event(Event::Start(e.clone()))?;
                }
            }
            Ok(Event::Empty(e)) => {
                let name = e.name();
                if in_cols && name.as_ref() == b"col" {
                    parse_col_event(&e, &mut col_array);
                } else if name.as_ref() == b"sheetData" {
                    if !saw_cols {
                        generate_cols(&mut writer, &mut col_array, unlocked_col_id)?;
                        saw_cols = true;
                    }
                    writer.write_event(Event::Empty(e.clone()))?;
                } else if name.as_ref() == b"c" {
                    writer.write_event(Event::Empty(e.clone()))?;
                } else if in_c {
                    c_children.push(Event::Empty(e.into_owned()));
                } else if !in_cols {
                    writer.write_event(Event::Empty(e.clone()))?;
                }
            }
            Ok(Event::End(e)) => {
                let name = e.name();
                if name.as_ref() == b"cols" {
                    in_cols = false;
                    generate_cols(&mut writer, &mut col_array, unlocked_col_id)?;
                } else if name.as_ref() == b"c" {
                    let start_ev = current_c_start.take().unwrap();
                    
                    if c_has_content {
                        let has_no_style = c_style_idx.is_none();
                        let actual_style = c_style_idx.unwrap_or(0);
                        let is_yellow = yellow_ids.contains(&actual_style);
                        let is_unlocked = unlocked_ids.contains(&actual_style);

                        if !is_yellow && (has_no_style || is_unlocked) {
                            let mut new_c = BytesStart::new("c");
                            for attr in start_ev.attributes().flatten() {
                                if attr.key.as_ref() != b"s" { new_c.push_attribute(attr); }
                            }
                            let lock_str = explicit_lock_id.to_string();
                            new_c.push_attribute(("s", lock_str.as_str()));
                            
                            writer.write_event(Event::Start(new_c))?;
                        } else {
                            writer.write_event(Event::Start(start_ev))?;
                        }
                    } else {
                        writer.write_event(Event::Start(start_ev))?;
                    }

                    for child in c_children.drain(..) { writer.write_event(child)?; }
                    writer.write_event(Event::End(e.clone()))?;
                    in_c = false;
                } else if in_c {
                    c_children.push(Event::End(e.into_owned()));
                } else if !in_cols {
                    writer.write_event(Event::End(e.clone()))?;
                }
            }
            Ok(Event::Text(e)) => {
                if in_c { c_children.push(Event::Text(e.into_owned())); } 
                else if !in_cols { writer.write_event(Event::Text(e.clone()))?; }
            }
            Ok(Event::Eof) => break,
            Err(e) => return Err(e.into()),
            Ok(ev) => {
                if in_c { c_children.push(ev.into_owned()); } 
                else if !in_cols { writer.write_event(ev)?; }
            }
        }
        buf.clear();
    }
    
    Ok(writer.into_inner().into_inner())
}

fn process_excel_file(input_path: &String, output_path: &String) -> Result<(), Box<dyn std::error::Error + Send + Sync>> {
    // std::fs::File is perfectly Send+Sync but standard Error needs to be bounded for rayon
    let file = File::open(input_path).map_err(|e| Box::new(e) as Box<dyn std::error::Error + Send + Sync>)?;
    let mut archive = ZipArchive::new(file).map_err(|e| Box::new(e) as Box<dyn std::error::Error + Send + Sync>)?;
    
    let out_file = File::create(output_path).map_err(|e| Box::new(e) as Box<dyn std::error::Error + Send + Sync>)?;
    let mut zip_writer = ZipWriter::new(out_file);
    let options = SimpleFileOptions::default().compression_method(zip::CompressionMethod::Deflated);

    let mut target_fill_id = None;
    let mut yellow_ids = HashSet::new();
    let mut unlocked_ids = HashSet::new();
    let mut new_style_id = 0;
    let mut explicit_lock_id = 0;
    let mut modified_styles_xml = Vec::new();

    if let Ok(mut styles_file) = archive.by_name("xl/styles.xml") {
        let mut content = Vec::new();
        styles_file.read_to_end(&mut content).unwrap_or_default();
        target_fill_id = find_fill_id_by_color(&content, "FFFAE5");
        
        let (new_xml, col_id, lock_id, y_ids, u_ids) = rewrite_styles_xml(&content, target_fill_id.unwrap_or(999999))?;
        modified_styles_xml = new_xml;
        new_style_id = col_id;
        explicit_lock_id = lock_id;
        yellow_ids = y_ids;
        unlocked_ids = u_ids;
    }

    for i in 0..archive.len() {
        let mut entry = archive.by_index(i).map_err(|e| Box::new(e) as Box<dyn std::error::Error + Send + Sync>)?;
        let name = entry.name().to_string();
        
        if name == "xl/styles.xml" {
            zip_writer.start_file(name, options).map_err(|e| Box::new(e) as Box<dyn std::error::Error + Send + Sync>)?;
            zip_writer.write_all(&modified_styles_xml).map_err(|e| Box::new(e) as Box<dyn std::error::Error + Send + Sync>)?;
        } else if name.starts_with("xl/worksheets/") && name.ends_with(".xml") {
            let mut content = Vec::new();
            entry.read_to_end(&mut content).map_err(|e| Box::new(e) as Box<dyn std::error::Error + Send + Sync>)?;
            
            let new_sheet_xml = rewrite_worksheet_xml(&content, &yellow_ids, &unlocked_ids, explicit_lock_id, new_style_id)?;
            
            zip_writer.start_file(name, options).map_err(|e| Box::new(e) as Box<dyn std::error::Error + Send + Sync>)?;
            zip_writer.write_all(&new_sheet_xml).map_err(|e| Box::new(e) as Box<dyn std::error::Error + Send + Sync>)?;
        } else {
            let mut bytes = Vec::new();
            entry.read_to_end(&mut bytes).map_err(|e| Box::new(e) as Box<dyn std::error::Error + Send + Sync>)?;
            zip_writer.start_file(name, options).map_err(|e| Box::new(e) as Box<dyn std::error::Error + Send + Sync>)?;
            zip_writer.write_all(&bytes).map_err(|e| Box::new(e) as Box<dyn std::error::Error + Send + Sync>)?;
        }
    }

    zip_writer.finish().map_err(|e| Box::new(e) as Box<dyn std::error::Error + Send + Sync>)?;
    Ok(())
}

fn main() -> Result<(), Box<dyn std::error::Error>> {
    let original_file = "../../testdata/Pruefvorlage/Pruefvorlage_2026_0004_001_deutsch_1.xlsx";
    let in_dir = "../../tmp/speed_in";
    let out_dir = "../../tmp/speed_out";

    println!("Bereite Verzeichnisse vor...");
    fs::create_dir_all(in_dir)?;
    fs::create_dir_all(out_dir)?;

    println!("Kopiere Ursprungsdatei 400 mal...");
    for i in 0..400 {
        let _ = fs::copy(original_file, format!("{}/test_{}.xlsx", in_dir, i));
    }

    // Wir erstellen eine Liste aller Aufgaben
    let mut tasks = Vec::new();
    for i in 0..400 {
        tasks.push((
            format!("{}/test_{}.xlsx", in_dir, i),
            format!("{}/test_{}.xlsx", out_dir, i)
        ));
    }

    println!("Starte PARALLELEN Speedtest: Verarbeite 400 Excel-Dateien gleichzeitig mit Rayon...");
    let start = Instant::now();

    // Der magische Aufruf: par_iter() teilt die Liste automatisch auf alle verfügbaren CPU-Kerne auf!
    tasks.par_iter().for_each(|(in_path, out_path)| {
        if let Err(e) = process_excel_file(in_path, out_path) {
            eprintln!("Fehler bei {}: {}", in_path, e);
        }
    });

    let duration = start.elapsed();
    println!("--------------------------------------------------");
    println!("FERTIG!");
    println!("Gesamtdauer für 400 Dateien (Parallel): {:?}", duration);
    println!("--------------------------------------------------");

    Ok(())
}
