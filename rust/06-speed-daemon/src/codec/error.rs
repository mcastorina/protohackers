use std;
use std::fmt::{self, Display};

use serde::{de, ser};

pub type Result<T> = std::result::Result<T, Error>;

// This is a bare-bones implementation. A real library would provide additional
// information in its error type, for example the line and column at which the
// error occurred, the byte offset into the input, or the current key being
// processed.
#[derive(Debug)]
pub enum Error {
    // One or more variants that can be created by data structures through the
    // `ser::Error` and `de::Error` traits. For example the Serialize impl for
    // Mutex<T> might return an error because the mutex is poisoned, or the
    // Deserialize impl for a struct may return an error because a required
    // field is missing.
    Message(String),

    // Zero or more variants that can be created directly by the Serializer and
    // Deserializer without going through `ser::Error` and `de::Error`. These
    // are specific to the format, in this case JSON.
    Eof,
    ExpectedAsciiCharacter,
    ExpectedSingleLengthString,
    TrailingBytes,
    UnsupportedType,
    StringTooLong,
    ArrayTooLong,
}

impl ser::Error for Error {
    fn custom<T: Display>(msg: T) -> Self {
        Error::Message(msg.to_string())
    }
}

impl de::Error for Error {
    fn custom<T: Display>(msg: T) -> Self {
        Error::Message(msg.to_string())
    }
}

impl Display for Error {
    fn fmt(&self, formatter: &mut fmt::Formatter) -> fmt::Result {
        match self {
            Error::Message(msg) => formatter.write_str(msg),
            Error::Eof => formatter.write_str("unexpected end of input"),
            Error::ExpectedAsciiCharacter => formatter.write_str("expected an ASCII character"),
            Error::ExpectedSingleLengthString => {
                formatter.write_str("expected a single ASCII character")
            }
            Error::TrailingBytes => formatter.write_str("trailing bytes"),
            Error::UnsupportedType => formatter.write_str("unsupported type"),
            Error::StringTooLong => {
                formatter.write_str("the provided string exceeds the max of 255 bytes")
            }
            Error::ArrayTooLong => {
                formatter.write_str("the provided slice or array exceeds the max of 255 elements")
            }
        }
    }
}

impl std::error::Error for Error {}
