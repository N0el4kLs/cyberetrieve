package hunter

import (
	"errors"
	"fmt"
	"strings"

	"github.com/N0el4kLs/cyberetrieve/sources"
	"github.com/projectdiscovery/gologger"
)

const (
	HUNTER   = "HUNTER"
	AUTH_URL = "https://hunter.qianxin.com/openApi/search?api-key="
)

var (
	SEARCH_URL = ""
)

type Provider struct {
}

// Name returns the name of the provider
func (p *Provider) Name() string {
	return HUNTER
}

// Auth checks if the provider is valid to use
func (p *Provider) Auth(s *sources.Session) bool {
	client := sources.DefaultClient
	if s.HunterKey == "" {
		return false
	}
	resp, err := client.Get(AUTH_URL+s.HunterKey, nil)
	if err != nil {
		return false
	}
	if !strings.Contains(resp.String(), `搜索内容不能为空`) {
		return false
	}
	SEARCH_URL = AUTH_URL + s.HunterKey
	return true
}

// Search the result with provider
func (p *Provider) Search(query *sources.Query) (chan *sources.Result, error) {
	results := make(chan *sources.Result)

	go func() {
		defer close(results)
		numberOfResult := 0
		pageSize := sources.DEFAULT_PAGE_SIZE
		if query.NumberOfQuery < pageSize {
			pageSize = query.NumberOfQuery
		}
		if query.NumberOfQuery > sources.DEFAULT_PAGE_SIZE_MAX || query.NumberOfQuery == -1 {
			pageSize = sources.DEFAULT_PAGE_SIZE_MAX / 2
		}

		if query.NumberOfQuery < 10 && query.NumberOfQuery != -1 {
			gologger.Warning().Label("Provider").
				Msgf("%s query number can't below 10, set query number to 10\n", p.Name())
			pageSize = 10
		}
		pageNumber := 1
		for {
			querySentence := query.Query

			//If AutoGrammar is on, use transferred grammar
			if query.HunterQuery != "" {
				querySentence = query.HunterQuery
			}
			gologger.Info().Msgf("Provider %s search grammar: %s \n", p.Name(), querySentence)
			queryFiled := NewHunterSearchFiled(querySentence, pageNumber, pageSize)
			pageNumber++
			currentSearchResult, err := p.query(queryFiled, results)
			if err != nil {
				gologger.Error().
					Label("Provider").
					Msgf("%s search error: %s\n", p.Name(), err)
				gologger.Info().Msgf("%s search done. You've found %d items\n", p.Name(), numberOfResult)
				break
			}

			if currentSearchResult == nil || len(currentSearchResult.Data.Arr) == 0 {
				gologger.Info().Msgf("%s search done. You've found %d items\n", p.Name(), numberOfResult)
				break
			}

			numberOfResult += len(currentSearchResult.Data.Arr)
			if numberOfResult >= query.NumberOfQuery && query.NumberOfQuery != -1 {
				gologger.Info().Label("Provider").Msgf("%s search done. You've found %d items\n", p.Name(), numberOfResult)
				break
			}
		}
	}()

	return results, nil
}

func (p *Provider) query(queryFiled HunterSearchFiled, results chan *sources.Result) (*HunterSearchResult, error) {
	searchUrl := fmt.Sprintf("%s%s", SEARCH_URL, hunterSearchTrans(queryFiled))
	resp, err := sources.DefaultClient.Get(searchUrl, nil)
	if err != nil || resp.StatusCode != 200 {
		gologger.Debug().Msgf("%s Search Error: %s \n", p.Name(), err)
		return nil, err
	}

	hunterSearchResult := &HunterSearchResult{}
	err = resp.Into(hunterSearchResult)
	if err != nil {
		gologger.Debug().Msgf("%s search result unmarshal error: %s \n", p.Name(), err)
		return nil, err
	}

	for _, item := range hunterSearchResult.Data.Arr {
		searchResult := &sources.Result{}
		searchResult.IP = item.IP
		searchResult.Port = item.Port
		searchResult.URL = item.URL
		searchResult.Domain = item.Domain
		searchResult.ICPUnit = item.Company
		searchResult.ICPLicence = item.Number

		results <- searchResult
	}

	return hunterSearchResult, nil
}

func ToHunterGrammer(s string) (string, error) {
	keywords := strings.Split(s, ":")
	keyword, search := keywords[0], keywords[1]
	switch keyword {
	case "ip":
		return fmt.Sprintf(`ip="%s"`, search), nil
	case "domain":
		return fmt.Sprintf(`domain="%s"`, search), nil
	case "header":
		return fmt.Sprintf(`header="%s"`, search), nil
	case "favicon":
		return fmt.Sprintf(`web.icon="%s"`, search), nil
	case "cert":
		return fmt.Sprintf(`cert="%s"`, search), nil
	default:
		return "", errors.New("transfer to hunter grammar false")
	}
}
