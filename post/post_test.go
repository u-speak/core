package post

import (
	"bytes"
	"crypto"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"golang.org/x/crypto/openpgp"
	"golang.org/x/crypto/openpgp/packet"
)

func post(t *testing.T) *Post {
	content := "foo"
	c := &packet.Config{
		DefaultHash: crypto.SHA256,
	}
	e, _ := openpgp.NewEntity("Test", "test", "test@example.com", c)
	_ = e.SerializePrivate(bytes.NewBuffer(nil), nil)
	buff := bytes.NewBuffer(nil)
	openpgp.DetachSignText(buff, e, strings.NewReader(content), c)
	sigp, err := packet.Read(buff)
	assert.NoError(t, err)
	sig := sigp.(*packet.Signature)
	p := &Post{Content: content, Pubkey: e, Signature: sig, Timestamp: time.Now().Unix()}
	return p
}

func TestVerify(t *testing.T) {
	p := post(t)
	assert.NoError(t, p.Verify())
	p.Content = "modified"
	assert.Error(t, p.Verify())
}

func TestSerializeable(t *testing.T) {
	p := post(t)
	assert.NoError(t, p.Verify())
	buff, err := p.Serialize()
	assert.NoError(t, err)
	p2 := &Post{}
	err = p2.Deserialize(buff)
	assert.NoError(t, err)
	assert.NoError(t, p2.Verify())
	assert.EqualValues(t, p.Timestamp, p2.Timestamp)
}

func TestImportExport(t *testing.T) {
	pubkey := `-----BEGIN PGP PUBLIC KEY BLOCK-----

mQENBFguKVUBCAD3UoSopfUdqg0L1B9AyG8IizpCEIANnKcxECzudGR7lXAgcsZB
tm8qA0kly2Lx0Y1KxmLLpuVEMOXCt7z60MKAYWNgOWhAsM/+AkDcB+o7m/wl+lc6
afRP8r8ve3h3B6uEAJ2+3Nj7AfDS4CAdS/W6fZqwkqxsjvhGv0Fu/KTy9qNsoB+Q
8574MdrEoO03UwciRZpWEsjR4Y+vevWxSx+ra4DfMBS3bn+B0sJshgybT3/RP27u
1DerxMdxSFTDRZqxQTCCT+h22cuVrMdrY1yoxInwRm3DiC51LNYg7x+THBbNCBB7
gR0t4cxNmm51+9IWaTcQ5+CqSkxthrvyHc5jABEBAAG0KERvbWluaWsgU8O8w58g
PGRvbWluaWsuc3Vlc3NAb3V0bG9vay5hdD6JATcEEwEIACEFAlguKVUCGwMFCwkI
BwIGFQgJCgsCBBYCAwECHgECF4AACgkQFCC6V2n2TSQ7dAgAs8f4XxOYi7jAUfF0
R8Bt03CV4mJ7BYm6JjNOlCo2sfOCQigHw5MWQbQhjUDX7J45FzpiACwuWeFPbwYL
OCOOhtDJ1yNdRRnFJtjbNndIiYG8VLfLC6OfV9wJuNGOR/sM2b7QFnGKxU5ZBe1E
WQ6DjWmBqz0lzcrdN4DUWIqa7Ydpr8/rgM9RaN848eNDCVilY2FF2/IRdAMaeZp/
TCX05varHOyGQIvPWKdQcFXPemElGxk7manWocPcPEzS6eOPDP0PZZhLP6CNCOL0
apx/QryGyPq4KAGXvavWY9fdgmx90H9jkP00nPfPHXHe5q4s3T2sQafeCZHmwWYw
yUskS7kBDQRYLilVAQgAzF1Ckgjb4v2VgyRDLQ4eydR64CXoZfaTzxW6cRBcZ9Q7
iRGVWaiOrhf5lagZhAnXt9Y3+BC/Z5v6LwxDbOmbkXcmtki1laP55Mb/wZeSCU8a
FkLff5iu3UMZXRz95DceDw93+c29nCxDmshKfDy/CjxNg78V26H/h0X/3qveVxt1
PfAFTnYjPaaXM5OJx/EiTArln8qyllQzBWdMJqfiYSSsyl8Bv2mTERWXvPchFa24
9fNTpOjJ0BNliGBSr9QWMZ4XM1xciPHAH4LzhelS7uLeX2gvDfyPSfsTfoM6R6LQ
hkTK9MJ7hQJ6YHktm0u/qk/62mev9Q7RrCabk2jiXQARAQABiQEfBBgBCAAJBQJY
LilVAhsMAAoJEBQguldp9k0kfPsIAJEiUH9etzuX6a8DLaHnAOYwZqz4tu7lDZxz
MBTvn72ZCUXM5PQMTljwemB1JrF0nnzxJjJ+yCzWYy5lKNHliycCk5r4h3MVDdrq
6cYKW0MNMH6aEni79MswKPUxdTTnWYrDSMmln6CzGuZk7L8UFPXo/MRdgLjZ8sKe
L+p1vVQLdAVhOd1hso1UyRbIb5lyGZ4M2fvE9w3LZJv816PRfufqCuBYG1QKl6qk
PBuRXD5EQJdgHy/2mIqXR4NaEGsLbdrv7qaaVJPmZ/Qh/4AdkoV6oSyia4HI03M2
bATkpmKdxC/qCgICRaXugOdehrLpoEEMSrHlc3xG3bmSvNa0U7Y=
=TF9i
-----END PGP PUBLIC KEY BLOCK-----
`
	p, err := asciiDecodeEntity(pubkey)
	assert.NoError(t, err)
	enc, err := asciiEncode(p, openpgp.PublicKeyType)
	assert.NoError(t, err)
	t.Log(enc)
}
