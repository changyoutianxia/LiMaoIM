package lmproto

import (
	"fmt"

	"github.com/pkg/errors"
)

// RecvPacket 收到消息的包
type RecvPacket struct {
	Framer
	Setting
	MsgKey      string // 用于验证此消息是否合法（仿中间人篡改）
	MessageID   int64  // 服务端的消息ID(全局唯一)
	MessageSeq  uint32 // 消息序列号 （用户唯一，有序递增）
	ClientMsgNo string // 客户端唯一标示
	Timestamp   int32  // 服务器消息时间戳(10位，到秒)
	FromUID     string // 发送者UID
	ChannelID   string // 频道ID
	ChannelType uint8  // 频道类型
	Payload     []byte // 消息内容
}

// GetPacketType 获得包类型
func (r *RecvPacket) GetPacketType() PacketType {
	return RECV
}

// VerityString 验证字符串
func (r *RecvPacket) VerityString() string {
	return fmt.Sprintf("%d%d%s%d%s%s%d%s", r.MessageID, r.MessageSeq, r.ClientMsgNo, r.Timestamp, r.FromUID, r.ChannelID, r.ChannelType, string(r.Payload))
}
func (r *RecvPacket) String() string {
	return fmt.Sprintf("Recv Header:%s Setting:%v MessageID:%d MessageSeq:%d Timestamp:%d FromUid:%s ChannelID:%s ChannelType:%d Payload:%s", r.Framer, r.Setting, r.MessageID, r.MessageSeq, r.Timestamp, r.FromUID, r.ChannelID, r.ChannelType, string(r.Payload))
}

func decodeRecv(frame Frame, data []byte, version uint8) (Frame, error) {
	dec := NewDecoder(data)
	recvPacket := &RecvPacket{}
	recvPacket.Framer = frame.(Framer)

	var err error
	if version > 3 {
		setting, err := dec.Uint8()
		if err != nil {
			return nil, errors.Wrap(err, "解码消息设置失败！")
		}
		recvPacket.Receipt = (setting >> 7 & 0x01) > 0
	}
	if version > 2 {
		// MsgKey
		if recvPacket.MsgKey, err = dec.String(); err != nil {
			return nil, errors.Wrap(err, "解码MsgKey失败！")
		}
	}

	// 消息全局唯一ID
	if recvPacket.MessageID, err = dec.Int64(); err != nil {
		return nil, errors.Wrap(err, "解码MessageId失败！")
	}
	// 消息序列号 （用户唯一，有序递增）
	if recvPacket.MessageSeq, err = dec.Uint32(); err != nil {
		return nil, errors.Wrap(err, "解码MessageSeq失败！")
	}
	if version > 1 {
		// 客户端唯一标示
		if recvPacket.ClientMsgNo, err = dec.String(); err != nil {
			return nil, errors.Wrap(err, "解码ClientMsgNo失败！")
		}
	}
	// 消息时间
	if recvPacket.Timestamp, err = dec.Int32(); err != nil {
		return nil, errors.Wrap(err, "解码Timestamp失败！")
	}
	// 频道ID
	if recvPacket.ChannelID, err = dec.String(); err != nil {
		return nil, errors.Wrap(err, "解码ChannelId失败！")
	}
	// 频道类型
	if recvPacket.ChannelType, err = dec.Uint8(); err != nil {
		return nil, errors.Wrap(err, "解码ChannelType失败！")
	}
	// 发送者
	if recvPacket.FromUID, err = dec.String(); err != nil {
		return nil, errors.Wrap(err, "解码FromUID失败！")
	}
	// payloadStartLen := 8 + 4 + 4 + uint32(len(recvPacket.ChannelID)+2) + 1 + uint32(len(recvPacket.FromUID)+2) // 消息ID长度 + 消息序列号长度 + 消息时间长度 +频道ID长度+字符串标示长度 + 频道类型长度 + 发送者uid长度
	// if version > 1 {
	// 	payloadStartLen += uint32(len(recvPacket.ClientMsgNo) + 2)
	// }
	// if version > 2 {
	// 	payloadStartLen += uint32(len(recvPacket.MsgKey) + 2)
	// }
	// if version > 3 {
	// 	payloadStartLen += 1 // 设置的长度
	// }
	// if uint32(len(data)) < payloadStartLen {
	// 	return nil, errors.New("解码RECV消息时失败！payload开始长度位置大于整个剩余数据长度！")
	// }
	// recvPacket.Payload = data[payloadStartLen:]
	if recvPacket.Payload, err = dec.BinaryAll(); err != nil {
		return nil, errors.Wrap(err, "解码payload失败！")
	}
	return recvPacket, err
}

func encodeRecv(frame Frame, version uint8) ([]byte, error) {

	recvPacket := frame.(*RecvPacket)
	enc := NewEncoder()
	if version > 3 {
		setting := encodeBool(recvPacket.Receipt) << 7
		enc.WriteByte(byte(setting))
	}
	if version > 2 {
		// MsgKey
		enc.WriteString(recvPacket.MsgKey)
	}
	// 消息唯一ID
	enc.WriteInt64(recvPacket.MessageID)
	// 消息有序ID
	enc.WriteUint32(recvPacket.MessageSeq)
	if version > 1 {
		// 客户端唯一标示
		enc.WriteString(recvPacket.ClientMsgNo)
	}
	// 消息时间戳
	enc.WriteInt32(recvPacket.Timestamp)
	// 频道ID
	enc.WriteString(recvPacket.ChannelID)
	// 频道类型
	enc.WriteUint8(recvPacket.ChannelType)
	// 发送者
	enc.WriteString(recvPacket.FromUID)
	// 消息内容
	enc.WriteBytes(recvPacket.Payload)
	return enc.Bytes(), nil
}
