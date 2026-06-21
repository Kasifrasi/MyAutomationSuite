fn main() {
    let inputs = ["hallo", "ha", "h", "2025a", "h2025a", "hallo123"];
    for i in inputs {
        println!("{:?} -> {:?}", i, folder_generator::format_project_name(i));
    }
}
