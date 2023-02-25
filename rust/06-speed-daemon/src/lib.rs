mod codec;
mod msg;

use msg::{DeserializeMessage, Message, SerializeMessage};
use std::error::Error;
use std::io::{BufRead, BufReader, BufWriter, Read, Write};
use std::sync::mpsc;
use std::{io, slice, thread, time};

struct Common;
struct Camera;
struct Dispatcher;

#[derive(Debug)]
struct Client<R: Read, W: Write, Kind = Common> {
    kind: std::marker::PhantomData<Kind>,
    rbuf: BufReader<R>,
    wbuf: BufWriter<W>,
    heartbeat: Option<(time::Duration, time::Instant)>,
}

impl<R: Read, W: Write> Client<R, W> {
    fn next_message(&mut self) -> Result<msg::IncomingMessage, Box<dyn Error>> {
        use msg::IncomingMessage::*;

        let mut id = 0;
        self.rbuf.read_exact(slice::from_mut(&mut id))?;

        Ok(match id {
            msg::WantHeartbeat::ID => WantHeartbeat(codec::from_reader(&mut self.rbuf)?),
            msg::IAmCamera::ID => IAmCamera(codec::from_reader(&mut self.rbuf)?),
            msg::IAmDispatcher::ID => IAmDispatcher(codec::from_reader(&mut self.rbuf)?),
            _ => return Err("unrecognized message".into()),
        })
    }

    fn into_camera(self) -> Client<R, W, Camera> {
        Client {
            kind: std::marker::PhantomData,
            rbuf: self.rbuf,
            wbuf: self.wbuf,
            heartbeat: self.heartbeat,
        }
    }

    fn into_dispatcher(self) -> Client<R, W, Dispatcher> {
        Client {
            kind: std::marker::PhantomData,
            rbuf: self.rbuf,
            wbuf: self.wbuf,
            heartbeat: self.heartbeat,
        }
    }
}

impl<R: Read, W: Write> Client<R, W, Camera> {
    fn next_message(&mut self) -> Result<msg::IncomingMessage, Box<dyn Error>> {
        use msg::IncomingMessage::*;

        let mut id = 0;
        self.rbuf.read_exact(slice::from_mut(&mut id))?;

        Ok(match id {
            msg::Plate::ID => Plate(codec::from_reader(&mut self.rbuf)?),
            msg::WantHeartbeat::ID => WantHeartbeat(codec::from_reader(&mut self.rbuf)?),
            _ => return Err("unrecognized message".into()),
        })
    }
}

impl<R: Read, W: Write> Client<R, W, Dispatcher> {
    fn next_message(&mut self) -> Result<msg::IncomingMessage, Box<dyn Error>> {
        use msg::IncomingMessage::*;

        let mut id = 0;
        self.rbuf.read_exact(slice::from_mut(&mut id))?;

        Ok(match id {
            msg::WantHeartbeat::ID => WantHeartbeat(codec::from_reader(&mut self.rbuf)?),
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
            heartbeat: None,
        }
    }
}

impl<R: Read, W: Write, Kind> Client<R, W, Kind> {
    fn want_heartbeat(&mut self, heartbeat: msg::WantHeartbeat) -> Result<(), Box<dyn Error>> {
        if self.heartbeat.is_some() {
            return Err("heartbeat already requested".into());
        }
        let period = heartbeat.interval.into();
        self.heartbeat = Some((period, time::Instant::now() - period));
        Ok(())
    }

    // TODO: run_once can return another implementation of Kind
    fn run_once(&mut self) -> Result<(), Box<dyn Error>> {
        if let Some((period, last)) = self.heartbeat {
            if last.elapsed() > period {
                msg::Heartbeat.to_writer(&mut self.wbuf)?;
                let mut next = last;
                while next + period < time::Instant::now() {
                    next += period;
                }
                self.heartbeat = Some((period, next));
            }
        }
        Ok(())
    }
}

#[test]
fn common_next_message() {
    let heartbeat = msg::WantHeartbeat {
        interval: msg::Decisecond(12345),
    };
    let camera = msg::IAmCamera {
        road: 0x4141,
        mile: 0xcafe,
        limit: 0xbabe,
    };
    let dispatcher = msg::IAmDispatcher {
        roads: vec![0xf00, 0xba6, 0xba2],
    };
    let input = codec::to_bytes(&(
        (msg::WantHeartbeat::ID, heartbeat.clone()),
        (msg::IAmCamera::ID, camera.clone()),
        (msg::IAmDispatcher::ID, dispatcher.clone()),
    ))
    .unwrap();

    let mut client = Client::new(&input[..], io::sink());
    assert_eq!(
        client.next_message().unwrap(),
        msg::IncomingMessage::WantHeartbeat(heartbeat)
    );
    assert_eq!(
        client.next_message().unwrap(),
        msg::IncomingMessage::IAmCamera(camera)
    );
    assert_eq!(
        client.next_message().unwrap(),
        msg::IncomingMessage::IAmDispatcher(dispatcher)
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
        (msg::Plate::ID, plates[0].clone()),
        (msg::Plate::ID, plates[1].clone()),
        (msg::Plate::ID, plates[2].clone()),
    ))
    .unwrap();

    let mut client = Client::new(&input[..], io::sink()).into_camera();
    assert_eq!(
        client.next_message().unwrap(),
        msg::IncomingMessage::Plate(plates[0].clone())
    );
    assert_eq!(
        client.next_message().unwrap(),
        msg::IncomingMessage::Plate(plates[1].clone())
    );
    assert_eq!(
        client.next_message().unwrap(),
        msg::IncomingMessage::Plate(plates[2].clone())
    );
}

#[test]
fn test_heartbeat() {
    let mut output = Vec::new();
    let mut client = Client::new(&[][..], &mut output);
    client
        .want_heartbeat(msg::WantHeartbeat {
            interval: msg::Decisecond(1),
        })
        .unwrap();
    client.run_once();
    client.run_once();
    thread::sleep(time::Duration::from_millis(100));
    client.run_once();
    client.run_once();
    drop(client);

    assert_eq!(output, vec![msg::Heartbeat::ID; 2]);
}
