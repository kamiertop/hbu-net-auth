// Codex+GPT5.5 生成

package srun

import (
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"unicode/utf16"
)

const (
	srunEncVer   = "srun_bx1"
	srunType     = "1"
	srunN        = "200"
	srunB64Alpha = "LVoJPiCN2R8G90yg+hmFHuacZ1OWMnrsSTXkYpUq/3dlbfKwv6xztjI7DeBE45QA"
)

func hmacMD5(data, key string) string {
	mac := hmac.New(md5.New, []byte(key))
	mac.Write([]byte(data))
	return hex.EncodeToString(mac.Sum(nil))
}

func sha1Hex(data string) string {
	sum := sha1.Sum([]byte(data))
	return hex.EncodeToString(sum[:])
}

func encodeUserInfo(username, password, ip, acid, token string) string {
	info := fmt.Sprintf(
		`{"username":"%s","password":"%s","ip":"%s","acid":"%s","enc_ver":"%s"}`,
		jsonEscape(username),
		jsonEscape(password),
		jsonEscape(ip),
		jsonEscape(acid),
		srunEncVer,
	)
	return "{SRBX1}" + customBase64(xencode(info, token), srunB64Alpha)
}

func xencode(value, key string) []byte {
	if value == "" {
		return nil
	}

	v := srunWords(value, true)
	k := srunWords(key, false)
	for len(k) < 4 {
		k = append(k, 0)
	}

	n := len(v) - 1
	z := v[n]
	c := uint32(0x9e3779b9)
	q := 6 + 52/(n+1)
	var d uint32

	for q > 0 {
		q--
		d += c
		e := (d >> 2) & 3
		for p := range n {
			y := v[p+1]
			m := (z >> 5) ^ (y << 2)
			m += (y >> 3) ^ (z << 4) ^ (d ^ y)
			m += k[(p&3)^int(e)] ^ z
			v[p] += m
			z = v[p]
		}

		y := v[0]
		m := (z >> 5) ^ (y << 2)
		m += (y >> 3) ^ (z << 4) ^ (d ^ y)
		m += k[(n&3)^int(e)] ^ z
		v[n] += m
		z = v[n]
	}

	return wordsToBytes(v)
}

func srunWords(value string, appendLen bool) []uint32 {
	units := utf16.Encode([]rune(value))
	words := make([]uint32, 0, (len(units)+3)/4+1)

	for i := 0; i < len(units); i += 4 {
		var word uint32
		for j := 0; j < 4 && i+j < len(units); j++ {
			word |= uint32(units[i+j]) << (j * 8)
		}
		words = append(words, word)
	}
	if appendLen {
		words = append(words, uint32(len(units)))
	}
	return words
}

func wordsToBytes(words []uint32) []byte {
	out := make([]byte, 0, len(words)*4)
	for _, word := range words {
		out = append(out,
			byte(word),
			byte(word>>8),
			byte(word>>16),
			byte(word>>24),
		)
	}
	return out
}

func customBase64(bytes []byte, alpha string) string {
	var out strings.Builder
	i := 0

	for ; i+3 <= len(bytes); i += 3 {
		b10 := uint32(bytes[i])<<16 | uint32(bytes[i+1])<<8 | uint32(bytes[i+2])
		out.WriteByte(alpha[b10>>18])
		out.WriteByte(alpha[(b10>>12)&63])
		out.WriteByte(alpha[(b10>>6)&63])
		out.WriteByte(alpha[b10&63])
	}

	switch len(bytes) - i {
	case 1:
		b10 := uint32(bytes[i]) << 16
		out.WriteByte(alpha[b10>>18])
		out.WriteByte(alpha[(b10>>12)&63])
		out.WriteString("==")
	case 2:
		b10 := uint32(bytes[i])<<16 | uint32(bytes[i+1])<<8
		out.WriteByte(alpha[b10>>18])
		out.WriteByte(alpha[(b10>>12)&63])
		out.WriteByte(alpha[(b10>>6)&63])
		out.WriteByte('=')
	}

	return out.String()
}

func jsonEscape(value string) string {
	bytes, _ := json.Marshal(value)
	return string(bytes[1 : len(bytes)-1])
}
