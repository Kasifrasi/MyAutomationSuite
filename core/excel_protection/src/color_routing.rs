use crate::ProtectionError;
use quick_xml::events::{BytesEnd, BytesStart, Event};
use quick_xml::reader::Reader;
use quick_xml::writer::Writer;
use std::collections::HashSet;
use std::fs::File;
use std::io::{Cursor, Read, Write};
use std::path::Path;
use zip::write::FileOptions;
use zip::{ZipArchive, ZipWriter};

pub fn apply_color_routing_protection(
    path: &Path,
    target_hex_color: &str,
) -> Result<(), ProtectionError> {
    let file = File::open(path)?;
    let mut archive = ZipArchive::new(file)?;

    // 1. Find fill ID for target color
    let mut target_fill_id = None;
    if let Ok(mut styles_file) = archive.by_name("xl/styles.xml") {
        let mut content = String::new();
        styles_file.read_to_string(&mut content).unwrap_or_default();
        target_fill_id = find_fill_id_by_color(&content, target_hex_color);
    }

    let target_fill_id = match target_fill_id {
        Some(id) => id,
        None => return Ok(()), // Color not found, nothing to do
    };

    // 2. Identify yellow style IDs and unlocked style IDs, and rewrite styles.xml
    let mut yellow_style_ids = HashSet::new();
    let mut unlocked_style_ids = HashSet::new();
    let mut modified_styles_xml = String::new();
    let mut explicit_lock_style_id = 0;

    if let Ok(mut styles_file) = archive.by_name("xl/styles.xml") {
        let mut content = String::new();
        styles_file.read_to_string(&mut content).unwrap_or_default();
        let (new_xml, lock_id) = rewrite_styles_xml(
            &content,
            target_fill_id,
            &mut yellow_style_ids,
            &mut unlocked_style_ids,
        )?;
        modified_styles_xml = new_xml;
        explicit_lock_style_id = lock_id;
    }

    // 3. Rewrite all files back to ZIP, applying sheet*.xml modification
    let temp_path = path.with_extension("tmp_routing");
    let out_file = File::create(&temp_path)?;
    let mut zip_writer = ZipWriter::new(out_file);

    for i in 0..archive.len() {
        let mut entry = archive.by_index(i)?;
        let name = entry.name().to_string();
        let compression = entry.compression();
        let unix_mode = entry.unix_mode();

        let options = FileOptions::<()>::default()
            .compression_method(compression)
            .unix_permissions(unix_mode.unwrap_or(0o644));

        if name == "xl/styles.xml" {
            zip_writer.start_file(&name, options)?;
            zip_writer.write_all(modified_styles_xml.as_bytes())?;
        } else if name.starts_with("xl/worksheets/") && name.ends_with(".xml") {
            let mut content = Vec::new();
            entry.read_to_end(&mut content)?;
            let new_xml = rewrite_worksheet_xml(
                &content,
                &yellow_style_ids,
                &unlocked_style_ids,
                explicit_lock_style_id,
            )?;
            zip_writer.start_file(&name, options)?;
            zip_writer.write_all(&new_xml)?;
        } else {
            zip_writer.raw_copy_file(entry)?;
        }
    }

    zip_writer.finish()?;
    std::fs::rename(&temp_path, path)?;
    Ok(())
}

pub fn find_fill_id_by_color(xml: &str, target_color: &str) -> Option<usize> {
    let mut reader = Reader::from_str(xml);
    let mut fill_idx = 0;
    let mut in_fill = false;
    let mut in_fills = false;
    let target_upper = target_color.to_uppercase();

    loop {
        match reader.read_event() {
            Ok(Event::Start(ref e)) => {
                let name = e.name();
                let name_ref = name.as_ref();
                if name_ref == b"fills" {
                    in_fills = true;
                } else if in_fills && name_ref == b"fill" {
                    in_fill = true;
                } else if in_fill && (name_ref == b"fgColor" || name_ref == b"bgColor") {
                    for attr in e.attributes().flatten() {
                        if attr.key.as_ref() == b"rgb" {
                            let val = std::str::from_utf8(&attr.value)
                                .unwrap_or("")
                                .to_uppercase();
                            if val.ends_with(&target_upper) {
                                return Some(fill_idx);
                            }
                        }
                    }
                }
            }
            Ok(Event::Empty(ref e)) => {
                let name = e.name();
                let name_ref = name.as_ref();
                if in_fills && name_ref == b"fill" {
                    fill_idx += 1;
                } else if in_fill && (name_ref == b"fgColor" || name_ref == b"bgColor") {
                    for attr in e.attributes().flatten() {
                        if attr.key.as_ref() == b"rgb" {
                            let val = std::str::from_utf8(&attr.value)
                                .unwrap_or("")
                                .to_uppercase();
                            if val.ends_with(&target_upper) {
                                return Some(fill_idx);
                            }
                        }
                    }
                }
            }
            Ok(Event::End(ref e)) => {
                let name = e.name();
                let name_ref = name.as_ref();
                if name_ref == b"fills" {
                    in_fills = false;
                } else if in_fills && name_ref == b"fill" {
                    fill_idx += 1;
                    in_fill = false;
                }
            }
            Ok(Event::Eof) | Err(_) => break,
            _ => {}
        }
    }
    None
}

pub fn rewrite_styles_xml(
    xml: &str,
    target_fill_id: usize,
    yellow_style_ids: &mut HashSet<usize>,
    unlocked_style_ids: &mut HashSet<usize>,
) -> Result<(String, usize), ProtectionError> {
    let mut reader = Reader::from_str(xml);
    let mut writer = Writer::new(Cursor::new(Vec::new()));

    let mut in_cell_xfs = false;
    let mut current_xf_idx = 0;

    let mut inside_xf = false;
    let mut current_xf_start = None;
    let mut current_xf_children = Vec::new();

    loop {
        match reader.read_event() {
            Ok(Event::Start(e)) => {
                let name = e.name();
                let name_ref = name.as_ref();
                if name_ref == b"cellXfs" {
                    in_cell_xfs = true;
                    let mut new_e = BytesStart::new("cellXfs");
                    for attr in e.attributes().flatten() {
                        if attr.key.as_ref() == b"count" {
                            if let Ok(s) = std::str::from_utf8(&attr.value) {
                                if let Ok(count) = s.parse::<usize>() {
                                    new_e.push_attribute((
                                        "count",
                                        (count + 1).to_string().as_str(),
                                    ));
                                    continue;
                                }
                            }
                        }
                        new_e.push_attribute(attr);
                    }
                    writer.write_event(Event::Start(new_e))?;
                } else if in_cell_xfs && name_ref == b"xf" {
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
                let name_ref = name.as_ref();
                if in_cell_xfs && name_ref == b"xf" {
                    current_xf_start = Some(e.into_owned());
                    current_xf_children.clear();

                    process_xf(
                        &mut writer,
                        current_xf_start.take().unwrap(),
                        &current_xf_children,
                        current_xf_idx,
                        target_fill_id,
                        yellow_style_ids,
                        unlocked_style_ids,
                    )?;

                    current_xf_idx += 1;
                    inside_xf = false;
                } else if inside_xf {
                    current_xf_children.push(Event::Empty(e.into_owned()));
                } else {
                    writer.write_event(Event::Empty(e.clone()))?;
                }
            }
            Ok(Event::End(e)) => {
                let name = e.name();
                let name_ref = name.as_ref();
                if name_ref == b"cellXfs" {
                    in_cell_xfs = false;

                    let new_xf = BytesStart::new("xf").with_attributes([
                        ("numFmtId", "0"),
                        ("fontId", "0"),
                        ("fillId", "0"),
                        ("borderId", "0"),
                        ("xfId", "0"),
                        ("applyProtection", "1"),
                    ]);
                    writer.write_event(Event::Start(new_xf))?;
                    writer.write_event(Event::Empty(
                        BytesStart::new("protection").with_attributes([("locked", "1")]),
                    ))?;
                    writer.write_event(Event::End(BytesEnd::new("xf")))?;

                    writer.write_event(Event::End(e.clone()))?;
                } else if inside_xf && name_ref == b"xf" {
                    process_xf(
                        &mut writer,
                        current_xf_start.take().unwrap(),
                        &current_xf_children,
                        current_xf_idx,
                        target_fill_id,
                        yellow_style_ids,
                        unlocked_style_ids,
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
    }

    let out = writer.into_inner().into_inner();
    Ok((String::from_utf8(out).unwrap_or_default(), current_xf_idx))
}

fn process_xf(
    writer: &mut Writer<Cursor<Vec<u8>>>,
    start_event: BytesStart<'static>,
    children: &[Event<'static>],
    current_xf_idx: usize,
    target_fill_id: usize,
    yellow_style_ids: &mut HashSet<usize>,
    unlocked_style_ids: &mut HashSet<usize>,
) -> Result<(), ProtectionError> {
    let mut fill_id = None;
    let mut original_locked = None;
    let mut alignment_children = Vec::new();
    let mut other_children = Vec::new();

    for attr in start_event.attributes().flatten() {
        if attr.key.as_ref() == b"fillId" {
            fill_id = std::str::from_utf8(&attr.value)
                .ok()
                .and_then(|s| s.parse::<usize>().ok());
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
                            original_locked =
                                Some(std::str::from_utf8(&attr.value).unwrap_or("").to_string());
                        }
                    }
                } else if name.as_ref() == b"alignment" {
                    in_alignment = matches!(child, Event::Start(_));
                    alignment_children.push(child.clone());
                } else {
                    if in_alignment {
                        alignment_children.push(child.clone());
                    } else {
                        other_children.push(child.clone());
                    }
                }
            }
            Event::End(e) => {
                let name = e.name();
                if name.as_ref() == b"alignment" {
                    alignment_children.push(child.clone());
                    in_alignment = false;
                } else if name.as_ref() != b"protection" {
                    if in_alignment {
                        alignment_children.push(child.clone());
                    } else {
                        other_children.push(child.clone());
                    }
                }
            }
            _ => {
                if in_alignment {
                    alignment_children.push(child.clone());
                } else {
                    other_children.push(child.clone());
                }
            }
        }
    }

    let is_target_fill = fill_id == Some(target_fill_id);
    let mut target_locked = "1";

    if is_target_fill {
        target_locked = "0";
        yellow_style_ids.insert(current_xf_idx);
        unlocked_style_ids.insert(current_xf_idx);
    } else if original_locked.as_deref() == Some("0") || original_locked.as_deref() == Some("false")
    {
        // preserve originally unlocked styles (like the column default)
        target_locked = "0";
        unlocked_style_ids.insert(current_xf_idx);
    }

    let mut new_start = BytesStart::new("xf");
    for attr in start_event.attributes().flatten() {
        if attr.key.as_ref() != b"applyProtection" {
            new_start.push_attribute(attr);
        }
    }
    new_start.push_attribute(("applyProtection", "1"));

    writer.write_event(Event::Start(new_start))?;

    for child in alignment_children {
        writer.write_event(child)?;
    }

    writer.write_event(Event::Empty(
        BytesStart::new("protection").with_attributes([("locked", target_locked)]),
    ))?;

    for child in other_children {
        writer.write_event(child)?;
    }

    writer.write_event(Event::End(BytesEnd::new("xf")))?;

    Ok(())
}

pub fn rewrite_worksheet_xml(
    xml: &[u8],
    yellow_style_ids: &HashSet<usize>,
    unlocked_style_ids: &HashSet<usize>,
    explicit_lock_style_id: usize,
) -> Result<Vec<u8>, ProtectionError> {
    let mut reader = Reader::from_reader(xml);
    let mut writer = Writer::new(Cursor::new(Vec::new()));

    let mut in_c = false;
    let mut current_c_start = None;
    let mut c_style_idx = None;
    let mut c_has_content = false;
    let mut c_children = Vec::new();

    loop {
        match reader.read_event() {
            Ok(Event::Start(e)) => {
                let name = e.name();
                let name_ref = name.as_ref();
                if name_ref == b"c" {
                    in_c = true;
                    c_style_idx = None;
                    for attr in e.attributes().flatten() {
                        if attr.key.as_ref() == b"s" {
                            c_style_idx = std::str::from_utf8(&attr.value)
                                .ok()
                                .and_then(|s| s.parse::<usize>().ok());
                        }
                    }
                    c_has_content = false;
                    current_c_start = Some(e.into_owned());
                    c_children.clear();
                } else if in_c {
                    if name_ref == b"v" || name_ref == b"f" || name_ref == b"is" {
                        c_has_content = true;
                    }
                    c_children.push(Event::Start(e.into_owned()));
                } else {
                    writer.write_event(Event::Start(e.clone()))?;
                }
            }
            Ok(Event::Empty(e)) => {
                let name = e.name();
                let name_ref = name.as_ref();
                if name_ref == b"c" {
                    writer.write_event(Event::Empty(e.clone()))?;
                } else if in_c {
                    if name_ref == b"v" || name_ref == b"f" || name_ref == b"is" {
                        c_has_content = true;
                    }
                    c_children.push(Event::Empty(e.into_owned()));
                } else {
                    writer.write_event(Event::Empty(e.clone()))?;
                }
            }
            Ok(Event::End(e)) => {
                let name = e.name();
                let name_ref = name.as_ref();
                if name_ref == b"c" {
                    in_c = false;
                    let mut start_e = current_c_start.take().unwrap();

                    if c_has_content {
                        let actual_style = c_style_idx.unwrap_or(0);
                        // If cell style is NOT yellow, and the style is an unlocked style (or it has no style, which falls back to column style which might be unlocked), force it to s="99".
                        // Wait, if it has no style (actual_style == 0) and style 0 is locked, it's fine.
                        // But wait! If it has NO style, it inherits the COLUMN style. So we must force s="99" to break the column inheritance!
                        let has_no_style = c_style_idx.is_none();
                        let is_yellow = yellow_style_ids.contains(&actual_style);
                        let is_unlocked = unlocked_style_ids.contains(&actual_style);

                        if !is_yellow && (is_unlocked || has_no_style) {
                            // Strip existing "s" attr and add explicit lock style
                            let mut new_start = BytesStart::new("c");
                            for attr in start_e.attributes().flatten() {
                                if attr.key.as_ref() != b"s" {
                                    new_start.push_attribute(attr);
                                }
                            }
                            new_start
                                .push_attribute(("s", explicit_lock_style_id.to_string().as_str()));
                            start_e = new_start;
                        }
                    }

                    writer.write_event(Event::Start(start_e))?;
                    for child in c_children.drain(..) {
                        writer.write_event(child)?;
                    }
                    writer.write_event(Event::End(e.clone()))?;
                } else if in_c {
                    c_children.push(Event::End(e.into_owned()));
                } else {
                    writer.write_event(Event::End(e.clone()))?;
                }
            }
            Ok(Event::Text(e)) => {
                if in_c {
                    c_children.push(Event::Text(e.into_owned()));
                } else {
                    writer.write_event(Event::Text(e.clone()))?;
                }
            }
            Ok(Event::Eof) => break,
            Err(e) => return Err(e.into()),
            Ok(ev) => {
                if in_c {
                    c_children.push(ev.into_owned());
                } else {
                    writer.write_event(ev)?;
                }
            }
        }
    }

    Ok(writer.into_inner().into_inner())
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_modify_styles_xml() {
        let xml = r#"<styleSheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main"><fills count="3"><fill><patternFill patternType="none"/></fill><fill><patternFill patternType="gray125"/></fill><fill><patternFill patternType="solid"><fgColor rgb="FFFAE5"/><bgColor indexed="64"/></patternFill></fill></fills><cellXfs count="3"><xf numFmtId="0" fontId="0" fillId="0" borderId="0" xfId="0"/><xf numFmtId="0" fontId="0" fillId="0" borderId="0" xfId="0" applyProtection="1"><protection locked="0"/></xf><xf numFmtId="0" fontId="0" fillId="2" borderId="0" xfId="0"/></cellXfs></styleSheet>"#;

        let target_id = find_fill_id_by_color(xml, "FFFAE5").unwrap();
        assert_eq!(target_id, 2);

        let mut yellow_ids = HashSet::new();
        let mut unlocked_ids = HashSet::new();

        let new_xml =
            rewrite_styles_xml(xml, target_id, &mut yellow_ids, &mut unlocked_ids).unwrap();

        println!("yellow_ids: {:?}", yellow_ids);
        assert!(yellow_ids.contains(&2));
        assert!(unlocked_ids.contains(&1));
        assert!(unlocked_ids.contains(&2));
        assert!(!unlocked_ids.contains(&0));

        assert!(new_xml.0.contains(r#"<xf numFmtId="0" fontId="0" fillId="0" borderId="0" xfId="0" applyProtection="1"><protection locked="1"/></xf>"#));
        assert!(new_xml.0.contains(r#"<xf numFmtId="0" fontId="0" fillId="0" borderId="0" xfId="0" applyProtection="1"><protection locked="0"/></xf>"#));
        assert!(new_xml.0.contains(r#"<xf numFmtId="0" fontId="0" fillId="2" borderId="0" xfId="0" applyProtection="1"><protection locked="0"/></xf>"#));
    }

    #[test]
    fn test_modify_worksheet_xml() {
        // C5 has s="2" (yellow style). B5 has content but no style (should get s="99").
        // D5 has content, s="1" (unlocked column style, should get s="99").
        let xml = br#"<worksheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main"><sheetData><row r="5"><c r="B5"><v>Projekt-Titel</v></c><c r="C5" s="2"><v>Eingabewert</v></c><c r="D5" s="1"><v>Another</v></c></row></sheetData></worksheet>"#;

        let mut yellow_ids = HashSet::new();
        yellow_ids.insert(2);
        let mut unlocked_ids = HashSet::new();
        unlocked_ids.insert(1);
        unlocked_ids.insert(2);

        let new_xml = rewrite_worksheet_xml(xml, &yellow_ids, &unlocked_ids, 99).unwrap();
        let new_str = String::from_utf8(new_xml).unwrap();

        assert!(new_str.contains(r#"<c r="B5" s="99"><v>Projekt-Titel</v></c>"#));
        assert!(new_str.contains(r#"<c r="C5" s="2"><v>Eingabewert</v></c>"#));
        assert!(new_str.contains(r#"<c r="D5" s="99"><v>Another</v></c>"#));
    }

    #[test]
    fn test_fixed_values() {
        let xml = br#"<worksheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main"><sheetData><row r="2"><c r="B2" s="1" t="s"><v>0</v></c></row><row r="3"><c r="B3" s="1"><v>42</v></c></row><row r="4"><c r="B4" t="str"><f>B3*2</f></c></row></sheetData></worksheet>"#;
        let mut yellow_ids = HashSet::new();
        let mut unlocked_ids = HashSet::new();
        unlocked_ids.insert(1); // Style 1 is unlocked

        let new_xml = rewrite_worksheet_xml(xml, &yellow_ids, &unlocked_ids, 99).unwrap();
        let new_str = String::from_utf8(new_xml).unwrap();

        println!("{}", new_str);
        assert!(
            new_str.contains(r#"<c r="B2" t="s" s="99">"#)
                || new_str.contains(r#"<c r="B2" s="99" t="s">"#)
        );
    }
}
