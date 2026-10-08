package main

import "fmt"

func Where[T any](source []T, pred func(T) bool) []T {
	var result []T
	for _, item := range source {
		if pred(item) {
			result = append(result, item)
		}
	}

	return result
}

func First[T any](source []T, pred func(T) bool) (T, error) {
	fmt.Println("source", source)
	var result T
	for _, item := range source {
		if pred(item) {
			return item, nil
		}
	}

	return result, fmt.Errorf("no item found")
}
