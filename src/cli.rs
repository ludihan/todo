use std::{
    env::{self, Args},
    fs::{self},
    path::PathBuf,
    process::{Command, ExitCode},
};

#[derive(Debug)]
pub struct Cli {
    target: PathBuf,
    // no command will only print the file
    command: Option<CliCommand>,
}

impl Cli {
    pub fn new(args: Args) -> Result<Self, String> {
        let mut target = dirs::data_dir().unwrap();
        target.push(env!("CARGO_PKG_NAME"));
        fs::create_dir_all(&target).unwrap();

        let args = args.skip(1);
        let mut args = args.peekable();

        match args.peek() {
            Some(arg) => {
                if arg.contains(".") || arg.contains("/") {
                    return Err("don't use dots or slashes".to_string());
                }

                if ["i", "c", "d", "h", "l", "D", "e"].contains(&arg.as_str()) {
                    target.push("default");
                } else {
                    target.push(arg);
                    args.next();
                }
            }
            None => {
                target.push("default");
                return Ok(Self {
                    target,
                    command: None,
                });
            }
        }

        let command = parse_command(args.collect())?;

        Ok(Self { target, command })
    }

    pub fn execute(self) -> ExitCode {
        match self.command {
            Some(CliCommand::Insert(ref value, index)) => {
                let mut file = fs::read_to_string(&self.target).unwrap_or_default();
                let mut lines: Vec<String> = if file.is_empty() {
                    vec![]
                } else {
                    file.lines().map(|l| l.to_string()).collect()
                };

                if let Some(i) = index {
                    if i == 0 || i > lines.len() + 1 {
                        eprintln!("invalid index: {i}");
                        return ExitCode::FAILURE;
                    }
                    lines.insert(i - 1, value.to_string());
                } else {
                    lines.push(value.to_string());
                }

                file = lines.join("\n") + "\n";
                if let Err(e) = fs::write(&self.target, file) {
                    eprintln!("failed to write file: {e}");
                    ExitCode::FAILURE
                } else {
                    self.print();
                    ExitCode::SUCCESS
                }
            }
            Some(CliCommand::Change(ref value, index)) => {
                let mut file = fs::read_to_string(&self.target).unwrap_or_default();
                let mut lines: Vec<String> = file.lines().map(|l| l.to_string()).collect();

                if lines.is_empty() {
                    eprintln!("no lines to change");
                    return ExitCode::FAILURE;
                }

                let i = index.unwrap_or(lines.len());
                if i == 0 || i > lines.len() {
                    eprintln!("invalid index: {i}");
                    return ExitCode::FAILURE;
                }

                lines[i - 1] = value.to_string();
                file = lines.join("\n") + "\n";
                if let Err(e) = fs::write(&self.target, file) {
                    eprintln!("failed to write file: {e}");
                    ExitCode::FAILURE
                } else {
                    self.print();
                    ExitCode::SUCCESS
                }
            }
            Some(CliCommand::Delete(Some(ref indices))) => {
                let mut file = fs::read_to_string(&self.target).unwrap_or_default();
                let mut lines: Vec<String> = file.lines().map(|l| l.to_string()).collect();

                if lines.is_empty() {
                    eprintln!("no lines to delete");
                    return ExitCode::FAILURE;
                }

                let mut indices = indices.clone();
                indices.sort_unstable();
                indices.dedup();

                for &i in indices.iter().rev() {
                    if i == 0 || i > lines.len() {
                        eprintln!("invalid index: {i}");
                        continue;
                    }
                    lines.remove(i - 1);
                }

                file = lines.join("\n") + "\n";
                if let Err(e) = fs::write(&self.target, file) {
                    eprintln!("failed to write file: {e}");
                    ExitCode::FAILURE
                } else {
                    self.print();
                    ExitCode::SUCCESS
                }
            }
            Some(CliCommand::Delete(None)) => {
                let mut file = fs::read_to_string(&self.target).unwrap_or_default();
                let mut lines: Vec<String> = file.lines().map(|l| l.to_string()).collect();

                if lines.pop().is_none() {
                    eprintln!("no lines to delete");
                    return ExitCode::FAILURE;
                }

                if lines.is_empty() {
                    file = lines.join("\n");
                } else {
                    file = lines.join("\n") + "\n";
                }
                if let Err(e) = fs::write(&self.target, file) {
                    eprintln!("failed to write file: {e}");
                    ExitCode::FAILURE
                } else {
                    self.print();
                    ExitCode::SUCCESS
                }
            }
            Some(CliCommand::DeleteFile) => {
                if self.target.exists() {
                    fs::remove_file(&self.target).unwrap();
                    ExitCode::SUCCESS
                } else {
                    eprintln!("file does not exist");
                    ExitCode::FAILURE
                }
            }
            Some(CliCommand::ListFiles) => {
                let files: Vec<_> = fs::read_dir(self.target.parent().unwrap())
                    .unwrap()
                    .collect();
                let width = ((files.len() as f64).log10() + 1.0) as usize;
                files.iter().enumerate().for_each(|(i, n)| {
                    println!(
                        "{:<width$}: {n}",
                        i + 1,
                        width = width,
                        n = n
                            .as_ref()
                            .unwrap()
                            .path()
                            .file_name()
                            .unwrap()
                            .to_str()
                            .unwrap()
                    )
                });
                ExitCode::SUCCESS
            }
            Some(CliCommand::Help) => {
                eprintln!("{HELP_MESSAGE}");
                ExitCode::SUCCESS
            }
            Some(CliCommand::Edit) => {
                let editor = env::var("VISUAL")
                    .or(env::var("EDITOR"))
                    .unwrap_or("nano".to_string());

                let result = Command::new(editor).arg(self.target).status();

                if result.is_err() {
                    eprintln!("failed to open editor");
                    ExitCode::FAILURE
                } else {
                    ExitCode::SUCCESS
                }
            }
            None => {
                self.print();
                ExitCode::SUCCESS
            }
        }
    }

    fn print(&self) {
        let file = fs::read_to_string(&self.target);
        if file.is_err() {
            fs::File::create(&self.target).unwrap();
            println!("you don't have any notes");
            return;
        }
        let file = file.unwrap();

        if file.is_empty() {
            println!("you don't have any notes");
            return;
        }
        let width = ((file.lines().count() as f64).log10() + 1.0) as usize;
        file.lines()
            .enumerate()
            .for_each(|(i, n)| println!("{:<width$}: {n}", i + 1, width = width));
    }
}

fn parse_command(command: Vec<String>) -> Result<Option<CliCommand>, String> {
    let mut command = command.iter().map(|s| s.as_str());

    match command.next() {
        Some("i") => {
            let value = command
                .next()
                .ok_or("you didn't provide a value".to_string())?;
            let index = command
                .next()
                .map(|i| {
                    i.parse::<usize>()
                        .map_err(|_| format!("'{i}' is not a valid index"))
                })
                .transpose()?;
            Ok(Some(CliCommand::Insert(value.to_string(), index)))
        }
        Some("c") => {
            let value = command
                .next()
                .ok_or("you didn't provide a value".to_string())?;
            let index = command
                .next()
                .map(|i| {
                    i.parse::<usize>()
                        .map_err(|_| format!("'{i}' is not a valid index"))
                })
                .transpose()?;
            Ok(Some(CliCommand::Change(value.to_string(), index)))
        }
        Some("d") => {
            let indices: Result<Vec<usize>, String> = command
                .map(|i| {
                    i.parse::<usize>()
                        .map_err(|_| format!("'{i}' is not a valid index"))
                })
                .collect();
            if indices.clone()?.is_empty() {
                Ok(Some(CliCommand::Delete(None)))
            } else {
                Ok(Some(CliCommand::Delete(Some(indices?))))
            }
        }
        Some("h") => Ok(Some(CliCommand::Help)),
        Some("l") => Ok(Some(CliCommand::ListFiles)),
        Some("D") => Ok(Some(CliCommand::DeleteFile)),
        Some("e") => Ok(Some(CliCommand::Edit)),
        Some(cmd) => Err(format!("{cmd} is not a valid command")),
        None => Ok(None),
    }
}

#[derive(Debug)]
enum CliCommand {
    Insert(String, Option<usize>),
    Change(String, Option<usize>),
    Delete(Option<Vec<usize>>),
    DeleteFile,
    ListFiles,
    Help,
    Edit,
}

const HELP_MESSAGE: &str = "\
todo is a tool for quick reminders and taking notes.

usage:
    todo [note] [options...]

if you don't specify a note, the application will use a default one

options:
    i <value> [index] Insert a value at a given index.
                      If an index is not provided, appends the value at the end of the list.
    c <value> [index] Change a value at a given index.
                      If an index is not provided, changes the last value.
    d [index...]      Delete values at given indices.
                      If no indices are provided, deletes the last value from the list.
    h                 Show this help message.
    l                 List all available notes.
    D                 Delete the specified note.
    e                 Edit a note using a text editor.
                      Will try to use the VISUAL and EDITOR environment variables.\
";
