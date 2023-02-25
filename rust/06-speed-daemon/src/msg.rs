use super::codec;
use std::io::{Read, Write};
use std::time;

#[derive(Clone, Debug, PartialEq, Eq, serde::Deserialize, serde::Serialize)]
pub struct Error(String);

#[derive(Clone, Debug, PartialEq, Eq, serde::Deserialize, serde::Serialize)]
pub struct Ticket {
    pub plate: String,
    pub road: u16,
    pub mile1: u32,
    pub timestamp1: u32,
    pub mile2: u32,
    pub timestamp2: u32,
    pub speed: u16,
}

#[derive(Clone, Debug, PartialEq, Eq, serde::Deserialize, serde::Serialize)]
pub struct Heartbeat;

#[derive(Clone, Debug, PartialEq, Eq, serde::Deserialize, serde::Serialize)]
pub struct Plate {
    pub plate: String,
    pub timestamp: u32,
}

#[derive(Clone, Debug, PartialEq, Eq, serde::Deserialize, serde::Serialize)]
pub struct WantHeartbeat {
    pub interval: Decisecond,
}

#[derive(Copy, Clone, Debug, PartialEq, Eq, serde::Deserialize, serde::Serialize)]
pub struct Decisecond(pub u32);

impl From<Decisecond> for time::Duration {
    fn from(d: Decisecond) -> Self {
        time::Duration::from_millis(d.0 as u64 * 100)
    }
}

#[derive(Clone, Debug, PartialEq, Eq, serde::Deserialize, serde::Serialize)]
pub struct IAmCamera {
    pub road: u16,
    pub mile: u16,
    pub limit: u16,
}

#[derive(Clone, Debug, PartialEq, Eq, serde::Deserialize, serde::Serialize)]
pub struct IAmDispatcher {
    pub roads: Vec<u16>,
}

pub trait Message {
    const ID: u8;
}

pub trait SerializeMessage: Message + serde::Serialize {
    fn to_writer<W: Write>(&self, w: W) -> Result<(), Box<dyn std::error::Error>> {
        codec::to_writer(w, &(Self::ID, &self))?;
        Ok(())
    }
}

pub trait DeserializeMessage<'de>: Message + serde::Deserialize<'de> {
    fn from_reader<R: Read>(&self, r: R) -> Result<Self, Box<dyn std::error::Error>> {
        let (id, t): (u8, Self) = codec::from_reader(r)?;
        if id != Self::ID {
            return Err("wrong ID".into());
        }
        Ok(t)
    }
}

#[derive(Debug, PartialEq, Eq)]
pub enum IncomingMessage {
    IAmCamera(IAmCamera),
    IAmDispatcher(IAmDispatcher),
    WantHeartbeat(WantHeartbeat),
    Plate(Plate),
}

#[derive(Debug, PartialEq, Eq)]
pub enum OutgoingMessage {
    Heartbeat,
    Error(Error),
    Ticket(Ticket),
}

macro_rules! impl_message {
    ($t:ty = $value:expr) => {
        impl SerializeMessage for $t {}
        impl DeserializeMessage<'_> for $t {}
        impl Message for $t {
            const ID: u8 = $value;
        }
    };
}

impl_message!(Error = 0x10);
impl_message!(Plate = 0x20);
impl_message!(Ticket = 0x21);
impl_message!(WantHeartbeat = 0x40);
impl_message!(Heartbeat = 0x41);
impl_message!(IAmCamera = 0x80);
impl_message!(IAmDispatcher = 0x81);

// Example implementation of deserializing an enum.
#[cfg(test)]
mod test {
    use super::*;

    #[derive(Debug, PartialEq, Eq)]
    pub enum IncomingMessage {
        IAmCamera(IAmCamera),
        IAmDispatcher(IAmDispatcher),
        WantHeartbeat(WantHeartbeat),
        Plate(Plate),
    }

    impl<'de> serde::Deserialize<'de> for IncomingMessage {
        fn deserialize<D>(deserializer: D) -> Result<Self, D::Error>
        where
            D: serde::Deserializer<'de>,
        {
            struct Visitor;

            impl<'de> serde::de::Visitor<'de> for Visitor {
                type Value = IncomingMessage;

                fn expecting(&self, formatter: &mut std::fmt::Formatter) -> std::fmt::Result {
                    formatter.write_str("tuple")
                }

                fn visit_seq<V>(self, mut seq: V) -> Result<IncomingMessage, V::Error>
                where
                    V: serde::de::SeqAccess<'de>,
                {
                    let id: u8 = seq
                        .next_element()?
                        .ok_or_else(|| serde::de::Error::invalid_length(0, &self))?;
                    Ok(match id {
                        IAmCamera::ID => IncomingMessage::IAmCamera(
                            seq.next_element()?
                                .ok_or_else(|| serde::de::Error::invalid_length(1, &self))?,
                        ),
                        IAmDispatcher::ID => IncomingMessage::IAmDispatcher(
                            seq.next_element()?
                                .ok_or_else(|| serde::de::Error::invalid_length(1, &self))?,
                        ),
                        WantHeartbeat::ID => IncomingMessage::WantHeartbeat(
                            seq.next_element()?
                                .ok_or_else(|| serde::de::Error::invalid_length(1, &self))?,
                        ),
                        Plate::ID => IncomingMessage::Plate(
                            seq.next_element()?
                                .ok_or_else(|| serde::de::Error::invalid_length(1, &self))?,
                        ),
                        _ => {
                            return Err(serde::de::Error::custom(format!("unrecognized id: {id}")))
                        }
                    })
                }
            }
            deserializer.deserialize_tuple(2, Visitor)
        }
    }

    #[test]
    fn test_deserialize_enum() {
        let bytes = [
            0x80, // IAMCamera ID
            0x00, 0x01, // road
            0x00, 0x02, // mile
            0x00, 0x03, // limit
        ];
        let expected = IncomingMessage::IAmCamera(IAmCamera {
            road: 1,
            mile: 2,
            limit: 3,
        });
        assert_eq!(expected, codec::from_bytes(&bytes[..]).unwrap());
    }
}
