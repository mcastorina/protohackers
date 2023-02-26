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
    fn next_message(&mut self) -> Result<Option<msg::IncomingMessage>, Box<dyn Error>> {
        use msg::IncomingMessage::*;

        let id = self.read_id()?;
        // If the client is too slow to provide all the bytes of a message, an error will be
        // returned and the order of messages possibly misaligned. That shouldn't matter though,
        // because an error here usually means the client should disconnect.
        id.map(|id| match id {
            msg::WantHeartbeat::ID => Ok(WantHeartbeat(self.read_message()?)),
            msg::IAmCamera::ID => Ok(IAmCamera(self.read_message()?)),
            msg::IAmDispatcher::ID => Ok(IAmDispatcher(self.read_message()?)),
            _ => Err("unrecognized message".into()),
        })
        .transpose()
    }

    // TODO: run_once can return another implementation of Kind
    fn run_once(&mut self) -> Result<(), Box<dyn Error>> {
        self.send_heartbeat()?;
        let msg = self.next_message()?;
        if msg.is_none() {
            // Nothing to do.
            return Ok(());
        }
        match msg.unwrap() {
            msg::IncomingMessage::WantHeartbeat(want_heartbeat) => {
                self.want_heartbeat(want_heartbeat)?
            }
            msg::IncomingMessage::IAmCamera(_) => {
                todo!()
            }
            msg::IncomingMessage::IAmDispatcher(_) => {
                todo!()
            }
            _ => unreachable!(),
        }
        Ok(())
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
    fn next_message(&mut self) -> Result<Option<msg::IncomingMessage>, Box<dyn Error>> {
        use msg::IncomingMessage::*;

        let id = self.read_id()?;
        // If the client is too slow to provide all the bytes of a message, an error will be
        // returned and the order of messages possibly misaligned. That shouldn't matter though,
        // because an error here usually means the client should disconnect.
        id.map(|id| match id {
            msg::Plate::ID => Ok(Plate(self.read_message()?)),
            msg::WantHeartbeat::ID => Ok(WantHeartbeat(self.read_message()?)),
            _ => Err("unrecognized message".into()),
        })
        .transpose()
    }
}

impl<R: Read, W: Write> Client<R, W, Dispatcher> {
    fn next_message(&mut self) -> Result<Option<msg::IncomingMessage>, Box<dyn Error>> {
        use msg::IncomingMessage::*;

        let id = self.read_id()?;
        // If the client is too slow to provide all the bytes of a message, an error will be
        // returned and the order of messages possibly misaligned. That shouldn't matter though,
        // because an error here usually means the client should disconnect.
        id.map(|id| match id {
            msg::WantHeartbeat::ID => Ok(WantHeartbeat(self.read_message()?)),
            _ => Err("unrecognized message".into()),
        })
        .transpose()
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

    fn send_heartbeat(&mut self) -> Result<(), Box<dyn Error>> {
        match self.heartbeat {
            None => (),
            Some((period, last)) if last.elapsed() < period => (),
            Some((period, last)) => {
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

    fn read_id(&mut self) -> Result<Option<u8>, Box<dyn Error>> {
        let mut id = 0;
        match self.rbuf.read_exact(slice::from_mut(&mut id)) {
            Ok(_) => Ok(Some(id)),
            Err(err) if err.kind() == io::ErrorKind::WouldBlock => Ok(None),
            Err(err) => Err(err.into()),
        }
    }

    fn read_message<'de, T: serde::Deserialize<'de>>(&mut self) -> Result<T, Box<dyn Error>> {
        // By this point we should already have received a message ID, so reading the message
        // shouldn't fail. A WouldBlock error means the client was too slow sending the data and we
        // partially read the message, losing bytes and getting in an inconsistent state. Since
        // this is soley for protohackers, this should be fine. Otherwise, a proper async library
        // should be used.
        let result = codec::from_reader(&mut self.rbuf);
        if let Err(codec::Error::IOError(ref err)) = result {
            if err.kind() == io::ErrorKind::WouldBlock {
                panic!("timed out reading message and lost data!");
            }
        }
        Ok(result?)
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
        client.next_message().unwrap().unwrap(),
        msg::IncomingMessage::WantHeartbeat(heartbeat)
    );
    assert_eq!(
        client.next_message().unwrap().unwrap(),
        msg::IncomingMessage::IAmCamera(camera)
    );
    assert_eq!(
        client.next_message().unwrap().unwrap(),
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
        client.next_message().unwrap().unwrap(),
        msg::IncomingMessage::Plate(plates[0].clone())
    );
    assert_eq!(
        client.next_message().unwrap().unwrap(),
        msg::IncomingMessage::Plate(plates[1].clone())
    );
    assert_eq!(
        client.next_message().unwrap().unwrap(),
        msg::IncomingMessage::Plate(plates[2].clone())
    );
}

#[test]
fn test_heartbeat() {
    let mut output = Vec::new();
    let mut client = Client::new(io::empty(), &mut output);
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
