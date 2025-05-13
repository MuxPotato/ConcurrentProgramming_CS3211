use std::error::Error;
use tokio::io::{AsyncBufReadExt, AsyncWriteExt, BufReader};
use tokio::net::{TcpListener, TcpStream};
use std::sync::mpsc;

use crate::task::Task;

pub trait ServerTrait {
    async fn start_server(
        &self,
        address: String,
        tx: mpsc::Sender<Result<(), Box<dyn Error + Send>>>,
    );
}

pub struct Server;

impl ServerTrait for Server {
    async fn start_server(
        &self,
        address: String,
        tx: mpsc::Sender<Result<(), Box<dyn Error + Send>>>,
    ) {
        println!("Starting the server");

        match TcpListener::bind(address).await {
            Ok(listener) => {
                tx.send(Ok(())).unwrap();
                loop {
                    match listener.accept().await {
                        Ok((socket, _)) => {
                            tokio::spawn(async move {
                                Self::handle_connection(socket).await;
                            });
                        }
                        Err(e) => {
                            eprintln!("Error accepting connection: {}", e);
                        }
                    }
                }
            },
            Err(e) => {
                println!("here {}", e);
                tx.send(Err(Box::new(e))).unwrap();
                return;
            }
        }
    }
}

impl Server {
    async fn handle_connection(mut stream: TcpStream) {
        loop {
            let mut buf_reader = BufReader::new(&mut stream);
            let mut line = String::new();
            match buf_reader.read_line(&mut line).await {
                Ok(0) => {
                    return;
                }
                Ok(_) => {
                    let response = Self::get_task_value(line).await;
                    if let Some(r) = response {
                        stream.write_all(&[r]).await.unwrap();
                    }
                }
                Err(e) => {
                    eprintln!("Unable to get command due to: {}", e);
                    return;
                }
            }
        }
    }

    async fn get_task_value(buf: String) -> Option<u8> {
        let try_parse = async || -> Result<u8, Box<dyn std::error::Error>> {
            let numbers: Vec<&str> = buf.trim().split(':').collect();
            let task_type = numbers.first().unwrap().parse::<u8>()?;
            let seed = numbers.last().unwrap().parse::<u64>()?;

            let result = Task::execute_async(task_type, seed).await;
            Ok(result)
        };

        match try_parse().await {
            Ok(r) => Some(r),
            Err(_) => None
        }
    }
}