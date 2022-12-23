use std::collections::HashMap;
use std::error::Error;
use std::io::{BufRead, BufReader, BufWriter, Write};
use std::net::{TcpListener, TcpStream};
use std::sync::{Arc, RwLock};
use std::thread;

type UserList = HashMap<String, BufWriter<TcpStream>>;

#[derive(Default, Clone)]
struct BudgetChat {
    user_list: Arc<RwLock<UserList>>,
}

impl BudgetChat {
    fn join(
        &mut self,
        name: &str,
        mut writer: BufWriter<TcpStream>,
    ) -> Result<Handle, Box<dyn Error + '_>> {
        // * Check if the name is taken.
        if self.user_list.read()?.contains_key(name) {
            return Err("invalid name".into());
        }

        // * Send list of users to this client.
        let _ = writer.write(b"* Users here: [")?;
        let _ = writer.write(self.current_users()?.join(", ").as_bytes())?;
        let _ = writer.write(b"]\n")?;
        writer.flush()?;

        // * Add the user to the user list.
        let _ = self.user_list.write()?.insert(name.to_string(), writer);

        Ok(Handle {
            chat: self.clone(),
            name: name.to_string(),
        })
    }

    fn current_users(&self) -> Result<Vec<String>, Box<dyn Error + '_>> {
        Ok(self.user_list.read()?.keys().cloned().collect::<Vec<_>>())
    }

    fn broadcast(&mut self, user: Option<&str>, msg: &str) -> Result<(), Box<dyn Error + '_>> {
        let msg = format!("* {}\n", msg.trim());
        for (name, writer) in self.user_list.write()?.iter_mut() {
            if user.is_some() && user.unwrap() == name {
                continue;
            }
            let _ = writer.write(msg.as_bytes())?;
            writer.flush()?;
        }
        Ok(())
    }
}

struct Handle {
    chat: BudgetChat,
    name: String,
}

impl Handle {
    fn send(&mut self, msg: &str) -> Result<(), Box<dyn Error + '_>> {
        let msg = format!("[{}] {}\n", self.name, msg.trim());
        for (user, writer) in self.chat.user_list.write()?.iter_mut() {
            if user == &self.name {
                continue;
            }
            let _ = writer.write(msg.as_bytes()).unwrap();
            writer.flush().unwrap();
        }
        Ok(())
    }
}

impl Drop for Handle {
    fn drop(&mut self) {
        if let Ok(mut user_list) = self.chat.user_list.write() {
            // * Remove user from users list.
            let _ = user_list.remove(&self.name);
        }
        // * Broadcast user left.
        let _ = self
            .chat
            .broadcast(None, &format!("{} has left", self.name));
    }
}

fn session(mut chat: BudgetChat, stream: TcpStream) -> Result<(), Box<dyn Error>> {
    let mut reader = BufReader::new(stream.try_clone()?);
    let mut writer = BufWriter::new(stream);

    // * Send intro message.
    let _ = writer.write(b"Sup, whatsyo name?\n")?;
    writer.flush()?;

    // * Set user's name.
    let mut name = String::new();
    reader.read_line(&mut name)?;
    let name = name.trim();

    // * Check name is alphanumeric.
    if name.is_empty() || !name.chars().all(|c| c.is_ascii_alphanumeric()) {
        return Err("invalid name".into());
    }

    // * Join the chatroom.
    let mut handle = chat.join(name, writer).map_err(|_| "error joining")?;

    // * Broadcast user joined.
    chat.broadcast(Some(name), &format!("{name} has joined"))
        .map_err(|_| "error broadcasting")?;

    // * Continuously receive messages and broadcast to the other users.
    let mut input = String::new();
    loop {
        input.clear();
        let nbytes = reader.read_line(&mut input)?;
        if nbytes == 0 {
            // * User has disconnected.
            break;
        }

        // * Broadcast message.
        handle.send(&input).map_err(|_| "error sending message")?;
    }
    Ok(())
}

fn main() {
    let listener = TcpListener::bind("0.0.0.0:1337").expect("could not bind to address");
    println!("listening on :1337");

    let chat: BudgetChat = Default::default();

    for stream in listener.incoming().filter_map(Result::ok) {
        println!("accepted new connection");

        let chat = chat.clone();
        thread::spawn(move || {
            let _ = session(chat, stream);
            println!("connection closed");
        });
    }
}
