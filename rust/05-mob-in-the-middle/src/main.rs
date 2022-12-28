use regex::Regex;
use std::io::{BufRead, BufReader, BufWriter, Read, Write};
use std::net::{Shutdown, TcpListener, TcpStream};
use std::thread;

const UPSTREAM_ADDRESS: &str = "chat.protohackers.com:16963";
const TONY_BOGUSCOIN_ADDRESS: &str = "7YWHMfk9JZe0LM0g1ZauHuiSxhI";

fn main() {
    let listener = TcpListener::bind(":::1337").expect("could not bind to address");
    println!("listening on :1337");

    for stream in listener.incoming().filter_map(Result::ok) {
        println!("accepted new connection");

        // Open a connection to upstream.
        let upstream = TcpStream::connect(UPSTREAM_ADDRESS).unwrap();

        // Create separate readers and writers.
        let (mut client_reader, mut client_writer) = split_stream(stream);
        let (mut upstream_reader, mut upstream_writer) = split_stream(upstream);

        // Asynchronously read from client and send to upstream.
        thread::spawn(move || {
            proxy(&mut client_reader, &mut upstream_writer);
            println!("client disconnected");
            let _ = upstream_writer.get_ref().shutdown(Shutdown::Both);
        });

        // Asynchronously read from upstream and send to client.
        thread::spawn(move || {
            proxy(&mut upstream_reader, &mut client_writer);
            let _ = client_writer.get_ref().shutdown(Shutdown::Both);
        });
    }
}

/// Proxy a read buffer into a write buffer, replacing all instances of a boguscoin with Tony's
/// boguscoin. If either a read or write operation errors, this function will return.
fn proxy<R: Read, W: Write>(reader: &mut BufReader<R>, writer: &mut BufWriter<W>) {
    let mut msg = String::new();
    loop {
        msg.clear();
        match reader.read_line(&mut msg) {
            // The connection closed.
            Ok(0) | Err(_) => break,
            // We read some data.
            _ => (),
        }
        msg = replace_boguscoin(msg);
        if writer.write_all(msg.as_bytes()).is_err() {
            break;
        }
        let _ = writer.flush();
    }
}

/// Split a TcpStream into a buffered reader and writer.
fn split_stream(stream: TcpStream) -> (BufReader<TcpStream>, BufWriter<TcpStream>) {
    let reader = BufReader::new(stream.try_clone().unwrap());
    let writer = BufWriter::new(stream);
    (reader, writer)
}

/// Replace all instances of a boguscoin with Tony's boguscoin.
fn replace_boguscoin(msg: impl AsRef<str>) -> String {
    let msg = msg.as_ref();
    let re = Regex::new(r"\b7[[:alnum:]]{25,34}\b").unwrap();
    let mut new_msg: String = msg.to_string();
    // Iterate over each match, filter out any matches that aren't on a space boundary, and replace
    // with Tony's boguscoin address. We need to iterate in reverse so the replacement doesn't
    // affect the next range.
    re.find_iter(msg.as_ref())
        .collect::<Vec<_>>()
        .iter()
        .rev()
        .filter(|mat| {
            let data = msg.as_bytes();
            if mat.start() != 0 && !(data[mat.start() - 1] as char).is_ascii_whitespace() {
                return false;
            }
            if mat.end() != data.len() && !(data[mat.end()] as char).is_ascii_whitespace() {
                return false;
            }
            true
        })
        .for_each(|mat| new_msg.replace_range(mat.range(), TONY_BOGUSCOIN_ADDRESS));
    new_msg
}

#[test]
fn test_replace_boguscoin() {
    assert_eq!(
        replace_boguscoin("7aaaaaaaaaaaaaaaaaaaaaaaaaa-GYUyb6MEuGSsxsdUO79xLjNMH8e-1234"),
        "7aaaaaaaaaaaaaaaaaaaaaaaaaa-GYUyb6MEuGSsxsdUO79xLjNMH8e-1234",
    );
    assert_eq!(
        replace_boguscoin(
            "7aaaaaaaaaaaaaaaaaaaaaaaaa 7bbbbbbbbbbbbbbbbbbbbbbbbb 7ccccccccccccccccccccccccc\n",
        ),
        "7YWHMfk9JZe0LM0g1ZauHuiSxhI 7YWHMfk9JZe0LM0g1ZauHuiSxhI 7YWHMfk9JZe0LM0g1ZauHuiSxhI\n",
    );
}
