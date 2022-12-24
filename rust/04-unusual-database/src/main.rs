use std::collections::HashMap;
use std::net::UdpSocket;

const VERSION: &str = "dev";

fn main() {
    #[cfg(not(feature = "fly"))]
    let socket = UdpSocket::bind(":::1337").expect("could not bind to address");
    #[cfg(feature = "fly")]
    let socket = UdpSocket::bind("fly-global-services:1337").expect("could not bind to address");

    println!("listening on udp {}", socket.local_addr().unwrap());

    let mut db: HashMap<String, String> = Default::default();

    let mut buf = [0; 1024];
    while let Ok((n, addr)) = socket.recv_from(&mut buf) {
        println!("received new packet");
        let request = String::from_utf8_lossy(&buf[..n]);
        if let Some((key, value)) = request.split_once('=') {
            db.insert(key.to_string(), value.to_string());
            continue;
        }
        let value = match request.as_ref() {
            "version" => VERSION,
            key => db.get(key).map(AsRef::as_ref).unwrap_or_default(),
        };
        let response = format!("{request}={value}");
        let _ = socket.send_to(response.as_bytes(), &addr);
    }
}
