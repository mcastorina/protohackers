mod codec;
mod msg;

use std::error::Error;
use std::io::{BufRead, BufReader, BufWriter, Read, Write};
use std::{io, slice};

struct Common;
struct Camera;
struct Dispatcher;

#[derive(Debug)]
struct Client<R: Read, W: Write, Kind = Common> {
    kind: std::marker::PhantomData<Kind>,
    rbuf: BufReader<R>,
    wbuf: BufWriter<W>,
}

#[derive(Debug, PartialEq, Eq)]
enum IncomingMessage {
    IAmCamera(msg::IAmCamera),
    IAmDispatcher(msg::IAmDispatcher),
    WantHeartbeat(msg::WantHeartbeat),
    Plate(msg::Plate),
}

#[derive(Debug, PartialEq, Eq)]
enum OutgoingMessage {
    Heartbeat,
    Error(String),
    Ticket(msg::Ticket),
}

impl<R: Read, W: Write> Client<R, W> {
    fn next_message(&mut self) -> Result<IncomingMessage, Box<dyn Error>> {
        use IncomingMessage::*;

        let mut id = 0;
        self.rbuf.read_exact(slice::from_mut(&mut id))?;

        Ok(match id {
            0x40 => WantHeartbeat(codec::from_reader(&mut self.rbuf)?),
            0x80 => IAmCamera(codec::from_reader(&mut self.rbuf)?),
            0x81 => IAmDispatcher(codec::from_reader(&mut self.rbuf)?),
            _ => return Err("unrecognized message".into()),
        })
    }

    fn into_camera(self) -> Client<R, W, Camera> {
        Client {
            kind: std::marker::PhantomData,
            rbuf: self.rbuf,
            wbuf: self.wbuf,
        }
    }

    fn into_dispatcher(self) -> Client<R, W, Dispatcher> {
        Client {
            kind: std::marker::PhantomData,
            rbuf: self.rbuf,
            wbuf: self.wbuf,
        }
    }
}

impl<R: Read, W: Write> Client<R, W, Camera> {
    fn next_message(&mut self) -> Result<IncomingMessage, Box<dyn Error>> {
        use IncomingMessage::*;

        let mut id = 0;
        self.rbuf.read_exact(slice::from_mut(&mut id))?;

        Ok(match id {
            0x20 => Plate(codec::from_reader(&mut self.rbuf)?),
            0x40 => WantHeartbeat(codec::from_reader(&mut self.rbuf)?),
            _ => return Err("unrecognized message".into()),
        })
    }
}

impl<R: Read, W: Write> Client<R, W, Dispatcher> {
    fn next_message(&mut self) -> Result<IncomingMessage, Box<dyn Error>> {
        use IncomingMessage::*;

        let mut id = 0;
        self.rbuf.read_exact(slice::from_mut(&mut id))?;

        Ok(match id {
            0x40 => WantHeartbeat(codec::from_reader(&mut self.rbuf)?),
            _ => return Err("unrecognized message".into()),
        })
    }
}

impl<R: Read, W: Write> Client<R, W> {
    fn new(r: R, w: W) -> Self {
        Self {
            kind: std::marker::PhantomData,
            rbuf: BufReader::new(r),
            wbuf: BufWriter::new(w),
        }
    }
}

#[test]
fn common_next_message() {
    let heartbeat = msg::WantHeartbeat { interval: 12345 };
    let camera = msg::IAmCamera {
        road: 0x4141,
        mile: 0xcafe,
        limit: 0xbabe,
    };
    let dispatcher = msg::IAmDispatcher {
        roads: vec![0xf00, 0xba6, 0xba2],
    };
    let input = codec::to_bytes(&(
        (0x40_u8, heartbeat.clone()),
        (0x80_u8, camera.clone()),
        (0x81_u8, dispatcher.clone()),
    ))
    .unwrap();

    let mut client = Client::new(&input[..], io::sink());
    assert_eq!(
        client.next_message().unwrap(),
        IncomingMessage::WantHeartbeat(heartbeat)
    );
    assert_eq!(
        client.next_message().unwrap(),
        IncomingMessage::IAmCamera(camera)
    );
    assert_eq!(
        client.next_message().unwrap(),
        IncomingMessage::IAmDispatcher(dispatcher)
    );
    assert!(client.next_message().is_err());
}

#[test]
fn into_camera() {
    let plates = vec![
        msg::Plate {
            plate: "hello".to_string(),
            timestamp: 1337,
        },
        msg::Plate {
            plate: "world".to_string(),
            timestamp: 7331,
        },
        msg::Plate {
            plate: "foo".to_string(),
            timestamp: 12345,
        },
    ];
    let input = codec::to_bytes(&(
        (0x20_u8, plates[0].clone()),
        (0x20_u8, plates[1].clone()),
        (0x20_u8, plates[2].clone()),
    ))
    .unwrap();

    let mut client = Client::new(&input[..], io::sink()).into_camera();
    assert_eq!(
        client.next_message().unwrap(),
        IncomingMessage::Plate(plates[0].clone())
    );
    assert_eq!(
        client.next_message().unwrap(),
        IncomingMessage::Plate(plates[1].clone())
    );
    assert_eq!(
        client.next_message().unwrap(),
        IncomingMessage::Plate(plates[2].clone())
    );
}
