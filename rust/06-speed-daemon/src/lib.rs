mod codec;

use std::error::Error;
use std::io::{self, BufRead, BufReader, BufWriter, Read, Write};
use std::iter::Peekable;

struct Common;
struct Camera;
struct Dispatcher;

#[derive(Debug)]
struct Client<R: Iterator<Item = Result<u8, io::Error>>, W: Write, Kind = Common> {
    kind: std::marker::PhantomData<Kind>,
    rbuf: Peekable<R>,
    wbuf: BufWriter<W>,
}

#[derive(Debug)]
enum IncomingMessage {
    IAmCamera { road: u16, mile: u16, limit: u16 },
    IAmDispatcher { roads: Vec<u16> },
    WantHeartbeat { interval: u32 },
    Plate { plate: String, timestamp: u32 },
}

#[derive(Debug)]
enum OutgoingMessage {
    Heartbeat,
    Error(String),
    Ticket {
        plate: String,
        road: u16,
        mile1: u32,
        timestamp1: u32,
        mile2: u32,
        timestamp2: u32,
        speed: u16,
    },
}

impl<R: Iterator<Item = Result<u8, io::Error>>, W: Write> Client<R, W> {
    fn next_message(&mut self) -> Result<IncomingMessage, Box<dyn Error>> {
        match self.rbuf.peek() {
            _ => return Err("oh no".into()),
        }
    }
}

impl<R: Read, W: Write> Client<io::Bytes<BufReader<R>>, W> {
    fn new(r: R, w: W) -> Self {
        Self {
            kind: std::marker::PhantomData,
            rbuf: BufReader::new(r).bytes().peekable(),
            wbuf: BufWriter::new(w),
        }
    }
}

#[cfg(test)]
mod test {
    use super::*;

    #[test]
    fn next_message() {
        let mut output = Vec::new();
        let mut client = Client::new("hello".as_bytes(), &mut output);

        panic!("{:?}", client.next_message());
    }
}

// 0x10: Error (Server->Client)
// 0x21: Ticket (Server->Client)
// 0x41: Heartbeat (Server->Client)
// 0x20: Plate (Client->Server)
// 0x40: WantHeartbeat (Client->Server)
// 0x80: IAmCamera (Client->Server)
// 0x81: IAmDispatcher (Client->Server)
