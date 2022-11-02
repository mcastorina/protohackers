package main

import (
	"context"
	"net"
)

func echo(ctx context.Context, conn net.Conn) error {
	data := make([]byte, 1024)
	for {
		totalRead, err := conn.Read(data)
		if err != nil {
			return err
		}
		if totalRead == 0 {
			return nil
		}
		for totalWrote := 0; totalWrote != totalRead; {
			wrote, err := conn.Write(data[totalWrote:totalRead])
			if err != nil {
				return err
			}
			totalWrote += wrote
		}
	}
}
