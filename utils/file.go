package utils

import (
	"bufio"
	"bytes"
	"io"
	"io/fs"
	"os"
)

func NewFileBuffer(path string) *FileBuffer {
	return &FileBuffer{Path: path}
}

type FileBuffer struct {
	Path    string
	Flag    int
	Mode    fs.FileMode
	Content *bytes.Buffer
}

func (f *FileBuffer) Write(content []byte) *FileBuffer {
	f.Content.Write(content)
	return f
}

func (f *FileBuffer) WriteLine(lineContent []byte) *FileBuffer {
	f.Content.Write(lineContent)
	f.Content.WriteByte(10)
	return f
}

func (f *FileBuffer) WriteString(content string) *FileBuffer {
	f.Content.WriteString(content)
	return f
}

func (f *FileBuffer) WriteStringLine(lineContent string) *FileBuffer {
	f.Content.WriteString(lineContent)
	f.Content.WriteByte(10)
	return f
}

// CountLine 统计文件行数
func (f *FileBuffer) CountLine() (count int, err error) {
	var (
		readSize      = 0
		buf           = make([]byte, 1024)
		file, openErr = os.Open(f.Path)
	)

	if openErr != nil {
		err = openErr
		return
	}
	defer file.Close()

	for {
		readSize, err = file.Read(buf)
		if err != nil {
			break
		}
		var buffPosition int
		for {
			i := bytes.IndexByte(buf[buffPosition:], '\n')
			if i == -1 || readSize == buffPosition {
				break
			}
			buffPosition += i + 1
			count++
		}
	}
	if readSize > 0 && count == 0 || count > 0 {
		count++
	}
	if err == io.EOF {
		return count, nil
	}

	return count, err
}

// ReadLastLine 读取最后 n 行
func (f *FileBuffer) ReadLastLine(n int) []byte {
	var (
		file, err = os.Open(f.Path)
		info, _   = file.Stat()
		filesize  = info.Size()
	)
	if err != nil {
		return []byte{}
	}
	// 最少读一行
	if n == 0 {
		n = 1
	}
	defer file.Close()
	result := make([]byte, 0, 32*1024)
	// 读取 n 行
	for n > 0 && filesize > 0 {
		filesize--
		file.Seek(filesize, 0)
		// 这个必须在里面分配
		char := make([]byte, 1)
		if _, err := file.Read(char); err != nil {
			break
		}
		// 倒序读，正序存
		result = append(char, result...)
		// 获取最后 n 行
		if char[0] == 10 {
			n--
		}
	}
	if result[0] == 10 {
		return result[1:]
	}
	return result
}

// ReadLastLineString 读取最后 n 行
// 返回字符串
func (f *FileBuffer) ReadLastLineString(n int) string {
	return BytesToString(f.ReadLastLine(n))
}

// Read 读取指定行数
func (f *FileBuffer) Read(n int) ([]byte, error) {
	file, err := os.Open(f.Path)
	if err != nil {
		return []byte{}, err
	}
	defer file.Close()
	// 按行读取
	if n > 0 {
		var (
			r   = bufio.NewReader(file)
			buf bytes.Buffer
		)
		for n > 0 {
			n--
			line, _, err := r.ReadLine()
			if err != nil {
				break
			}
			buf.Write(line)
			buf.WriteByte(10)
		}
		return bytes.TrimRight(buf.Bytes(), "\n"), nil
	}
	// 全部读取
	return io.ReadAll(file)
}

// ReadRange 从指定位置开始读取n行
func (f FileBuffer) ReadRange(start, n int) []byte {
	var (
		file, _ = os.Open(f.Path)
		scanner = bufio.NewScanner(file)
		buf     bytes.Buffer
	)
	defer file.Close()
	// 计算机从 0 开始计数
	start--
	for scanner.Scan() {
		// 从指定行数 + n 读取行完毕
		if 1 > start|n {
			break
		}
		// 扣除开始行数
		if start > 0 {
			start--
			continue
		}
		// 正式开始读取
		buf.Write(scanner.Bytes())
		buf.WriteByte(10)
		// 扣除剩余行数
		n--
	}
	return bytes.TrimRight(buf.Bytes(), "\n")
}
