mod notes;

use std::process::ExitCode;

fn main() -> ExitCode {
    let mut args = std::env::args();
    args.next();
    let cmd = args.next().unwrap_or("h".to_string());

    match cmd.as_str() {
        "l" => notes::list(),
        "i" => notes::insert(args),
        "c" => notes::change(args),
        "d" => notes::delete(args),
        "e" => notes::edit(),
        "h" => notes::help(),
        _ => {
            eprintln!("{cmd} is not a valid command");
            ExitCode::FAILURE
        }
    }
}
