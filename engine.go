package cyberetrieve

import (
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/N0el4kLs/cyberetrieve/sources"
	"github.com/N0el4kLs/cyberetrieve/sources/fofa"
	"github.com/N0el4kLs/cyberetrieve/sources/hunter"
	"github.com/N0el4kLs/cyberetrieve/sources/quake"

	"github.com/projectdiscovery/gologger"
)

type EngineMode uint8

const (
	ModeQuake EngineMode = 1 << (8 - 1 - iota)
	ModeFofa
	ModeHunter
)

// EngineOption is a type for setting options for the engine
type EngineOption func(c *CyberRetrieveEngine)

// CyberRetrieveEngine is the main struct for the engine
type CyberRetrieveEngine struct {
	// Query is the query struct
	// that contains the different engines query to search
	Query *sources.Query

	//
	channelBuffer int

	// searchMode is the mode of the search engine
	// e.g. quake, fofa
	searchMode EngineMode

	// sessions is the session for the providers
	sessions *sources.Session

	// providers is the list of providers
	providers []sources.Provider

	// providerLock is the lock for the providers
	providerWg *sync.WaitGroup

	// resultChannel is the channel for results
	resultChannel chan sources.Result

	// resultSlice is the slice of results
	resultSlice []sources.Result

	// mutex for save result into resultSlice
	mutex *sync.Mutex

	// isAutoGrammar is the flag to enable auto grammar,
	// which will transform the default query into the corresponding format
	isAutoGrammar bool

	// isDeepSearch is the flag to enable deep search
	// which will cost many points to detect corresponding targets as many as possible
	// now not support in Query.Query,
	// only support in Query.QuakeQuery, Query.FofaQuery, Query.HunterQuery
	// when use deep search mode, unlimited query number
	isDeepSearch bool
}

// NewCyberRetrieveEngine creates a new cyber retrieve engine
func NewCyberRetrieveEngine(query sources.Query, session sources.Session, engineOptions ...EngineOption) *CyberRetrieveEngine {
	bufSize := query.NumberOfQuery
	if bufSize < 0 {
		gologger.Fatal().Msgf("query number can't below 0")
	}

	if bufSize > 1000 {
		bufSize = bufSize / 50
	} else if 3 < bufSize && bufSize < 500 {
		bufSize = bufSize / 3
	}

	engine := &CyberRetrieveEngine{
		Query:         &query,
		channelBuffer: bufSize,
		sessions:      &session,
		providers:     make([]sources.Provider, 0, 2),
		providerWg:    &sync.WaitGroup{},
		resultChannel: make(chan sources.Result, bufSize),
		mutex:         &sync.Mutex{},
		isAutoGrammar: false,
		isDeepSearch:  false,
	}
	if len(engineOptions) != 0 {
		for _, opt := range engineOptions {
			opt(engine)
		}
	}

	// set deep search grammar
	if engine.isDeepSearch {

		// when use deep search mode, unlimited query number
		engine.Query.NumberOfQuery = -1

		// Todo only handle single keyword now, need to handle multiple keywords later
		if strings.Contains(engine.Query.Query, "domain") {
			d := strings.Split(engine.Query.Query, ":")[1]
			// remove the double quote
			d = d[1 : len(d)-1]
			engine.Query.Query = fmt.Sprintf(`%s OR cert:"%s"`, engine.Query.Query, d)
		}

		if engine.searchMode&ModeQuake == ModeQuake {
			if strings.Contains(engine.Query.QuakeQuery, "domain") {
				d := strings.Split(engine.Query.QuakeQuery, ":")[1]
				// remove the double quote
				d = d[1 : len(d)-1]
				engine.Query.QuakeQuery = fmt.Sprintf(`%s OR cert:"%s"`, engine.Query.QuakeQuery, d)
			}
		}

		if engine.searchMode&ModeFofa == ModeFofa {
			if strings.Contains(engine.Query.FofaQuery, "domain") {
				d := strings.Split(engine.Query.FofaQuery, "=")[1]
				// remove the double quote
				d = d[1 : len(d)-1]
				engine.Query.FofaQuery = fmt.Sprintf(`%s || cert="%s"`, engine.Query.FofaQuery, d)
			}
		}

		if engine.searchMode&ModeHunter == ModeHunter {
			if strings.Contains(engine.Query.HunterQuery, "domain") {
				d := strings.Split(engine.Query.HunterQuery, "=")[1]
				// remove the double quote
				d = d[1 : len(d)-1]
				engine.Query.HunterQuery = fmt.Sprintf(`%s || cert="%s"`, engine.Query.HunterQuery, d)
			}
		}
	}

	return engine
}

// WithFofaSearch this function is used to set the search mode to fofa
func WithFofaSearch() EngineOption {
	return func(c *CyberRetrieveEngine) {
		c.searchMode = c.searchMode | ModeFofa
	}
}

// WithQuakeSearch this function is used to set the search mode to quake
func WithQuakeSearch() EngineOption {
	return func(c *CyberRetrieveEngine) {
		c.searchMode = c.searchMode | ModeQuake
	}
}

func WithHunterSearch() EngineOption {
	return func(c *CyberRetrieveEngine) {
		c.searchMode = c.searchMode | ModeHunter
	}
}

// WithAutoGrammar this function is used to set the auto grammar option
func WithAutoGrammar() EngineOption {
	return func(c *CyberRetrieveEngine) {
		c.isAutoGrammar = true
	}
}

// WithDeepSearch this function is used to set the deep search option
func WithDeepSearch() EngineOption {
	return func(c *CyberRetrieveEngine) {
		c.isDeepSearch = true
	}
}

// RetrieveWithChannel this function return the result with chan Result  type
func (c *CyberRetrieveEngine) RetrieveWithChannel() (chan sources.Result, error) {
	if err := c.retrieve(); err != nil {
		// Todo handle this return value if err happened in retrieve
		return c.resultChannel, err
	} else {
		return c.resultChannel, nil
	}
}

// RetrieveResult this function return the result with []sources.Result slice type
func (c *CyberRetrieveEngine) RetrieveResult() ([]sources.Result, error) {
	// avoiding resultChannel block
	go func() {
		for {
			_, ok := <-c.resultChannel
			if !ok {
				break
			}
		}
	}()

	if err := c.retrieve(); err != nil {
		// Todo handle this return value if err happened in retrieve
		return c.resultSlice, err
	} else {
		return c.resultSlice, nil
	}
}

func (c *CyberRetrieveEngine) retrieve() error {
	if err := c.checkSession(); err != nil {
		return err
	}

	var (
		tmpRstsBroker = make(chan *sources.Result, c.channelBuffer)
	)

	for _, prd := range c.providers {
		c.providerWg.Add(1)

		go func(provider sources.Provider) {
			defer c.providerWg.Done()

			query := c.Query
			queryMap := map[string]string{
				quake.QUAKE:   query.QuakeQuery,
				fofa.FOFA:     query.FofaQuery,
				hunter.HUNTER: query.HunterQuery,
			}

			// if autoGrammar is on, and corresponding engine's query is empty,
			// then transfer the default query into the corresponding format
			if c.isAutoGrammar && queryMap[provider.Name()] == "" {
				prdGrammar := c.autoGrammar(query.Query, provider.Name())
				if prdGrammar != "" {
					switch provider.Name() {
					case quake.QUAKE:
						query.QuakeQuery = prdGrammar
					case fofa.FOFA:
						query.FofaQuery = prdGrammar
					case hunter.HUNTER:
						query.HunterQuery = prdGrammar
					}
					//gologger.Info().Msgf("Provider %s search grammar: %s\n", provider.Name(), prdGrammar)
				}
			}

			rstChannel, err := provider.Search(query)
			if err != nil {
				// Todo do something to handle the error
				return
			}
			for result := range rstChannel {
				tmpRstsBroker <- result
			}
		}(prd)
	}

	go func() {
		c.providerWg.Wait()
		close(tmpRstsBroker)
	}()

	// Unique results
	tmpList := make(map[sources.Result]struct{})
	for item := range tmpRstsBroker {
		if _, ok := tmpList[*item]; !ok {
			tmpList[*item] = struct{}{}
			c.resultChannel <- *item

			c.mutex.Lock()
			c.resultSlice = append(c.resultSlice, *item)
			c.mutex.Unlock()
		}
	}
	close(c.resultChannel)

	return nil
}

// check if the session is validated or not
func (c *CyberRetrieveEngine) checkSession() error {
	var (
		err       error = nil
		engineNum int   = 0
	)
	if c.searchMode&ModeQuake == ModeQuake {
		provider := &quake.Provider{}
		gologger.Info().Msgf("Check %s authorization,wait a second...\n", provider.Name())
		if ok := provider.Auth(c.sessions); !ok {
			errorMsg := fmt.Sprintf("%s auth err, please check your quake token", provider.Name())
			err = errors.New(errorMsg)
		} else {
			c.providers = append(c.providers, provider)
			engineNum++
		}
	}
	if c.searchMode&ModeFofa == ModeFofa {
		provider := &fofa.Provider{}
		gologger.Info().Msgf("Check %s authorization,wait a second...\n", provider.Name())
		if ok := provider.Auth(c.sessions); !ok {
			errorMsg := fmt.Sprintf("%s auth err, please check your quake token", provider.Name())
			err = errors.New(errorMsg)
		} else {
			c.providers = append(c.providers, provider)
			engineNum++
		}
	}
	if c.searchMode&ModeHunter == ModeHunter {
		provider := &hunter.Provider{}
		gologger.Info().Msgf("Check %s authorization,wait a second...\n", provider.Name())
		if ok := provider.Auth(c.sessions); !ok {
			errorMsg := fmt.Sprintf("%s auth err, please check your quake token", provider.Name())
			err = errors.New(errorMsg)
		} else {
			c.providers = append(c.providers, provider)
			engineNum++
		}
	}
	if engineNum == 0 {
		err = errors.New("please choose a search engine")
	} else if engineNum > 1 {
		c.isAutoGrammar = true
	}

	gologger.Info().Msgf("All search engine authorization check done...\n")

	// If have at least one engine can be used, just warning the unauthed engine and return nil
	if engineNum != 0 && err != nil {
		gologger.Warning().Msgf("%s\n", err)
		return nil
	} else {
		return err
	}
}

func (c *CyberRetrieveEngine) autoGrammar(query, name string) string {
	// Todo When you add new search engine, add case condition here
	switch name {
	case quake.QUAKE:
		result, err := quake.ToQuakeGrammar(query)
		if err != nil {
			gologger.Error().Msg(err.Error())
		}
		return result
	case fofa.FOFA:
		result, err := fofa.ToFofaGrammar(query)
		if err != nil {
			gologger.Error().Msgf(err.Error())
		}
		return result
	case hunter.HUNTER:
		result, err := hunter.ToHunterGrammer(query)
		if err != nil {
			gologger.Error().Msgf(err.Error())
		}
		return result
	default:
		return ""
	}
}
