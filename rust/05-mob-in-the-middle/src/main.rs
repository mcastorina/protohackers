use regex::Regex;
use std::io::{BufRead, BufReader, BufWriter, Write};
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
            // Create client stream reader and writer.
            let mut client_reader = BufReader::new(stream.try_clone().unwrap());
            let mut client_writer = BufWriter::new(stream);

            // Open a connection to upstream.
            let upstream = TcpStream::connect(UPSTREAM_ADDRESS).unwrap();
            upstream
                .set_read_timeout(Some(Duration::from_millis(200)))
                .unwrap();
            let mut upstream_reader = BufReader::new(upstream.try_clone().unwrap());
            let mut upstream_writer = BufWriter::new(upstream);

            // Create a signal to indicate the client has disconnected and we should disconnect the
            // upstream connection.
            let done_signal: Arc<RwLock<bool>> = Default::default();

            // Asynchronously read from client and send to upstream.
            let done = done_signal.clone();
            let h1 = thread::spawn(move || {
                let mut msg = String::new();
                loop {
                    msg.clear();
                    if let Ok(0) = client_reader.read_line(&mut msg) {
                        break;
                    }
                    msg = replace_boguscoin(msg);
                    let _ = upstream_writer.write(msg.as_bytes()).unwrap();
                    let _ = upstream_writer.flush().unwrap();
                }
                *done.write().unwrap() = true;
                println!("client disconnected");
            });

            // Asynchronously read from upstream and send to client.
            let done = done_signal.clone();
            let h2 = thread::spawn(move || {
                let mut msg = String::new();
                loop {
                    msg.clear();
                    match upstream_reader.read_line(&mut msg) {
                        // The upstream closed the connection.
                        Ok(0) => break,
                        // We have a read timeout so check to see if we should disconnect.
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
                    let _ = client_writer.write(msg.as_bytes()).unwrap();
                    let _ = client_writer.flush().unwrap();
                }
                println!("upstream disconnected");
            });

            h1.join().unwrap();
            h2.join().unwrap();
        });
    }
}

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
