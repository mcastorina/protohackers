use std::collections::HashMap;
use std::io::{BufRead, BufReader, BufWriter, Write};
use std::net::{TcpListener, TcpStream};
use std::sync::{Arc, RwLock};
use std::thread;

fn main() {
    let listener = TcpListener::bind("0.0.0.0:1337").expect("could not bind to address");
    println!("listening on :1337");

    let user_list: Arc<RwLock<HashMap<String, BufWriter<TcpStream>>>> = Default::default();

    for stream in listener.incoming().filter_map(Result::ok) {
        println!("accepted new connection");

        let user_list = Arc::clone(&user_list);
        // let messages_tx = messages_tx.clone();
        thread::spawn(move || {
            let mut reader = BufReader::new(stream.try_clone().unwrap());
            let mut writer = BufWriter::new(stream);

            // * Send intro message.
            writer.write(b"Sup, whatsyo name?\n").unwrap();
            writer.flush().unwrap();

            // * Set user's name.
            let mut name = String::new();
            reader.read_line(&mut name).unwrap();
            let name = name.trim().to_string();

            // * Check name is alphanumeric.
            if name.len() == 0 || !name.chars().all(|c| c.is_ascii_alphanumeric()) {
                return;
            }

            // * Join the chatroom.
            {
                let mut user_list = user_list.write().unwrap();

                // * Check if the name is taken.
                if user_list.contains_key(&name) {
                    return;
                }

                // * Send list of users to this client.
                writer.write(b"* Users here: [").unwrap();
                let joined_users = user_list
                    .keys()
                    .map(AsRef::as_ref)
                    .collect::<Vec<_>>()
                    .join(", ");
                writer.write(joined_users.as_bytes()).unwrap();
                writer.write(b"]\n").unwrap();
                writer.flush().unwrap();

                // * Broadcast user joined.
                let msg = format!("* {name} has joined\n");
                for (_, writer) in user_list.iter_mut() {
                    writer.write(msg.as_bytes()).unwrap();
                    writer.flush().unwrap();
                }

                // * Add the user to the user list.
                let _ = user_list.insert(name.clone(), writer);
            }

            // * Continuously receive messages and broadcast to the other users.
            let mut input = String::new();
            loop {
                input.clear();
                let result = reader.read_line(&mut input);
                match result {
                    Ok(0) | Err(_) => break,
                    _ => (),
                }

                // * Broadcast message.
                let msg = format!("[{name}] {input}");
                for (user, writer) in user_list.write().unwrap().iter_mut() {
                    if user == &name {
                        continue;
                    }
                    writer.write(msg.as_bytes()).unwrap();
                    writer.flush().unwrap();
                }
            }

            // * Disconnect user.
            {
                let mut user_list = user_list.write().unwrap();

                // * Remove user from users list.
                let _ = user_list.remove(&name);

                // * Broadcast user left.
                let msg = format!("* {name} has left\n");
                for (_, writer) in user_list.iter_mut() {
                    writer.write(msg.as_bytes()).unwrap();
                    writer.flush().unwrap();
                }
            }

            println!("connection closed");
        });
    }
}
