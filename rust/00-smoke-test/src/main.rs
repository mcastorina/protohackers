use std::io::{self, BufReader, BufWriter};
use std::net::{TcpListener, TcpStream};
use std::thread;

fn main() {
    let listener = TcpListener::bind("0.0.0.0:1337").expect("could not bind to address");
    println!("listening on :1337");
    listener
        .incoming()
        .filter_map(|stream| stream.ok())
        .inspect(|_| println!("accepted new connection"))
        .for_each(|stream| {
            thread::spawn(|| {
                echo(stream);
                println!("connection closed");
            });
        })
}

fn echo(stream: TcpStream) {
    let mut reader = BufReader::new(&stream);
    let mut writer = BufWriter::new(&stream);
    io::copy(&mut reader, &mut writer).unwrap();
}
