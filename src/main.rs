mod cli;

use std::{
    env::{self},
    process::ExitCode,
};

fn main() -> ExitCode {
    let args = env::args();
    match cli::Cli::new(args) {
        Ok(o) => o.execute(),
        Err(e) => {
            eprintln!("{e}");
            ExitCode::FAILURE
        }
    }
}
