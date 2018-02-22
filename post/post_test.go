package post

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"golang.org/x/crypto/openpgp"
	"golang.org/x/crypto/openpgp/armor"
	"golang.org/x/crypto/openpgp/packet"
)

func post(t *testing.T) *Post {
	content := "foo"
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	assert.NoError(t, err)
	privkey := packet.NewRSAPrivateKey(time.Now(), key)
	buff := bytes.NewBuffer(nil)
	e := &openpgp.Entity{
		PrivateKey: privkey,
		PrimaryKey: &privkey.PublicKey,
	}
	err = openpgp.ArmoredDetachSignText(buff, e, strings.NewReader(content), nil)
	assert.NoError(t, err)
	block, err := armor.Decode(buff)
	assert.Equal(t, block.Type, openpgp.SignatureType)
	reader := packet.NewReader(block.Body)
	pkt, err := reader.Next()
	assert.NoError(t, err)
	sig, ok := pkt.(*packet.Signature)
	assert.True(t, ok)
	p := &Post{Content: content, Pubkey: &privkey.PublicKey, Signature: sig, Timestamp: time.Now().Unix()}
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
