package util

import (
    "bufio"
    "bytes"
    "encoding/binary"
)
//for tcp msg processing

func Encode(message string) ([]byte, error) {
    //read message length
    var length int32 = int32(len(message))
    var pkg *bytes.Buffer = new(bytes.Buffer)
    //write header
    err := binary.Write(pkg, binary.LittleEndian, length)
    if err != nil {
       return nil, err
    }
    //write message body
    err = binary.Write(pkg, binary.LittleEndian, []byte(message))
    if err != nil {
       return nil, err
    }
    return pkg.Bytes(), nil
}

func Decode(reader *bufio.Reader) (string, error) {
    //read msg length
    lengthByte, _ := reader.Peek(4)
    lengthBuff := bytes.NewBuffer(lengthByte)
    var length int32
    err := binary.Read(lengthBuff, binary.LittleEndian, &length)
    if err != nil {
       return "", err
    }
    if int32(reader.Buffered()) < length + 4 {
       return "", err
    }

    //read msg body
    pack := make([]byte, int( 4 + length))
    _, err = reader.Read(pack)
    if err != nil {
       return "", err
    }
    return string(pack[4:]), nil
}

