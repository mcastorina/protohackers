package main

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"net"
	"strings"
	"sync"
)

func main() {
	server, err := NewServer()
	if err != nil {
		log.Fatal(err)
	}

	chat := BudgetChat{
		users: make(map[string]User),
	}

	for conn := range server.Connections() {
		server.Handle(conn, func(conn net.Conn) {
			user := NewUser(conn)

			user.Send("Whatsyo name, delicate?")
			user.name, _ = user.Receive()

			if err := chat.Join(user); err != nil {
				user.Send(fmt.Sprintf("lol no thanks: %s", err.Error()))
				return
			}
			defer chat.Leave(user)

			// start reading/writing incoming/outgoing messages
			for {
				msg, err := user.Receive()
				if err != nil {
					log.Println(err)
					break
				}
				chat.Broadcast(user.name, msg)
			}

		})
	}
	server.Wait()
}

type BudgetChat struct {
	users     map[string]User
	usersLock sync.RWMutex
}

func (b *BudgetChat) Broadcast(name, msg string) {
	b.usersLock.RLock()
	defer b.usersLock.RUnlock()

	for _, user := range b.users {
		if user.name == name {
			continue
		}
		user.Send(fmt.Sprintf("[%s] %s", name, msg))
	}
}

func (b *BudgetChat) Join(user User) error {
	// check if name has non-alpha numeric characters or is empty
	for _, r := range user.name {
		if (r < 'a' || r > 'z') && (r < 'A' || r > 'Z') && (r < '0' || r > '9') {
			return errors.New("Name is not alphanumeric")
		}
	}

	// check if name is empty
	if user.name == "" {
		return errors.New("Name is empty")
	}

	// check if name is taken
	b.usersLock.Lock()
	defer b.usersLock.Unlock()

	if _, ok := b.users[user.name]; ok {
		return errors.New("Name already taken")
	}

	b.users[user.name] = user

	// send presence message
	var userNames []string
	for _, u := range b.users {
		if u.name == user.name {
			continue
		}
		userNames = append(userNames, u.name)

		u.Send(fmt.Sprintf("* %s has joined the room", user.name))
	}

	user.Send(fmt.Sprintf("* The room contains: %v", userNames))

	return nil
}

func (b *BudgetChat) Leave(user User) {
	b.usersLock.Lock()
	defer b.usersLock.Unlock()
	delete(b.users, user.name)

	// send leave message
	for _, u := range b.users {
		u.Send(fmt.Sprintf("* %s has left the room", user.name))
	}
}

type User struct {
	conn net.Conn
	rbuf *bufio.Reader
	name string
}

func NewUser(conn net.Conn) User {
	return User{
		conn: conn,
		rbuf: bufio.NewReader(conn),
	}
}

func (u *User) Send(msg string) {
	fmt.Fprintln(u.conn, msg)
}

func (u *User) Receive() (string, error) {
	s, err := u.rbuf.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimRight(s, "\n\r"), nil
}
