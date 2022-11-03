use std::io::{self, BufReader, BufWriter};
use std::net::{TcpListener, TcpStream};
use std::thread;

// Deep inside Initrode Global's enterprise management framework lies a
// component that writes data to a server and expects to read the same data
// back. (Think of it as a kind of distributed system delay-line memory). We
// need you to write the server to echo the data back.

// Accept TCP connections.

// Whenever you receive data from a client, send it back unmodified.

// Make sure you don't mangle binary data, and that you can handle at least 5
// simultaneous clients.

// Once the client has finished sending data to you it shuts down its sending
// side. Once you've reached end-of-file on your receiving side, and sent back
// all the data you've received, close the socket so that the client knows
// you've finished. (This point trips up a lot of proxy software, such as
// ngrok; if you're using a proxy and you can't work out why you're failing the
// check, try hosting your server in the cloud instead).

// Your program will implement the TCP Echo Service from RFC 862.

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
