// DO NOT MODIFY THIS FILE
use std::io::{BufReader, Read, Write};
use std::net::TcpStream;
use std::sync::mpsc;
use std::thread;
use std::time::Instant;

use rand::{RngCore, SeedableRng};

use crate::task::TOTAL_TASK_TYPE;

pub trait ClientTrait {
    fn start_client(
        &self,
        initial_seed: u64,
        total_clients: usize,
        total_messages_per_client: usize,
        address: String,
    );
}

pub struct Client;

impl ClientTrait for Client {
    fn start_client(
        &self,
        initial_seed: u64,
        total_clients: usize,
        total_messages_per_client: usize,
        address: String,
    ) {
        let start_time = Instant::now();
        println!("Starting client benchmarking with {} client(s)", total_clients);

        let (sender, receiver) = mpsc::channel();

        let mut handles = vec![];

        for i in 0..total_clients {
            let sender_clone = sender.clone();
            let thread_seed = initial_seed + 3211 * i as u64;

            let cloned_address = address.clone();
            let handle = thread::spawn(move || {
                let mut rng = rand_chacha::ChaCha20Rng::seed_from_u64(thread_seed);
                let mut current_seed_result = rng.next_u64();

                if let Ok(mut stream) = TcpStream::connect(cloned_address) {
                    for _ in 0..total_messages_per_client {
                        let request = format!("{}:{}\n", rng.next_u64() as usize % TOTAL_TASK_TYPE, current_seed_result);

                        match stream.write(request.as_bytes()) {
                            Ok(_) => {}
                            Err(_) => panic!("Unable to write task request to server")
                        }

                        let buf_reader = BufReader::new(&mut stream);
                        let data = Self::read_message(buf_reader).unwrap();
                        let new_seed = *data.first().unwrap();
                        current_seed_result = (new_seed as u64).wrapping_add(rng.next_u64());
                    }

                    sender_clone.send(current_seed_result).expect("unable to send data");
                } else {
                    panic!("Unable to connect to server: timeout");
                }
            });

            handles.push(handle);
        }

        for h in handles {
            h.join().unwrap();
        }

        drop(sender);

        let results: Vec<u64> = receiver.iter().collect();
        let final_sum = results
            .iter()
            .fold(0 as u64, |a, &b| a.wrapping_add(b));
        println!("Successfully collected results from all clients: {}", final_sum);

        let end_time = Instant::now();
        let elapsed_time = end_time - start_time;
        println!("Elapsed time for all clients to finish: {:?}", elapsed_time);
    }
}

impl Client {
    fn read_message<R: Read>(mut reader: R) -> Result<Vec<u8>, Box<dyn std::error::Error>> {
        let data_length = 1;
        let mut buf: Vec<u8> = vec![];
        let mut packet: Vec<u8> = vec![0; data_length];
        loop {
            reader.read_exact(packet.as_mut_slice())?;
            buf.append(&mut packet);
            return Ok(buf);
        }
    }
}
