package frame

import (
	"bytes"
	"errors"
	"fmt"

	"image"
	"image/jpeg"
	"log"
	"reflect"

	"github.com/disintegration/imaging"
)

type FrameProcessor struct {
	Chain map[string]interface{}
}

var StubImaging = map[string]interface{}{
	"Grayscale":        imaging.Grayscale,
	"AdjustBrightness": imaging.AdjustBrightness,
	"AdjustSaturation": imaging.AdjustSaturation,
	"AdjustContrast":   imaging.AdjustContrast,
	"Blur":             imaging.Blur,
	"AdjustGamma":      imaging.AdjustGamma,
}

func New(chain map[string]interface{}) *FrameProcessor {
	chProcessor := &FrameProcessor{
		Chain: chain,
	}
	err := chProcessor.validateChain()
	if err != nil {
		return nil
	}
	return chProcessor
}

func (p *FrameProcessor) validateChain() error {
	for k, _ := range p.Chain {
		_, ok := StubImaging[k]
		if !ok {
			return errors.New("error: invalid chain " + k)
		}
	}
	return nil
}

// Process chain
func (p *FrameProcessor) ProcessChain(img image.Image) image.Image {
	var dst image.Image = img
	for k, v := range p.Chain {
		fmt.Printf("Processing chain %s", k)
		dst = p.executeStep(dst, k, v)

		buf := new(bytes.Buffer)
		err := jpeg.Encode(buf, dst, nil)
		if err != nil {
			log.Fatal(err)
		}
	}
	return dst
}

// ExecuteStep
func (p *FrameProcessor) executeStep(img image.Image, step string, value interface{}) image.Image {
	fmt.Printf("running step: %s", step)
	var resA interface{}
	var err error
	if value == 0 {
		resA, err = Call(step, img)
	} else {
		resA, err = Call(step, img, value)
	}
	if err != nil {
		log.Fatalf("failed to build image: %v", err)
	}
	return resA.(image.Image)
}

// Call method by string name
func Call(funcName string, params ...interface{}) (result interface{}, err error) {
	f := reflect.ValueOf(StubImaging[funcName])
	if len(params) != f.Type().NumIn() {
		err = errors.New("the number of params is out of index")
		return
	}
	in := make([]reflect.Value, len(params))
	for k, param := range params {
		in[k] = reflect.ValueOf(param)
	}
	var res []reflect.Value
	res = f.Call(in)
	result = res[0].Interface()
	return
}
