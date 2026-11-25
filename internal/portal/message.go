package portal

import (
	"encoding/binary"
	"strings"

	"github.com/akrck02/go-renderer/failure"
)

// D-Bus sends messages as bytes: a fixed header, a list of header fields, then the body, each
// value aligned to its own size. Only what the portal needs is written here: strings, booleans,
// numbers, byte arrays and dictionaries of variants; any value can be read, to skip what is not
// asked for. The specification: https://dbus.freedesktop.org/doc/dbus-specification.html

// The kinds of message.
const (
	methodCall   byte = 1
	methodReturn byte = 2
	errorReply   byte = 3
	signal       byte = 4
)

// The header fields a message names, by their codes in the specification.
const (
	fieldPath        byte = 1
	fieldInterface   byte = 2
	fieldMember      byte = 3
	fieldErrorName   byte = 4
	fieldReplySerial byte = 5
	fieldDestination byte = 6
	fieldSignature   byte = 8
)

// littleEndian is the first byte of a message written little-endian, as every message here is.
const littleEndian = 'l'

// fixedHeaderSize is the size of the header before its fields: the byte order, the kind, the
// flags, the version, the body's length, the serial and the fields' length.
const fixedHeaderSize = 16

// message is a D-Bus message: its kind, its serial, the header fields read, and its body.
type message struct {
	kind                                           byte
	serial, replySerial                            uint32
	path, iface, member, errorName, signatureField string
	body                                           []any
}

// encoder writes values as D-Bus lays them out.
type encoder struct{ data []byte }

// align adds zero bytes until the length is a multiple of size.
func (writer *encoder) align(size int) {
	for len(writer.data)%size != 0 {
		writer.data = append(writer.data, 0)
	}
}

func (writer *encoder) byteValue(value byte) { writer.data = append(writer.data, value) }

func (writer *encoder) uint32Value(value uint32) {
	writer.align(4)
	writer.data = binary.LittleEndian.AppendUint32(writer.data, value)
}

func (writer *encoder) boolean(value bool) {
	if value {
		writer.uint32Value(1)
		return
	}
	writer.uint32Value(0)
}

// text writes a string, or an object path: its length, its bytes and a zero byte.
func (writer *encoder) text(value string) {
	writer.uint32Value(uint32(len(value)))
	writer.data = append(append(writer.data, value...), 0)
}

// signature writes a signature: its length in one byte, its letters and a zero byte.
func (writer *encoder) signature(value string) {
	writer.data = append(append(append(writer.data, byte(len(value))), value...), 0)
}

// array writes an array whose elements align to elementAlignment, with what elements writes.
func (writer *encoder) array(elementAlignment int, elements func()) {
	writer.uint32Value(0)
	lengthAt := len(writer.data) - 4
	writer.align(elementAlignment)
	start := len(writer.data)
	elements()
	binary.LittleEndian.PutUint32(writer.data[lengthAt:], uint32(len(writer.data)-start))
}

// option writes one entry of a dictionary of variants (a{sv}): its key, then the value's
// signature and the value, written by value.
func (writer *encoder) option(key, valueSignature string, value func()) {
	writer.align(8)
	writer.text(key)
	writer.signature(valueSignature)
	value()
}

// headerField writes one header field: its code and its value as a variant.
func (writer *encoder) headerField(code byte, valueSignature string, value func()) {
	writer.align(8)
	writer.byteValue(code)
	writer.signature(valueSignature)
	value()
}

// methodCallBytes is a method call, with its body already written and the body's signature.
func methodCallBytes(serial uint32, destination, path, iface, member, bodySignature string, body []byte) []byte {
	writer := &encoder{}
	writer.data = append(writer.data, littleEndian, methodCall, 0, 1)
	writer.uint32Value(uint32(len(body)))
	writer.uint32Value(serial)
	writer.array(8, func() {
		writer.headerField(fieldPath, "o", func() { writer.text(path) })
		writer.headerField(fieldInterface, "s", func() { writer.text(iface) })
		writer.headerField(fieldMember, "s", func() { writer.text(member) })
		writer.headerField(fieldDestination, "s", func() { writer.text(destination) })
		if bodySignature != "" {
			writer.headerField(fieldSignature, "g", func() { writer.signature(bodySignature) })
		}
	})
	writer.align(8)
	return append(writer.data, body...)
}

// decoder reads values as D-Bus lays them out; ok turns false at the first that does not fit.
type decoder struct {
	data     []byte
	position int
	ok       bool
}

// align skips to the next multiple of size.
func (reader *decoder) align(size int) {
	for reader.position%size != 0 {
		reader.position++
	}
}

// take gives the next count bytes, or nothing once they run out.
func (reader *decoder) take(count int) []byte {
	if !reader.ok || count < 0 || reader.position+count > len(reader.data) {
		reader.ok = false
		return nil
	}
	taken := reader.data[reader.position : reader.position+count]
	reader.position += count
	return taken
}

// uint32Value reads 4 bytes, aligned.
func (reader *decoder) uint32Value() uint32 {
	reader.align(4)
	if taken := reader.take(4); taken != nil {
		return binary.LittleEndian.Uint32(taken)
	}
	return 0
}

// text reads a string or an object path, and the zero byte after it.
func (reader *decoder) text() string {
	length := reader.uint32Value()
	taken := reader.take(int(length) + 1)
	if taken == nil {
		return ""
	}
	return string(taken[:length])
}

// signature reads a signature, and the zero byte after it.
func (reader *decoder) signature() string {
	length := reader.take(1)
	if length == nil {
		return ""
	}
	taken := reader.take(int(length[0]) + 1)
	if taken == nil {
		return ""
	}
	return string(taken[:length[0]])
}

// value reads one value of a complete type, which a signature starts with, and gives back the
// rest of the signature. Arrays and structs come as []any, a dictionary entry as a [2]any.
func (reader *decoder) value(signature string) (any, string) {
	if signature == "" {
		reader.ok = false
		return nil, ""
	}
	kind, rest := signature[0], signature[1:]
	switch kind {
	case 'y':
		if taken := reader.take(1); taken != nil {
			return taken[0], rest
		}
		return byte(0), rest
	case 'b':
		return reader.uint32Value() != 0, rest
	case 'n', 'q':
		reader.align(2)
		if taken := reader.take(2); taken != nil {
			return binary.LittleEndian.Uint16(taken), rest
		}
		return uint16(0), rest
	case 'i', 'u', 'h':
		return reader.uint32Value(), rest
	case 'x', 't', 'd':
		reader.align(8)
		if taken := reader.take(8); taken != nil {
			return binary.LittleEndian.Uint64(taken), rest
		}
		return uint64(0), rest
	case 's', 'o':
		return reader.text(), rest
	case 'g':
		return reader.signature(), rest
	case 'v':
		inner := reader.signature()
		value, _ := reader.value(inner)
		return value, rest
	case 'a':
		return reader.array(rest)
	case '(':
		return reader.structure(rest, ')')
	case '{':
		values, after := reader.structure(rest, '}')
		if len(values) != 2 {
			reader.ok = false
			return [2]any{}, after
		}
		return [2]any{values[0], values[1]}, after
	}
	reader.ok = false
	return nil, ""
}

// array reads an array whose element type starts the signature, and gives back the rest after it.
func (reader *decoder) array(signature string) (any, string) {
	element := completeType(signature)
	length := int(reader.uint32Value())
	reader.align(alignment(element))
	end := reader.position + length
	var values []any
	for reader.ok && reader.position < end {
		value, _ := reader.value(element)
		values = append(values, value)
	}
	if reader.position != end {
		reader.ok = false
	}
	return values, signature[len(element):]
}

// structure reads the values of a struct, or a dictionary entry, up to its closing letter.
func (reader *decoder) structure(signature string, closing byte) ([]any, string) {
	reader.align(8)
	var values []any
	for reader.ok && signature != "" && signature[0] != closing {
		var value any
		value, signature = reader.value(signature)
		values = append(values, value)
	}
	if signature == "" {
		reader.ok = false
		return values, ""
	}
	return values, signature[1:]
}

// completeType is the first complete type of a signature: "as" of "asu", "{sv}" of "{sv}".
func completeType(signature string) string {
	if signature == "" {
		return ""
	}
	switch signature[0] {
	case 'a':
		return "a" + completeType(signature[1:])
	case '(', '{':
		depth := 0
		for index := 0; index < len(signature); index++ {
			switch signature[index] {
			case '(', '{':
				depth++
			case ')', '}':
				depth--
				if depth == 0 {
					return signature[:index+1]
				}
			}
		}
		return signature
	}
	return signature[:1]
}

// alignment is the boundary a value of a type starts on.
func alignment(kind string) int {
	if kind == "" {
		return 1
	}
	switch kind[0] {
	case 'n', 'q':
		return 2
	case 'b', 'i', 'u', 'h', 's', 'o', 'a':
		return 4
	case 'x', 't', 'd', '(', '{':
		return 8
	}
	return 1
}

// bodyLengths reads, from a message's fixed header, how long its fields and its body are, and
// tells whether the header is one this package can read.
func bodyLengths(header []byte) (fieldsLength, bodyLength int, ok bool) {
	if len(header) < fixedHeaderSize || header[0] != littleEndian {
		return 0, 0, false
	}
	return int(binary.LittleEndian.Uint32(header[12:])), int(binary.LittleEndian.Uint32(header[4:])), true
}

// parseMessage reads a whole message: its header, padded to 8 bytes, then its body.
func parseMessage(data []byte, bodyStart int) (message, error) {
	reader := &decoder{data: data[:bodyStart], position: 12, ok: true}
	read := message{kind: data[1], serial: binary.LittleEndian.Uint32(data[8:])}
	fields, _ := reader.value("a(yv)")
	for _, field := range asList(fields) {
		pair := asList(field)
		if len(pair) != 2 {
			continue
		}
		code, _ := pair[0].(byte)
		switch value := pair[1].(type) {
		case string:
			read.setField(code, value)
		case uint32:
			if code == fieldReplySerial {
				read.replySerial = value
			}
		}
	}
	if !reader.ok {
		return message{}, BadMessage
	}
	body := &decoder{data: data[bodyStart:], ok: true}
	for signature := read.signatureField; signature != "" && body.ok; {
		var value any
		value, signature = body.value(signature)
		read.body = append(read.body, value)
	}
	if !body.ok {
		return message{}, failure.Wrap(BadMessage, read.member+" "+read.signatureField, nil)
	}
	return read, nil
}

// setField keeps one of the text header fields.
func (read *message) setField(code byte, value string) {
	switch code {
	case fieldPath:
		read.path = value
	case fieldInterface:
		read.iface = value
	case fieldMember:
		read.member = value
	case fieldErrorName:
		read.errorName = value
	case fieldSignature:
		read.signatureField = value
	}
}

// asList is an array or a struct read, or nothing for any other value.
func asList(value any) []any {
	list, _ := value.([]any)
	return list
}

// escapedName is a bus name as the portal writes it in a request's path: ":1.42" is "1_42".
func escapedName(name string) string {
	return strings.ReplaceAll(strings.TrimPrefix(name, ":"), ".", "_")
}
