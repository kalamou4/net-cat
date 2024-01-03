package server

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strings"
	"sync"
	"time"
)

var (
	Port            string
	maxConnectMutex sync.Mutex
	userName        = ""
)

func ServerTCP() {
	listener, err := net.Listen("tcp", ":"+Port)
	CatchError(err)

	fmt.Println("Listening on the port :" + Port)

	defer listener.Close()

	for {
		conn, err := listener.Accept()
		CatchError(err)
		go IncommingConnections(conn)

	}
}

func IncommingConnections(conn net.Conn) {
	var cnxA int
	var tab []net.Conn
	if cnxA > 10 {
		return
	} else {
		fmt.Fprint(conn,string(WelcomeMessage()))
		//conn.Write(WelcomeMessage())
		for userName == "" {
			conn.Write([]byte("[ENTER YOUR NAME]: "))
			userName = Reader(conn)
		}

		maxConnectMutex.Lock()
		cnxA++
		maxConnectMutex.Unlock()

		tab = append(tab, conn)
		go Writer(conn, tab)
	}

}

func Reader(conn net.Conn) string {

	// Read data from the client
	netData, err := bufio.NewReader(conn).ReadString('\n')
	netData = strings.Trim(netData, "\n")

	if err != nil {
		if err == io.EOF {
			return "/logout"
		} else {
			log.Fatal("Error:", err)
		}
	}
	return netData
}

func Writer(conn net.Conn, tab []net.Conn) {

		time := time.Now().Format("01-01-1889 13:45:45")
		writer := bufio.NewWriter(conn)
		scanner := bufio.NewScanner(conn)

		for scanner.Scan() {
			message := scanner.Text()

			_, err := writer.WriteString("[" + time + "][" + userName + "]:" + message)
			if err != io.EOF {
				CatchError(err)
			}
		}
		writer.Flush()
}

func WelcomeMessage() []byte {
	file, err := os.Open("./pingoin.txt")
	CatchError(err)

	defer file.Close()

	buffer := make([]byte, 1024)
	n, err := file.Read(buffer)
	CatchError(err)

	return buffer[:n]
}

func CatchError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
