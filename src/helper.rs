use std::process::Command;
use std::thread;
use std::env;
use std::io::{self, Write};

fn main() {
    let default_insts = 8;
    let num_instances: usize = match env::args().nth(1) {
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
    let go_args: Vec<&str> = input.trim().split_whitespace().collect();
    let go_program = "./load-checker.exe";
    println!(
        "Starting {} instances with arguments: {:?}",
        num_instances, go_args
    );
    let mut handles = vec![];
    for _ in 0..num_instances {
        let go_args_clone = go_args.clone();
        let handle = thread::spawn(move || {
            match Command::new(go_program).args(go_args_clone).spawn() {
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