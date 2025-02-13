package server

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"sync"
	"time"
)

var (
	Port            string
	maxConnectMutex sync.Mutex
	userName        string
	cnxA            int
	loggedOut       string
)

var users = make(map[string]bool)
var clients = make(map[net.Conn]string)
var timer = time.Now().Format("2006-01-02 15:04:05")

func ServerTCP() {
	listener, err := net.Listen("tcp", ":"+Port)
	CatchError(err)

	fmt.Println("Listening on the port :" + Port)
	EmptyFile()

	defer listener.Close()

	for {
		conn, err := listener.Accept()
		CatchError(err)
		go IncommingConnections(conn)

	}
}

func IncommingConnections(conn net.Conn) {

	if cnxA > 10 {
		return
	} else {
		fmt.Fprintln(conn, string(WelcomeMessage()))
		userName = ""
		for userName == "" {
			fmt.Fprint(conn, "[ENTER YOUR NAME]: ")
			userName = Reader(conn)
			if CheckUsername(strings.ToLower(userName), conn) || CorrectTheUsername(userName, conn) {
				userName = ""
			}
		}

		maxConnectMutex.Lock()
		cnxA++
		clients[conn] = strings.ToLower(userName)
		users[userName] = true
		maxConnectMutex.Unlock()
		LogSignal(conn)
		UpdateChat(conn)
		go BroadcastMessage(conn)
	}
}

func Reader(conn net.Conn) string {
	netData, err := bufio.NewReader(conn).ReadString('\n')
	netData = strings.Trim(netData, "\n")

	if err != nil {
		//if not a logged user will be deleted
		if err == io.EOF || CheckUsername(loggedOut, conn) {
			loggedOut = clients[conn]
			clients[conn] = "" // Set the username to an empty string to indicate disconnection
			users[loggedOut] = false
			return loggedOut
		} else {
			fmt.Println(err)
		}
	}
	return netData
}

func MessageWriter(conn net.Conn) string {
	var msg string
	for msg == "" {
		fmt.Fprint(conn, "["+timer+"]["+clients[conn]+"]:")
		msg = Reader(conn)
	}
	return msg
}

func BroadcastMessage(sender net.Conn) {
	for {
		var msg = MessageWriter(sender)

		// Check for disconnection
		for conn := range clients {
			if clients[conn] == "" {
				LogLogout(conn, loggedOut)
				return
			}
		}

		// Iterate over all connected clients and send the message
		for conn := range clients {
			if conn != sender {
				_, err := fmt.Fprintln(conn, "\n["+timer+"]["+clients[sender]+"]:"+msg)
				if err != nil {
					fmt.Println("Error writing to connection:", err)
					LogLogout(conn, loggedOut) // Disconnect client on write error
					return
				}
			}
		}

		saved := "[" + timer + "][" + clients[sender] + "]:" + msg
		SaveMessage(saved)

		// Write to the sender's connection
		for conn := range clients {
			if conn != sender {
				_, err := fmt.Fprint(conn, "["+timer+"]["+clients[conn]+"]:")
				CatchError(err)
			}
		}
		time.Sleep(1000 * time.Millisecond)

	}
}

func LogSignal(loger net.Conn) {
	for conn := range clients {
		if conn != loger {
			_, err := fmt.Fprintln(conn, "\n"+clients[loger]+" has joined our chat...")
			CatchError(err)
			_, err = fmt.Fprint(conn, "["+timer+"]["+clients[conn]+"]:")
			CatchError(err)
		}
	}
}

func LogLogout(disconnect net.Conn, loggedUser string) {

	for conn := range clients {
		if conn != disconnect && clients[conn] != "" {
			fmt.Fprintln(conn, "\n"+loggedUser+" has left our chat...")
			fmt.Fprint(conn, "["+timer+"]["+clients[conn]+"]:")
		}
	}

	// Remove the disconnected client from the clients map
	maxConnectMutex.Lock()
	delete(clients, disconnect)
	cnxA--
	maxConnectMutex.Unlock()
}

func UpdateChat(conn net.Conn) {

	file, err := os.Open("./savedChat.txt")
	CatchError(err)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		_, err := fmt.Fprintln(conn, scanner.Text())
		CatchError(err)
	}
	defer file.Close()
}

func EmptyFile() {
	filepath := "./savedChat.txt"
	// Open the file with read-write mode and truncate the content to zero bytes
	file, err := os.OpenFile(filepath, os.O_RDWR|os.O_TRUNC, 0666)
	CatchError(err)
	defer file.Close()
}

func SaveMessage(msg string) {
	file, err := os.OpenFile("./savedChat.txt", os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0666)
	CatchError(err)

	_, err = io.WriteString(file, msg+"\n")
	CatchError(err)

	defer file.Close()
}

func CheckUsername(user string, conn net.Conn) bool {
	if users[user] {
		fmt.Fprintln(conn, "This username has been already used!")
		return true
	}
	return false
}

func CorrectTheUsername(s string, conn net.Conn) bool {
	for _, val := range s {
		if val < 0 || (val > 9 && val < 'A') || (val > 'Z' && val < 'a') || val > 'z' {
			fmt.Fprintln(conn, "Username must be alphanumeric!")
			return true
		}
	}
	if len(s) > 20 {
		fmt.Fprintln(conn, "Too much characters!")
		return true
	}
	return false
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
		fmt.Println(err)
		return
	}
}
