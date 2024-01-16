package qwen_client

import "fmt"

const QWEN_DASHSCOPE_URL = "https://dashscope.aliyuncs.com/api/v1/services/aigc/text-generation/generation"

type Qwen_Model string

const QWEN_TURBO Qwen_Model = "qwen-turbo"
const QWEN_PLUS Qwen_Model = "qwen-plus"
const QWEN_MAX Qwen_Model = "qwen-max"
const QWEN_MAX_1201 Qwen_Model = "qwen-max-1201"
const QWEN_MAX_LONGCONTEXT Qwen_Model = "qwen-max-longcontext"

type Model struct{}

func ChoseQwenModel(model string) Qwen_Model {
	m := Model{}
	switch model {
	case "qwen-turbo":
		return m.QwenTurbo()
	case "qwen-plus":
		return m.QwenPlus()
	case "qwen-max":
		return m.QwenMax()
	case "qwen-max-1201":
		return m.QwenMax1201()
	case "qwen-max-longcontext":
		return m.QwenMaxLongContext()
	default:
		fmt.Println("target model not found, use default model: qwen-turbo")
		return m.QwenTurbo()
	}
}

func (m *Model) QwenTurbo() Qwen_Model {
	return QWEN_TURBO
}

func (m *Model) QwenPlus() Qwen_Model {
	return QWEN_PLUS
}

func (m *Model) QwenMax() Qwen_Model {
	return QWEN_MAX
}

func (m *Model) QwenMax1201() Qwen_Model {
	return QWEN_MAX_1201
}

func (m *Model) QwenMaxLongContext() Qwen_Model {
	return QWEN_MAX_LONGCONTEXT
}

