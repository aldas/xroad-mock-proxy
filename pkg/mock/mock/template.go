package mock

import (
	"crypto/md5"
	"fmt"
	"io"
	"time"
)

var names = []string{
	"Alfa",
	"Bravo",
	"Charlie",
	"Delta",
	"Echo",
	"Foxtrot",
	"Golf",
	"Hotel",
	"India",
	"Juliett",
	"Kilo",
	"Lima",
	"Mike",
	"November",
	"Oscar",
	"Papa",
	"Quebec",
	"Romeo",
	"Sierra",
	"Tango",
	"Uniform",
	"Victor",
	"Whiskey",
	"X-ray",
	"Yankee",
	"Zulu",
}

type templateVars struct {
	md5 []byte

	MD5          string
	Identity     string
	Gender       string
	Now          time.Time
	StartOfToday time.Time
	Timestamp    int64
}

func fromIdentity(identity string) templateVars {
	md5Sum := identityMD5(identity)

	now := time.Now()
	return templateVars{
		md5: md5Sum,
		MD5: fmt.Sprintf("%x", md5Sum),

		Identity:     identity,
		Now:          now,
		StartOfToday: now.Truncate(24 * time.Hour),
		Timestamp:    now.Unix(),
	}
}

func (v templateVars) IDName1() string {
	return v.IDNameNth(1)
}
func (v templateVars) IDName2() string {
	return v.IDNameNth(2)
}

func (v templateVars) IDNameNth(useNthMD5Byte int) string {
	nLen := len(names)
	if useNthMD5Byte >= md5.Size {
		useNthMD5Byte = md5.Size - 1
	}

	md5Number := int(v.md5[useNthMD5Byte])
	if md5Number == 0 {
		return names[0]
	}

	place := 0
	if nLen > md5Number {
		place = nLen % md5Number
	} else {
		place = md5Number % nLen
	}

	return names[place]
}

func identityMD5(identity string) []byte {
	h := md5.New()
	io.WriteString(h, identity)
	return h.Sum(nil)
}

func (v templateVars) IDNvl2(IDnthChar int, valueIfOdd string, valueIfEven string) string {
	nth := IDnthChar
	size := len(v.Identity)
	if nth >= size {
		nth = size - 1
	} else if nth < 0 {
		nth = 0
	}

	char := v.Identity[nth]

	if char%2 == 0 {
		return valueIfEven
	}
	return valueIfOdd
}
