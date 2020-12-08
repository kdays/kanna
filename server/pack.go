package server

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"strings"
)

type DataPack struct {
	Op     string
	Data   []interface{}
	bMulti bool
	sMsgId string
}

func NewDataPack(op string) *DataPack {
	return &DataPack{
		Op: op,
	}
}

func (p *DataPack) SetMsgId(msgId string) {
	p.sMsgId = msgId
}

func (p *DataPack) Multi() {
	p.bMulti = true
}

func (p *DataPack) SetData(v []interface{}) {
	p.Data = v
}

func (p *DataPack) PushData(vals ...interface{}) {
	var d []interface{}
	for _, val := range vals {
		d = append(d, val)
	}

	if p.bMulti {
		p.Data = append(p.Data, d)
	} else {
		p.Data = d
	}
}

func (p *DataPack) Pack() []byte {
	msgFooter := ""
	if p.sMsgId != "" {
		msgFooter = p.sMsgId
	}

	var txt string
	if p.bMulti {
		s := p.Op + "{"
		for k, v := range p.Data {
			sMap := []string{}
			for _, vv := range v.([]interface{}) {
				sMap = append(sMap, toStr(vv))
			}

			s += strings.Join(sMap, "\t")
			if len(p.Data)-1 != k {
				s += "\n"
			}
		}

		txt = s + "}" + msgFooter
	} else {
		sMap := []string{}
		for _, v := range p.Data {
			sMap = append(sMap, toStr(v))
		}

		txt = p.Op + "{" + strings.Join(sMap, "\t") + "}" + msgFooter
	}

	txt = txt + "\n"
	var res []byte
	for _, v := range txt {
		if v == 0 { // why fill \u0000？
			continue
		}
		res = append(res, byte(v))
	}

	dataBuff := bytes.NewBuffer([]byte{})
	if err := binary.Write(dataBuff, binary.BigEndian, uint32(len(res))); err != nil {
		log.Println(err)
		return []byte{}
	}

	if err := binary.Write(dataBuff, binary.BigEndian, res); err != nil {
		log.Println(err)
		return []byte{}
	}

	return dataBuff.Bytes()
}

type OpCmd struct {
	Op    string
	MsgId string
	Args  []string
}

func ParseOp(b []byte) *OpCmd {
	s := strings.TrimSpace(string(b[:]))
	begin := strings.Index(s, "{")
	end := strings.Index(s, "}")

	if end < begin {
		log.Println("sth maybe wrong, end < begin", s)
		return nil
	}

	if begin == -1 || end == -1 {
		return nil
	}

	msgId := strings.Replace(s[end+1:], "\n", "", -1)

	return &OpCmd{
		Op:    s[0:begin],
		Args:  strings.Split(s[begin+1:end], ","),
		MsgId: msgId,
	}
}

func toStr(v interface{}) string {
	return fmt.Sprintf("%v", v)
}
