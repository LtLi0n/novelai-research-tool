package main

import (
	linkedList "container/list"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"unicode"

	"github.com/chzyer/readline"
	"github.com/fatih/color"
	"github.com/jdkato/prose/v2"
	"github.com/wbrown/gpt_bpe"
	novelai_api "github.com/wbrown/novelai-research-tool/novelai-api"
	"github.com/wbrown/novelai-research-tool/utils"
)

type Handler struct {
	Adventure *Adventure
	Model     *AdventureModel
	Config    *novelai_api.NaiGenerateParams
	API       novelai_api.NovelAiAPI
	MaxTokens uint16
}

type Adventure struct {
	LastContext         string
	ContextPrefix       string
	contextPrefixTokens *gpt_bpe.Tokens
	ContextHistorySteps *linkedList.List
	Scenario            *AdventureScenario
}

type ContextHistoryStep struct {
	Request  *gpt_bpe.Tokens
	Response *gpt_bpe.Tokens
}

var models map[string]*AdventureModel

type AdventureModel struct {
	Model            *string
	Name             *string
	Encoder          *gpt_bpe.GPTEncoder
	OpusContextLimit uint16
	ActionTokens     *gpt_bpe.Tokens
	Configs          map[string]*novelai_api.NaiGenerateParams
	Scenarios        map[string]*AdventureScenario
}
type AdventureModelJson struct {
	Model            *string         `json:"model"`
	Name             *string         `json:"name"`
	VocabId          *string         `json:"vocabId"`
	ActionTokens     *gpt_bpe.Tokens `json:"actionTokens"`
	OpusContextLimit *uint16         `json:"opusContextLimit"`
}

type AdventureScenario struct {
	Id           *string
	PrefixTokens *gpt_bpe.Tokens
	Actions      *[]AdventureScenarioAction
}
type AdventureScenarioAction struct {
	Input  *gpt_bpe.Tokens
	Output *gpt_bpe.Tokens
}
type AdventureScenarioJson struct {
	Id      *string `json:"id"`
	Prefix  *string `json:"prefix"`
	Actions []struct {
		Input  *string `json:"input"`
		Output *string `json:"output"`
	} `json:"actions"`
}

// Builds token context.
func (handler *Handler) BuildContextTokens() *gpt_bpe.Tokens {
	adventure := handler.Adventure
	model := handler.Model

	tokens := gpt_bpe.Tokens{}
	tokens = append(tokens, *adventure.Scenario.PrefixTokens...)
	total_n := len(*adventure.Scenario.PrefixTokens)

	// find total token count for ongoing context
	for e := adventure.ContextHistorySteps.Front(); e != nil; e = e.Next() {
		historyStep := e.Value.(*ContextHistoryStep)

		total_n += len(*model.ActionTokens)
		total_n += len(*historyStep.Request)
		if historyStep.Response != nil {
			total_n += len(*historyStep.Response)
		}
	}

	// remove context steps from the start, until it can fit
	for total_n > int(model.OpusContextLimit) {
		e := adventure.ContextHistorySteps.Front()
		historyStep := e.Value.(*ContextHistoryStep)

		total_n -= len(*model.ActionTokens)
		total_n -= len(*historyStep.Request)
		total_n -= len(*historyStep.Response)

		adventure.ContextHistorySteps.Remove(e)
	}

	// build context
	for e := adventure.ContextHistorySteps.Front(); e != nil; e = e.Next() {
		historyStep := e.Value.(*ContextHistoryStep)

		tokens = append(tokens, *model.ActionTokens...)
		tokens = append(tokens, *historyStep.Request...)
		if historyStep.Response != nil {
			tokens = append(tokens, *historyStep.Response...)
		}
	}

	return &tokens
}

func (handler *Handler) BuildContextStr() *string {
	tokens := handler.BuildContextTokens()
	str := handler.Model.Encoder.Decode(tokens)
	return &str
}

func selectInteractiveNaiModel() string {
	models := []string{"6B-v4", "clio-v1"}

	rl, err := readline.New("> ")
	if err != nil {
		panic(err)
	}
	defer rl.Close()
	rl.ResetHistory()

	for {
		fmt.Println("Select model:")
		for i, val := range models {
			fmt.Println(fmt.Sprint(i+1), "-", val)
		}
		fmt.Println("q - to quit")
		selectedIdxStr, err := rl.Readline()
		if err != nil {
			utils.TryScreenClear()
			fmt.Print("Bad selection.\n\n")
			continue
		}
		if selectedIdxStr == "q" {
			os.Exit(0)
		}
		selectedIdxRaw, err := strconv.ParseInt(selectedIdxStr, 10, 32)
		if err != nil {
			utils.TryScreenClear()
			fmt.Print("Bad selection.\n\n")
			continue
		}
		selectedIdx := int(selectedIdxRaw)

		selectedIdx--

		if selectedIdx >= 0 && selectedIdx < len(models) {
			utils.TryScreenClear()
			return models[selectedIdx]
		} else {
			fmt.Println("Bad selection.")
		}
	}
}

func (model *AdventureModel) NewHandler(config *novelai_api.NaiGenerateParams, scenario *AdventureScenario) (handler *Handler) {
	handler = &Handler{
		API:       novelai_api.NewNovelAiAPI(),
		Model:     model,
		Config:    config,
		MaxTokens: model.OpusContextLimit - uint16(*config.MaxLength),
		Adventure: &Adventure{
			ContextHistorySteps: linkedList.New().Init(),
			Scenario:            scenario,
		},
	}

	// add scenario tokens to context
	for _, scenarioAction := range *scenario.Actions {
		handler.Adventure.ContextHistorySteps.PushBack(
			&ContextHistoryStep{
				Request:  scenarioAction.Input,
				Response: scenarioAction.Output,
			})
	}

	return handler
}

func writeAllText(path string, content *string) {
	f, err := os.Create(path)
	if err != nil {
		log.Fatal(err)
	}

	_, err = f.WriteString(*content)
	if err != nil {
		log.Fatal(err)
	}
	f.Sync()
}

func (handler *Handler) start() {
	defer color.Set(color.Reset)
	adventure := handler.Adventure
	model := handler.Model

	utils.PrintLn(*handler.BuildContextStr(), color.FgYellow)
	rl, err := readline.New("> You ")
	if err != nil {
		panic(err)
	}
	defer rl.Close()

	var output string
	log.SetOutput(rl.Stderr())
	for {
		color.Set(color.BgHiWhite, color.FgBlack, color.BlinkSlow)
		rl.ResetHistory()
		line, err := rl.Readline()
		color.Set(color.Reset)
		if err != nil { // io.EOF
			break
		}
		if line == "$ctx" {
			utils.PrintLn("\n"+adventure.LastContext+"\n", color.FgCyan)
			writeAllText("context.txt", &adventure.LastContext)
			continue
		}
		if unicode.IsUpper(rune(line[0])) {
			line = string(unicode.ToLower(rune(line[0]))) + line[1:]
		}
		line = " You " + line + "\n"
		inputTokens := model.Encoder.Encode(&line)
		adventure.ContextHistorySteps.PushBack(&ContextHistoryStep{Request: inputTokens})
		contextTokens := handler.BuildContextTokens()
		adventure.LastContext = model.Encoder.Decode(contextTokens)

		resp := handler.API.GenerateWithParams(&adventure.LastContext, *handler.Config)
		output = resp.Response
		doc, err := prose.NewDocument(output)
		if err != nil {
			log.Fatal(err)
		}
		processed := make([]string, 0)

		for _, sent := range doc.Sentences() {
			lastCharRune := []rune(sent.Text[len(sent.Text)-1:])[0]
			quoteEnding := lastCharRune == '\'' || lastCharRune == '"'
			quote_n := 0

			// if last rune was quote based, make sure output ending doesn't contain an unfinished quote.
			if quoteEnding {
				for _, c := range sent.Text {
					if c == lastCharRune {
						quote_n++
					}
				}
				if quote_n%2 != 0 {
					continue
				}
			} else if lastCharRune != '.' && lastCharRune != '!' && lastCharRune != '?' {
				continue
			}

			processed = append(processed, sent.Text)
		}

		output = strings.Join(processed, " ") + "\n"
		adventure.LastContext += output
		outputTokens := model.Encoder.Encode(&output)
		historyStep := adventure.ContextHistorySteps.Back().Value.(*ContextHistoryStep)
		historyStep.Response = outputTokens

		utils.PrintLn(output, color.FgWhite)
	}
}

func main() {
	InitAdventureModels()
	model := models["Clio"]
	handler := model.NewHandler(model.Configs["default.json"], model.Scenarios["default"])
	handler.start()
}
