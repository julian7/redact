package files_test

import (
	"bytes"
	"encoding/binary"
	"io"
	"testing"
	"testing/iotest"

	"github.com/julian7/redact/encoder"
	"github.com/pkg/errors"
)

var (
	samplePlaintext = "Lorem ipsum dolor sit amet, consectetur adipiscing " +
		"elit. Fusce odio lacus, feugiat a elit ut, hendrerit venenatis en" +
		"im. Duis vehicula, purus nec cursus iaculis, purus magna elementu" +
		"m ipsum, in suscipit risus nulla non magna. Maecenas pharetra vul" +
		"putate condimentum. Vivamus in eros scelerisque, maximus ex nec, " +
		"tincidunt ex. Mauris aliquam lectus libero, sit amet tincidunt du" +
		"i suscipit vel. Integer interdum imperdiet felis. Cras egestas in" +
		"terdum iaculis. Phasellus id dui aliquam sem lacinia tristique id" +
		" sit amet lectus. Quisque scelerisque magna vel finibus viverra. " +
		"Praesent vestibulum nisi et nulla fermentum tincidunt. Quisque ul" +
		"tricies enim vel leo rhoncus tristique. Phasellus at justo id est" +
		" laoreet volutpat. Nullam eleifend, mi sed congue ultrices, neque" +
		" purus euismod velit, id mattis odio nunc quis lectus."

	sampleCiphertext = "\x00REDACTED\x00\x00\x00\x00\x00\x00\x00\x00\x01" +
		"\xd9\x00\xaeo\x8ezk\n\xa6^\x97\t4\xf0+\x12\x04J\r\xf3\xa8\xafG\x16{3z\xfaj\xbe\x97\x9f\xff\xa9\xa2\b\xac\xad\xe52\"R\xf6\x1d\xb8\aC\xc0\xd2" +
		"\xcb=P\xf7X\x869D&2A\xad\x81\xe9\xa4\xc4\xe5xl.\xf0'ϗ\u007f \xdc<\xba-v\x13V\xb2~\x1c\xdfX\xb6\x833g\xe6\x9dMh\x19o\x02z\x01\x04\xcc\xecIt\x9eDh\xa4R\xd5" +
		"\xe5\xed\x9f1\xfe\x0fy\xd5\x18\x98}`\xecq\xadqK^\xa4\xc6[\x04E\x01E\x97D\xa24\xf0\x94u\xba\xdf\xc75ad7\xc4\xfc9\x81\xcf\xd2J\xba땻U<a5\x97\x06\xd4\xed\x87" +
		"Z\xa0\xc0\xf7\xcf\b\x1awYx8J\t\x88\"\xbc\xf64\xf9\x03lU\xabX!܂\x96\xfbG\xa6FO\xab\xf8\x04\xc4Q\x18\x03e;\x13\xd1m\xd3gt\rem¸\xa2,\xbe\xc3\fB<E\xea\xe8\x01" +
		"܈xA\xe8\x95;\xf4\xf1To>\xad\x86\xb0\xcc\x01\xc3%\xd0\x13`\xaa\x17\x19\x17\x10\xb6O&\xe1\xc9\xc9z,P̓\x1c\x17kZ\xe8\xf9\xb7\u007f\x12\x89s\xefz:as\xa6\bs\xf1" +
		"\x01.\xb5z\x01d\x92\xe5]\x06\xc4\xf8\x12E\xd8NbT\xfer\xb3\n-?O\xe0\xe6\xa7q\x80$6\xf9\x80\x9b\x91\xbfPW\x86\xed\"\xae\xbbϰ\xa7\xa6\xaa(\xed{\xed\rw\x93\xdd" +
		"\f\xa0\xbb\xd4TM\xe6\xa0G\xd0\xe8k\x94,\xbb£N\x15,\xf5U{\x0e\xaf\xea\xab\x0e\x15\b(m\x8d\x06{\xec\x97B\xba\xbd\x82\xd4Տ$\xa2\x8f)\xc0\x05y\xa6\tv\xd4*x\x15" +
		"\x8b\xban\x18\xf6.0*#\xae0\xb1a\x167\xa2*\xed\t\x9e\x95\xa4\xa5\xd0H\xfe\xff\xddFl\x1a\xff\xc3\xfdC\x1c55\x10\x9d\xa1W\xca <&)D\xb1\xb4Z\xf8W\xee\b\xf2\xdc4" +
		"\t\xcb\x01\xcf˟_\xd8Ŏ\x16\x04\xb9\x0e\x9fɫ\x8fM'\xff\x16\x1er\xf0x\xc2I\x852\x00\x16\x95\x97$\x90\x84*$#\xd1\x18\x8fu\xee\xecS?\x00\x93\xf3\xf5(\x87\xf5_" +
		"\xceg\x88\n)\xdfVt\xe0uܤGX\x04j\xb8E\xf3\xe64\xd3;%\xd4\xd0>\xd0!\x1a\x94\xe8|\x11\x88\x1ct\xb4\xe6\r\x8b]J\xb9\xc8\x12\x14\x04K\x9a\tu\x16\x8f\xda\x15\xbc" +
		"\xd6]S\xef`h\x9d\x89N\xb7\x8cR\xa1IF\x9ei@#\xe0\x99\x16n\xb8\x8b\xed_\xa0\x92\x80\x127\xb9'\xec\xa0\xe7o\xa7\xd85\xb0\xe3(\xf9\x980\xce\xe0F\x82\x99}\xc5" +
		"\xe4/:P6\xe3F\x97\xb9@\xe6\x9b\xe3n\x15\x80,\xe1\xe1\xec>\xe2%\xb63Yt\xfd\xcf\x1f\xf4\xa4\vs`-]\a\xb1ܴV\xb7\rQ\xb4\xcdX\x81\xcfpA\x81W\xf2ӟaQIW6\xcb\xfc" +
		"\xd2\xe3\xb1\xe3\xd89\xf7\xef\xd9\xe0YNL\x12\x13)\x03\xf7\xe5lX\xec\x0e\xb0j~\xd3Ҽ\xcdt k\xac\xbf\x98\a&h\xbb\xaa\xf9\xea+\x14\xcbHT/M\xdaU^\xda,\x99Z\xbf'p" +
		"\x03\xe2\x9e\xef}\xb9\b\x03\x92\x8e\x01>S\xb0\xb4Hh\xc8S\x1f/\x10\x90\xb1͇[Fu\xbb9%n\x97\xf7\\q\x0f\x1e)\x14\xbc\xcb\xd7+Y\xc9C7\\H&\xec\x8eq\x8b\xbbP\x8d\xf9" +
		"\xfc\xc1\xa9D\\\xc7\xd4Bm\x01\b\x81t\x10v\x16\xef\xda\x14\xeb\n\x9b\x8f\xaaS\x17;\xdeM\xc8y\xeb"
)

func genCiphertextHeader(encType, epoch int) []byte {
	out := bytes.NewBuffer(nil)
	_, _ = out.WriteString("\x00REDACTED\x00")
	_ = binary.Write(out, binary.BigEndian, uint32(encType))
	_ = binary.Write(out, binary.BigEndian, uint32(epoch))
	return out.Bytes()
}

func ensureFailingEncoder() int {
	failEncoderID := 50000
	_ = encoder.RegisterEncoder(failEncoderID, newFailingEncoder)
	return failEncoderID
}

type failingReader struct{}

func (r *failingReader) Read(p []byte) (n int, err error) {
	return 0, io.ErrUnexpectedEOF
}

type failingWriter struct{}

func (r *failingWriter) Write(p []byte) (n int, err error) {
	return 0, io.ErrUnexpectedEOF
}

func TimeoutWriter(r io.Writer) io.Writer { return &timeoutWriter{r, 0} }

type timeoutWriter struct {
	r     io.Writer
	count int
}

func (r *timeoutWriter) Write(p []byte) (int, error) {
	r.count++
	if r.count == 2 {
		return 0, iotest.ErrTimeout
	}
	return r.r.Write(p)
}

type failingEncoder struct{}

func (e *failingEncoder) Encode([]byte) ([]byte, error) {
	return nil, errors.New("failing encoder error: cannot encode")
}
func (e *failingEncoder) Decode([]byte) ([]byte, error) {
	return nil, errors.New("failing encoder error: cannot decode")
}
func newFailingEncoder(aes, hmac []byte) (encoder.Encoder, error) {
	return &failingEncoder{}, nil
}

func TestEncode(t *testing.T) {
	failEncoderID := ensureFailingEncoder()
	k, err := genGitRepo()
	if err != nil {
		t.Error(err)
		return
	}
	if err := writeKey(k); err != nil {
		t.Error(err)
		return
	}
	if err := k.Load(); err != nil {
		t.Error(err)
		return
	}
	tt := []struct {
		name    string
		enctype uint32
		epoch   uint32
		reader  io.Reader
		writer  io.Writer
		err     string
		output  string
	}{
		{
			name:    "no encoder",
			enctype: 65535,
			epoch:   1,
			reader:  bytes.NewReader([]byte(samplePlaintext)),
			writer:  bytes.NewBuffer(nil),
			err:     "setting up encoder: invalid encoding type 65535",
			output:  "",
		},
		{
			name:    "no key",
			enctype: 0,
			epoch:   65535,
			reader:  bytes.NewReader([]byte(samplePlaintext)),
			writer:  bytes.NewBuffer(nil),
			err:     "encoding stream: key version 65535 not found",
			output:  "",
		},
		{
			name:    "error reading",
			enctype: 0,
			epoch:   1,
			reader:  &failingReader{},
			writer:  bytes.NewBuffer(nil),
			err:     "reading input stream: unexpected EOF",
			output:  "",
		},
		{
			name:    "encoder error",
			enctype: uint32(failEncoderID),
			epoch:   1,
			reader:  bytes.NewReader([]byte(samplePlaintext)),
			writer:  bytes.NewBuffer(nil),
			err:     "encoding stream: failing encoder error: cannot encode",
			output:  "",
		},
		{
			name:    "error writing header",
			enctype: 0,
			epoch:   1,
			reader:  bytes.NewReader([]byte(samplePlaintext)),
			writer:  &failingWriter{},
			err:     "writing file header: unexpected EOF",
			output:  "",
		},
		{
			name:    "error body",
			enctype: 0,
			epoch:   1,
			reader:  bytes.NewReader([]byte(samplePlaintext)),
			writer:  TimeoutWriter(bytes.NewBuffer(nil)),
			err:     "writing encoded stream: timeout",
			output:  "",
		},
		{
			name:    "successful",
			enctype: 0,
			epoch:   1,
			reader:  bytes.NewReader([]byte(samplePlaintext)),
			writer:  bytes.NewBuffer(nil),
			err:     "",
			output:  sampleCiphertext,
		},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			err := k.Encode(tc.enctype, tc.epoch, tc.reader, tc.writer)
			if testerr := checkError(tc.err, err); testerr != nil {
				t.Error(testerr)
			}
			if err != nil {
				return
			}
			outbuf, ok := tc.writer.(*bytes.Buffer)
			if !ok {
				t.Error("couldn't capture encoded output")
				return
			}
			if err := checkString(tc.output, outbuf.String()); err != nil {
				t.Error(err)
			}
		})
	}
}

func TestDecode(t *testing.T) {
	failEncoderID := ensureFailingEncoder()
	k, err := genGitRepo()
	if err != nil {
		t.Error(err)
		return
	}
	if err := writeKey(k); err != nil {
		t.Error(err)
		return
	}
	if err := k.Load(); err != nil {
		t.Error(err)
		return
	}

	tt := []struct {
		name   string
		reader io.Reader
		writer io.Writer
		err    string
		output string
	}{
		{
			name:   "unencrypted",
			reader: bytes.NewReader([]byte(samplePlaintext)),
			writer: bytes.NewBuffer(nil),
			err:    "invalid file preamble",
			output: "",
		},
		{
			name:   "no key",
			reader: bytes.NewReader(genCiphertextHeader(0, 65535)),
			writer: bytes.NewBuffer(nil),
			err:    "retrieving key: key version 65535 not found",
			output: "",
		},
		{
			name:   "no encoder",
			reader: bytes.NewReader(genCiphertextHeader(65535, 1)),
			writer: bytes.NewBuffer(nil),
			err:    "setting up encoder: invalid encoding type 65535",
			output: "",
		},
		{
			name:   "error reading preamble",
			reader: &failingReader{},
			writer: bytes.NewBuffer(nil),
			err:    "reading file header: unexpected EOF",
			output: "",
		},
		{
			name:   "error reading body",
			reader: iotest.TimeoutReader(bytes.NewReader([]byte(sampleCiphertext))),
			writer: bytes.NewBuffer(nil),
			err:    "reading stream: timeout",
			output: "",
		},
		{
			name:   "decoder error",
			reader: bytes.NewReader(genCiphertextHeader(failEncoderID, 1)),
			writer: bytes.NewBuffer(nil),
			err:    "decoding stream: failing encoder error: cannot decode",
			output: "",
		},
		{
			name:   "error writing header",
			reader: bytes.NewReader([]byte(sampleCiphertext)),
			writer: &failingWriter{},
			err:    "writing decoded stream: unexpected EOF",
			output: "",
		},
		{
			name:   "successful",
			reader: bytes.NewReader([]byte(sampleCiphertext)),
			writer: bytes.NewBuffer(nil),
			err:    "",
			output: samplePlaintext,
		},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			err := k.Decode(tc.reader, tc.writer)
			if testerr := checkError(tc.err, err); testerr != nil {
				t.Error(testerr)
			}
			if err != nil {
				return
			}
			outbuf, ok := tc.writer.(*bytes.Buffer)
			if !ok {
				t.Error("couldn't capture encoded output")
				return
			}
			if err := checkString(tc.output, outbuf.String()); err != nil {
				t.Error(err)
			}
		})
	}

	reader := bytes.NewReader([]byte(sampleCiphertext))
	writer := bytes.NewBuffer(nil)

	if err := k.Decode(reader, writer); err != nil {
		t.Error(err)
		return
	}
	if err := checkString(samplePlaintext, writer.String()); err != nil {
		t.Error(err)
		return
	}
}

func TestFileStatus(t *testing.T) {
	testFN := "testfile.txt"
	tt := []struct {
		name      string
		contents  string
		encrypted bool
		epoch     uint32
	}{
		{"plaintext", samplePlaintext, false, 0},
		{"encrypted", sampleCiphertext, true, 1},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			k, err := genGitRepo()
			if err != nil {
				t.Error(err)
				return
			}
			if err := writeKey(k); err != nil {
				t.Error(err)
				return
			}
			if err := k.Load(); err != nil {
				t.Error(err)
				return
			}
			if err := writeFile(k, testFN, 0644, tc.contents); err != nil {
				t.Error(err)
				return
			}
			reader, err := k.Open(testFN)
			if err != nil {
				t.Error(err)
				return
			}
			ok, epoch := k.FileStatus(reader)
			reader.Close()
			if ok != tc.encrypted {
				t.Errorf("expected %s file", tc.name)
			}
			if epoch != tc.epoch {
				t.Errorf("expected epoch == %d; received: %d", tc.epoch, epoch)
			}
		})
	}
}
