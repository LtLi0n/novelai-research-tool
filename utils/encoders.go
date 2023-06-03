package utils

import (
	"github.com/wbrown/gpt_bpe"
)

var (
	encoderGPT2        *gpt_bpe.GPTEncoder
	encoderPile        *gpt_bpe.GPTEncoder
	encoderCLIP        *gpt_bpe.GPTEncoder
	encoderNerdstashV1 *gpt_bpe.GPTEncoder
	encoderNerdstashV2 *gpt_bpe.GPTEncoder
)

func GetEncoderByVocabId(vocabId *string) *gpt_bpe.GPTEncoder {
	switch *vocabId {
	case gpt_bpe.VOCAB_ID_GPT2:
		return GetEncoderGPT2()
	case gpt_bpe.VOCAB_ID_PILE:
		return GetEncoderPile()
	case gpt_bpe.VOCAB_ID_CLIP:
		return GetEncoderCLIP()
	case gpt_bpe.VOCAB_ID_NERDSTASH_V1:
		return GetEncoderNerdstashV1()
	case gpt_bpe.VOCAB_ID_NERDSTASH_V2:
		return GetEncoderNerdstashV2()
	default:
		return nil
	}
}

func GetEncoderGPT2() *gpt_bpe.GPTEncoder {
	if encoderGPT2 == nil {
		encoder := gpt_bpe.NewGPT2Encoder()
		encoderGPT2 = &encoder
	}
	return encoderGPT2
}

func GetEncoderPile() *gpt_bpe.GPTEncoder {
	if encoderPile == nil {
		encoder := gpt_bpe.NewPileEncoder()
		encoderPile = &encoder
	}
	return encoderPile
}

func GetEncoderCLIP() *gpt_bpe.GPTEncoder {
	if encoderPile == nil {
		encoder := gpt_bpe.NewCLIPEncoder()
		encoderCLIP = &encoder
	}
	return encoderCLIP
}

func GetEncoderNerdstashV1() *gpt_bpe.GPTEncoder {
	if encoderPile == nil {
		encoder := gpt_bpe.NewNerdstashV1Encoder()
		encoderNerdstashV1 = &encoder
	}
	return encoderNerdstashV1
}

func GetEncoderNerdstashV2() *gpt_bpe.GPTEncoder {
	if encoderPile == nil {
		encoder := gpt_bpe.NewNerdstashV2Encoder()
		encoderNerdstashV2 = &encoder
	}
	return encoderNerdstashV2
}
