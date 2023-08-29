package cache

import (
	"bufio"
	"os"
)

type Ticket struct {
	index uint32
}

type Cache struct {
	data []string
	size uint32
}

func NewCache() *Cache {
	return &Cache{size: 0}
}

func NewTicket() *Ticket {
	return &Ticket{index: 0}
}

func (ticket *Ticket) Index(max uint32) (index uint32) {
	index = ticket.index
	if index == max-1 {
		ticket.index = 0
		return
	}
	ticket.index += 1
	return
}

func (cache *Cache) Load(path string) error {

	file, err := os.Open(path) //open the file
	if err != nil {
		//fmt.Printf("Error opening file: %s, %v", path, err)
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file) //scan the contents of a file and print line by line

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			break
		}
		cache.data = append(cache.data, line)
		cache.size++
	}
	err = scanner.Err()
	if err != nil {
		//fmt.Println("Error reading from file:", err) //print error if scanning is not done properly
	}

	return err
}
func (cache *Cache) Next(ticket *Ticket) string {
	index := ticket.Index(cache.size)
	return cache.data[index]
}
