use excel_protection::color_routing::*;

#[test]
fn test_debug() {
    let xml = r#"<styleSheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main"><fills count="3"><fill><patternFill patternType="none"/></fill><fill><patternFill patternType="gray125"/></fill><fill><patternFill patternType="solid"><fgColor rgb="FFFFFAE5"/></patternFill></fill></fills><cellXfs count="3"><xf numFmtId="0" fontId="0" fillId="0" borderId="0" xfId="0"></xf><xf numFmtId="0" fontId="0" fillId="0" borderId="0" xfId="0" applyAlignment="false" applyProtection="true"><alignment></alignment><protection hidden="false" locked="false"></protection></xf><xf numFmtId="0" fontId="0" fillId="2" borderId="0" xfId="0" applyFill="true" applyAlignment="false"><alignment></alignment></xf></cellXfs></styleSheet>"#;

    let target_id = find_fill_id_by_color(xml, "FFFAE5").unwrap();
    let mut yellow_ids = std::collections::HashSet::new();
    let mut unlocked_ids = std::collections::HashSet::new();

    let new_xml = rewrite_styles_xml(xml, target_id, &mut yellow_ids, &mut unlocked_ids).unwrap();
    
    println!("yellow: {:?}", yellow_ids);
    println!("unlocked: {:?}", unlocked_ids);
    println!("new_xml: {}", new_xml.0);
    assert!(unlocked_ids.contains(&1));
}
