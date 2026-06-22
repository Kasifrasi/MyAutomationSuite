fn main() {
    println!("cargo:rerun-if-changed=../sidecars/Excelize/generator.exe");

    let config = slint_build::CompilerConfiguration::new().with_style("fluent".into());

    slint_build::compile_with_config("ui/main.slint", config).unwrap();
}
