package portal

import (
	"bufio"
	"encoding/hex"
	"io"
	"net"
	"strconv"
	"strings"

	"github.com/akrck02/go-renderer/failure"
)

// bus is a connection to the session's message bus: its name there, the serial of the last
// message sent, and the signals read while a reply was awaited.
type bus struct {
	connection net.Conn
	reader     *bufio.Reader
	name       string
	serial     uint32
	waiting    []message
}

// dial connects to the bus at an address as D-Bus writes them ("unix:path=/run/user/1000/bus",
// or "unix:abstract=…"), lets the bus know who this program is, and says hello.
func dial(address string, userID int) (*bus, error) {
	network, place, found := socketOf(address)
	if !found {
		return nil, failure.Wrap(NoBus, address, nil)
	}
	connection, err := net.Dial(network, place)
	if err != nil {
		return nil, failure.Wrap(NoBus, address, err)
	}
	opened := &bus{connection: connection, reader: bufio.NewReader(connection)}
	if err := opened.authenticate(userID); err != nil {
		connection.Close() // the bus refused it already: there is nothing left to close cleanly
		return nil, err
	}
	reply, err := opened.call("org.freedesktop.DBus", "/org/freedesktop/DBus", "org.freedesktop.DBus", "Hello", "", nil)
	if err != nil {
		connection.Close() // as above
		return nil, err
	}
	if len(reply.body) > 0 {
		opened.name, _ = reply.body[0].(string)
	}
	return opened, nil
}

// socketOf is the first Unix socket of an address the bus gives: the network for net.Dial, and
// its path, an abstract one starting with @.
func socketOf(address string) (network, place string, found bool) {
	for _, each := range strings.Split(address, ";") {
		kind, keys, cut := strings.Cut(each, ":")
		if !cut || kind != "unix" {
			continue
		}
		for _, pair := range strings.Split(keys, ",") {
			key, value, _ := strings.Cut(pair, "=")
			switch key {
			case "path":
				return "unix", value, true
			case "abstract":
				return "unix", "@" + value, true
			}
		}
	}
	return "", "", false
}

// authenticate tells the bus this program runs as the user, as every program of the session does.
func (opened *bus) authenticate(userID int) error {
	hexUser := hex.EncodeToString([]byte(strconv.Itoa(userID)))
	if _, err := io.WriteString(opened.connection, "\x00AUTH EXTERNAL "+hexUser+"\r\n"); err != nil {
		return failure.Wrap(NoBus, "", err)
	}
	answer, err := opened.reader.ReadString('\n')
	if err != nil || !strings.HasPrefix(answer, "OK ") {
		return failure.Wrap(Refused, strings.TrimSpace(answer), err)
	}
	if _, err := io.WriteString(opened.connection, "BEGIN\r\n"); err != nil {
		return failure.Wrap(NoBus, "", err)
	}
	return nil
}

// call calls a method and waits for its reply; the signals read meanwhile wait for nextSignal.
func (opened *bus) call(destination, path, iface, member, signature string, body []byte) (message, error) {
	opened.serial++
	serial := opened.serial
	if _, err := opened.connection.Write(methodCallBytes(serial, destination, path, iface, member, signature, body)); err != nil {
		return message{}, failure.Wrap(NoBus, member, err)
	}
	for {
		read, err := opened.read()
		if err != nil {
			return message{}, err
		}
		switch {
		case read.kind == signal:
			opened.waiting = append(opened.waiting, read)
		case read.replySerial != serial:
			continue // a reply to nothing asked here
		case read.kind == errorReply:
			return message{}, failure.Wrap(CallFailed, read.errorName, nil)
		case read.kind == methodReturn:
			return read, nil
		}
	}
}

// nextSignal waits for the signal of a member sent from an object path.
func (opened *bus) nextSignal(path, member string) (message, error) {
	for index, each := range opened.waiting {
		if each.path == path && each.member == member {
			opened.waiting = append(opened.waiting[:index], opened.waiting[index+1:]...)
			return each, nil
		}
	}
	for {
		read, err := opened.read()
		if err != nil {
			return message{}, err
		}
		if read.kind == signal && read.path == path && read.member == member {
			return read, nil
		}
	}
}

// read reads the next message from the bus.
func (opened *bus) read() (message, error) {
	header := make([]byte, fixedHeaderSize)
	if _, err := io.ReadFull(opened.reader, header); err != nil {
		return message{}, failure.Wrap(NoBus, "", err)
	}
	fieldsLength, bodyLength, ok := bodyLengths(header)
	if !ok {
		return message{}, BadMessage
	}
	bodyStart := fixedHeaderSize + fieldsLength
	for bodyStart%8 != 0 {
		bodyStart++
	}
	data := make([]byte, bodyStart+bodyLength)
	copy(data, header)
	if _, err := io.ReadFull(opened.reader, data[fixedHeaderSize:]); err != nil {
		return message{}, failure.Wrap(NoBus, "", err)
	}
	return parseMessage(data, bodyStart)
}

// close ends the connection.
func (opened *bus) close() {
	opened.connection.Close() // done with the bus: an error closing it changes nothing here
}
