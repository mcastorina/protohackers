package main

import (
	"errors"
	"fmt"
	"log"
	"net"
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
			user := User{conn: conn}

			user.Send("Whatsyo name, delicate?")
			user.name, _ = user.Receive()

			if err := chat.Join(user); err != nil {
				user.Send(fmt.Sprintf("lol no thanks: %s", err.Error()))
				return
			}

			// start reading/writing incoming/outgoing messages
			var wg sync.WaitGroup
			wg.Add(1)
			go func() {
				for {
					s, err := user.Receive()
					if err != nil {
						break
					}

					fmt.Printf("TODO broadcast %s", s)
				}

				wg.Done()
			}()

			wg.Wait()
		})
	}
	server.Wait()
}

type BudgetChat struct {
	users     map[string]User
	usersLock sync.RWMutex
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

type User struct {
	conn net.Conn
	name string
}

func (u *User) Send(msg string) {
	fmt.Fprintln(u.conn, msg)
}

func (u *User) Receive() (string, error) {
	var s string
	_, err := fmt.Fscanln(u.conn, &s)
	if err != nil {
		return "", err
	}
	return s, nil
}
