package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"slices"
)

const (
	CONFIG        string = "config.json"
	TEMPLATES_DIR string = "templates"
)

var (
	configuration  *Configuration    = new(Configuration)
	templatesCache map[string]string = make(map[string]string)
	outputDir      string            = ""
)

type Configuration struct {
	PagesToBuild []string  `json:"pagesToBuild"`
	Head         *Template `json:"head"`
	Body         *Template `json:"body"`
	BodyHeader   *Template `json:"body_header"`
	BodyFooter   *Template `json:"body_footer"`
	Pages        []*Page   `json:"pages"`
}

type Page struct {
	Id             string          `json:"id"`
	ContentRefFile string          `json:"content"`
	ContentParams  *TemplateParams `json:"contentParams"`
	HeadParams     *TemplateParams `json:"headParams"`
	HeaderParams   *TemplateParams `json:"headerParams"`
	FooterParams   *TemplateParams `json:"footerParams"`
}

type Template struct {
	RefFile string          `json:"template"`
	Params  *TemplateParams `json:"templateParams"`
}

type TemplateParams struct {
	Data []*TemplateParam
}

type TemplateParam struct {
	Key   string
	Value string
}

func (tp *TemplateParams) UnmarshalJSON(buf []byte) error {
	tp.Data = make([]*TemplateParam, 0)
	str := strings.Trim(strings.TrimSpace(string(buf)), "{}\r\n")
	if str == "" {
		return nil
	}
	lines := strings.Split(str, ",\r\n")
	for _, line := range lines {
		pair := strings.SplitN(line, ":", 2)
		key := strings.Trim(strings.TrimSpace(pair[0]), `"`+"\r\n")
		value := strings.Trim(strings.TrimSpace(pair[1]), `"`+"\r\n")

		// fmt.Printf("%s = %s\n", key, value)
		tp.Data = append(tp.Data, &TemplateParam{
			Key:   key,
			Value: value,
		})
	}

	return nil
}

func main() {
	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}
	outputDir = filepath.Dir(filepath.Dir(ex))
	if err != nil {
		panic(err)
	}
	fmt.Printf("Building to %s\n", outputDir)

	readConfig(CONFIG)

	for _, pageId := range configuration.PagesToBuild {
		idx := slices.IndexFunc(configuration.Pages, func(el *Page) bool {
			return el.Id == pageId
		})
		if idx < 0 {
			fmt.Printf("Page {%s} not found in config! Failed to build.\n", pageId)
			continue
		}

		buildPage(configuration.Pages[idx])
	}
}

func readConfig(filename string) {
	// filename = filepath.Join(configuration.ExecDirectory, filename)
	file, _ := os.Open(filename)
	defer file.Close()
	decoder := json.NewDecoder(file)
	err := decoder.Decode(&configuration)
	if err != nil {
		panic(err)
	}
}

func buildPage(page *Page) {
	outputName := page.Id + ".html"
	head := buildHead(page.HeadParams.Data)
	content := compileTemplate(
		page.ContentRefFile,
		page.ContentParams.Data,
		nil,
	)

	var headerParams []*TemplateParam
	if page.HeaderParams != nil {
		headerParams = page.HeaderParams.Data
	}
	body_header := buildHeader(headerParams)

	var footerParams []*TemplateParam
	if page.HeaderParams != nil {
		footerParams = page.FooterParams.Data
	}
	body_footer := buildFooter(footerParams)
	body := compileTemplate(
		configuration.Body.RefFile,
		nil,
		[]*TemplateParam{
			{
				Key:   "$body_header",
				Value: body_header,
			},
			{
				Key:   "$body_footer",
				Value: body_footer,
			},
			{
				Key:   "$body_content",
				Value: content,
			},
		},
	)
	fullPage := fmt.Sprintf(`<!DOCTYPE html><html>%s%s</html>`, head, body)

	exportHTML(outputName, fullPage)
}

func buildHead(params []*TemplateParam) string {
	content := compileTemplate(
		configuration.Head.RefFile,
		params,
		configuration.Head.Params.Data,
	)
	return content
}

func buildHeader(params []*TemplateParam) string {
	content := compileTemplate(
		configuration.BodyHeader.RefFile,
		params,
		configuration.BodyHeader.Params.Data,
	)
	return content
}

func buildFooter(params []*TemplateParam) string {
	year, _, _ := time.Now().Date()
	params = append(params, &TemplateParam{Key: "$year", Value: strconv.Itoa(year)})
	content := compileTemplate(
		configuration.BodyFooter.RefFile,
		params,
		configuration.BodyFooter.Params.Data,
	)
	return content
}

func compileTemplate(templateFile string, params, defaults []*TemplateParam) string {
	content := readTemplateFile(templateFile)
	paramsMerged := mergeParams(params, defaults)
	for _, par := range paramsMerged {
		content = strings.ReplaceAll(content, par.Key, par.Value)
	}
	return content
}

func readTemplateFile(filename string) string {
	// filename = filepath.Join(configuration.ExecDirectory, filename)
	if v, ok := templatesCache[filename]; ok {
		return v
	}

	file, _ := os.Open(filepath.Join(TEMPLATES_DIR, filename))
	defer file.Close()
	content, err := io.ReadAll(file)
	if err != nil {
		panic(err)
	}
	templatesCache[filename] = string(content)

	return templatesCache[filename]
}

func mergeParams(params, defaults []*TemplateParam) []*TemplateParam {
	if len(params) == 0 {
		return defaults
	}
	if len(defaults) == 0 {
		return params
	}
	paramsMap := make(map[string]*TemplateParam)
	// -- Set defaults
	for _, par := range defaults {
		paramsMap[par.Key] = par
	}
	// -- Overwrite with params
	for _, par := range params {
		paramsMap[par.Key] = par
	}

	merged := make([]*TemplateParam, 0, len(paramsMap))
	for _, v := range paramsMap {
		merged = append(merged, v)
	}
	return merged
}

func exportHTML(filename, content string) {
	file, err := os.Create(filepath.Join(outputDir, filename))
	if err != nil {
		panic(err)
	}
	defer file.Close()
	if _, err := file.WriteString(content); err != nil {
		panic(err)
	}
}
