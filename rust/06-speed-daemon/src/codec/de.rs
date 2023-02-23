use std::str;

use serde::de::{self, DeserializeSeed, MapAccess, SeqAccess, Visitor};
use serde::Deserialize;

use super::error::{Error, Result};

pub struct Deserializer<'de> {
    // This string starts with the input data and characters are truncated off
    // the beginning as data is parsed.
    input: &'de [u8],
}

impl<'de> Deserializer<'de> {
    pub fn from_bytes(input: &'de [u8]) -> Self {
        Deserializer { input }
    }
}

// Deserialize from a slice of bytes.
pub fn from_bytes<'a, T>(s: &'a [u8]) -> Result<T>
where
    T: Deserialize<'a>,
{
    let mut deserializer = Deserializer::from_bytes(s);
    let t = T::deserialize(&mut deserializer)?;
    if deserializer.input.is_empty() {
        Ok(t)
    } else {
        Err(Error::TrailingBytes)
    }
}

impl<'de> Deserializer<'de> {
    // Look at the first character in the input without consuming it.
    fn peek_byte(&mut self) -> Result<u8> {
        self.input.iter().next().ok_or(Error::Eof).cloned()
    }

    // Consume the first character in the input.
    fn next_byte(&mut self) -> Result<u8> {
        let b = self.peek_byte()?;
        self.input = &self.input[1..];
        Ok(b)
    }

    // Parse a length prefix string of ASCII characters.
    fn parse_ascii(&mut self) -> Result<&'de str> {
        let len = self.next_byte()? as usize;
        if self.input.len() < len {
            return Err(Error::Eof);
        }
        let s = str::from_utf8(&self.input[..len]).map_err(|_| Error::ExpectedAsciiCharacter)?;
        if !s.is_ascii() {
            return Err(Error::ExpectedAsciiCharacter);
        }
        self.input = &self.input[len..];
        Ok(s)
    }

    // Parse a single u8.
    fn parse_u8(&mut self) -> Result<u8> {
        self.next_byte()
    }

    // Parse a big-endian encoded u16.
    fn parse_u16(&mut self) -> Result<u16> {
        if self.input.len() < 2 {
            return Err(Error::Eof);
        }
        let bytes = &self.input[..2];
        self.input = &self.input[2..];
        Ok(u16::from_be_bytes(bytes.try_into().unwrap()))
    }

    // Parse a big-endian encoded u32.
    fn parse_u32(&mut self) -> Result<u32> {
        if self.input.len() < 4 {
            return Err(Error::Eof);
        }
        let bytes = &self.input[..4];
        self.input = &self.input[4..];
        Ok(u32::from_be_bytes(bytes.try_into().unwrap()))
    }
}

impl<'de, 'a> de::Deserializer<'de> for &'a mut Deserializer<'de> {
    type Error = Error;

    fn deserialize_any<V>(self, _visitor: V) -> Result<V::Value>
    where
        V: Visitor<'de>,
    {
        unimplemented!()
    }

    fn deserialize_bool<V>(self, _visitor: V) -> Result<V::Value>
    where
        V: Visitor<'de>,
    {
        unimplemented!()
    }

    fn deserialize_i8<V>(self, _visitor: V) -> Result<V::Value>
    where
        V: Visitor<'de>,
    {
        unimplemented!()
    }

    fn deserialize_i16<V>(self, _visitor: V) -> Result<V::Value>
    where
        V: Visitor<'de>,
    {
        unimplemented!()
    }

    fn deserialize_i32<V>(self, _visitor: V) -> Result<V::Value>
    where
        V: Visitor<'de>,
    {
        unimplemented!()
    }

    fn deserialize_i64<V>(self, _visitor: V) -> Result<V::Value>
    where
        V: Visitor<'de>,
    {
        unimplemented!()
    }

    fn deserialize_u8<V>(self, visitor: V) -> Result<V::Value>
    where
        V: Visitor<'de>,
    {
        visitor.visit_u8(self.parse_u8()?)
    }

    fn deserialize_u16<V>(self, visitor: V) -> Result<V::Value>
    where
        V: Visitor<'de>,
    {
        visitor.visit_u16(self.parse_u16()?)
    }

    fn deserialize_u32<V>(self, visitor: V) -> Result<V::Value>
    where
        V: Visitor<'de>,
    {
        visitor.visit_u32(self.parse_u32()?)
    }

    fn deserialize_u64<V>(self, _visitor: V) -> Result<V::Value>
    where
        V: Visitor<'de>,
    {
        unimplemented!()
    }

    fn deserialize_f32<V>(self, _visitor: V) -> Result<V::Value>
    where
        V: Visitor<'de>,
    {
        unimplemented!()
    }

    fn deserialize_f64<V>(self, _visitor: V) -> Result<V::Value>
    where
        V: Visitor<'de>,
    {
        unimplemented!()
    }

    fn deserialize_char<V>(self, visitor: V) -> Result<V::Value>
    where
        V: Visitor<'de>,
    {
        // Parse a string, check that it is one character, call `visit_char`.
        let s = self.parse_ascii()?;
        if s.len() != 1 {
            return Err(Error::ExpectedSingleLengthString);
        }
        visitor.visit_char(s.bytes().next().unwrap() as char)
    }

    fn deserialize_str<V>(self, visitor: V) -> Result<V::Value>
    where
        V: Visitor<'de>,
    {
        visitor.visit_borrowed_str(self.parse_ascii()?)
    }

    fn deserialize_string<V>(self, visitor: V) -> Result<V::Value>
    where
        V: Visitor<'de>,
    {
        self.deserialize_str(visitor)
    }

    fn deserialize_bytes<V>(self, visitor: V) -> Result<V::Value>
    where
        V: Visitor<'de>,
    {
        let len = self.next_byte()? as usize;
        visitor.visit_seq(LengthPrefix::new(self, len))
    }

    fn deserialize_byte_buf<V>(self, visitor: V) -> Result<V::Value>
    where
        V: Visitor<'de>,
    {
        self.deserialize_bytes(visitor)
    }

    fn deserialize_option<V>(self, visitor: V) -> Result<V::Value>
    where
        V: Visitor<'de>,
    {
        visitor.visit_some(self)
    }

    fn deserialize_unit<V>(self, visitor: V) -> Result<V::Value>
    where
        V: Visitor<'de>,
    {
        visitor.visit_unit()
    }

    fn deserialize_unit_struct<V>(self, _name: &'static str, visitor: V) -> Result<V::Value>
    where
        V: Visitor<'de>,
    {
        self.deserialize_unit(visitor)
    }

    fn deserialize_newtype_struct<V>(self, _name: &'static str, visitor: V) -> Result<V::Value>
    where
        V: Visitor<'de>,
    {
        visitor.visit_newtype_struct(self)
    }

    fn deserialize_seq<V>(self, visitor: V) -> Result<V::Value>
    where
        V: Visitor<'de>,
    {
        let len = self.next_byte()? as usize;
        visitor.visit_seq(LengthPrefix::new(self, len))
    }

    fn deserialize_tuple<V>(self, len: usize, visitor: V) -> Result<V::Value>
    where
        V: Visitor<'de>,
    {
        visitor.visit_seq(LengthPrefix::new(self, len))
    }

    // Tuple structs look just like sequences in JSON.
    fn deserialize_tuple_struct<V>(
        self,
        _name: &'static str,
        _len: usize,
        visitor: V,
    ) -> Result<V::Value>
    where
        V: Visitor<'de>,
    {
        self.deserialize_seq(visitor)
    }

    // Much like `deserialize_seq` but calls the visitors `visit_map` method
    // with a `MapAccess` implementation, rather than the visitor's `visit_seq`
    // method with a `SeqAccess` implementation.
    fn deserialize_map<V>(self, _visitor: V) -> Result<V::Value>
    where
        V: Visitor<'de>,
    {
        unimplemented!()
    }

    fn deserialize_struct<V>(
        self,
        _name: &'static str,
        fields: &'static [&'static str],
        visitor: V,
    ) -> Result<V::Value>
    where
        V: Visitor<'de>,
    {
        visitor.visit_seq(LengthPrefix::new(self, fields.len()))
    }

    fn deserialize_enum<V>(
        self,
        _name: &'static str,
        _variants: &'static [&'static str],
        _visitor: V,
    ) -> Result<V::Value>
    where
        V: Visitor<'de>,
    {
        unimplemented!()
    }

    fn deserialize_identifier<V>(self, visitor: V) -> Result<V::Value>
    where
        V: Visitor<'de>,
    {
        self.deserialize_str(visitor)
    }

    fn deserialize_ignored_any<V>(self, _visitor: V) -> Result<V::Value>
    where
        V: Visitor<'de>,
    {
        unimplemented!()
    }
}

struct LengthPrefix<'a, 'de: 'a> {
    de: &'a mut Deserializer<'de>,
    count: usize,
}

impl<'a, 'de> LengthPrefix<'a, 'de> {
    fn new(de: &'a mut Deserializer<'de>, count: usize) -> Self {
        LengthPrefix { de, count }
    }
}

// `SeqAccess` is provided to the `Visitor` to give it the ability to iterate
// through elements of the sequence.
impl<'de, 'a> SeqAccess<'de> for LengthPrefix<'a, 'de> {
    type Error = Error;

    fn next_element_seed<T>(&mut self, seed: T) -> Result<Option<T::Value>>
    where
        T: DeserializeSeed<'de>,
    {
        if self.count == 0 {
            return Ok(None);
        }
        self.count -= 1;
        seed.deserialize(&mut *self.de).map(Some)
    }
}

////////////////////////////////////////////////////////////////////////////////

#[test]
fn test_struct() {
    #[derive(Deserialize, PartialEq, Debug)]
    struct Test {
        int: u32,
        seq: Vec<u16>,
        msg: Option<String>,
    }
    let bytes = [
        0x00, 0x00, 0x05, 0x39, // 1337
        3,    // array length
        0x00, 0x00, // item 0
        0x00, 0x01, // item 1
        0x00, 0x02, // item 2
        5,    // string length
        b'h', b'e', b'l', b'l', b'o',
    ];
    let expected = Test {
        int: 1337,
        seq: vec![0, 1, 2],
        msg: Some("hello".to_string()),
    };
    assert_eq!(expected, from_bytes(&bytes[..]).unwrap());
}

#[test]
fn test_tuple() {
    let bytes = [
        0x00, 0x00, 0x05, 0x39, // 1337
        3,    // array length
        0x00, 0x00, // item 0
        0x00, 0x01, // item 1
        0x00, 0x02, // item 2
        5,    // string length
        b'h', b'e', b'l', b'l', b'o',
    ];
    let expected: (u32, Vec<u16>, String) = (1337, vec![0, 1, 2], "hello".to_string());
    assert_eq!(expected, from_bytes(&bytes[..]).unwrap());
}

#[test]
fn test_u32() {
    let bytes = [0x00, 0x00, 0x05, 0x39];
    let expected: u32 = 1337;
    assert_eq!(expected, from_bytes(&bytes[..]).unwrap());
}

#[test]
fn test_u32s() {
    let bytes = [2, 0x00, 0x00, 0x05, 0x39, 0xca, 0xfe, 0xba, 0xbe];
    let expected = vec![1337, 0xcafebabe];
    assert_eq!(expected, from_bytes::<Vec<u32>>(&bytes[..]).unwrap());
}
