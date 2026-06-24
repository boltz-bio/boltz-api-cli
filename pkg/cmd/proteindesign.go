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

var proteinDesignRetrieve = cli.Command{
	Name:    "retrieve",
	Usage:   "Retrieve a design run by ID, including progress and status",
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
	Action:          handleProteinDesignRetrieve,
	HideHelpCommand: true,
}

var proteinDesignList = cli.Command{
	Name:    "list",
	Usage:   "List protein design runs, optionally filtered by workspace",
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
	Action:          handleProteinDesignList,
	HideHelpCommand: true,
}

var proteinDesignDeleteData = cli.Command{
	Name:    "delete-data",
	Usage:   "Permanently delete the input, output, and result data associated with this\ndesign run. The design run record itself is retained with a `data_deleted_at`\ntimestamp. This action is irreversible.",
	Suggest: true,
	Flags: []cli.Flag{
		&requestflag.Flag[string]{
			Name:      "id",
			Required:  true,
			PathParam: "id",
		},
	},
	Action:          handleProteinDesignDeleteData,
	HideHelpCommand: true,
}

var proteinDesignEstimateCost = cli.Command{
	Name:    "estimate-cost",
	Usage:   "Estimate the cost of a protein design run without creating any resource or\nconsuming GPU.",
	Suggest: true,
	Flags: []cli.Flag{
		&requestflag.Flag[map[string]any]{
			Name:     "binder-specification",
			Usage:    "Binder specification for protein design. Use no_template for sequence-defined binders, structure_template for uploaded binder structures, or boltz_curated for Boltz-managed nanobody and antibody defaults.",
			Required: true,
			BodyPath: "binder_specification",
		},
		&requestflag.Flag[int64]{
			Name:     "num-proteins",
			Usage:    "Number of protein designs to generate. Must be between 10 and 1,000,000.",
			Required: true,
			BodyPath: "num_proteins",
		},
		&requestflag.Flag[map[string]any]{
			Name:     "target",
			Usage:    "Target specification (structure template or template-free)",
			Required: true,
			BodyPath: "target",
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
	Action:          handleProteinDesignEstimateCost,
	HideHelpCommand: true,
}

var proteinDesignListResults = cli.Command{
	Name:    "list-results",
	Usage:   "Retrieve paginated results from a protein design run",
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
	Action:          handleProteinDesignListResults,
	HideHelpCommand: true,
}

var proteinDesignResume = cli.Command{
	Name:    "resume",
	Usage:   "Resume a stopped protein design run from its last checkpoint",
	Suggest: true,
	Flags: []cli.Flag{
		&requestflag.Flag[string]{
			Name:      "id",
			Required:  true,
			PathParam: "id",
		},
	},
	Action:          handleProteinDesignResume,
	HideHelpCommand: true,
}

var proteinDesignStart = cli.Command{
	Name:    "start",
	Usage:   "Create a new design run that generates novel protein binder candidates",
	Suggest: true,
	Flags: []cli.Flag{
		&requestflag.Flag[map[string]any]{
			Name:     "binder-specification",
			Usage:    "Binder specification for protein design. Use no_template for sequence-defined binders, structure_template for uploaded binder structures, or boltz_curated for Boltz-managed nanobody and antibody defaults.",
			Required: true,
			BodyPath: "binder_specification",
		},
		&requestflag.Flag[int64]{
			Name:     "num-proteins",
			Usage:    "Number of protein designs to generate. Must be between 10 and 1,000,000.",
			Required: true,
			BodyPath: "num_proteins",
		},
		&requestflag.Flag[map[string]any]{
			Name:     "target",
			Usage:    "Target specification (structure template or template-free)",
			Required: true,
			BodyPath: "target",
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
	Action:          handleProteinDesignStart,
	HideHelpCommand: true,
}

var proteinDesignStop = cli.Command{
	Name:    "stop",
	Usage:   "Stop an in-progress protein design run early",
	Suggest: true,
	Flags: []cli.Flag{
		&requestflag.Flag[string]{
			Name:      "id",
			Required:  true,
			PathParam: "id",
		},
	},
	Action:          handleProteinDesignStop,
	HideHelpCommand: true,
}

func handleProteinDesignRetrieve(ctx context.Context, cmd *cli.Command) error {
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

	params := boltzapi.ProteinDesignGetParams{}

	var res []byte
	options = append(options, option.WithResponseBodyInto(&res))
	_, err = client.Protein.Design.Get(
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
		Title:          "protein:design retrieve",
		Transform:      transform,
	})
}

func handleProteinDesignList(ctx context.Context, cmd *cli.Command) error {
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

	params := boltzapi.ProteinDesignListParams{}

	format := cmd.Root().String("format")
	explicitFormat := cmd.Root().IsSet("format")
	transform := cmd.Root().String("transform")
	if format == "raw" {
		var res []byte
		options = append(options, option.WithResponseBodyInto(&res))
		_, err = client.Protein.Design.List(ctx, params, options...)
		if err != nil {
			return err
		}
		obj := gjson.ParseBytes(res)
		return ShowJSON(obj, ShowJSONOpts{
			ExplicitFormat: explicitFormat,
			Format:         format,
			RawOutput:      cmd.Root().Bool("raw-output"),
			Title:          "protein:design list",
			Transform:      transform,
		})
	} else {
		iter := client.Protein.Design.ListAutoPaging(ctx, params, options...)
		maxItems := int64(-1)
		if cmd.IsSet("max-items") {
			maxItems = cmd.Value("max-items").(int64)
		}
		return ShowJSONIterator(iter, maxItems, ShowJSONOpts{
			ExplicitFormat: explicitFormat,
			Format:         format,
			RawOutput:      cmd.Root().Bool("raw-output"),
			Title:          "protein:design list",
			Transform:      transform,
		})
	}
}

func handleProteinDesignDeleteData(ctx context.Context, cmd *cli.Command) error {
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
	_, err = client.Protein.Design.DeleteData(ctx, cmd.Value("id").(string), options...)
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
		Title:          "protein:design delete-data",
		Transform:      transform,
	})
}

func handleProteinDesignEstimateCost(ctx context.Context, cmd *cli.Command) error {
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

	params := boltzapi.ProteinDesignEstimateCostParams{}

	var res []byte
	options = append(options, option.WithResponseBodyInto(&res))
	_, err = client.Protein.Design.EstimateCost(ctx, params, options...)
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
		Title:          "protein:design estimate-cost",
		Transform:      transform,
	})
}

func handleProteinDesignListResults(ctx context.Context, cmd *cli.Command) error {
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

	params := boltzapi.ProteinDesignListResultsParams{}

	format := cmd.Root().String("format")
	explicitFormat := cmd.Root().IsSet("format")
	transform := cmd.Root().String("transform")
	if format == "raw" {
		var res []byte
		options = append(options, option.WithResponseBodyInto(&res))
		_, err = client.Protein.Design.ListResults(
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
			Title:          "protein:design list-results",
			Transform:      transform,
		})
	} else {
		iter := client.Protein.Design.ListResultsAutoPaging(
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
			Title:          "protein:design list-results",
			Transform:      transform,
		})
	}
}

func handleProteinDesignResume(ctx context.Context, cmd *cli.Command) error {
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
	_, err = client.Protein.Design.Resume(ctx, cmd.Value("id").(string), options...)
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
		Title:          "protein:design resume",
		Transform:      transform,
	})
}

func handleProteinDesignStart(ctx context.Context, cmd *cli.Command) error {
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

	params := boltzapi.ProteinDesignStartParams{}

	var res []byte
	options = append(options, option.WithResponseBodyInto(&res))
	_, err = client.Protein.Design.Start(ctx, params, options...)
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
		Title:          "protein:design start",
		Transform:      transform,
	})
}

func handleProteinDesignStop(ctx context.Context, cmd *cli.Command) error {
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
	_, err = client.Protein.Design.Stop(ctx, cmd.Value("id").(string), options...)
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
		Title:          "protein:design stop",
		Transform:      transform,
	})
}
