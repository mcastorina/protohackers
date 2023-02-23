use std::io::{BufRead, BufReader, BufWriter, Read, Write};
use std::net::{Shutdown, TcpListener, TcpStream};
use std::thread;

fn main() {
    let listener = TcpListener::bind(("::", 1337)).expect("could not bind to address");
    println!("listening on :1337");

    for stream in listener.incoming().filter_map(Result::ok) {
        println!("accepted new connection");
    }
}

/// Split a TcpStream into a buffered reader and writer.
fn split_stream(stream: TcpStream) -> (BufReader<TcpStream>, BufWriter<TcpStream>) {
    let reader = BufReader::new(stream.try_clone().unwrap());
    let writer = BufWriter::new(stream);
    (reader, writer)
}
