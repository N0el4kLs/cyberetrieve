package quake

import (
	"errors"
	"fmt"
	"strings"

	"github.com/N0el4kLs/cyberetrieve/sources"

	"github.com/projectdiscovery/gologger"
)

const (
	QUAKE             = "QUAKE"
	AUTH_URL          = "https://quake.360.cn/api/v3/user/info"
	SEARCH_URL        = "https://quake.360.cn/api/v3/search/quake_service"
	DEFAULT_PAGE_SIZE = 30
)

var (
	quakeHeader = map[string]string{
		"X-QuakeToken": "",
	}
)

type Provider struct {
}

// Name returns the name of the provider
func (p *Provider) Name() string {
	return QUAKE
}

// Auth checks if the provider is valid to use
func (p *Provider) Auth(s *sources.Session) bool {
	client := sources.DefaultClient
	if s.QuakeToken == "" {
		return false
	}
	quakeHeader["X-QuakeToken"] = s.QuakeToken
	resp, err := client.Get(AUTH_URL, quakeHeader)
	if err != nil {
		return false
	}
	//fmt.Printf("Quake Auth Resp: %s \n", resp.String())
	if !strings.Contains(resp.String(), `"message":"Successful."`) {
		return false
	}
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
			pageSize = 100 // Todo handle the unlimited query number
		}
		for {
			querySentence := query.Query
			// If AutoGrammar is on, use transferred grammar
			if query.QuakeQuery != "" {
				querySentence = query.QuakeQuery
			}
			queryFiled := NewQuakeSearchFiled(querySentence, numberOfResult, pageSize)
			currentSearchResult, err := p.query(queryFiled, results)
			if err != nil { // todo need refactor error handle
				gologger.Error().
					Label("Provider").
					Msgf("Quake search error: %s\n", err)
				gologger.Info().Msgf("Quake search done. You've found %d items\n", numberOfResult)
				break
			}

			if currentSearchResult == nil || len(currentSearchResult.Data) == 0 {
				gologger.Info().Msgf("Quake search done. You've found %d items\n", numberOfResult)
				break
			}

			numberOfResult += len(currentSearchResult.Data)

			if isOverSize(numberOfResult, query.NumberOfQuery, currentSearchResult) {
				gologger.Info().Label("Provider").Msgf("Quake search done. You've found %d items\n", numberOfResult)
				break
			}
		}
	}()

	return results, nil
}

func (p *Provider) query(queryFiled *QuakeSearchFiled, results chan *sources.Result) (*QuakeSearchResult, error) {
	header := map[string]string{
		"X-QuakeToken": quakeHeader["X-QuakeToken"],
		"Content-Type": "application/json",
	}
	resp, err := sources.DefaultClient.Post(SEARCH_URL, header, queryFiled)
	//gologger.Debug().Msgf("Quake Search Resp: %s \n", resp.String())
	if err != nil {
		gologger.Debug().Msgf("Quake Search Error: %s \n", err)
		return nil, err
	}

	quakeSearchResults := NewQuakeSearchResult()
	err = resp.Into(quakeSearchResults)
	//gologger.Debug().Msgf("Quake Search Result: %#v \n", quakeSearchResults)
	if err != nil {
		gologger.Debug().Msgf("Quake search result unmarshal error: %s \n", err)
		return nil, err
	}
	if !strings.Contains(quakeSearchResults.Message, "Successful") {
		return nil, errors.New(quakeSearchResults.Message)
	}
	for _, item := range quakeSearchResults.Data {
		searchResult := &sources.Result{}

		searchResult.IP = item.IP
		searchResult.Port = item.Port
		searchResult.Host = item.Hostname
		if len(item.Service.Http.HttpLoadUrl) == 1 {
			searchResult.URL = item.Service.Http.HttpLoadUrl[0]
		}
		searchResult.Domain = item.Domain
		if searchResult.URL == "" { // if url is empty, use domain and port to generate url
			var d string
			if searchResult.Domain != "" {
				d = searchResult.Domain
			} else if searchResult.Host != "" {
				d = searchResult.Host
			}
			if d != "" {
				searchResult.URL = fmt.Sprintf("http://%s:%d", d, searchResult.Port)
			}
		}
		// Todo more effective way to get icp info
		if len(item.Service.Http.Icp) > 0 {
			searchResult.ICPUnit = item.Service.Http.Icp["main_licence"].(map[string]interface{})["unit"].(string)
			searchResult.ICPLicence = item.Service.Http.Icp["main_licence"].(map[string]interface{})["licence"].(string)
		}

		gologger.Debug().Msgf("%#v \n", searchResult)

		results <- searchResult
	}

	return quakeSearchResults, nil
}

// If number of result is more than number of query, return true
func isOverSize(numberOfResult, numberOfQuery int, currentSearchResult *QuakeSearchResult) bool {
	var overSize = false

	if numberOfResult >= numberOfQuery && numberOfQuery != -1 {
		overSize = true
	}

	if currentSearchResult.Meta.Pagination.Count > 0 && numberOfResult > currentSearchResult.Meta.Pagination.Total {
		overSize = true
	}

	return overSize
}

func ToQuakeGrammar(s string) (string, error) {
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
		if st, err := parse2QuakeKeywords(strings.TrimSpace(q)); err != nil {
			return "", err
		} else {
			query += st + " AND "
		}
	}

	// handle the latest suffix
	if strings.HasSuffix(query, "AND ") {
		query = query[0 : len(query)-5]
	}
	return query, nil
}

func parse2QuakeKeywords(s string) (string, error) {
	var (
		notCondition bool
		query        string
		err          error
	)

	if strings.HasPrefix(strings.ToLower(s), "not ") {
		notCondition = true
		s = s[4:]
	}
	keywords := strings.Split(s, ":")
	keyword, search := keywords[0], keywords[1]
	switch keyword {
	case "ip":
		query = fmt.Sprintf("ip: %s", search)
	case "domain":
		if notCondition && search == `""` {
			query = "is_domain: true"
			notCondition = false
		} else {
			query = fmt.Sprintf("domain: %s", search)
		}
	case "header":
		query = fmt.Sprintf("headers: %s", search)
	case "favicon":
		query = fmt.Sprintf("favicon: %s", search)
	case "cert":
		query = fmt.Sprintf("cert: %s", search)
	case "title":
		query = fmt.Sprintf("title: %s", search)
	case "body":
		query = fmt.Sprintf("body: %s", search)
	default:
		query = ""
		err = errors.New("transfer to QUAKE grammar false")
	}
	if err != nil {
		return query, err
	}

	if notCondition {
		query = fmt.Sprintf("(NOT %s)", query)
	}
	return query, nil
}
