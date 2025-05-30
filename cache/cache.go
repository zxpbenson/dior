package cache

import (
	"bufio"
	"os"
)

type Ticket struct {
	index uint32
}

type Cache struct {
	data      [][]byte
	size      uint32
	bufSizeMb int
}

func NewCache(bufSizeMb int) *Cache {
	return &Cache{size: 0, bufSizeMb: bufSizeMb}
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
	// 设置 buffer 上限为 10MB
	buf := make([]byte, 0, 1024*1024) // 初始 buffer 为 1MB
	scanner.Buffer(buf, cache.bufSizeMb*1024*1024)

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			break
		}
		cache.data = append(cache.data, []byte(line))
		cache.size++
	}
	err = scanner.Err()
	if err != nil {
		//fmt.Println("Error reading from file:", err) //print error if scanning is not done properly
	}

	return err
}
func (cache *Cache) Next(ticket *Ticket) []byte {
	index := ticket.Index(cache.size)
	return cache.data[index]
}
