package ghttp

import (
	"bytes"
	"io"
	"io/ioutil"
	"mime/multipart"
)

type MultipartDataType int8

//MultipartDataType 的枚举
const (
	MultipartDataTypeFile MultipartDataType = iota
	MultipartDataTypeFormData
	MultipartDataTypeContent
)

//用于 post multipart 的接口
type IMultipart interface {
	//实现 Read函数
	io.Reader

	//获取这次的 boundary 字符串
	ContentType() string
}

type EasyMultipart struct {
	//multipart 内容
	buf *bytes.Buffer

	//随机生成的用于multipart 的 boundary字符串
	contentType string
}

func (m *EasyMultipart) ContentType() string {
	return m.contentType
}

func (m *EasyMultipart) Read(p []byte) (n int, err error) {
	return m.buf.Read(p)
}

//用于 `MultipartBuilder` 中的内容记录
type MultipartDataContent struct {
	Type    MultipartDataType
	Content []byte
}

//用于构造一个 multipart
type MultipartBuilder struct {
	content map[string]*MultipartDataContent
}

//初始化一个 multipart 构造器
func NewMultipartBuilder() *MultipartBuilder {
	return &MultipartBuilder{
		content: make(map[string]*MultipartDataContent),
	}
}

//添加文件
//name upload时的name
//fileName 文件的路径+文件名
func (m *MultipartBuilder) AddFile(name, fileName string) *MultipartBuilder {
	m.content[name] = &MultipartDataContent{
		Type:    MultipartDataTypeFile,
		Content: []byte(fileName),
	}
	return m
}

//添加form-data数据
func (m *MultipartBuilder) AddFromData(name, value string) *MultipartBuilder {
	m.content[name] = &MultipartDataContent{
		Type:    MultipartDataTypeFormData,
		Content: []byte(value),
	}
	return m
}

//以map的形式 添加form-data数据, k=>form-data的name, v=>form-data的Valve
func (m *MultipartBuilder) FromData(value map[string]string) *MultipartBuilder {
	for k, v := range value {
		m.content[k] = &MultipartDataContent{
			Type:    MultipartDataTypeFormData,
			Content: []byte(v),
		}
	}
	return m
}

//添加[]byte, name, form-data的name
func (m *MultipartBuilder) AddBytes(name string, bytes []byte) *MultipartBuilder {
	m.content[name] = &MultipartDataContent{
		Type:    MultipartDataTypeContent,
		Content: bytes,
	}
	return m
}

//构造 MultipartData
func (m *MultipartBuilder) Builder() (*EasyMultipart, error) {
	buf := new(bytes.Buffer)
	mulWriter := multipart.NewWriter(buf)
	defer mulWriter.Close()

	for name, content := range m.content {
		switch content.Type {
		case MultipartDataTypeFile:
			//文件类型,读取文件,写入buf
			formFile, err := mulWriter.CreateFormFile(name, name)
			if err != nil {
				return nil, err
			}
			file, err := ioutil.ReadFile(string(content.Content))
			if err != nil {
				return nil, err
			}
			_, err = formFile.Write(file)
			if err != nil {
				return nil, err
			}
		case MultipartDataTypeFormData:
			//form-data类型,直接写入
			err := mulWriter.WriteField(name, string(content.Content))
			if err != nil {
				return nil, err
			}
		case MultipartDataTypeContent:
			//[]byte 类型,当做文件写入
			formFile, err := mulWriter.CreateFormFile(name, name)
			if err != nil {
				return nil, err
			}
			_, err = formFile.Write(content.Content)
			if err != nil {
				return nil, err
			}
		}
	}
	return &EasyMultipart{
		buf:         buf,
		contentType: mulWriter.FormDataContentType(),
	}, nil
}
