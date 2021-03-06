package evaluator

import (
	"testing"

	"github.com/butlermatt/monlox/lexer"
	"github.com/butlermatt/monlox/object"
	"github.com/butlermatt/monlox/parser"
)

func TestEvalNumberExpression(t *testing.T) {
	tests := []struct {
		input    string
		expected float32
	}{
		{"5", 5},
		{"10.45", 10.45},
		{"-5", -5},
		{"-10.45", -10.45},
		{"5 + 5 + 5 + 5 - 10", 10},
		{"2 * 2 * 2 * 2 * 2", 32},
		{"-50 + 100 + -50", 0},
		{"5.5 * 2 + 10", 21},
		{"20 + 2 * -10", 0},
		{"50 / 2 * 2 + 10", 60},
		{"2 * 2.5 * 10", 50},
		{"2 * (5 + 10)", 30},
		{"3 * 3 * 3 + 10", 37},
		{"(5 + 10 * 2 + 15 / 3) * 2 + -10", 50},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		testNumberObject(t, evaluated, tt.expected)
	}
}

func TestEvalBooleanExpression(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"true", true},
		{"false", false},
		{"1 < 2", true},
		{"1 > 2", false},
		{"1 < 1", false},
		{"1 > 1", false},
		{"1 == 1", true},
		{"1 != 1", false},
		{"1 != 2", true},
		{"1 <= 1", true},
		{"1 >= 1", true},
		{"1 <= 2", true},
		{"1 >= 2", false},
		{"true == true", true},
		{"true == false", false},
		{"false == false", true},
		{"true != true", false},
		{"true != false", true},
		{"(1 < 2) == true", true},
		{"(1 < 2) == false", false},
		{"(1 >= 2) == true", false},
		{"(1 >= 2) == false", true},
		{"true or true", true},
		{"true or false", true},
		{"false or true", true},
		{"true and true", true},
		{"true and false", false},
		{"false and false", false},
		{"false and true", false},
		{`"hello" == "hello"`, true},
		{`"hello" == "world"`, false},
		{`"hello" != "hello"`, false},
		{`"hello" != "world"`, true},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		testBooleanObject(t, evaluated, tt.expected)
	}
}

func TestBangOperator(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"!true", false},
		{"!false", true},
		{"!5", false},
		{"!!true", true},
		{"!!false", false},
		{"!!5", true},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		testBooleanObject(t, evaluated, tt.expected)
	}
}

func TestIfElseExpression(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		{"if (true) { 10 }", 10},
		{"if (false) { 10 }", nil},
		{"if (1) { 10 }", 10},
		{"if (1 < 2) { 10 }", 10},
		{"if (1 >= 2) { 10 }", nil},
		{"if (1 >= 2) { 10 } else { 20 }", 20},
		{"if (1 == 2 or 1 <= 2) { 10 } else { 20 }", 10},
		{"if (true and 1 == 1) { 10 } else { 20 }", 10},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		number, ok := tt.expected.(int)
		if ok {
			testNumberObject(t, evaluated, float32(number))
		} else {
			testNullObject(t, evaluated)
		}
	}
}

func TestReturnStatements(t *testing.T) {
	tests := []struct {
		input    string
		expected float32
	}{
		{"return 10;", 10},
		{"return 10; 9;", 10},
		{"return 2 * 5; 9;", 10},
		{"9; return 2 * 5; 9;", 10},
		{"if (10 > 1) { if (10 > 1) { return 10; } return 1; }", 10},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		testNumberObject(t, evaluated, tt.expected)
	}
}

func TestErrorHandling(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"5 + true;", "on line 1: type mismatch: NUMBER + BOOLEAN"},
		{"5 + true; 5;", "on line 1: type mismatch: NUMBER + BOOLEAN"},
		{"-true;", "on line 1: unknown operator: -BOOLEAN"},
		{"true + false;", "on line 1: unknown operator: BOOLEAN + BOOLEAN"},
		{"5; true + false; 5", "on line 1: unknown operator: BOOLEAN + BOOLEAN"},
		{"if (10 > 1) { true + false; }", "on line 1: unknown operator: BOOLEAN + BOOLEAN"},
		{"if (1 == true) { 10 }", "on line 1: type mismatch: NUMBER == BOOLEAN"},
		{`
if (10 > 1) {
   if (10 > 1) { 
     return true + false 
   } 
   return 1; 
}`, "on line 4: unknown operator: BOOLEAN + BOOLEAN"},
		{"foobar", "on line 1: identifier not found: foobar"},
		{`"Hello" - "World"`, "on line 1: unknown operator: STRING - STRING"},
		{`{"name": "Monkey"}[fn(x) { x }];`, "on line 1: unusable as hash key: FUNCTION"},
	}

	for i, tt := range tests {
		evaluated := testEval(tt.input)

		errObj, ok := evaluated.(*object.Error)
		if !ok {
			t.Errorf("test %d: wrong type returned. expected=*object.Error, got=%T (%+v)", i+1, evaluated, evaluated)
			continue
		}

		if errObj.Message != tt.expected {
			t.Errorf("test %d: wrong error message. expected=%q, got=%q", i+1, tt.expected, errObj.Message)
		}
	}
}

func TestLetStatements(t *testing.T) {
	tests := []struct {
		input    string
		expected float32
	}{
		{"let a = 5; a;", 5},
		{"let a = 5 * 5; a;", 25},
		{"let a = 5; let b = a; b;", 5},
		{"let a = 5; let b = a; let c = a + b + 5; c;", 15},
		{"let a = 10; let a = a * 2; a;", 20},
	}

	for _, tt := range tests {
		testNumberObject(t, testEval(tt.input), tt.expected)
	}
}

func TestFunctionObject(t *testing.T) {
	input := "fn(x) { x + 2; };"

	evaluated := testEval(input)
	fn, ok := evaluated.(*object.Function)
	if !ok {
		t.Fatalf("object is not expected type. expected=*object.Function, got=%T (%+[1]v)", evaluated)
	}

	if len(fn.Parameters) != 1 {
		t.Fatalf("function has wrong number of parameters. expected=%d, got=%d", 1, len(fn.Parameters))
	}

	if fn.Parameters[0].String() != "x" {
		t.Fatalf("parameter is wrong value. expected=%q, got=%q", "x", fn.Parameters[0].String())
	}

	expectedBody := "(x + 2)"

	if fn.Body.String() != expectedBody {
		t.Fatalf("body is wrong value. expected=%q, got=%q", expectedBody, fn.Body.String())
	}
}

func TestFunctionApplication(t *testing.T) {
	tests := []struct {
		input    string
		expected float32
	}{
		{"let identity = fn(x) { x; }; identity(5);", 5},
		{"let identity = fn(x) { return x; }; identity(4.5);", 4.5},
		{"let double = fn(x) { x * 2; }; double(5.5);", 11},
		{"let double = fn(x) { x * 2; }; double(-10.5)", -21},
		{"let add = fn(x, y) { x + y; }; add(5, 5);", 10},
		{"let add = fn(x, y) { x + y; }; add(5 + 5, add(5, 5));", 20},
		{"fn(x) { x; }(5)", 5},
	}

	for _, tt := range tests {
		testNumberObject(t, testEval(tt.input), tt.expected)
	}
}

func TestClosures(t *testing.T) {
	input := `let newAdder = fn(x) {
  fn(y) { x + y };
};

let addTwo = newAdder(2);
addTwo(2);`

	testNumberObject(t, testEval(input), 4)
}

func TestStringLiteral(t *testing.T) {
	input := `"Hello World!"`

	evaluated := testEval(input)
	str, ok := evaluated.(*object.String)
	if !ok {
		t.Fatalf("object is wrong type. expected=*object.String, got=%T (%+[1]v)", evaluated)
	}

	if str.Value != "Hello World!" {
		t.Errorf("String as wrong value. expected=%q, got=%q", "Hello World!", str.Value)
	}
}

func TestStringConcatenation(t *testing.T) {
	input := `"Hello" + " " + "World!"`

	evaluated := testEval(input)
	str, ok := evaluated.(*object.String)
	if !ok {
		t.Fatalf("object is wrong type. expected=*object.String, got=%T (%+[1]v)", evaluated)
	}

	if str.Value != "Hello World!" {
		t.Errorf("String has wrong value. expected=%q, got=%q", "Hello World!", str.Value)
	}
}

func TestBuiltinFunctions(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		{`len("")`, float32(0)},
		{`len("four")`, float32(4)},
		{`len("hello world")`, float32(11)},
		{`len(1)`, "on line 1: argument to `len` not supported. got=NUMBER"},
		{`len("one", "two")`, "on line 1: wrong number of arguments. expected=1, got=2"},
		{`len([1, 2, 3])`, float32(3)},
		{`len([])`, float32(0)},
		{`first([1, 2, 3])`, float32(1)},
		{`first([])`, nil},
		{`first(1)`, "on line 1: argument to `first` must be ARRAY, got=NUMBER"},
		{`last([1, 2, 3])`, float32(3)},
		{`last([])`, nil},
		{`last(1)`, "on line 1: argument to `last` must be ARRAY, got=NUMBER"},
		{`rest([1, 2, 3])`, []float32{2, 3}},
		{`rest([])`, nil},
		{`push([], 1)`, []float32{1}},
		{`push(1, 1)`, "on line 1: first argument to `push` must be ARRAY, got=NUMBER"},
		{`puts("hello", "world!")`, nil},
	}

	for i, tt := range tests {
		evaluated := testEval(tt.input)

		switch expected := tt.expected.(type) {
		case nil:
			testNullObject(t, evaluated)
		case float32:
			testNumberObject(t, evaluated, float32(expected))
		case string:
			errObj, ok := evaluated.(*object.Error)
			if !ok {
				t.Errorf("test %d: object is unexpected type. expected=*object.Error, got=%T (%+[2]v", i, evaluated)
				continue
			}
			if errObj.Message != expected {
				t.Errorf("unexpected error message. expected=%q, got=%q", expected, errObj.Message)
			}
		case []int:
			array, ok := evaluated.(*object.Array)
			if !ok {
				t.Errorf("obj not Array. got=%T (%+v)", evaluated, evaluated)
				continue
			}

			if len(array.Elements) != len(expected) {
				t.Errorf("wrong num of elements. want=%d, got=%d",
					len(expected), len(array.Elements))
				continue
			}

			for i, expectedElem := range expected {
				testNumberObject(t, array.Elements[i], float32(expectedElem))
			}
		}
	}
}

func TestArrayLiterals(t *testing.T) {
	input := "[1, 2 * 2, 3 + 3]"

	evaluated := testEval(input)
	result, ok := evaluated.(*object.Array)
	if !ok {
		t.Fatalf("object is wrong type. expected=*object.Array, got=%T (%+[1]v)", evaluated)
	}

	if len(result.Elements) != 3 {
		t.Fatalf("array has wrong number of elements. expected=%d, got=%d", 3, len(result.Elements))
	}

	testNumberObject(t, result.Elements[0], 1)
	testNumberObject(t, result.Elements[1], 4)
	testNumberObject(t, result.Elements[2], 6)
}

func TestArrayIndexExpressions(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		{"[1, 2, 3][0]", float32(1)},
		{"[1, 2, 3][1]", float32(2)},
		{"[1, 2, 3][2]", float32(3)},
		{"let i = 0; [1][i];", float32(1)},
		{"[1, 2, 3][1 + 1];", float32(3)},
		{"let myArray = [1, 2, 3]; myArray[2];", float32(3)},
		{"let myArray = [1, 2, 3]; myArray[0] + myArray[1] + myArray[2];", float32(6)},
		{"let myArray = [1, 2, 3]; let i = myArray[0]; myArray[i]", float32(2)},
		{"[1, 2, 3][3]", nil},
		{"[1, 2, 3][-1]", nil},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		number, ok := tt.expected.(float32)
		if ok {
			testNumberObject(t, evaluated, number)
		} else {
			testNullObject(t, evaluated)
		}
	}
}

func TestHashLiterals(t *testing.T) {
	input := `let two = "two";
{
	"one": 10 - 9,
	two: 1 + 1,
	"thr" + "ee": 6 / 2,
	4: 4,
	true: 5,
	false: 6
}`

	evaluated := testEval(input)
	result, ok := evaluated.(*object.Hash)
	if !ok {
		t.Fatalf("eval returned wrong type. expected=*object.Hash, got=%T (%+[1]v)", evaluated)
	}

	expected := map[object.HashKey]float32{
		(&object.String{Value: "one"}).HashKey():   1,
		(&object.String{Value: "two"}).HashKey():   2,
		(&object.String{Value: "three"}).HashKey(): 3,
		(&object.Number{Value: 4}).HashKey():       4,
		True.HashKey():                             5,
		False.HashKey():                            6,
	}

	if len(result.Pairs) != len(expected) {
		t.Fatalf("Hash has wrong number of elements. expected=%d, got=%d", len(expected), len(result.Pairs))
	}

	for exKey, exVal := range expected {
		p, ok := result.Pairs[exKey]
		if !ok {
			t.Errorf("no pair for given key %q", exKey)
		}

		testNumberObject(t, p.Value, exVal)
	}
}

func TestHashIndexExpressions(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		{`{"foo": 5}["foo"]`, float32(5)},
		{`{"foo": 5}["bar"]`, nil},
		{`let key = "foo"; {"foo": 5}[key]`, float32(5)},
		{`{}["foo"]`, nil},
		{`{5: 5}[5]`, float32(5)},
		{`{true: 5}[true]`, float32(5)},
		{`{false: 5}[false]`, float32(5)},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		number, ok := tt.expected.(float32)
		if ok {
			testNumberObject(t, evaluated, number)
		} else {
			testNullObject(t, evaluated)
		}
	}
}

func testNullObject(t *testing.T, obj object.Object) bool {
	if obj != Null {
		t.Errorf("object is not expected type. expected=Null, got=%T (%+v)", obj, obj)
		return false
	}

	return true
}

func testNumberObject(t *testing.T, obj object.Object, expected float32) bool {
	result, ok := obj.(*object.Number)
	if !ok {
		t.Errorf("object is not expected type. expected=*object.Number, got=%T, (%+v)", obj, obj)
		return false
	}
	if result.Value != expected {
		t.Errorf("object has wrong value. expected=%v, got=%v", expected, result.Value)
		return false
	}

	return true
}

func testBooleanObject(t *testing.T, obj object.Object, expected bool) bool {
	result, ok := obj.(*object.Boolean)
	if !ok {
		t.Errorf("object is not expected type. expected=*object.Boolean, got=%T (%+v)", obj, obj)
		return false
	}
	if result.Value != expected {
		t.Errorf("object has wrong value. expected=%v, got=%v", expected, result.Value)
		return false
	}

	return true
}

func testEval(input string) object.Object {
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	env := object.NewEnvironment()

	return Eval(program, env)
}
