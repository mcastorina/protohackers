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

#[derive(Debug, serde::Deserialize)]
struct Request {
    method: String,
    number: f64,
}

fn prime(stream: TcpStream) {
    let mut reader = BufReader::new(&stream);
    let mut writer = BufWriter::new(&stream);
    let requests = serde_jsonlines::JsonLinesReader::new(&mut reader);

    for request in requests.read_all() {
        let request: Request = request.unwrap();
        // calculate isPrime
        let response = Response {
            method: "isPrime".to_string(),
            prime: is_prime(request.number),
        };
        serde_json::to_writer(&mut writer, &response).unwrap();
        writer.write_all(b"\n").unwrap();
        writer.flush().unwrap();
    }
}

fn is_prime(num: f64) -> bool {
    if num.fract() != 0. {
        return false;
    }

    primes::is_prime(num as u64)
}

#[derive(Debug, Clone, serde::Serialize)]
struct Response {
    method: String,
    #[serde(rename = "isPrime")]
    prime: bool,
}
