package libs

import (
	"bufio"
	"bytes"
	"io"

	"github.com/Laisky/go-utils"
	"github.com/ugorji/go/codec"
	"go.uber.org/zap"
)

type TinyFluentRecord struct {
	Timestamp uint64
	Data      map[string]interface{}
}

type FluentEncoder struct {
	wrap, batchWrap               []interface{}
	encoder, internalBatchEncoder *codec.Encoder
	msgBuf                        *bytes.Buffer
	msgWriter                     *bufio.Writer
	tmpMsg                        *FluentMsg
}

func NewFluentEncoder(writer io.Writer) *FluentEncoder {
	enc := &FluentEncoder{
		wrap:      []interface{}{0, []TinyFluentRecord{TinyFluentRecord{}}},
		encoder:   codec.NewEncoder(writer, NewCodec()),
		msgBuf:    &bytes.Buffer{},
		batchWrap: []interface{}{nil, nil, nil},
	}
	enc.msgWriter = bufio.NewWriter(enc.msgBuf)
	enc.internalBatchEncoder = codec.NewEncoder(enc.msgWriter, NewCodec())
	return enc
}

func (e *FluentEncoder) Encode(msg *FluentMsg) error {
	e.wrap[0] = msg.Tag
	e.wrap[1].([]TinyFluentRecord)[0].Data = msg.Message
	return e.encoder.Encode(e.wrap)
}

func (e *FluentEncoder) EncodeBatch(tag string, msgBatch []*FluentMsg) (err error) {
	for _, e.tmpMsg = range msgBatch {
		e.wrap[1] = e.tmpMsg.Message
		if err = e.internalBatchEncoder.Encode(e.wrap); err != nil {
			utils.Logger.Error("try to encode msg got error", zap.String("tag", tag))
		}
	}

	e.batchWrap[0] = tag
	e.msgWriter.Flush()
	e.batchWrap[1] = e.msgBuf.Bytes()
	e.msgBuf.Reset()
	return e.encoder.Encode(e.batchWrap)
}

type Decoder struct {
	wrap    []interface{}
	decoder *codec.Decoder
}

func NewDecoder(reader io.Reader) *Decoder {
	return &Decoder{
		wrap:    []interface{}{nil, nil, nil},
		decoder: codec.NewDecoder(reader, NewCodec()),
	}
}

func (d *Decoder) Decode(msg *FluentMsg) (err error) {
	d.wrap[2] = make(map[string]interface{}) // create new map, avoid influenced by old data
	if err = d.decoder.Decode(&d.wrap); err != nil {
		return err
	}

	msg.Tag = string(d.wrap[0].([]byte))
	msg.Message = d.wrap[2].(map[string]interface{})
	return nil
}
