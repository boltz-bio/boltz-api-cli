// File generated from our OpenAPI spec by Stainless. See CONTRIBUTING.md for details.

package cmd

import (
	"context"
	"fmt"
	"github.com/boltz-bio/boltz-api-go"

	"github.com/boltz-bio/boltz-api-cli/internal/apiquery"
	"github.com/boltz-bio/boltz-api-cli/internal/requestflag"
	"github.com/boltz-bio/boltz-api-go/option"
	"github.com/tidwall/gjson"
	"github.com/urfave/cli/v3"
)

var predictionsStructureAndBindingRetrieve = cli.Command{
	Name:    "retrieve",
	Usage:   "Retrieve a prediction by ID, including its status and results.",
	Suggest: true,
	Flags: []cli.Flag{
		&requestflag.Flag[string]{
			Name:      "id",
			Required:  true,
			PathParam: "id",
		},
		&requestflag.Flag[string]{
			Name:      "workspace-id",
			Usage:     "Workspace ID. Only used with admin API keys. Ignored (or validated) for workspace-scoped keys.",
			QueryPath: "workspace_id",
		},
	},
	Action:          handlePredictionsStructureAndBindingRetrieve,
	HideHelpCommand: true,
}

var predictionsStructureAndBindingList = cli.Command{
	Name:    "list",
	Usage:   "List structure and binding predictions, optionally filtered by workspace",
	Suggest: true,
	Flags: []cli.Flag{
		&requestflag.Flag[string]{
			Name:      "after-id",
			Usage:     "Return results after this ID",
			QueryPath: "after_id",
		},
		&requestflag.Flag[string]{
			Name:      "before-id",
			Usage:     "Return results before this ID",
			QueryPath: "before_id",
		},
		&requestflag.Flag[int64]{
			Name:      "limit",
			Usage:     "Max items to return. Defaults to 100.",
			Default:   100,
			QueryPath: "limit",
		},
		&requestflag.Flag[string]{
			Name:      "workspace-id",
			Usage:     "Filter by workspace ID. Only used with admin API keys. If not provided, defaults to the workspace associated with the API key, or the default workspace for admin keys.",
			QueryPath: "workspace_id",
		},
		&requestflag.Flag[int64]{
			Name:  "max-items",
			Usage: "The maximum number of items to return (use -1 for unlimited).",
		},
	},
	Action:          handlePredictionsStructureAndBindingList,
	HideHelpCommand: true,
}

var predictionsStructureAndBindingDeleteData = cli.Command{
	Name:    "delete-data",
	Usage:   "Permanently delete the input, output, and result data associated with this\nprediction. The prediction record itself is retained with a `data_deleted_at`\ntimestamp. This action is irreversible.",
	Suggest: true,
	Flags: []cli.Flag{
		&requestflag.Flag[string]{
			Name:      "id",
			Required:  true,
			PathParam: "id",
		},
	},
	Action:          handlePredictionsStructureAndBindingDeleteData,
	HideHelpCommand: true,
}

var predictionsStructureAndBindingEstimateCost = requestflag.WithInnerFlags(cli.Command{
	Name:    "estimate-cost",
	Usage:   "Estimate the cost of a prediction without creating any resource or consuming\nGPU.",
	Suggest: true,
	Flags: []cli.Flag{
		&requestflag.Flag[map[string]any]{
			Name:     "input",
			Required: true,
			BodyPath: "input",
		},
		&requestflag.Flag[string]{
			Name:     "model",
			Usage:    "Model to use for prediction",
			Default:  "boltz-2.1",
			Const:    true,
			BodyPath: "model",
		},
		&requestflag.Flag[string]{
			Name:     "idempotency-key",
			Usage:    "Client-provided key to prevent duplicate submissions on retries",
			BodyPath: "idempotency_key",
		},
		&requestflag.Flag[string]{
			Name:     "workspace-id",
			Usage:    "Target workspace ID (admin keys only; ignored for workspace keys)",
			BodyPath: "workspace_id",
		},
	},
	Action:          handlePredictionsStructureAndBindingEstimateCost,
	HideHelpCommand: true,
}, map[string][]requestflag.HasOuterFlag{
	"input": {
		&requestflag.InnerFlag[[]map[string]any]{
			Name:       "input.entities",
			Usage:      "Entities (proteins, RNA, DNA, ligands) forming the complex to predict. Order determines chain assignment.",
			InnerField: "entities",
		},
		&requestflag.InnerFlag[map[string]any]{
			Name:       "input.binding",
			InnerField: "binding",
		},
		&requestflag.InnerFlag[[]map[string]any]{
			Name:       "input.bonds",
			Usage:      "Bond constraints between atoms. Atom-level ligand references currently support ligand_ccd only; ligand_smiles is unsupported.",
			InnerField: "bonds",
		},
		&requestflag.InnerFlag[[]map[string]any]{
			Name:       "input.constraints",
			Usage:      "Structural constraints (pocket and contact). Atom-level ligand references currently support ligand_ccd only; ligand_smiles is unsupported.",
			InnerField: "constraints",
		},
		&requestflag.InnerFlag[map[string]any]{
			Name:       "input.model-options",
			InnerField: "model_options",
		},
		&requestflag.InnerFlag[int64]{
			Name:       "input.num-samples",
			Usage:      "Number of structure samples to generate",
			InnerField: "num_samples",
		},
		&requestflag.InnerFlag[[]map[string]any]{
			Name:       "input.templates",
			Usage:      "Template structure files to guide protein-chain prediction. Supports up to 4 CIF or PDB templates from HTTPS URLs or base64 uploads. Use template_chains to map request chains to template-file chains.",
			InnerField: "templates",
		},
	},
})

var predictionsStructureAndBindingStart = requestflag.WithInnerFlags(cli.Command{
	Name:    "start",
	Usage:   "Submit a prediction job that produces 3D structure coordinates and confidence\nscores for the input molecular complex, with optional binding metrics. Protein\nentities can use automatic MSA generation, custom A3M/CSV MSAs, or empty MSA\nmode. Boltz-2.1 predictions can also include up to 4 CIF or PDB templates to\nguide protein-chain geometry.",
	Suggest: true,
	Flags: []cli.Flag{
		&requestflag.Flag[map[string]any]{
			Name:     "input",
			Required: true,
			BodyPath: "input",
		},
		&requestflag.Flag[string]{
			Name:     "model",
			Usage:    "Model to use for prediction",
			Default:  "boltz-2.1",
			Const:    true,
			BodyPath: "model",
		},
		&requestflag.Flag[string]{
			Name:     "idempotency-key",
			Usage:    "Client-provided key to prevent duplicate submissions on retries",
			BodyPath: "idempotency_key",
		},
		&requestflag.Flag[string]{
			Name:     "workspace-id",
			Usage:    "Target workspace ID (admin keys only; ignored for workspace keys)",
			BodyPath: "workspace_id",
		},
	},
	Action:          handlePredictionsStructureAndBindingStart,
	HideHelpCommand: true,
}, map[string][]requestflag.HasOuterFlag{
	"input": {
		&requestflag.InnerFlag[[]map[string]any]{
			Name:       "input.entities",
			Usage:      "Entities (proteins, RNA, DNA, ligands) forming the complex to predict. Order determines chain assignment.",
			InnerField: "entities",
		},
		&requestflag.InnerFlag[map[string]any]{
			Name:       "input.binding",
			InnerField: "binding",
		},
		&requestflag.InnerFlag[[]map[string]any]{
			Name:       "input.bonds",
			Usage:      "Bond constraints between atoms. Atom-level ligand references currently support ligand_ccd only; ligand_smiles is unsupported.",
			InnerField: "bonds",
		},
		&requestflag.InnerFlag[[]map[string]any]{
			Name:       "input.constraints",
			Usage:      "Structural constraints (pocket and contact). Atom-level ligand references currently support ligand_ccd only; ligand_smiles is unsupported.",
			InnerField: "constraints",
		},
		&requestflag.InnerFlag[map[string]any]{
			Name:       "input.model-options",
			InnerField: "model_options",
		},
		&requestflag.InnerFlag[int64]{
			Name:       "input.num-samples",
			Usage:      "Number of structure samples to generate",
			InnerField: "num_samples",
		},
		&requestflag.InnerFlag[[]map[string]any]{
			Name:       "input.templates",
			Usage:      "Template structure files to guide protein-chain prediction. Supports up to 4 CIF or PDB templates from HTTPS URLs or base64 uploads. Use template_chains to map request chains to template-file chains.",
			InnerField: "templates",
		},
	},
})

var predictionsStructureAndBindingTokenCount = requestflag.WithInnerFlags(cli.Command{
	Name:    "token-count",
	Usage:   "Returns the total token count for a prospective Boltz2 complex along with the\nmaximum token count the caller's compute config will accept. Pure introspection\n— never rejects when the complex is too large; still rejects on malformed inputs\nthe model could not interpret. Use the sandbox prediction endpoint if you need\nthe cap enforced.",
	Suggest: true,
	Flags: []cli.Flag{
		&requestflag.Flag[map[string]any]{
			Name:     "input",
			Required: true,
			BodyPath: "input",
		},
		&requestflag.Flag[string]{
			Name:     "model",
			Usage:    "Model to use for prediction",
			Default:  "boltz-2.1",
			Const:    true,
			BodyPath: "model",
		},
		&requestflag.Flag[string]{
			Name:     "idempotency-key",
			Usage:    "Client-provided key to prevent duplicate submissions on retries",
			BodyPath: "idempotency_key",
		},
		&requestflag.Flag[string]{
			Name:     "workspace-id",
			Usage:    "Target workspace ID (admin keys only; ignored for workspace keys)",
			BodyPath: "workspace_id",
		},
	},
	Action:          handlePredictionsStructureAndBindingTokenCount,
	HideHelpCommand: true,
}, map[string][]requestflag.HasOuterFlag{
	"input": {
		&requestflag.InnerFlag[[]map[string]any]{
			Name:       "input.entities",
			Usage:      "Entities (proteins, RNA, DNA, ligands) forming the complex to predict. Order determines chain assignment.",
			InnerField: "entities",
		},
		&requestflag.InnerFlag[map[string]any]{
			Name:       "input.binding",
			InnerField: "binding",
		},
		&requestflag.InnerFlag[[]map[string]any]{
			Name:       "input.bonds",
			Usage:      "Bond constraints between atoms. Atom-level ligand references currently support ligand_ccd only; ligand_smiles is unsupported.",
			InnerField: "bonds",
		},
		&requestflag.InnerFlag[[]map[string]any]{
			Name:       "input.constraints",
			Usage:      "Structural constraints (pocket and contact). Atom-level ligand references currently support ligand_ccd only; ligand_smiles is unsupported.",
			InnerField: "constraints",
		},
		&requestflag.InnerFlag[map[string]any]{
			Name:       "input.model-options",
			InnerField: "model_options",
		},
		&requestflag.InnerFlag[int64]{
			Name:       "input.num-samples",
			Usage:      "Number of structure samples to generate",
			InnerField: "num_samples",
		},
		&requestflag.InnerFlag[[]map[string]any]{
			Name:       "input.templates",
			Usage:      "Template structure files to guide protein-chain prediction. Supports up to 4 CIF or PDB templates from HTTPS URLs or base64 uploads. Use template_chains to map request chains to template-file chains.",
			InnerField: "templates",
		},
	},
})

func handlePredictionsStructureAndBindingRetrieve(ctx context.Context, cmd *cli.Command) error {
	client := boltzapi.NewClient(getDefaultRequestOptions(cmd)...)
	unusedArgs := cmd.Args().Slice()
	if !cmd.IsSet("id") && len(unusedArgs) > 0 {
		cmd.Set("id", unusedArgs[0])
		unusedArgs = unusedArgs[1:]
	}
	if len(unusedArgs) > 0 {
		return fmt.Errorf("Unexpected extra arguments: %v", unusedArgs)
	}

	options, err := flagOptions(
		cmd,
		apiquery.NestedQueryFormatBrackets,
		apiquery.ArrayQueryFormatRepeat,
		EmptyBody,
		false,
	)
	if err != nil {
		return err
	}

	params := boltzapi.PredictionStructureAndBindingGetParams{}

	var res []byte
	options = append(options, option.WithResponseBodyInto(&res))
	_, err = client.Predictions.StructureAndBinding.Get(
		ctx,
		cmd.Value("id").(string),
		params,
		options...,
	)
	if err != nil {
		return err
	}

	obj := gjson.ParseBytes(res)
	format := cmd.Root().String("format")
	explicitFormat := cmd.Root().IsSet("format")
	transform := cmd.Root().String("transform")
	return ShowJSON(obj, ShowJSONOpts{
		ExplicitFormat: explicitFormat,
		Format:         format,
		RawOutput:      cmd.Root().Bool("raw-output"),
		Title:          "predictions:structure-and-binding retrieve",
		Transform:      transform,
	})
}

func handlePredictionsStructureAndBindingList(ctx context.Context, cmd *cli.Command) error {
	client := boltzapi.NewClient(getDefaultRequestOptions(cmd)...)
	unusedArgs := cmd.Args().Slice()

	if len(unusedArgs) > 0 {
		return fmt.Errorf("Unexpected extra arguments: %v", unusedArgs)
	}

	options, err := flagOptions(
		cmd,
		apiquery.NestedQueryFormatBrackets,
		apiquery.ArrayQueryFormatRepeat,
		EmptyBody,
		false,
	)
	if err != nil {
		return err
	}

	params := boltzapi.PredictionStructureAndBindingListParams{}

	format := cmd.Root().String("format")
	explicitFormat := cmd.Root().IsSet("format")
	transform := cmd.Root().String("transform")
	if format == "raw" {
		var res []byte
		options = append(options, option.WithResponseBodyInto(&res))
		_, err = client.Predictions.StructureAndBinding.List(ctx, params, options...)
		if err != nil {
			return err
		}
		obj := gjson.ParseBytes(res)
		return ShowJSON(obj, ShowJSONOpts{
			ExplicitFormat: explicitFormat,
			Format:         format,
			RawOutput:      cmd.Root().Bool("raw-output"),
			Title:          "predictions:structure-and-binding list",
			Transform:      transform,
		})
	} else {
		iter := client.Predictions.StructureAndBinding.ListAutoPaging(ctx, params, options...)
		maxItems := int64(-1)
		if cmd.IsSet("max-items") {
			maxItems = cmd.Value("max-items").(int64)
		}
		return ShowJSONIterator(iter, maxItems, ShowJSONOpts{
			ExplicitFormat: explicitFormat,
			Format:         format,
			RawOutput:      cmd.Root().Bool("raw-output"),
			Title:          "predictions:structure-and-binding list",
			Transform:      transform,
		})
	}
}

func handlePredictionsStructureAndBindingDeleteData(ctx context.Context, cmd *cli.Command) error {
	client := boltzapi.NewClient(getDefaultRequestOptions(cmd)...)
	unusedArgs := cmd.Args().Slice()
	if !cmd.IsSet("id") && len(unusedArgs) > 0 {
		cmd.Set("id", unusedArgs[0])
		unusedArgs = unusedArgs[1:]
	}
	if len(unusedArgs) > 0 {
		return fmt.Errorf("Unexpected extra arguments: %v", unusedArgs)
	}

	options, err := flagOptions(
		cmd,
		apiquery.NestedQueryFormatBrackets,
		apiquery.ArrayQueryFormatRepeat,
		EmptyBody,
		false,
	)
	if err != nil {
		return err
	}

	var res []byte
	options = append(options, option.WithResponseBodyInto(&res))
	_, err = client.Predictions.StructureAndBinding.DeleteData(ctx, cmd.Value("id").(string), options...)
	if err != nil {
		return err
	}

	obj := gjson.ParseBytes(res)
	format := cmd.Root().String("format")
	explicitFormat := cmd.Root().IsSet("format")
	transform := cmd.Root().String("transform")
	return ShowJSON(obj, ShowJSONOpts{
		ExplicitFormat: explicitFormat,
		Format:         format,
		RawOutput:      cmd.Root().Bool("raw-output"),
		Title:          "predictions:structure-and-binding delete-data",
		Transform:      transform,
	})
}

func handlePredictionsStructureAndBindingEstimateCost(ctx context.Context, cmd *cli.Command) error {
	client := boltzapi.NewClient(getDefaultRequestOptions(cmd)...)
	unusedArgs := cmd.Args().Slice()

	if len(unusedArgs) > 0 {
		return fmt.Errorf("Unexpected extra arguments: %v", unusedArgs)
	}

	options, err := flagOptions(
		cmd,
		apiquery.NestedQueryFormatBrackets,
		apiquery.ArrayQueryFormatRepeat,
		ApplicationJSON,
		false,
	)
	if err != nil {
		return err
	}

	params := boltzapi.PredictionStructureAndBindingEstimateCostParams{}

	var res []byte
	options = append(options, option.WithResponseBodyInto(&res))
	_, err = client.Predictions.StructureAndBinding.EstimateCost(ctx, params, options...)
	if err != nil {
		return err
	}

	obj := gjson.ParseBytes(res)
	format := cmd.Root().String("format")
	explicitFormat := cmd.Root().IsSet("format")
	transform := cmd.Root().String("transform")
	return ShowJSON(obj, ShowJSONOpts{
		ExplicitFormat: explicitFormat,
		Format:         format,
		RawOutput:      cmd.Root().Bool("raw-output"),
		Title:          "predictions:structure-and-binding estimate-cost",
		Transform:      transform,
	})
}

func handlePredictionsStructureAndBindingStart(ctx context.Context, cmd *cli.Command) error {
	client := boltzapi.NewClient(getDefaultRequestOptions(cmd)...)
	unusedArgs := cmd.Args().Slice()

	if len(unusedArgs) > 0 {
		return fmt.Errorf("Unexpected extra arguments: %v", unusedArgs)
	}

	options, err := flagOptions(
		cmd,
		apiquery.NestedQueryFormatBrackets,
		apiquery.ArrayQueryFormatRepeat,
		ApplicationJSON,
		false,
	)
	if err != nil {
		return err
	}

	params := boltzapi.PredictionStructureAndBindingStartParams{}

	var res []byte
	options = append(options, option.WithResponseBodyInto(&res))
	_, err = client.Predictions.StructureAndBinding.Start(ctx, params, options...)
	if err != nil {
		return err
	}

	obj := gjson.ParseBytes(res)
	format := cmd.Root().String("format")
	explicitFormat := cmd.Root().IsSet("format")
	transform := cmd.Root().String("transform")
	return ShowJSON(obj, ShowJSONOpts{
		ExplicitFormat: explicitFormat,
		Format:         format,
		RawOutput:      cmd.Root().Bool("raw-output"),
		Title:          "predictions:structure-and-binding start",
		Transform:      transform,
	})
}

func handlePredictionsStructureAndBindingTokenCount(ctx context.Context, cmd *cli.Command) error {
	client := boltzapi.NewClient(getDefaultRequestOptions(cmd)...)
	unusedArgs := cmd.Args().Slice()

	if len(unusedArgs) > 0 {
		return fmt.Errorf("Unexpected extra arguments: %v", unusedArgs)
	}

	options, err := flagOptions(
		cmd,
		apiquery.NestedQueryFormatBrackets,
		apiquery.ArrayQueryFormatRepeat,
		ApplicationJSON,
		false,
	)
	if err != nil {
		return err
	}

	params := boltzapi.PredictionStructureAndBindingTokenCountParams{}

	var res []byte
	options = append(options, option.WithResponseBodyInto(&res))
	_, err = client.Predictions.StructureAndBinding.TokenCount(ctx, params, options...)
	if err != nil {
		return err
	}

	obj := gjson.ParseBytes(res)
	format := cmd.Root().String("format")
	explicitFormat := cmd.Root().IsSet("format")
	transform := cmd.Root().String("transform")
	return ShowJSON(obj, ShowJSONOpts{
		ExplicitFormat: explicitFormat,
		Format:         format,
		RawOutput:      cmd.Root().Bool("raw-output"),
		Title:          "predictions:structure-and-binding token-count",
		Transform:      transform,
	})
}
