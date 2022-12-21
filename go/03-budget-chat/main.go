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
			user.name = user.Receive()

			if err := chat.Join(user); err != nil {
				user.Send("lol no thanks")
				return
			}
		})
	}
	server.Wait()
}

type BudgetChat struct {
	users     map[string]User
	usersLock sync.RWMutex
}

func (b *BudgetChat) Join(user User) error {
	return errors.New("oh no")
}

type User struct {
	conn net.Conn
	name string
}

func (u *User) Send(msg string) {
	fmt.Fprintln(u.conn, msg)
}

func (u *User) Receive() string {
	var s string
	_, _ = fmt.Fscanln(u.conn, &s)
	return s
}
