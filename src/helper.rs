use std::process::Command;
use std::thread;
use std::env;
use std::io::{self, Write};

fn main() {
    let go_program = match env::args().nth(1) {
        Some(program) => program,
        None => {
            eprintln!("Load checker program path must be specified as the first argument.");
            return;
        }
    };
    let default_insts = 8;
    let num_instances: usize = match env::args().nth(2) {
        Some(arg) => match arg.parse::<usize>() {
            Ok(n) if n > 0 => n,
            Ok(_) => {
                eprintln!("Invalid number of instances. Using default value of {}.", default_insts);
                default_insts
            }
            Err(_) => {
                eprintln!("Error parsing number of instances. Using default value of {}.", default_insts);
                default_insts
            }
        },
        None => default_insts,
    };
    print!("Enter arguments for the load checking program (e.g., -url example.com -requests 50): ");
    io::stdout().flush().unwrap();
    let mut input = String::new();
    io::stdin().read_line(&mut input).expect("Failed to read input");
    let go_args: Vec<String> = input.trim().split_whitespace().map(String::from).collect();
    println!(
        "Starting {} instances of the load checker program '{}' with arguments: {:?}",
        num_instances, go_program, go_args
    );
    let mut handles = vec![];
    for _ in 0..num_instances {
        let go_args_clone = go_args.clone();
        let go_program_clone = go_program.clone(); 
        let handle = thread::spawn(move || {
            match Command::new(&go_program_clone).args(go_args_clone).spawn() {
                Ok(mut child) => {
                    if let Err(e) = child.wait() {
                        eprintln!("Error waiting for child process: {}", e);
                    }
                }
                Err(e) => eprintln!("Failed to start load checking program: {}", e),
            }
        });
        handles.push(handle);
    }
    for handle in handles {
        handle.join().unwrap();
    }
    println!("All instances completed");
}