package fofa

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/N0el4kLs/cyberetrieve/sources"

	"github.com/projectdiscovery/gologger"
)

const (
	FOFA              = "FOFA"
	BASE_URL          = "https://fofa.info/api/v1/"
	AUTH_URL          = "https://fofa.info/api/v1/info/my?key=%s"
	DEFAULT_PAGE_SIZE = 30
)

var (
	AUTHED_SEARCH_URL = BASE_URL + "search/all?key=%s"
)

type Provider struct {
}

// Name returns the name of the provider
func (p *Provider) Name() string {
	return FOFA
}

// Auth checks if the provider is valid to use
func (p *Provider) Auth(s *sources.Session) bool {
	if s.FofaKey == "" {
		return false
	}
	infoUrl := fmt.Sprintf(AUTH_URL, s.FofaKey)
	client := sources.DefaultClient
	resp, err := client.Get(infoUrl, nil)
	if err != nil {
		return false
	}
	//fmt.Printf("Fofa Auth Result: %s \n", resp.String())
	if !strings.Contains(resp.String(), `"error":false`) {
		return false
	}
	AUTHED_SEARCH_URL = fmt.Sprintf(AUTHED_SEARCH_URL, s.FofaKey)
	return true
}

// Search the result with provider
func (p *Provider) Search(query *sources.Query) (chan *sources.Result, error) {
	results := make(chan *sources.Result)
	go func() {
		defer close(results)
		numberOfResult := 0
		pageSize := DEFAULT_PAGE_SIZE
		if query.NumberOfQuery < DEFAULT_PAGE_SIZE {
			pageSize = query.NumberOfQuery
		}
		if query.NumberOfQuery == -1 {
			pageSize = 100 // todo need refactor
		}
		page := 1
		for {
			querySentence := query.Query
			// If AutoGrammar is on, use transferred grammar
			if query.FofaQuery != "" {
				querySentence = query.FofaQuery
			}
			queryFiled := NewFofaSearchFiled(querySentence, page, pageSize)

			currentSearchResult := p.query(queryFiled, results)
			numberOfResult += len(currentSearchResult.Results)

			if !isValidResult(currentSearchResult) || isOverSize(numberOfResult, query.NumberOfQuery) {
				// todo need refactor error handle chain
				gologger.Info().Label("Provider").Msgf("Fofa search done. You've found %d items\n", numberOfResult)
				break
			}
			page++
		}
	}()

	return results, nil
}

func (p *Provider) query(queryFiled *FofaSearchFiled, results chan *sources.Result) *FofaSearchResult {
	SEARCH_URL := AUTHED_SEARCH_URL +
		"&qbase64=%s" +
		"&page=%d" +
		"&size=%d" +
		"&full=%s" +
		"&fields=%s"

	searchUrl := fmt.Sprintf(SEARCH_URL,
		queryFiled.Query,
		queryFiled.Page,
		queryFiled.Size,
		queryFiled.Full,
		queryFiled.Fields,
	)
	resp, err := sources.DefaultClient.Get(searchUrl, nil)
	if err != nil {
		gologger.Debug().Msgf("Fofa Search Error: %s \n", err)
		return nil
	}
	fofaSearchResults := &FofaSearchResult{}
	err = resp.Into(fofaSearchResults)
	//gologger.Debug().Msgf("Fofa Search Result: %#v \n", fofaSearchResults)
	if err != nil {
		gologger.Debug().Msgf("Fofa search result unmarshal error: %s \n", err)
		return nil
	}

	for _, item := range fofaSearchResults.Results {
		searchResult := &sources.Result{}
		searchResult.IP = item[0]
		searchResult.Host = item[1]
		searchResult.Port, _ = strconv.Atoi(item[2])
		searchResult.Domain = item[3]
		//protocol = item[4]
		var url string
		if strings.HasPrefix(item[1], "http") {
			url = item[1]
		} else {
			url = fmt.Sprintf("%s://%s", item[4], item[1])
		}
		searchResult.URL = url

		results <- searchResult
	}

	return fofaSearchResults
}

// isOverSize check if the number of result is over the number of query which is required
func isOverSize(numberOfResult, numberOfQuery int) bool {
	var isOver = false
	if numberOfResult >= numberOfQuery && numberOfQuery != -1 {
		isOver = true
	}

	return isOver
}

// isValidResult check the result is valid or not
// If result is valid, return true
func isValidResult(result *FofaSearchResult) bool {
	if result == nil {
		return false
	}
	if result.Size == 0 {
		return false
	}
	if len(result.Results) == 0 {
		return false
	}
	return true
}

func ToFofaGrammar(s string) (string, error) {
	var (
		query      string
		subQueries []string
	)
	if strings.Contains(s, "&&") {
		subQueries = strings.Split(s, "&&")
	} else {
		subQueries = append(subQueries, s)
	}

	for _, q := range subQueries {
		if st, err := parse2FofaKeywords(strings.TrimSpace(q)); err != nil {
			return "", err
		} else {
			query += st + " && "
		}
	}

	// handle the latest suffix
	if strings.HasSuffix(query, "&& ") {
		query = query[0 : len(query)-4]
	}
	return query, nil
}

func parse2FofaKeywords(s string) (string, error) {
	var (
		equalSymbol = "="
	)

	if strings.HasPrefix(strings.ToLower(s), "not ") {
		equalSymbol = "!="
		s = s[4:]
	}

	keywords := strings.Split(s, ":")
	keyword, search := keywords[0], keywords[1]
	switch keyword {
	case "ip":
		return fmt.Sprintf("ip%s%s", equalSymbol, search), nil
	case "domain":
		return fmt.Sprintf("domain%s%s", equalSymbol, search), nil
	case "header":
		return fmt.Sprintf("headers%s%s", equalSymbol, search), nil
	case "title":
		return fmt.Sprintf("title%s%s", equalSymbol, search), nil
	case "body":
		return fmt.Sprintf("body%s%s", equalSymbol, search), nil
	default:
		return "", errors.New("transfer to FOFA grammar false")
	}
}
