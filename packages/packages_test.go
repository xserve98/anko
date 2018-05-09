package packages

import (
	"testing"

	"github.com/mattn/anko/parser"
	"github.com/mattn/anko/vm"
)

type Test struct {
	Script         string
	ParseError     error
	ParseErrorFunc *func(*testing.T, error)
	EnvSetupFunc   *func(*testing.T, *vm.Env)
	Types          map[string]interface{}
	Input          map[string]interface{}
	RunError       error
	RunErrorFunc   *func(*testing.T, error)
	RunOutput      interface{}
	Output         map[string]interface{}
}

type TestingOptions struct {
	EnvSetupFunc *func(*testing.T, *vm.Env)
}

// RunTests runs VM tests
func RunTests(t *testing.T, tests []Test, testingOptions *TestingOptions) {
	for _, test := range tests {
		RunTest(t, test, testingOptions)
	}
}

// RunTest runs a VM test
func RunTest(t *testing.T, test Test, testingOptions *TestingOptions) {
	stmts, err := parser.ParseSrc(test.Script)
	if test.ParseErrorFunc != nil {
		(*test.ParseErrorFunc)(t, err)
	} else if err != nil && test.ParseError != nil {
		if err.Error() != test.ParseError.Error() {
			t.Errorf("ParseSrc error - received: %v - expected: %v - script: %v", err, test.ParseError, test.Script)
			return
		}
	} else if err != test.ParseError {
		t.Errorf("ParseSrc error - received: %v - expected: %v - script: %v", err, test.ParseError, test.Script)
		return
	}
	// Note: Still want to run the code even after a parse error to see what happens

	env := vm.NewEnv()
	if testingOptions != nil && testingOptions.EnvSetupFunc != nil {
		(*testingOptions.EnvSetupFunc)(t, env)
	}
	if test.EnvSetupFunc != nil {
		(*test.EnvSetupFunc)(t, env)
	}

	for typeName, typeValue := range test.Types {
		err = env.DefineType(typeName, typeValue)
		if err != nil {
			t.Errorf("DefineType error: %v - typeName: %v - script: %v", err, typeName, test.Script)
			return
		}
	}

	for inputName, inputValue := range test.Input {
		err = env.Define(inputName, inputValue)
		if err != nil {
			t.Errorf("Define error: %v - inputName: %v - script: %v", err, inputName, test.Script)
			return
		}
	}

	var value interface{}
	value, err = vm.Run(stmts, env)
	if test.RunErrorFunc != nil {
		(*test.RunErrorFunc)(t, err)
	} else if err != nil && test.RunError != nil {
		if err.Error() != test.RunError.Error() {
			t.Errorf("Run error - received: %v - expected: %v - script: %v", err, test.RunError, test.Script)
			return
		}
	} else if err != test.RunError {
		t.Errorf("Run error - received: %v - expected: %v - script: %v", err, test.RunError, test.Script)
		return
	}

	if !vm.ValueEqual(value, test.RunOutput) {
		t.Errorf("Run output - received: %#v - expected: %#v - script: %v", value, test.RunOutput, test.Script)
		t.Errorf("received type: %T - expected: %T", value, test.RunOutput)
		return
	}

	for outputName, outputValue := range test.Output {
		value, err = env.Get(outputName)
		if err != nil {
			t.Errorf("Get error: %v - outputName: %v - script: %v", err, outputName, test.Script)
			return
		}

		if !vm.ValueEqual(value, outputValue) {
			t.Errorf("outputName %v - received: %#v - expected: %#v - script: %v", outputName, value, outputValue, test.Script)
			t.Errorf("received type: %T - expected: %T", value, outputValue)
			continue
		}
	}

	env.Destroy()
}
