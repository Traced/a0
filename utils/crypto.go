package utils

import (
	"bytes"
	"crypto/des"
	"crypto/md5"
	"encoding/base64"
	"errors"
	"fmt"
)

func MD5(s string) string {
	return fmt.Sprintf("%x", md5.Sum(StringToBytes(s)))
}

// Base64Encode 编码
func Base64Encode(b []byte) string {
	return base64.StdEncoding.EncodeToString(b)
}

// Base64EncodeByte 编码成字节数组
func Base64EncodeByte(b []byte) []byte {
	return StringToBytes(Base64Encode(b))
}

// Base64EncodeString 编码成字符串
func Base64EncodeString(s string) string {
	return Base64Encode([]byte(s))
}

// Base64Decode 解码
func Base64Decode(s string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(s)
}

// Base64DecodeToString 解码成字符串
func Base64DecodeToString(s string) string {
	if v, err := Base64Decode(s); err == nil {
		return BytesToString(v)
	}
	return ""
}

// Base64DecodeToByte 解码为字节数组
func Base64DecodeToByte(s string) []byte {
	return StringToBytes(Base64DecodeToString(s))
}

// Base64DecodeByteToByte 解码字节数组到字节数组
func Base64DecodeByteToByte(b []byte) []byte {
	return Base64DecodeToByte(BytesToString(b))
}

// Base64DecodeStringToByte 解码字符串到字节数组
func Base64DecodeStringToByte(s string) []byte {
	return Base64DecodeToByte(s)
}

func IsBase64(s string) bool {
	b, _ := Base64Decode(s)
	return Base64EncodeString(string(b)) == s
}

func IsBase64String(s string) bool {
	return IsBase64(s)
}

func IsBase64Byte(b []byte) bool {
	return IsBase64(string(b))
}

func PKCS5Padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}

func PKCS5UnPadding(origData []byte) []byte {
	length := len(origData)
	unpadding := int(origData[length-1])
	return origData[:(length - unpadding)]
}

func ZeroPadding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padtext := bytes.Repeat([]byte{0}, padding)
	return append(ciphertext, padtext...)
}

func ZeroUnPadding(origData []byte) []byte {
	return bytes.TrimFunc(origData,
		func(r rune) bool {
			return r == rune(0)
		})
}

func DesEncrypt(src, key []byte) ([]byte, error) {
	block, err := des.NewCipher(key)
	if err != nil {
		return nil, err
	}
	bs := block.BlockSize()
	src = ZeroPadding(src, bs)
	// src = PKCS5Padding(src, bs)
	if len(src)%bs != 0 {
		return nil, errors.New("Need a multiple of the blocksize")
	}
	out := make([]byte, len(src))
	dst := out
	for len(src) > 0 {
		block.Encrypt(dst, src[:bs])
		src = src[bs:]
		dst = dst[bs:]
	}
	return out, nil
}

func DesDecrypt(src, key []byte) ([]byte, error) {
	block, err := des.NewCipher(key)
	if err != nil {
		return nil, err
	}
	out := make([]byte, len(src))
	dst := out
	bs := block.BlockSize()
	if len(src)%bs != 0 {
		return nil, errors.New("crypto/cipher: input not full blocks")
	}
	for len(src) > 0 {
		block.Decrypt(dst, src[:bs])
		src = src[bs:]
		dst = dst[bs:]
	}
	out = ZeroUnPadding(out)
	// out = PKCS5UnPadding(out)
	return out, nil
}
