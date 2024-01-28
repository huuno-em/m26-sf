package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"time"
)

const (
	bufferSize = 10              // Размер буфера
	interval   = 3 * time.Second // Интервал опустошения буфера
)

func filterNegative(in <-chan int) <-chan int {
	out := make(chan int)

	go func() {
		defer close(out)
		for num := range in {
			if num >= 0 {
				out <- num
			}
		}
	}()

	return out
}

func filterNonMultipleOfThree(in <-chan int) <-chan int {
	out := make(chan int)

	go func() {
		defer close(out)
		for num := range in {
			if num%3 == 0 && num != 0 {
				out <- num
			}
		}
	}()

	return out
}

func bufferProcessor(in <-chan int, bufferSize int, interval time.Duration) <-chan []int {
	out := make(chan []int)
	buffer := make([]int, 0, bufferSize)

	go func() {
		defer close(out)
		timer := time.NewTicker(interval)
		defer timer.Stop()

		for {
			select {
			case num, ok := <-in:
				if !ok {
					if len(buffer) > 0 {
						out <- buffer
						log.Printf("Buffered data: %v\n", buffer)
					}
					return
				}
				buffer = append(buffer, num)
				if len(buffer) == bufferSize {
					out <- buffer
					log.Printf("Buffered data: %v\n", buffer)
					buffer = make([]int, 0, bufferSize)
				}
			case <-timer.C:
				if len(buffer) > 0 {
					out <- buffer
					log.Printf("Buffered data: %v\n", buffer)
					buffer = make([]int, 0, bufferSize)
				}
			}
		}
	}()

	return out
}

func sourceData(out chan<- int) {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		// Фильтрация нечисловых данных
		num := 0
		_, err := fmt.Sscan(scanner.Text(), &num)
		if err == nil {
			out <- num
			log.Printf("Received data: %d\n", num)
		} else {
			log.Printf("Invalid number: %s\n", scanner.Text())
		}
	}
	close(out)
}

func consumer(in <-chan []int) {
	for buffer := range in {
		log.Printf("Processed data: %v\n", buffer)
	}
}

func main() {
	// Создаем каналы
	dataSource := make(chan int)
	filteredNegative := filterNegative(dataSource)
	filteredNonMultipleOfThree := filterNonMultipleOfThree(filteredNegative)
	bufferedData := bufferProcessor(filteredNonMultipleOfThree, bufferSize, interval)

	// Запускаем источник данных
	go sourceData(dataSource)

	// Запускаем потребителя данных
	consumer(bufferedData)
}
