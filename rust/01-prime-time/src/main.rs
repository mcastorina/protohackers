use prime_time::{PrimeTime, PrimeTimeOutput};
use std::io::{BufReader, BufWriter, Write};
use std::net::{TcpListener, TcpStream};
use std::thread;

fn main() {
    let listener = TcpListener::bind("0.0.0.0:1337").expect("could not bind to address");
    println!("listening on :1337");
    for stream in listener.incoming() {
        println!("accepted new connection");
        thread::spawn(|| {
            prime(stream.unwrap());
            println!("connection closed");
        });
    }
}

fn prime(stream: TcpStream) {
    let mut reader = BufReader::new(&stream);
    let mut writer = BufWriter::new(&stream);

    let requests = serde_jsonlines::JsonLinesReader::new(&mut reader);
    for request in requests.read_all() {
        match PrimeTime::try_from(request).map(PrimeTimeOutput::from) {
            Ok(response) => {
                let _ = serde_json::to_writer(&mut writer, &response);
            }
            Err(error) => {
                let _ = serde_json::to_writer(&mut writer, &error);
            }
        }
        let _ = writer.write_all(b"\n");
        let _ = writer.flush();
    }
}
