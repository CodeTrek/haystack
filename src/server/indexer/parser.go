package indexer

import (
	"search-indexer/running"
	"sync"
)

type Parser struct {
}

func NewParser() *Parser {
	return &Parser{}
}

func (p *Parser) Start(wg *sync.WaitGroup) {
	wg.Add(1)
	go p.run(wg)
}

func (p *Parser) run(wg *sync.WaitGroup) {
	defer wg.Done()
	running.WaitingForShutdown()
}
