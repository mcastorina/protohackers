use serde::{ser, Serialize};

use super::error::{Error, Result};

pub struct Serializer {
    // This vector starts empty and bytes are appended as values are serialized.
    output: Vec<u8>,
}

pub fn to_bytes<T>(value: &T) -> Result<Vec<u8>>
where
    T: Serialize,
{
    let mut serializer = Serializer { output: Vec::new() };
    value.serialize(&mut serializer)?;
    Ok(serializer.output)
}

impl<'a> ser::Serializer for &'a mut Serializer {
    type Ok = ();
    type Error = Error;

    type SerializeSeq = Self;
    type SerializeTuple = Self;
    type SerializeTupleStruct = Self;
    type SerializeTupleVariant = Self;
    type SerializeMap = Self;
    type SerializeStruct = Self;
    type SerializeStructVariant = Self;

    fn serialize_bool(self, _v: bool) -> Result<()> {
        Err(Error::UnsupportedType)
    }

    fn serialize_i8(self, _v: i8) -> Result<()> {
        Err(Error::UnsupportedType)
    }

    fn serialize_i16(self, _v: i16) -> Result<()> {
        Err(Error::UnsupportedType)
    }

    fn serialize_i32(self, _v: i32) -> Result<()> {
        Err(Error::UnsupportedType)
    }

    fn serialize_i64(self, _v: i64) -> Result<()> {
        Err(Error::UnsupportedType)
    }

    fn serialize_u8(self, v: u8) -> Result<()> {
        self.output.push(v);
        Ok(())
    }

    fn serialize_u16(self, v: u16) -> Result<()> {
        self.output.extend(v.to_be_bytes());
        Ok(())
    }

    fn serialize_u32(self, v: u32) -> Result<()> {
        self.output.extend(v.to_be_bytes());
        Ok(())
    }

    fn serialize_u64(self, _v: u64) -> Result<()> {
        Err(Error::UnsupportedType)
    }

    fn serialize_f32(self, _v: f32) -> Result<()> {
        Err(Error::UnsupportedType)
    }

    fn serialize_f64(self, _v: f64) -> Result<()> {
        Err(Error::UnsupportedType)
    }

    // Serialize a char as a single-character string.
    fn serialize_char(self, v: char) -> Result<()> {
        self.serialize_str(&v.to_string())
    }

    fn serialize_str(self, v: &str) -> Result<()> {
        let bytes = v.as_bytes();
        if bytes.len() > 255 {
            return Err(Error::StringTooLong);
        }
        if !v.is_ascii() {
            return Err(Error::ExpectedAsciiCharacter);
        }
        self.output.push(bytes.len() as u8);
        self.output.extend(bytes);
        Ok(())
    }

    fn serialize_bytes(self, v: &[u8]) -> Result<()> {
        if v.len() > 255 {
            return Err(Error::ArrayTooLong);
        }
        self.output.push(v.len() as u8);
        self.output.extend(v);
        Ok(())
    }

    // An absent optional is represented as the JSON `null`.
    fn serialize_none(self) -> Result<()> {
        Ok(())
    }

    fn serialize_some<T>(self, value: &T) -> Result<()>
    where
        T: ?Sized + Serialize,
    {
        value.serialize(self)
    }

    fn serialize_unit(self) -> Result<()> {
        Ok(())
    }

    fn serialize_unit_struct(self, _name: &'static str) -> Result<()> {
        Ok(())
    }

    fn serialize_unit_variant(
        self,
        _name: &'static str,
        _variant_index: u32,
        _variant: &'static str,
    ) -> Result<()> {
        Ok(())
    }

    fn serialize_newtype_struct<T>(self, _name: &'static str, value: &T) -> Result<()>
    where
        T: ?Sized + Serialize,
    {
        value.serialize(self)
    }

    fn serialize_newtype_variant<T>(
        self,
        _name: &'static str,
        _variant_index: u32,
        _variant: &'static str,
        value: &T,
    ) -> Result<()>
    where
        T: ?Sized + Serialize,
    {
        value.serialize(&mut *self)
    }

    fn serialize_seq(self, len: Option<usize>) -> Result<Self::SerializeSeq> {
        match len {
            Some(len) if len < 256 => {
                self.output.push(len as u8);
                Ok(self)
            }
            Some(_) => Err(Error::ArrayTooLong),
            None => Err(Error::UnsupportedType),
        }
    }

    fn serialize_tuple(self, _len: usize) -> Result<Self::SerializeTuple> {
        Ok(self)
    }

    fn serialize_tuple_struct(
        self,
        _name: &'static str,
        _len: usize,
    ) -> Result<Self::SerializeTupleStruct> {
        Ok(self)
    }

    fn serialize_tuple_variant(
        self,
        _name: &'static str,
        _variant_index: u32,
        variant: &'static str,
        _len: usize,
    ) -> Result<Self::SerializeTupleVariant> {
        variant.serialize(&mut *self)?;
        Ok(self)
    }

    fn serialize_map(self, _len: Option<usize>) -> Result<Self::SerializeMap> {
        Err(Error::UnsupportedType)
    }

    fn serialize_struct(self, _name: &'static str, _len: usize) -> Result<Self::SerializeStruct> {
        Ok(self)
    }

    fn serialize_struct_variant(
        self,
        _name: &'static str,
        _variant_index: u32,
        _variant: &'static str,
        _len: usize,
    ) -> Result<Self::SerializeStructVariant> {
        Ok(self)
    }
}

impl<'a> ser::SerializeSeq for &'a mut Serializer {
    type Ok = ();
    type Error = Error;

    // Serialize a single element of the sequence.
    fn serialize_element<T>(&mut self, value: &T) -> Result<()>
    where
        T: ?Sized + Serialize,
    {
        value.serialize(&mut **self)
    }

    // Close the sequence.
    fn end(self) -> Result<()> {
        Ok(())
    }
}

impl<'a> ser::SerializeTuple for &'a mut Serializer {
    type Ok = ();
    type Error = Error;

    fn serialize_element<T>(&mut self, value: &T) -> Result<()>
    where
        T: ?Sized + Serialize,
    {
        value.serialize(&mut **self)
    }

    fn end(self) -> Result<()> {
        Ok(())
    }
}

impl<'a> ser::SerializeTupleStruct for &'a mut Serializer {
    type Ok = ();
    type Error = Error;

    fn serialize_field<T>(&mut self, value: &T) -> Result<()>
    where
        T: ?Sized + Serialize,
    {
        value.serialize(&mut **self)
    }

    fn end(self) -> Result<()> {
        Ok(())
    }
}

impl<'a> ser::SerializeTupleVariant for &'a mut Serializer {
    type Ok = ();
    type Error = Error;

    fn serialize_field<T>(&mut self, value: &T) -> Result<()>
    where
        T: ?Sized + Serialize,
    {
        value.serialize(&mut **self)
    }

    fn end(self) -> Result<()> {
        Ok(())
    }
}

impl<'a> ser::SerializeMap for &'a mut Serializer {
    type Ok = ();
    type Error = Error;

    fn serialize_key<T>(&mut self, _key: &T) -> Result<()>
    where
        T: ?Sized + Serialize,
    {
        Ok(())
    }

    fn serialize_value<T>(&mut self, value: &T) -> Result<()>
    where
        T: ?Sized + Serialize,
    {
        value.serialize(&mut **self)
    }

    fn end(self) -> Result<()> {
        Ok(())
    }
}

impl<'a> ser::SerializeStruct for &'a mut Serializer {
    type Ok = ();
    type Error = Error;

    fn serialize_field<T>(&mut self, _key: &'static str, value: &T) -> Result<()>
    where
        T: ?Sized + Serialize,
    {
        value.serialize(&mut **self)
    }

    fn end(self) -> Result<()> {
        Ok(())
    }
}

impl<'a> ser::SerializeStructVariant for &'a mut Serializer {
    type Ok = ();
    type Error = Error;

    fn serialize_field<T>(&mut self, _key: &'static str, value: &T) -> Result<()>
    where
        T: ?Sized + Serialize,
    {
        value.serialize(&mut **self)
    }

    fn end(self) -> Result<()> {
        Ok(())
    }
}

////////////////////////////////////////////////////////////////////////////////

#[test]
fn test_struct() {
    #[derive(Serialize)]
    struct Test {
        int: u32,
        seq: Vec<u16>,
        msg: Option<&'static str>,
    }
    let test = Test {
        int: 1337,
        seq: vec![0, 1, 2],
        msg: Some("hello"),
    };
    let expected = [
        0x00, 0x00, 0x05, 0x39, // 1337
        3,    // array length
        0x00, 0x00, // item 0
        0x00, 0x01, // item 1
        0x00, 0x02, // item 2
        5,    // string length
        b'h', b'e', b'l', b'l', b'o',
    ];
    assert_eq!(to_bytes(&test).unwrap(), expected);
}

#[test]
fn test_none() {
    #[derive(Serialize)]
    struct Test {
        int: u32,
        seq: Vec<u16>,
        msg: Option<&'static str>,
    }
    let test = Test {
        int: 1337,
        seq: vec![0, 1, 2],
        msg: None,
    };
    let expected = [
        0x00, 0x00, 0x05, 0x39, // 1337
        3,    // array length
        0x00, 0x00, // item 0
        0x00, 0x01, // item 1
        0x00, 0x02, // item 2
    ];
    assert_eq!(to_bytes(&test).unwrap(), expected);
}

#[test]
fn test_tuple() {
    let test: (u32, Vec<u16>, String) = (1337, vec![0, 1, 2], "hello".to_string());
    let expected = [
        0x00, 0x00, 0x05, 0x39, // 1337
        3,    // array length
        0x00, 0x00, // item 0
        0x00, 0x01, // item 1
        0x00, 0x02, // item 2
        5,    // string length
        b'h', b'e', b'l', b'l', b'o',
    ];
    assert_eq!(to_bytes(&test).unwrap(), expected);
}

#[test]
fn test_option() {
    let test = Some("hello");
    let expected = [
        5, // string length
        b'h', b'e', b'l', b'l', b'o',
    ];
    assert_eq!(to_bytes(&test).unwrap(), expected);
}
