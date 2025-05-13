use std::error::Error;
use tokio::io::{AsyncBufReadExt, AsyncWriteExt};
use tokio::io::{BufReader};
use tokio::net::{TcpListener, TcpStream};
use std::sync::mpsc;
use crate::task::Task;
use crate::task::TaskType;
use tokio_rayon::spawn;

pub trait ServerTrait {
    async fn start_server(
        &self,
        address: String,
        tx: mpsc::Sender<Result<(), Box<dyn Error + Send>>>,
    ) -> Result<(), Box<dyn Error>>;
}

pub struct Server;

/*
    Note: Compared to the original code, where there was a match statement to see if result is 
    OK or Error, here the function itself returns either (), or Box dyn error. So the error that
    is threw by .await? is propagated to whoever called this functions to handle them

    By using ?, any error from `.await` is returned to the caller as `Err(Box<dyn Error>)`,
    so we don’t need to manually match on Ok/Err here.
*/
impl ServerTrait for Server {
    async fn start_server(
        &self,
        address: String,
        tx: mpsc::Sender<Result<(), Box<dyn Error + Send>>>,
    ) -> Result<(), Box<dyn Error>> {
        //println!("Starting the server");
        let listener = TcpListener::bind(address).await?;

        tx.send(Ok(())).unwrap();

        loop {
            let (stream, _socket_addr) = listener.accept().await?;
            tokio::spawn(async move {
                if let Err(e) = Self::handle_connection(stream).await {
                    eprintln!("Connection handler error: {}", e);
                }
            });
        }
    }
}

impl Server {
    async fn handle_connection(mut stream: TcpStream) -> Result<(), Box<dyn Error>> {
        loop {
            let mut buf_reader = BufReader::new(&mut stream);
            let mut line = String::new();
            
            let bytes_read = buf_reader.read_line(&mut line).await?;
            if bytes_read == 0 {
                return Ok(());
            }
            
            let response = Self::get_task_value(line).await;
            
            /*
                write() atempt to write up to buf.len() bytes, but may write fewer (and return the bytes it wrote)
                write_all() is from tokio::io::AsyncWriteExt, doing the loop (of writing) under the hood, retrying till it finishes, so its for async version
            */
            if let Some(r) = response {
                stream.write_all(&[r]).await?;
            }
        }
    }

    async fn get_task_value(buf: String) -> Option<u8> {
        //println!("get_task_value received buf: {:?}", buf);
        
        let numbers: Vec<&str> = buf.trim().split(':').collect();
        let task_type = numbers.first().unwrap().parse::<u8>().ok()?;
        let seed = numbers.last().unwrap().parse::<u64>().ok()?;
        
        if task_type == TaskType::IOIntensiveTask as u8 {
            let result = Task::execute_async(task_type, seed).await;
            Some(result)
        } else {
            // runs on your global Rayon pool
            let result: u8 = spawn(move || {
                Task::execute(task_type, seed)
            })
            .await;
            Some(result)
        }
    }
}
