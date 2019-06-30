package encoder_test

import (
	"testing"

	"github.com/julian7/redact/encoder"
)

const (
	sampleAES       = "0123456789abcdefghijklmnopqrstuv"
	sampleHMAC      = "0123456789abcdefghijklmnopqrstuv0123456789abcdefghijklmnopqrstuv"
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

	sampleCiphertext = "\x9f\x13\xb8\x192\xaa\xef\xdbmX%\xbb\xe9\x01ߦ\x12\x16\xc7\x19\x0e\xbb\xe6?\x83\xe4F\xfa\x81\xab0I\x8f\xd8@\xa7\xb4\x10\x16\x19\xbenb|\xad\xb9$_\xddt\xd1跭7\x12\xad^\x83Uq\xd8\x1e\xd0к\fŔ=\xfdj\xff\x8e\r\x0e=\xa5N\x90\xa3\bz\xc8{\u007fPc\xef\xc4\xee\b\xe3ۄj֞n7\xa3\x89X$\x13v5\x1f\vX4\x12ә\x14u֥\xd2\b\xdb\x1e\x9b\x87\x98d\x97r\x88\x98\xd8*\xe3)N\x01{\xa7\xe8Msǹ\x93\xe2?\xd6bΊ0\xd6qz\x01*l\vI\b\x85\x0564S\r\xc1y~\x9d`4\xf5\x16\xe2:\xaaI\t\x88\xa3u\xeb\x13\x90\x8fF\x17\xe7wT\xec\xc0J[E\xa2\xd17\xe1\xaa\xf3\xadwQ\xfa\xcb}\x1ei\xea6\xe6h\\\x1c\xa9\x12\xe7\x84\xc2\xe1\x9e}8l\xf8\xae\x8ff\xc1\xb9\x9fP}/\xca\xf3\xc0\x04\x86\xe2\x9dI\xf1w\xb7\x18RѤ\x9f\x8c\xc2Լ\xfe4\xdaAZpU\x1bt\xfa\xc2y\x9b<F\xa6\xfc\x93\xf1\xca\xf2\x10\x9a\xea\xeaI\xefO\x8c\x14>\xf1\x9e\x19\xb3\x1bn\x15o\x97\xafv:g5o\x9c=\x97\xa1\x968_\x0e\xf7$w7\xef\xf6j\xc4$d(\xd8\xd5(\x93\x96\x14\xbf\xb2\xaap\tC\x90M\xee\xe9z\xd9|^Gƫ+\xd2\xe2:\x9f%%\xdb\xf1s\x85p\xbcB1л\xdc\x1a\x00#`D\x00ĺ{\x16y)<\x81 \xed\xdaфc\xbd\xe0D\xab\x0fR(\xba\x88\x9e\xb9E\x16\xc8$h~\xc4\x0e}~\xd7\xeffN~\x06\"\xe5\xf52\xbd\x91\xb5ŧ\x98\xe6J\xe9\\%;ѽ\xaa\x1d\x05\xd1O\x87\xcf\xfe\xac\xfa\xb2\x1a\xcb\xcaٌI~E\xf5\xc9Q\x924]\xb1\xb0\\\xa9\x1c+\xff\xd8ޣ\x04\xa1\xa0V\x1b\xd0k8\xc4>\xfa\t3\xee\v\xb7\xe7^A\xd7\xe3\x8d\xdf\xfc\x05\x82d\fX>\xccF\xf3\x96\x1f _\x8c\x91\x9c\xf8n\n-H\xc5\xe9\x1d\x86\x9f2MN\t\xd1\xf1\xe4\xe6\xfb!N\xb3\xe5\x02\xc7\xdbՑ(<,\x19]\xd7n\x98뼯\xe1\xc7\xf7L\x06\x83\x01\x04m7\xb6ﲦ\xabۍ\xd5|\xa9\xd6*\xba\xab\xce)\xbc>Pb0f0\x16}N7Ԏ\xfc\x91dE\xfe\xe7˜\x9b\u007f\x91\xb3\xeb\xfe\x1d\xa0\xcfD#\x82\b\xba\xeaV\xb4\x96\xac\x9do\x9a@ܴ\u007f\xb0\xc3٣\x85\x1di2\x95\xde#\x95b\x0e\xfc\xbb\x1b.\xe02IVL\xdb\xc9\xeaf`\x99<\xcd)\xfc\x89%dg\xb3\xf0߫\xf3\x91\\\x0e\u0099\xda\xd6\xfcv\x85\xd2\xfa&\x94\xe1\x18\xbe?_\n\x82>D\"\x96<\x8dg\x1d\xdc\xf0\x9b\a\x81*\xf8\xa1\x9d?\xe8x`zת\x15\x84ƍ\x9c<\xd5\x11\xcd\xe53\xed\xc0j\x929\u007f\x19\xdb\xc6,\xfcUd\x9d1zϞ\xa0\xa3\xc9\xe4\x16Q\x8d\x18\xe2~\xb5\x1a\xf8XF\xb0G\x04?2\xf9\x8d\xe4\rr\x15\x8d9̪\a k}\xfa?i\a\xad\xaf\xfbbƯ\\\xd7\xfd\xd1\x1c9|\x02\x19#\xae\fHp\xf3\x16\x9b%\xc8\x1c\x83\xb5\xf1\xcdU\xfbxw\xf4\xb5\x83$\x88 1e\x81#B宣\x93vt\xe4\\\x1aV\x18\xa2\x1cHw\xbc\xcc\x02=!\x93֫\x11{\xff\xe3iZKe\xd0\xecJB"
)

func TestEncode(t *testing.T) {
	enc, err := encoder.NewAES256GCM96([]byte(sampleAES), []byte(sampleHMAC))
	if err != nil {
		t.Errorf("cannot create encoder: %v", err)
		return
	}
	ret, err := enc.Encode([]byte(samplePlaintext))
	if err != nil {
		t.Errorf("cannot encode: %v", err)
		return
	}
	if string(ret) != sampleCiphertext {
		t.Errorf("Encrypted message not matching: %q", ret)
	}
}

func TestDecode(t *testing.T) {
	enc, err := encoder.NewAES256GCM96([]byte(sampleAES), []byte(sampleHMAC))
	if err != nil {
		t.Errorf("cannot create encoder: %v", err)
		return
	}
	ret, err := enc.Decode([]byte(sampleCiphertext))
	if err != nil {
		t.Errorf("cannot decode: %v", err)
		return
	}
	if string(ret) != samplePlaintext {
		t.Errorf("Decrypted message not matching: %q", ret)
	}
}