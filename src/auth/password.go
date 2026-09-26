package auth

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

const (
	argonMemory  = 64 * 1024
	argonTime    = 1
	argonThreads = 4
	argonKeyLen  = 32
	argonSaltLen = 16
)

var ErrHashTidakValid = errors.New("hash sandi tidak valid")

func HashSandi(sandi string) (string, error) {
	salt := make([]byte, argonSaltLen)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}
	sum := argon2.IDKey([]byte(sandi), salt, argonTime, argonMemory, argonThreads, argonKeyLen)
	return fmt.Sprintf("$argon2id$v=19$m=%d,t=%d,p=%d$%s$%s",
		argonMemory, argonTime, argonThreads,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(sum),
	), nil
}

func CocokkanSandi(encoded, sandi string) (bool, error) {
	salt, sum, err := pecahHash(encoded)
	if err != nil {
		return false, err
	}
	hitung := argon2.IDKey([]byte(sandi), salt, argonTime, argonMemory, argonThreads, argonKeyLen)
	if len(hitung) != len(sum) {
		return false, nil
	}
	return subtle.ConstantTimeCompare(hitung, sum) == 1, nil
}

func pecahHash(encoded string) (salt, sum []byte, err error) {
	bagian := strings.Split(encoded, "$")
	if len(bagian) != 6 || bagian[1] != "argon2id" {
		return nil, nil, ErrHashTidakValid
	}
	salt, err = base64.RawStdEncoding.DecodeString(bagian[4])
	if err != nil {
		return nil, nil, ErrHashTidakValid
	}
	sum, err = base64.RawStdEncoding.DecodeString(bagian[5])
	if err != nil {
		return nil, nil, ErrHashTidakValid
	}
	return salt, sum, nil
}
