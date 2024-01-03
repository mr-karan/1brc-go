package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
)

var (
	numWorkers           int
	measurementsFilePath string
	chunkSize            int64
)

type Stats struct {
	Min, Mean, Max float64
	Count          int
}

func init() {
	flag.StringVar(&measurementsFilePath, "file", "", "Path to the measurements file")
	flag.Int64Var(&chunkSize, "chunksize", 512*1024, "Size of each file chunk in bytes")
	flag.Parse()

	if measurementsFilePath == "" {
		fmt.Println("Error: Measurements file path is required")
		os.Exit(1)
	}

	numWorkers = runtime.NumCPU()
	runtime.GOMAXPROCS(numWorkers)
}

func main() {
	file, err := os.Open(measurementsFilePath)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	resultsChan := make(chan map[string]Stats, numWorkers)
	var wg sync.WaitGroup
	var aggWg sync.WaitGroup

	aggWg.Add(1)
	finalResults := make(map[string]Stats)

	// Start a separate goroutine for aggregation
	go func() {
		defer aggWg.Done()
		for workerResult := range resultsChan {
			for station, stats := range workerResult {
				finalStats, exists := finalResults[station]
				if !exists {
					finalResults[station] = stats
					continue
				}
				finalStats.Min = min(finalStats.Min, stats.Min)
				finalStats.Max = max(finalStats.Max, stats.Max)
				totalCount := finalStats.Count + stats.Count
				finalStats.Mean = (finalStats.Mean*float64(finalStats.Count) + stats.Mean*float64(stats.Count)) / float64(totalCount)
				finalStats.Count = totalCount
				finalResults[station] = finalStats
			}
		}
	}()

	buf := make([]byte, chunkSize)
	leftover := make([]byte, 0, chunkSize)

	go func() {
		for {
			bytesRead, err := file.Read(buf)
			if bytesRead > 0 {
				chunk := make([]byte, bytesRead)
				copy(chunk, buf[:bytesRead])
				validChunk, newLeftover := processChunk(chunk, leftover)
				leftover = newLeftover
				if len(validChunk) > 0 {
					wg.Add(1)
					go processChunkData(validChunk, resultsChan, &wg)
				}
			}
			if err != nil {
				break
			}
		}
		wg.Wait()
		close(resultsChan)
	}()

	aggWg.Wait()

	// Print results
	printStats(finalResults)
}

func processChunk(chunk, leftover []byte) (validChunk, newLeftover []byte) {
	firstNewline := -1
	lastNewline := -1
	for i, b := range chunk {
		if b == '\n' {
			if firstNewline == -1 {
				firstNewline = i
			}
			lastNewline = i
		}
	}
	if firstNewline != -1 {
		validChunk = append(leftover, chunk[:lastNewline+1]...)
		newLeftover = make([]byte, len(chunk[lastNewline+1:]))
		copy(newLeftover, chunk[lastNewline+1:])
	} else {
		newLeftover = append(leftover, chunk...)
	}
	return validChunk, newLeftover
}

func processChunkData(chunk []byte, resultsChan chan<- map[string]Stats, wg *sync.WaitGroup) {
	defer wg.Done()

	stationStats := make(map[string]Stats)
	scanner := bufio.NewScanner(strings.NewReader(string(chunk)))

	for scanner.Scan() {
		line := scanner.Text()

		// Find the index of the delimiter
		delimiterIndex := strings.Index(line, ";")
		if delimiterIndex == -1 {
			continue // Delimiter not found, skip this line
		}

		// Extract the station name and temperature string
		station := line[:delimiterIndex]
		tempStr := line[delimiterIndex+1:]

		// Convert the temperature string to a float
		temp, err := strconv.ParseFloat(tempStr, 64)
		if err != nil {
			continue // Invalid temperature value, skip this line
		}

		// Update the statistics for the station
		stats, exists := stationStats[station]
		if !exists {
			stats = Stats{Min: temp, Max: temp}
		}
		stats.Count++
		stats.Min = min(stats.Min, temp)
		stats.Max = max(stats.Max, temp)
		stats.Mean += (temp - stats.Mean) / float64(stats.Count)
		stationStats[station] = stats
	}

	// Send the computed stats to resultsChan
	resultsChan <- stationStats
}

func min(a, b float64) float64 {
	if a == 0 || a > b {
		return b
	}
	return a
}

func max(a, b float64) float64 {
	if a < b {
		return b
	}
	return a
}

func printStats(statsMap map[string]Stats) {
	var stations []string
	for station := range statsMap {
		stations = append(stations, station)
	}
	sort.Strings(stations)

	fmt.Print("{")
	for i, station := range stations {
		stats := statsMap[station]
		fmt.Printf("%s=%.1f/%.1f/%.1f", station, stats.Min, stats.Mean, stats.Max)
		if i < len(stations)-1 {
			fmt.Print(", ")
		}
	}
	fmt.Println("}")
}
