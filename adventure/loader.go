package main

import (
	"encoding/json"
	"io/ioutil"
	"os"

	novelai_api "github.com/wbrown/novelai-research-tool/novelai-api"
	"github.com/wbrown/novelai-research-tool/utils"
)

func InitAdventureModels() {
	wd, err := os.Getwd()
	utils.HandleError(err)

	modelsDir := wd + "\\models"
	fi_models, err := ioutil.ReadDir(modelsDir)
	utils.HandleError(err)

	models = make(map[string]*AdventureModel, 0)

	// iterate through available models: models/...
	for _, fi_model := range fi_models {
		if !fi_model.IsDir() {
			continue
		}

		path_model := modelsDir + "\\" + fi_model.Name()
		path_configs := path_model + "\\configs"
		path_scenarios := path_model + "\\scenarios"

		// read info.json: models/[model]/info.json
		var model_json AdventureModelJson

		model_json_bin, err := os.ReadFile(path_model + "\\info.json")
		utils.HandleError(err)

		err = json.Unmarshal(model_json_bin, &model_json)
		utils.HandleError(err)

		// create adventure model
		model := AdventureModel{
			Model:            model_json.Model,
			Name:             model_json.Name,
			Encoder:          utils.GetEncoderByVocabId(model_json.VocabId),
			ActionTokens:     model_json.ActionTokens,
			OpusContextLimit: *model_json.OpusContextLimit,
			Configs:          make(map[string]*novelai_api.NaiGenerateParams),
			Scenarios:        make(map[string]*AdventureScenario),
		}

		// read defined configs: models/[model]/configs/...
		fi_configs, err := ioutil.ReadDir(path_configs)
		utils.HandleError(err)
		for _, fi_config := range fi_configs {
			path := path_configs + "\\" + fi_config.Name()

			config_json_bin, err := os.ReadFile(path)
			utils.HandleError(err)

			var config_json novelai_api.NaiGenerateParams
			err = json.Unmarshal(config_json_bin, &config_json)
			utils.HandleError(err)

			model.Configs[fi_config.Name()] = &config_json
		}

		// read defined scenarios: models/[model]/scenarios/...
		fi_scenarios, err := ioutil.ReadDir(path_scenarios)
		utils.HandleError(err)
		for _, fi_scenario := range fi_scenarios {
			path := path_scenarios + "\\" + fi_scenario.Name()

			scenario_json_bin, err := os.ReadFile(path)
			utils.HandleError(err)

			var scenario_json AdventureScenarioJson
			err = json.Unmarshal(scenario_json_bin, &scenario_json)
			utils.HandleError(err)

			// scenario actions are tokenized
			scenario_actions := make([]AdventureScenarioAction, 0)
			for _, scenario_action_json := range scenario_json.Actions {
				scenario_action := AdventureScenarioAction{
					Input:  model.Encoder.Encode(scenario_action_json.Input),
					Output: model.Encoder.Encode(scenario_action_json.Output),
				}
				scenario_actions = append(scenario_actions, scenario_action)
			}

			scenario := AdventureScenario{
				Id:           scenario_json.Id,
				PrefixTokens: model.Encoder.Encode(scenario_json.Prefix),
				Actions:      &scenario_actions,
			}

			if _, ok := model.Scenarios[*scenario.Id]; ok {
				// TODO: show paths for prev and current for easy resolve!
				panic("scenario duplicate with id: ")
			}
			model.Scenarios[*scenario.Id] = &scenario
		}
		models[*model.Name] = &model
	}
}
