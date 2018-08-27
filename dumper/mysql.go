package dumper

import (
	"bytes"
	"encoding/binary"
	"strings"

	"github.com/k1LoW/tcprxy/logger"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	comQuery       = 0x03
	comStmtPrepare = 0x16
	comStmtExecute = 0x17
)

const (
	clientLongPassword uint32 = 1 << iota
	clientFoundRows
	clientLongFlag
	clientConnectWithDB
	clientNoSchema
	clientCompress
	clientODBC
	clientLocalFiles
	clientIgnoreSpace
	clientProtocol41
	clientInteractive
	clientSSL
	clientIgnoreSIGPIPE
	clientTransactions
	clientReserved
	clientSecureConnection
	clientMultiStatements
	clientMultiResults
	clientPSMultiResults
	clientPluginAuth
	clientConnectAttrs
	clientPluginAuthLenEncClientData
	clientCanHandleExpiredPasswords
	clientSessionTrack
	clientDeprecateEOF
)

// MysqlDumper struct
type MysqlDumper struct {
	logger *zap.Logger
}

// NewMysqlDumper returns a MysqlDumper
func NewMysqlDumper() *MysqlDumper {
	dumper := &MysqlDumper{
		logger: logger.NewQueryLogger(),
	}
	return dumper
}

// Dump query of MySQL
func (m *MysqlDumper) Dump(in []byte, direction Direction, persistent *DumpValues, additional []DumpValue) error {
	if direction == RemoteToClient {
		return nil
	}
	if len(in) < 6 {
		return nil
	}

	if len(in) >= 37 {
		clientCapabilities := binary.LittleEndian.Uint32(in[4:8])

		// Protocol::HandshakeResponse41
		if clientCapabilities&clientProtocol41 > 0 && bytes.Compare(in[13:36], []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}) == 0 {
			buff := bytes.NewBuffer(in[36:])
			readed, _ := buff.ReadString(0x00)
			username := strings.Trim(readed, "\x00")
			persistent.Values = append(persistent.Values, DumpValue{
				Key:   "username",
				Value: username,
			})
			fields := []zapcore.Field{}
			for _, kv := range persistent.Values {
				fields = append(fields, zap.Any(kv.Key, kv.Value))
			}
			for _, kv := range additional {
				fields = append(fields, zap.Any(kv.Key, kv.Value))
			}
			return nil
		}
	}

	seqNum := int64(in[3])
	commandID := in[4]
	if commandID != comQuery && commandID != comStmtPrepare && commandID != comStmtExecute {
		return nil
	}
	query := strings.Trim(string(in[5:]), "\x00")
	fields := []zapcore.Field{
		zap.Int64("seq_num", seqNum),
		zap.String("command_id", string(commandID)),
	}
	for _, kv := range persistent.Values {
		fields = append(fields, zap.Any(kv.Key, kv.Value))
	}
	for _, kv := range additional {
		fields = append(fields, zap.Any(kv.Key, kv.Value))
	}

	m.logger.Info(query, fields...)
	return nil
}
