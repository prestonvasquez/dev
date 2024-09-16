package main

import (
	"fmt"

	"github.com/Knetic/govaluate"
)

type Document struct {
	EncodedName string
	Name        string
	Tags        []string
	Size        int64
}

// evaluateExpression takes a string expression and evaluates it against the document
func evaluateExpression(expString string, doc Document) (bool, error) {
	// Map to hold the document's fields for evaluation
	parameters := make(map[string]interface{})
	parameters["name"] = doc.Name
	parameters["size"] = doc.Size

	tagSet := make(map[string]bool)
	for _, tag := range doc.Tags {
		tagSet[tag] = true
	}

	// Custom function to check if the document has the specified tag
	functions := map[string]govaluate.ExpressionFunction{
		"tag": func(args ...interface{}) (interface{}, error) {
			for _, arg := range args {
				if tagSet[arg.(string)] {
					return true, nil
				}
			}

			return false, nil
		},
	}

	expression, err := govaluate.NewEvaluableExpressionWithFunctions(expString, functions)
	if err != nil {
		return false, err
	}

	// Create an expression object
	//expr, err := govaluate.NewEvaluableExpression(expression)
	//if err != nil {
	//	return false, err
	//}

	// Evaluate the expression against the document
	result, err := expression.Evaluate(parameters)
	if err != nil {
		return false, err
	}

	// Convert the result to a boolean value
	return result.(bool), nil
}

func main() {
	// Example document
	doc := Document{
		Name: "x",
		Tags: []string{"y", "z"},
		Size: 4,
	}

	// Example expression, using single quotes and proper logical operators
	expression := "name == 'x' && tag('x')"

	// Evaluate the expression
	result, err := evaluateExpression(expression, doc)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Println("Match:", result)
}
