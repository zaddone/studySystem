package main
import(
	"encoding/gob"
	"bytes"
)
type SelfUser struct {
	UserName string
	NickName string
}

type Msg struct {
	FromUserName string
	ToUserName string
	Content string
	old string
	MsgId string
	MsgType float64
	CreateTime float64
}

func (self *Msg) LoadByte (db []byte) error {
	return gob.NewDecoder(bytes.NewBuffer(db)).Decode(self)
}

func (self *Msg) ToByte () []byte {
	var network bytes.Buffer
	err := gob.NewEncoder(&network).Encode(self)
	if err != nil {
		panic(err)
	}
	return network.Bytes()
	//self.ToByte()
}
type MsgList struct{
	AddMsgList []*Msg
}
type Member struct {
	NickName string
	RemarkName string
	UserName string
}
type MemberList struct{
	MemberList []*Member
}
