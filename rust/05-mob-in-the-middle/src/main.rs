use regex::Regex;
use std::io::{BufRead, BufReader, BufWriter, Read, Write};
use std::net::{TcpListener, TcpStream};
use std::sync::{Arc, RwLock};
use std::thread;
use std::time::Duration;

const UPSTREAM_ADDRESS: &str = "chat.protohackers.com:16963";
const TONY_BOGUSCOIN_ADDRESS: &str = "7YWHMfk9JZe0LM0g1ZauHuiSxhI";

fn main() {
    let listener = TcpListener::bind(":::1337").expect("could not bind to address");
    println!("listening on :1337");

    for stream in listener.incoming().filter_map(Result::ok) {
        println!("accepted new connection");

        thread::spawn(move || {
            // Open a connection to upstream.
            let upstream = TcpStream::connect(UPSTREAM_ADDRESS).unwrap();
            upstream
                .set_read_timeout(Some(Duration::from_millis(200)))
                .unwrap();

            // Create separate readers and writers.
            let (client_reader, client_writer) = split_stream(stream);
            let (upstream_reader, upstream_writer) = split_stream(upstream);

            // Create a signal to indicate the client has disconnected and we should disconnect the
            // upstream connection.
            let done_signal: Arc<RwLock<bool>> = Default::default();

            // Asynchronously read from client and send to upstream.
            let done = done_signal.clone();
            let h1 = thread::spawn(move || {
                proxy(client_reader, upstream_writer, done.clone());
                println!("client disconnected");
                *done.write().unwrap() = true;
            });

            // Asynchronously read from upstream and send to client.
            let h2 = thread::spawn(move || {
                proxy(upstream_reader, client_writer, done_signal);
            });

            // Wait for both threads to finish.
            h1.join().unwrap();
            h2.join().unwrap();
        });
    }
}

/// Proxy a read buffer into a write buffer, replacing all instances of a boguscoin with Tony's
/// boguscoin. If the reader errors (for e.g. from a timeout), the done mutex is checked to exit
/// the loop.
fn proxy<R: Read, W: Write>(
    mut reader: BufReader<R>,
    mut writer: BufWriter<W>,
    done: Arc<RwLock<bool>>,
) {
    let mut msg = String::new();
    loop {
        msg.clear();
        match reader.read_line(&mut msg) {
            // The connection closed.
            Ok(0) => break,
            // Possibly a read timeout - check to see if we should exit.
            Err(_) => {
                if *done.read().unwrap() {
                    break;
                }
                continue;
            }
            // We read some data.
            _ => (),
        }
        msg = replace_boguscoin(msg);
        let _ = writer.write(msg.as_bytes()).unwrap();
        writer.flush().unwrap();
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
