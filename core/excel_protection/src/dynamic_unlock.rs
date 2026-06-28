use crate::ProtectionError;
use quick_xml::events::{BytesEnd, BytesStart, Event};
use quick_xml::reader::Reader;
use quick_xml::writer::Writer;
use std::collections::HashSet;
use std::io::{Cursor, Write};

#[derive(Clone, PartialEq)]
struct ColData {
    style: String,
    attrs: Vec<(String, String)>,
}

pub fn rewrite_styles_xml(
    xml: &[u8],
    target_hex_color: &str,
) -> Result<(Vec<u8>, usize, usize, HashSet<usize>, HashSet<usize>), ProtectionError> {
    let mut reader = Reader::from_reader(xml);
    let mut writer = Writer::new(Cursor::new(Vec::with_capacity(xml.len() + 1024)));

    let target_upper = target_hex_color.to_uppercase();
    let mut target_fill_id = None;
    let mut fill_idx = 0;
    let mut in_fills = false;
    let mut in_fill = false;

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
                let name_ref = name.as_ref();
                if name_ref == b"fills" {
                    in_fills = true;
                    writer.write_event(Event::Start(e.clone()))?;
                } else if in_fills && name_ref == b"fill" {
                    in_fill = true;
                    writer.write_event(Event::Start(e.clone()))?;
                } else if in_fill && (name_ref == b"fgColor" || name_ref == b"bgColor") {
                    for attr in e.attributes().flatten() {
                        if attr.key.as_ref() == b"rgb" {
                            let val = std::str::from_utf8(&attr.value)
                                .unwrap_or("")
                                .to_uppercase();
                            if val.ends_with(&target_upper) {
                                target_fill_id = Some(fill_idx);
                            }
                        }
                    }
                    writer.write_event(Event::Start(e.clone()))?;
                } else if name_ref == b"cellXfs" {
                    in_cell_xfs = true;
                    let mut new_e = BytesStart::new("cellXfs");
                    for attr in e.attributes().flatten() {
                        if attr.key.as_ref() == b"count" {
                            if let Ok(s) = std::str::from_utf8(&attr.value) {
                                if let Ok(count) = s.parse::<usize>() {
                                    count_val = count;
                                    new_e.push_attribute((
                                        "count",
                                        (count + 2).to_string().as_str(),
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
                if in_fills && name_ref == b"fill" {
                    fill_idx += 1;
                    writer.write_event(Event::Empty(e.clone()))?;
                } else if in_fill && (name_ref == b"fgColor" || name_ref == b"bgColor") {
                    for attr in e.attributes().flatten() {
                        if attr.key.as_ref() == b"rgb" {
                            let val = std::str::from_utf8(&attr.value)
                                .unwrap_or("")
                                .to_uppercase();
                            if val.ends_with(&target_upper) {
                                target_fill_id = Some(fill_idx);
                            }
                        }
                    }
                    writer.write_event(Event::Empty(e.clone()))?;
                } else if in_cell_xfs && name_ref == b"xf" {
                    process_xf(
                        &mut writer,
                        e.into_owned(),
                        &[],
                        target_fill_id,
                        current_xf_idx,
                        &mut yellow_ids,
                        &mut unlocked_ids,
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
                let name_ref = name.as_ref();
                if name_ref == b"fills" {
                    in_fills = false;
                    writer.write_event(Event::End(e.clone()))?;
                } else if in_fills && name_ref == b"fill" {
                    fill_idx += 1;
                    in_fill = false;
                    writer.write_event(Event::End(e.clone()))?;
                } else if name_ref == b"cellXfs" {
                    in_cell_xfs = false;

                    let new_xf_unlocked = BytesStart::new("xf").with_attributes([
                        ("numFmtId", "0"),
                        ("fontId", "0"),
                        ("fillId", "0"),
                        ("borderId", "0"),
                        ("xfId", "0"),
                        ("applyAlignment", "false"),
                        ("applyProtection", "true"),
                    ]);
                    writer.write_event(Event::Start(new_xf_unlocked))?;
                    writer.write_event(Event::Empty(BytesStart::new("alignment")))?;
                    writer.write_event(Event::Empty(
                        BytesStart::new("protection")
                            .with_attributes([("hidden", "false"), ("locked", "0")]),
                    ))?;
                    writer.write_event(Event::End(BytesEnd::new("xf")))?;

                    let new_xf_locked = BytesStart::new("xf").with_attributes([
                        ("numFmtId", "0"),
                        ("fontId", "0"),
                        ("fillId", "0"),
                        ("borderId", "0"),
                        ("xfId", "0"),
                        ("applyAlignment", "false"),
                        ("applyProtection", "true"),
                    ]);
                    writer.write_event(Event::Start(new_xf_locked))?;
                    writer.write_event(Event::Empty(BytesStart::new("alignment")))?;
                    writer.write_event(Event::Empty(
                        BytesStart::new("protection")
                            .with_attributes([("hidden", "false"), ("locked", "1")]),
                    ))?;
                    writer.write_event(Event::End(BytesEnd::new("xf")))?;

                    writer.write_event(Event::End(e.clone()))?;
                } else if inside_xf && name_ref == b"xf" {
                    process_xf(
                        &mut writer,
                        current_xf_start.take().unwrap(),
                        &current_xf_children,
                        target_fill_id,
                        current_xf_idx,
                        &mut yellow_ids,
                        &mut unlocked_ids,
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
    Ok((
        out,
        unlocked_col_id,
        explicit_lock_id,
        yellow_ids,
        unlocked_ids,
    ))
}

fn process_xf(
    writer: &mut Writer<Cursor<Vec<u8>>>,
    start_event: BytesStart<'static>,
    children: &[Event<'static>],
    target_fill_id: Option<usize>,
    current_xf_idx: usize,
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

    let is_target_fill = target_fill_id.is_some() && fill_id == target_fill_id;
    let mut target_locked = "1";

    if is_target_fill {
        target_locked = "0";
        yellow_style_ids.insert(current_xf_idx);
        unlocked_style_ids.insert(current_xf_idx);
    } else if original_locked.as_deref() == Some("0") || original_locked.as_deref() == Some("false")
    {
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

fn parse_col_event(e: &BytesStart, col_array: &mut Vec<Option<ColData>>) {
    let mut min = 0;
    let mut max = 0;
    let mut style = String::new();
    let mut attrs = Vec::new();

    for attr in e.attributes().flatten() {
        if let (Ok(k), Ok(v)) = (
            std::str::from_utf8(attr.key.as_ref()),
            std::str::from_utf8(attr.value.as_ref()),
        ) {
            if k == "min" {
                min = v.parse().unwrap_or(0);
            } else if k == "max" {
                max = v.parse().unwrap_or(0);
            } else if k == "style" {
                style = v.to_string();
            } else {
                attrs.push((k.to_string(), v.to_string()));
            }
        }
    }

    if min > 0 && max >= min {
        let capped_max = std::cmp::min(max, 16384);
        for i in min..=capped_max {
            col_array[i] = Some(ColData {
                style: style.clone(),
                attrs: attrs.clone(),
            });
        }
    }
}

fn generate_cols(
    writer: &mut Writer<Cursor<Vec<u8>>>,
    col_array: &mut Vec<Option<ColData>>,
    unlocked_col_id: usize,
    columns_to_unlock: &[u32],
) -> Result<(), ProtectionError> {
    for &col_idx in columns_to_unlock {
        let i = col_idx as usize;
        if i == 0 || i > 16384 {
            continue;
        }
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
            Some(data) => match &mut current_group {
                Some((_, current_max, current_data)) if current_data == data => {
                    *current_max = i;
                }
                _ => {
                    if let Some(g) = current_group.take() {
                        groups.push(g);
                    }
                    current_group = Some((i, i, data.clone()));
                }
            },
            None => {
                if let Some(g) = current_group.take() {
                    groups.push(g);
                }
            }
        }
    }
    if let Some(g) = current_group.take() {
        groups.push(g);
    }

    if !groups.is_empty() {
        writer.write_event(Event::Start(BytesStart::new("cols")))?;
        for (min, max, data) in groups {
            let mut c = BytesStart::new("col");
            let min_str = min.to_string();
            let max_str = max.to_string();
            c.push_attribute(("min", min_str.as_str()));
            c.push_attribute(("max", max_str.as_str()));
            if !data.style.is_empty() {
                c.push_attribute(("style", data.style.as_str()));
            }
            for (k, v) in &data.attrs {
                c.push_attribute((k.as_str(), v.as_str()));
            }
            writer.write_event(Event::Empty(c))?;
        }
        writer.write_event(Event::End(BytesEnd::new("cols")))?;
    }
    Ok(())
}

pub fn rewrite_worksheet_xml(
    xml: &[u8],
    yellow_style_ids: &HashSet<usize>,
    unlocked_style_ids: &HashSet<usize>,
    explicit_lock_style_id: usize,
    unlocked_col_id: usize,
    columns_to_unlock: &[u32],
) -> Result<Vec<u8>, ProtectionError> {
    let mut reader = Reader::from_reader(xml);
    let mut writer = Writer::new(Cursor::new(Vec::with_capacity(xml.len() + 4096)));

    let mut in_c = false;
    let mut in_cols = false;
    let mut has_written_cols = false;

    let mut cell_buffer = Vec::with_capacity(512);

    let mut c_style_idx = None;
    let mut c_has_content = false;
    let mut current_c_start = None;

    let mut col_array: Vec<Option<ColData>> = vec![None; 16385];
    let mut buf = Vec::new();

    loop {
        match reader.read_event_into(&mut buf) {
            Ok(Event::Start(e)) => {
                let name = e.name();
                let name_ref = name.as_ref();
                if name_ref == b"cols" {
                    in_cols = true;
                } else if in_cols && name_ref == b"col" {
                    parse_col_event(&e, &mut col_array);
                } else if name_ref == b"sheetData" {
                    if !has_written_cols {
                        generate_cols(
                            &mut writer,
                            &mut col_array,
                            unlocked_col_id,
                            columns_to_unlock,
                        )?;
                        has_written_cols = true;
                    }
                    writer.write_event(Event::Start(e.clone()))?;
                } else if name_ref == b"c" {
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
                    cell_buffer.clear();
                } else if in_c {
                    if name_ref == b"v" || name_ref == b"f" || name_ref == b"is" {
                        c_has_content = true;
                    }
                    let mut cell_writer = Writer::new(&mut cell_buffer);
                    cell_writer.write_event(Event::Start(e.clone())).unwrap();
                } else {
                    writer.write_event(Event::Start(e.clone()))?;
                }
            }
            Ok(Event::Empty(e)) => {
                let name = e.name();
                let name_ref = name.as_ref();
                if in_cols && name_ref == b"col" {
                    parse_col_event(&e, &mut col_array);
                } else if in_cols {
                    // skip other tags in cols
                } else if name_ref == b"sheetData" {
                    if !has_written_cols {
                        generate_cols(
                            &mut writer,
                            &mut col_array,
                            unlocked_col_id,
                            columns_to_unlock,
                        )?;
                        has_written_cols = true;
                    }
                    writer.write_event(Event::Empty(e.clone()))?;
                } else if name_ref == b"c" {
                    writer.write_event(Event::Empty(e.clone()))?;
                } else if in_c {
                    if name_ref == b"v" || name_ref == b"f" || name_ref == b"is" {
                        c_has_content = true;
                    }
                    let mut cell_writer = Writer::new(&mut cell_buffer);
                    cell_writer.write_event(Event::Empty(e.clone())).unwrap();
                } else {
                    writer.write_event(Event::Empty(e.clone()))?;
                }
            }
            Ok(Event::End(e)) => {
                let name = e.name();
                let name_ref = name.as_ref();
                if name_ref == b"cols" {
                    in_cols = false;
                    generate_cols(
                        &mut writer,
                        &mut col_array,
                        unlocked_col_id,
                        columns_to_unlock,
                    )?;
                    has_written_cols = true;
                } else if in_cols {
                    // skip
                } else if name_ref == b"c" {
                    in_c = false;
                    let mut start_e = current_c_start.take().unwrap();

                    if c_has_content {
                        let actual_style = c_style_idx.unwrap_or(0);
                        let has_no_style = c_style_idx.is_none();
                        let is_yellow = yellow_style_ids.contains(&actual_style);
                        let is_unlocked = unlocked_style_ids.contains(&actual_style);

                        if !is_yellow && (is_unlocked || has_no_style) {
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
                    writer.get_mut().write_all(&cell_buffer).unwrap();
                    writer.write_event(Event::End(e.clone()))?;
                } else if in_c {
                    let mut cell_writer = Writer::new(&mut cell_buffer);
                    cell_writer.write_event(Event::End(e.clone())).unwrap();
                } else {
                    writer.write_event(Event::End(e.clone()))?;
                }
            }
            Ok(Event::Text(e)) => {
                if in_cols {
                    // skip
                } else if in_c {
                    let mut cell_writer = Writer::new(&mut cell_buffer);
                    cell_writer.write_event(Event::Text(e.clone())).unwrap();
                } else {
                    writer.write_event(Event::Text(e.clone()))?;
                }
            }
            Ok(Event::Eof) => break,
            Err(e) => return Err(e.into()),
            Ok(ev) => {
                if in_cols {
                    // skip
                } else if in_c {
                    let mut cell_writer = Writer::new(&mut cell_buffer);
                    cell_writer.write_event(ev.into_owned()).unwrap();
                } else {
                    writer.write_event(ev)?;
                }
            }
        }
        buf.clear();
    }

    Ok(writer.into_inner().into_inner())
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_modify_styles_xml() {
        let xml = br#"<styleSheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main"><fills count="3"><fill><patternFill patternType="none"/></fill><fill><patternFill patternType="gray125"/></fill><fill><patternFill patternType="solid"><fgColor rgb="FFFAE5"/><bgColor indexed="64"/></patternFill></fill></fills><cellXfs count="3"><xf numFmtId="0" fontId="0" fillId="0" borderId="0" xfId="0"/><xf numFmtId="0" fontId="0" fillId="0" borderId="0" xfId="0" applyProtection="1"><protection locked="0"/></xf><xf numFmtId="0" fontId="0" fillId="2" borderId="0" xfId="0"/></cellXfs></styleSheet>"#;

        let (new_xml, unlocked_col_id, lock_id, yellow_ids, unlocked_ids) =
            rewrite_styles_xml(xml, "FFFAE5").unwrap();

        assert_eq!(unlocked_col_id, 3);
        assert_eq!(lock_id, 4);

        println!("yellow_ids: {:?}", yellow_ids);
        assert!(yellow_ids.contains(&2));
        assert!(unlocked_ids.contains(&1));
        assert!(unlocked_ids.contains(&2));
        assert!(!unlocked_ids.contains(&0));

        let new_xml_str = String::from_utf8(new_xml).unwrap();
        assert!(new_xml_str.contains(r#"<xf numFmtId="0" fontId="0" fillId="0" borderId="0" xfId="0" applyProtection="1"><protection locked="1"/></xf>"#));
        assert!(new_xml_str.contains(r#"<xf numFmtId="0" fontId="0" fillId="0" borderId="0" xfId="0" applyProtection="1"><protection locked="0"/></xf>"#));
        assert!(new_xml_str.contains(r#"<xf numFmtId="0" fontId="0" fillId="2" borderId="0" xfId="0" applyProtection="1"><protection locked="0"/></xf>"#));
    }

    #[test]
    fn test_modify_worksheet_xml() {
        let xml = br#"<worksheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main"><sheetData><row r="5"><c r="B5"><v>Projekt-Titel</v></c><c r="C5" s="2"><v>Eingabewert</v></c><c r="D5" s="1"><v>Another</v></c></row></sheetData></worksheet>"#;

        let mut yellow_ids = HashSet::new();
        yellow_ids.insert(2);
        let mut unlocked_ids = HashSet::new();
        unlocked_ids.insert(1);
        unlocked_ids.insert(2);

        let new_xml =
            rewrite_worksheet_xml(xml, &yellow_ids, &unlocked_ids, 99, 98, &[1, 2, 3]).unwrap();
        let new_str = String::from_utf8(new_xml).unwrap();

        assert!(new_str.contains(r#"<c r="B5" s="99"><v>Projekt-Titel</v></c>"#));
        assert!(new_str.contains(r#"<c r="C5" s="2"><v>Eingabewert</v></c>"#));
        assert!(new_str.contains(r#"<c r="D5" s="99"><v>Another</v></c>"#));
        assert!(new_str.contains(r#"<col min="1" max="3" style="98""#));
    }

    #[test]
    fn test_fixed_values() {
        let xml = br#"<worksheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main"><sheetData><row r="2"><c r="B2" s="1" t="s"><v>0</v></c></row><row r="3"><c r="B3" s="1"><v>42</v></c></row><row r="4"><c r="B4" t="str"><f>B3*2</f></c></row></sheetData></worksheet>"#;
        let yellow_ids = HashSet::new();
        let mut unlocked_ids = HashSet::new();
        unlocked_ids.insert(1); // Style 1 is unlocked

        let new_xml = rewrite_worksheet_xml(xml, &yellow_ids, &unlocked_ids, 99, 98, &[]).unwrap();
        let new_str = String::from_utf8(new_xml).unwrap();

        println!("{}", new_str);
        assert!(
            new_str.contains(r#"<c r="B2" t="s" s="99">"#)
                || new_str.contains(r#"<c r="B2" s="99" t="s">"#)
        );
        assert!(
            new_str.contains(r#"<c r="B4" t="str" s="99">"#)
                || new_str.contains(r#"<c r="B4" s="99" t="str">"#)
        );
    }
}
