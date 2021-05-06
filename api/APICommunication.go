package api

import (
	"/.."
	"net"
)

func listen() {
	l, err := net.Listen("tcp", main.Conf.ApiAddressDHT)
}
