use std::{
    env::{self, Args},
    fs::{self},
    path::PathBuf,
    process::{Command, ExitCode},
};

fn get_path() -> PathBuf {
    let mut path = env::home_dir().unwrap_or(PathBuf::from("."));
    path.push(".todo");
    path
}

fn read_file() -> Vec<String> {
    let path = get_path();
    fs::read_to_string(path)
        .unwrap_or_default()
        .lines()
        .filter(|s| *s != "")
        .map(|s| s.to_string())
        .collect()
}

fn write_notes(notes: &[String]) -> ExitCode {
    let path = get_path();
    let result = fs::write(path, notes.join("\n"));
    if let Err(_) = result {
        eprintln!("couldn't write to file");
        ExitCode::FAILURE
    } else {
        ExitCode::SUCCESS
    }
}

pub(crate) fn list() -> ExitCode {
    let notes = read_file();
    if notes.is_empty() {
        println!("you don't have any notes");
        return ExitCode::SUCCESS;
    }
    let width = ((notes.len() as f64).log10()+1.0) as usize;
    notes
        .iter()
        .enumerate()
        .for_each(|(i, n)| println!("{:<width$}: {n}", i + 1, width=width));
    ExitCode::SUCCESS
}

pub(crate) fn insert(mut args: Args) -> ExitCode {
    let mut notes = read_file();
    let value = args.next();
    let index = args.next();

    match value {
        None => {
            eprintln!("you didn't provide a value");
            ExitCode::FAILURE
        }
        Some(v) if v.is_empty() => {
            eprintln!("you didn't provide a value");
            ExitCode::FAILURE
        }
        Some(value) => match index {
            Some(index) => match parse_indices(&[index], true) {
                Some(parsed) => {
                    let index = parsed.first().unwrap();
                    if *index == notes.len() - 1 {
                        notes.push(value);
                        write_notes(&notes);
                        ExitCode::SUCCESS
                    } else {
                        notes.insert(index - 1, value);
                        write_notes(&notes);
                        ExitCode::SUCCESS
                    }
                }
                None => ExitCode::FAILURE,
            },
            None => {
                notes.push(value);
                write_notes(&notes)
            }
        },
    }
}

pub(crate) fn change(mut args: Args) -> ExitCode {
    let mut notes = read_file();
    let value = args.next();
    let index = args.next();

    match value {
        None => {
            eprintln!("you didn't provide a value");
            ExitCode::FAILURE
        }
        Some(value) if value.is_empty() => {
            eprintln!("you didn't provide a value");
            ExitCode::FAILURE
        }
        Some(value) => match index {
            Some(index) => match parse_indices(&[index], false) {
                Some(parsed) => {
                    let i = parsed.first().unwrap();
                    notes[i - 1] = value;
                    write_notes(&notes);
                    ExitCode::SUCCESS
                }
                None => ExitCode::FAILURE,
            },
            None => match notes.last_mut() {
                Some(last) => {
                    *last = value;
                    write_notes(&notes);
                    ExitCode::SUCCESS
                }
                None => {
                    eprintln!("there are no notes to be changed");
                    ExitCode::FAILURE
                }
            },
        },
    }
}

pub(crate) fn delete(args: Args) -> ExitCode {
    let mut notes = read_file();
    let args_vec: Vec<String> = args.collect();
    match parse_indices(&args_vec, false) {
        Some(vec) => {
            if vec.is_empty() {
                notes.remove(notes.len() - 1);
                write_notes(&notes)
            } else {
                let updated_notes: Vec<String> = notes
                    .into_iter()
                    .enumerate()
                    .filter_map(|(i, note)| {
                        if vec.contains(&(i + 1)) {
                            None
                        } else {
                            Some(note)
                        }
                    })
                    .collect();

                write_notes(&updated_notes)
            }
        }
        None => ExitCode::FAILURE,
    }
}

pub(crate) fn edit() -> ExitCode {
    let path = get_path();
    let editor = env::var("VISUAL")
        .or(env::var("EDITOR"))
        .unwrap_or("nano".to_string());

    let result = Command::new(editor).arg(path).status();
    if let Err(_) = result {
        eprintln!("failed to open editor");
        ExitCode::FAILURE
    } else {
        ExitCode::SUCCESS
    }
}

pub(crate) fn help() -> ExitCode {
    eprintln!("{HELP_MESSAGE}");
    ExitCode::FAILURE
}

fn parse_indices(args: &[String], inclusive: bool) -> Option<Vec<usize>> {
    let notes = read_file();
    let parsed_indices: Result<Vec<_>, _> = args.iter().map(|x| x.parse::<usize>()).collect();

    match parsed_indices {
        Ok(vec) => {
            if let Some(biggest) = vec.iter().max() {
                if if inclusive {
                    *biggest > notes.len() + 1
                } else {
                    *biggest > notes.len()
                } {
                    eprintln!("index {biggest} is bigger than the file");
                    None
                } else {
                    if vec.iter().any(|&x| x <= 0) {
                        eprintln!("null or negative indices are not allowed");
                        None
                    } else {
                        Some(vec)
                    }
                }
            } else {
                Some(vec)
            }
        }
        Err(e) => {
            eprintln!("not a proper index: {e}");
            None
        }
    }
}

const HELP_MESSAGE: &str = "\
todo is a tool for quick reminders and taking notes.
options:
  l    List all notes.
  i <value> [index]
       Insert a note. If an index is provided, writes the note at that position.
  c <value> [index]
       Change the last note from the database, if an index is provided, changes
       the note at that index.
  d [index...]
       Delete the last note from the list. Optionally, accepts a list of indices
       to delete.
  h    Show this help message.
  e    Edit your notes using a text editor, uses the VISUAL and EDITOR
       environment variables.\
";
