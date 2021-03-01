package main

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
)

func Precedence(C byte) int {
	switch C {
	case '*':
		fallthrough
	case '/':
		return 2
	case '+':
		fallthrough
	case '-':
		return 1
	default:
		return 0
	}
}

func IsOperator(C byte) bool {
	return Precedence(C) > 0
}

func InfixToPostfix(Infix string) (Result []string, Err error) {
	Result = nil
	Err = nil

	OperandBuffer := make([]byte, 0)
	Postfix := make([]string, 0)
	OperatorStack := make([]byte, 0)
	for I := 0; I < len(Infix); I++ {
		C := Infix[I]
		if IsOperator(C) {
			Postfix = append(Postfix, string(OperandBuffer))
			OperandBuffer = OperandBuffer[:0]
			for len(OperatorStack) != 0 {
				LastIndex := len(OperatorStack) - 1
				LastOperator := OperatorStack[LastIndex]
				if Precedence(C) > Precedence(LastOperator) {
					break
				} else {
					Postfix = append(Postfix, string(LastOperator))
					OperatorStack = OperatorStack[:LastIndex]
				}
			}
			OperatorStack = append(OperatorStack, C)

		} else if C == '(' {
			Postfix = append(Postfix, string(OperandBuffer))
			OperandBuffer = OperandBuffer[:0]
			OperatorStack = append(OperatorStack, C)
		} else if C == ')' {
			Postfix = append(Postfix, string(OperandBuffer))
			OperandBuffer = OperandBuffer[:0]

			FoundOpenBraces := false
			// Pop stack until '('
			for len(OperatorStack) != 0 {
				LastIndex := len(OperatorStack) - 1
				LastOperator := OperatorStack[LastIndex]
				if LastOperator == '(' {
					FoundOpenBraces = true
					break
				}
				Postfix = append(Postfix, string(LastOperator))
			}
			if !FoundOpenBraces {
				Err = fmt.Errorf("[InfixToPostfix] Invalid Expression\n")
				return
			}
		} else {
			OperandBuffer = append(OperandBuffer, C)
		}
	}
	Postfix = append(Postfix, string(OperandBuffer))
	OperandBuffer = OperandBuffer[:0]

	for len(OperatorStack) != 0 {
		LastIndex := len(OperatorStack) - 1
		LastOperator := OperatorStack[LastIndex]
		OperatorStack = OperatorStack[:LastIndex]
		Postfix = append(Postfix, string(LastOperator))
	}

	Result = Postfix
	return
}

func Operate(L int, R int, Op byte) (Result int, Err error) {
	Err = nil
	Result = 0
	switch Op {
	case '*':
		Result = L * R
	case '/':
		if R == 0 {
			Err = fmt.Errorf("[Operate] R is 0\n")
			return
		}
		Result = L / R
	case '+':
		Result = L + R
	case '-':
		Result = L - R
	default:
		Err = fmt.Errorf("[Operate] Invalid path\n")
	}
	return
}

type Evaluate_Postfix_Result struct {
	Sum      int
	DiceInfo string
}

// Returns: Success, Sum, Rolls
func EvaluatePostfix(Postfix []string) (Result *Evaluate_Postfix_Result, Err error) {
	Result = nil
	Err = nil

	OperandStack := make([]int, 0, 10)
	DiceInfo := string(make([]byte, 0, DiscordMessageMaxChars))

	for _, Str := range Postfix {
		if IsOperator(Str[0]) {
			LastIndex := len(OperandStack) - 1
			if LastIndex == 0 {
				Err = fmt.Errorf("[EvaluatePostfix] Not enough values in stack\n")
				return
			}
			B := OperandStack[LastIndex]
			A := OperandStack[LastIndex-1]
			OperandStack = OperandStack[:LastIndex-1]
			C, OperateErr := Operate(B, A, Str[0])
			if OperateErr != nil {
				Err = fmt.Errorf("[EvaluatePostfix] Issues operating: %v", OperateErr)
				return
			}
			OperandStack = append(OperandStack, C)
		} else {
			// Attempt to evaluate the string
			// We either accept a raw number, or a number seperated by 'd'
			// (e.g. 1d6)
			SplitArr := strings.Split(Str, "d")
			if len(SplitArr) < 2 {
				SplitArr = strings.Split(Str, "D")
			}

			// From here, we only deal with SplitArr len == 1 OR 2.
			// All other cases is an error
			if len(SplitArr) == 1 {
				// Normal case: Raw numbers
				Value, ConvErr := strconv.Atoi(SplitArr[0])
				if ConvErr != nil {
					Err = fmt.Errorf("[EvaluatePostfix] %s\n", ConvErr)
					return
				}
				OperandStack = append(OperandStack, Value)
			} else if len(SplitArr) == 2 {
				// Dice roll case
				Times, ConvErr := strconv.Atoi(SplitArr[0])
				if ConvErr != nil {
					Err = fmt.Errorf("[EvaluatePostfix] %s\n", ConvErr)
					return
				}
				Sides, ConvErr := strconv.Atoi(SplitArr[1])
				if ConvErr != nil {
					Err = fmt.Errorf("[EvaluatePostfix] %s\n", ConvErr)
					return
				}
				DiceInfo += fmt.Sprintf("%dd%d = ", Times, Sides)
				Sum := 0
				for TimeIndex := 0; TimeIndex < Times; TimeIndex++ {
					RollResult := rand.Intn(Sides) + 1
					Sum += RollResult
					if TimeIndex != Times-1 {
						// If not the last one
						DiceInfo += fmt.Sprintf("%d + ", RollResult)
					} else {
						if TimeIndex == 0 {
							// If there's only one entry
							DiceInfo += fmt.Sprintf("%d\n", Sum)
						} else {
							DiceInfo += fmt.Sprintf("%d = %d\n", RollResult, Sum)
						}
					}
				}
				OperandStack = append(OperandStack, Sum)
			} else {
				Err = fmt.Errorf("[EvaluatePostfix] Invalid split\n")
				return
			}

		}
	}

	if len(OperandStack) == 1 {
		Result = &Evaluate_Postfix_Result{}
		Result.Sum = OperandStack[0]
		Result.DiceInfo = DiceInfo
		return
	}

	Err = fmt.Errorf("[EvaluatePostfix] Final OperandStack issue!")
	return
}
