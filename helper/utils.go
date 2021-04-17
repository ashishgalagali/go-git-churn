package helper

import (
	"fmt"
	"log"
	"os"
)

func init() {
	fmt.Println("Checking/creating outputs folder")
	if _, err := os.Stat("outputs"); os.IsNotExist(err) {
		fmt.Println("Creating outputs folder")

		os.MkdirAll("outputs", 0777)
	}
}

func UniqueElements(input []string) []string {
	u := make([]string, 0, len(input))
	m := make(map[string]bool)

	for _, val := range input {
		if _, ok := m[val]; !ok {
			m[val] = true
			u = append(u, val)
		}
	}

	return u
}

func AppendToFile(fileName, text string) {
	f, err := os.OpenFile(fileName,
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Println(err)
	}
	defer f.Close()
	if _, err := f.WriteString(text); err != nil {
		log.Println(err)
	}
}
