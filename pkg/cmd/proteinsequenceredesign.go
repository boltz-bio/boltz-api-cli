// File generated from our OpenAPI spec by Stainless. See CONTRIBUTING.md for details.

package cmd

import (
	"context"
	"fmt"

	"github.com/boltz-bio/boltz-api-cli/internal/apiquery"
	"github.com/boltz-bio/boltz-api-cli/internal/requestflag"
	"github.com/boltz-bio/boltz-api-go"
	"github.com/boltz-bio/boltz-api-go/option"
	"github.com/tidwall/gjson"
	"github.com/urfave/cli/v3"
)

var proteinSequenceRedesignRetrieve = cli.Command{
	Name:    "retrieve",
	Usage:   "Retrieve a sequence redesign run by ID, including progress and status",
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
	Action:          handleProteinSequenceRedesignRetrieve,
	HideHelpCommand: true,
}

var proteinSequenceRedesignList = cli.Command{
	Name:    "list",
	Usage:   "List protein sequence redesign runs, optionally filtered by workspace",
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
	Action:          handleProteinSequenceRedesignList,
	HideHelpCommand: true,
}

var proteinSequenceRedesignDeleteData = cli.Command{
	Name:    "delete-data",
	Usage:   "Permanently delete the input, output, and result data associated with this\nsequence redesign run. The sequence redesign run record itself is retained with\na `data_deleted_at` timestamp. This action is irreversible.",
	Suggest: true,
	Flags: []cli.Flag{
		&requestflag.Flag[string]{
			Name:      "id",
			Required:  true,
			PathParam: "id",
		},
	},
	Action:          handleProteinSequenceRedesignDeleteData,
	HideHelpCommand: true,
}

var proteinSequenceRedesignEstimateCost = cli.Command{
	Name:    "estimate-cost",
	Usage:   "Estimate the cost of a protein sequence redesign run without creating any\nresource or consuming GPU.",
	Suggest: true,
	Flags: []cli.Flag{
		&requestflag.Flag[[]map[string]any]{
			Name:     "entity",
			Usage:    "Every chain in the input CIF, assigned exactly once as target or binder.",
			Required: true,
			BodyPath: "entities",
		},
		&requestflag.Flag[int64]{
			Name:     "num-proteins",
			Usage:    "Number of unique filter-passing redesigned proteins to generate.",
			Required: true,
			BodyPath: "num_proteins",
		},
		&requestflag.Flag[map[string]any]{
			Name:     "structure",
			Usage:    "How to provide a CIF structure file. URLs are auto-detected; base64 uploads must use chemical/x-cif media type.",
			Required: true,
			BodyPath: "structure",
		},
		&requestflag.Flag[string]{
			Name:     "type",
			Usage:    `Allowed values: "binder".`,
			Default:  "binder",
			Const:    true,
			BodyPath: "type",
		},
		&requestflag.Flag[[]map[string]any]{
			Name:     "global-design-filter",
			Usage:    "Filters applied to every redesigned region. When omitted, cysteine is excluded. Pass [] to disable global filters.",
			BodyPath: "global_design_filters",
		},
		&requestflag.Flag[string]{
			Name:     "idempotency-key",
			BodyPath: "idempotency_key",
		},
		&requestflag.Flag[string]{
			Name:     "workspace-id",
			Usage:    "Workspace to run this redesign in.",
			BodyPath: "workspace_id",
		},
	},
	Action:          handleProteinSequenceRedesignEstimateCost,
	HideHelpCommand: true,
}

var proteinSequenceRedesignListResults = cli.Command{
	Name:    "list-results",
	Usage:   "Retrieve paginated results from a protein sequence redesign run",
	Suggest: true,
	Flags: []cli.Flag{
		&requestflag.Flag[string]{
			Name:      "id",
			Required:  true,
			PathParam: "id",
		},
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
		&requestflag.Flag[string]{
			Name:      "ids",
			Usage:     "Comma-separated list of result IDs to filter by (max 200). Only results whose ID matches one of these is returned; missing IDs are silently skipped. Composes with `limit`, `after_id`, and `before_id` — the filter is applied before pagination.",
			QueryPath: "ids",
		},
		&requestflag.Flag[int64]{
			Name:      "limit",
			Usage:     "Max results to return. Defaults to 100.",
			Default:   100,
			QueryPath: "limit",
		},
		&requestflag.Flag[string]{
			Name:      "workspace-id",
			Usage:     "Workspace ID. Only used with admin API keys. Ignored (or validated) for workspace-scoped keys.",
			QueryPath: "workspace_id",
		},
		&requestflag.Flag[int64]{
			Name:  "max-items",
			Usage: "The maximum number of items to return (use -1 for unlimited).",
		},
	},
	Action:          handleProteinSequenceRedesignListResults,
	HideHelpCommand: true,
}

var proteinSequenceRedesignResume = cli.Command{
	Name:    "resume",
	Usage:   "Resume a stopped protein sequence redesign run from its last checkpoint",
	Suggest: true,
	Flags: []cli.Flag{
		&requestflag.Flag[string]{
			Name:      "id",
			Required:  true,
			PathParam: "id",
		},
	},
	Action:          handleProteinSequenceRedesignResume,
	HideHelpCommand: true,
}

var proteinSequenceRedesignStart = cli.Command{
	Name:    "start",
	Usage:   "Create a protein sequence redesign run from selected residues in a fixed input\nstructure",
	Suggest: true,
	Flags: []cli.Flag{
		&requestflag.Flag[[]map[string]any]{
			Name:     "entity",
			Usage:    "Every chain in the input CIF, assigned exactly once as target or binder.",
			Required: true,
			BodyPath: "entities",
		},
		&requestflag.Flag[int64]{
			Name:     "num-proteins",
			Usage:    "Number of unique filter-passing redesigned proteins to generate.",
			Required: true,
			BodyPath: "num_proteins",
		},
		&requestflag.Flag[map[string]any]{
			Name:     "structure",
			Usage:    "How to provide a CIF structure file. URLs are auto-detected; base64 uploads must use chemical/x-cif media type.",
			Required: true,
			BodyPath: "structure",
		},
		&requestflag.Flag[string]{
			Name:     "type",
			Usage:    `Allowed values: "binder".`,
			Default:  "binder",
			Const:    true,
			BodyPath: "type",
		},
		&requestflag.Flag[[]map[string]any]{
			Name:     "global-design-filter",
			Usage:    "Filters applied to every redesigned region. When omitted, cysteine is excluded. Pass [] to disable global filters.",
			BodyPath: "global_design_filters",
		},
		&requestflag.Flag[string]{
			Name:     "idempotency-key",
			BodyPath: "idempotency_key",
		},
		&requestflag.Flag[string]{
			Name:     "workspace-id",
			Usage:    "Workspace to run this redesign in.",
			BodyPath: "workspace_id",
		},
	},
	Action:          handleProteinSequenceRedesignStart,
	HideHelpCommand: true,
}

var proteinSequenceRedesignStop = cli.Command{
	Name:    "stop",
	Usage:   "Stop an in-progress protein sequence redesign run early",
	Suggest: true,
	Flags: []cli.Flag{
		&requestflag.Flag[string]{
			Name:      "id",
			Required:  true,
			PathParam: "id",
		},
	},
	Action:          handleProteinSequenceRedesignStop,
	HideHelpCommand: true,
}

func handleProteinSequenceRedesignRetrieve(ctx context.Context, cmd *cli.Command) error {
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

	params := boltzapi.ProteinSequenceRedesignGetParams{}

	var res []byte
	options = append(options, option.WithResponseBodyInto(&res))
	_, err = client.Protein.SequenceRedesign.Get(
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
		Title:          "protein:sequence-redesign retrieve",
		Transform:      transform,
	})
}

func handleProteinSequenceRedesignList(ctx context.Context, cmd *cli.Command) error {
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

	params := boltzapi.ProteinSequenceRedesignListParams{}

	format := cmd.Root().String("format")
	explicitFormat := cmd.Root().IsSet("format")
	transform := cmd.Root().String("transform")
	if format == "raw" {
		var res []byte
		options = append(options, option.WithResponseBodyInto(&res))
		_, err = client.Protein.SequenceRedesign.List(ctx, params, options...)
		if err != nil {
			return err
		}
		obj := gjson.ParseBytes(res)
		return ShowJSON(obj, ShowJSONOpts{
			ExplicitFormat: explicitFormat,
			Format:         format,
			RawOutput:      cmd.Root().Bool("raw-output"),
			Title:          "protein:sequence-redesign list",
			Transform:      transform,
		})
	} else {
		iter := client.Protein.SequenceRedesign.ListAutoPaging(ctx, params, options...)
		maxItems := int64(-1)
		if cmd.IsSet("max-items") {
			maxItems = cmd.Value("max-items").(int64)
		}
		return ShowJSONIterator(iter, maxItems, ShowJSONOpts{
			ExplicitFormat: explicitFormat,
			Format:         format,
			RawOutput:      cmd.Root().Bool("raw-output"),
			Title:          "protein:sequence-redesign list",
			Transform:      transform,
		})
	}
}

func handleProteinSequenceRedesignDeleteData(ctx context.Context, cmd *cli.Command) error {
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
	_, err = client.Protein.SequenceRedesign.DeleteData(ctx, cmd.Value("id").(string), options...)
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
		Title:          "protein:sequence-redesign delete-data",
		Transform:      transform,
	})
}

func handleProteinSequenceRedesignEstimateCost(ctx context.Context, cmd *cli.Command) error {
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

	params := boltzapi.ProteinSequenceRedesignEstimateCostParams{}

	var res []byte
	options = append(options, option.WithResponseBodyInto(&res))
	_, err = client.Protein.SequenceRedesign.EstimateCost(ctx, params, options...)
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
		Title:          "protein:sequence-redesign estimate-cost",
		Transform:      transform,
	})
}

func handleProteinSequenceRedesignListResults(ctx context.Context, cmd *cli.Command) error {
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

	params := boltzapi.ProteinSequenceRedesignListResultsParams{}

	format := cmd.Root().String("format")
	explicitFormat := cmd.Root().IsSet("format")
	transform := cmd.Root().String("transform")
	if format == "raw" {
		var res []byte
		options = append(options, option.WithResponseBodyInto(&res))
		_, err = client.Protein.SequenceRedesign.ListResults(
			ctx,
			cmd.Value("id").(string),
			params,
			options...,
		)
		if err != nil {
			return err
		}
		obj := gjson.ParseBytes(res)
		return ShowJSON(obj, ShowJSONOpts{
			ExplicitFormat: explicitFormat,
			Format:         format,
			RawOutput:      cmd.Root().Bool("raw-output"),
			Title:          "protein:sequence-redesign list-results",
			Transform:      transform,
		})
	} else {
		iter := client.Protein.SequenceRedesign.ListResultsAutoPaging(
			ctx,
			cmd.Value("id").(string),
			params,
			options...,
		)
		maxItems := int64(-1)
		if cmd.IsSet("max-items") {
			maxItems = cmd.Value("max-items").(int64)
		}
		return ShowJSONIterator(iter, maxItems, ShowJSONOpts{
			ExplicitFormat: explicitFormat,
			Format:         format,
			RawOutput:      cmd.Root().Bool("raw-output"),
			Title:          "protein:sequence-redesign list-results",
			Transform:      transform,
		})
	}
}

func handleProteinSequenceRedesignResume(ctx context.Context, cmd *cli.Command) error {
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
	_, err = client.Protein.SequenceRedesign.Resume(ctx, cmd.Value("id").(string), options...)
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
		Title:          "protein:sequence-redesign resume",
		Transform:      transform,
	})
}

func handleProteinSequenceRedesignStart(ctx context.Context, cmd *cli.Command) error {
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

	params := boltzapi.ProteinSequenceRedesignStartParams{}

	var res []byte
	options = append(options, option.WithResponseBodyInto(&res))
	_, err = client.Protein.SequenceRedesign.Start(ctx, params, options...)
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
		Title:          "protein:sequence-redesign start",
		Transform:      transform,
	})
}

func handleProteinSequenceRedesignStop(ctx context.Context, cmd *cli.Command) error {
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
	_, err = client.Protein.SequenceRedesign.Stop(ctx, cmd.Value("id").(string), options...)
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
		Title:          "protein:sequence-redesign stop",
		Transform:      transform,
	})
}
