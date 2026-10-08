package main

import (
	"bufio"
	"fmt"
	"maps"
	"os"
	"regexp"
	"slices"
	"strings"
	"time"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Please provide a file path as an argument.")
		return
	}

	filePath := os.Args[1]
	fmt.Println("File path provided:", filePath)

	_, err := os.Stat(filePath)
	if err != nil && os.IsNotExist(err) {
		fmt.Println("File does not exist:", filePath)
		os.Exit(1)
	}

	fmt.Println("File exists:", filePath)

	bytes, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Println("Error reading file:", err)
		os.Exit(1)
	}

	text := string(bytes)
	fmt.Println("File length:", len(bytes)/1024, "KB")

	lines := strings.Split(strings.TrimSpace(text), "\n")

	dict := make(map[int][]string)
	for i, line := range lines {
		//formatted := fmt.Sprintf("%d. %s", i, line)
		//fmt.Println(formatted)
		chunks := strings.Split(line, " ")
		dict[i] = chunks
	}

	foo := maps.Keys(dict)
	//sortedKeys := slices.Collect(foo)
	//slices.Sort(sortedKeys)

	sortedKeys := slices.Sorted(foo)
	for k := range sortedKeys {
		v := dict[k]
		fmt.Printf("%d ", k)
		for _, word := range v {
			fmt.Printf("%v ", word)
			//time.Sleep(5 * time.Millisecond)
		}
		fmt.Println()
	}

	//using bufio
	wordCount, err := countWords(filePath)
	if err != nil {
		fmt.Println("Error counting words:", err)
		os.Exit(1)
	}

	for word, count := range wordCount {
		fmt.Printf("%s: %d\n", word, count)
		time.Sleep(10 * time.Millisecond)
	}

	os.Exit(0)

}

func countWords(filePath string) (map[string]int, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %s", filePath)
	}
	defer file.Close()

	wordCount := map[string]int{}
	scanner := bufio.NewScanner(file)
	regex := regexp.MustCompile(`\w+`) // Regex to match a word
	for scanner.Scan() {
		line := scanner.Text()
		//parts := strings.Split(line, " ")
		parts := regex.FindAllString(line, -1)
		for _, word := range parts {
			word = strings.ToLower(word)
			wordCount[word] += 1
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to scan file: %s", filePath)
	}

	return wordCount, nil
}
