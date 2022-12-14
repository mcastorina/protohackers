use std::error::Error;
use std::io::{BufReader, BufWriter, Read, Write};
use std::net::{TcpListener, TcpStream};
use std::ops::RangeInclusive;
use std::thread;

fn main() {
    let listener = TcpListener::bind("0.0.0.0:1337").expect("could not bind to address");
    println!("listening on :1337");
    for stream in listener.incoming().filter_map(Result::ok) {
        println!("accepted new connection");
        thread::spawn(|| {
            let _ = mean(stream);
            println!("connection closed");
        });
    }
}

fn mean(stream: TcpStream) -> Result<(), Box<dyn Error>> {
    let mut reader = BufReader::new(&stream);
    let mut writer = BufWriter::new(&stream);

    let mut prices = Vec::new();
    let mut buf = [0; 9];
    loop {
        let request: Request = reader
            .read_exact(&mut buf)
            .map_err(|err| Box::new(err) as Box<dyn Error>)
            .and_then(|_| buf.try_into())?;
        match request {
            Request::Insert { timestamp, price } => {
                prices.push((timestamp, price));
            }
            Request::Query { time_range } => {
                let prices: Vec<_> = prices
                    .iter()
                    .filter_map(|(ts, price)| time_range.contains(ts).then_some(price))
                    .copied()
                    .collect();
                let mean = i32::checked_div(prices.iter().sum(), prices.len() as i32).unwrap_or(0);
                let _ = writer.write(&mean.to_be_bytes());
                let _ = writer.flush();
            }
        }
    }
}

#[derive(Debug)]
enum Request {
    Insert { timestamp: i32, price: i32 },
    Query { time_range: RangeInclusive<i32> },
}

impl TryFrom<[u8; 9]> for Request {
    type Error = Box<dyn Error>;

    fn try_from(slice: [u8; 9]) -> Result<Self, Self::Error> {
        let n1 = i32::from_be_bytes(slice[1..5].try_into()?);
        let n2 = i32::from_be_bytes(slice[5..].try_into()?);
        match slice[0] as char {
            'I' => Ok(Request::Insert {
                timestamp: n1,
                price: n2,
            }),
            'Q' => Ok(Request::Query {
                time_range: n1..=n2,
            }),
            _ => Err("unrecognized request".into()),
        }
    }
}
