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

var shareLinksCreate = cli.Command{
	Name:    "create",
	Usage:   "Create an unauthenticated, read-only share link covering one or more predictions\nand/or pipelines that all live in the same workspace. The returned `id` is the\nbearer credential — treat it as a secret.",
	Suggest: true,
	Flags: []cli.Flag{
		&requestflag.Flag[string]{
			Name:     "expires-at",
			Required: true,
			BodyPath: "expires_at",
		},
		&requestflag.Flag[[]string]{
			Name:     "pipeline-id",
			Usage:    "Pipelines to expose through the share link. Must belong to the resolved workspace. Up to 100 entries.",
			BodyPath: "pipeline_ids",
		},
		&requestflag.Flag[[]string]{
			Name:     "prediction-id",
			Usage:    "Predictions to expose through the share link. Must belong to the resolved workspace. Up to 100 entries.",
			BodyPath: "prediction_ids",
		},
		&requestflag.Flag[string]{
			Name:     "workspace-id",
			Usage:    "Workspace ID. Only used with admin API keys. Ignored (or validated) for workspace-scoped keys.",
			BodyPath: "workspace_id",
		},
	},
	Action:          handleShareLinksCreate,
	HideHelpCommand: true,
}

var shareLinksRetrieve = cli.Command{
	Name:    "retrieve",
	Usage:   "Read the predictions and pipelines exposed by a share link. No authentication is\nrequired — the share link ID itself is the access credential. Returns 404\nindistinguishably for unknown, expired, or revoked links.",
	Suggest: true,
	Flags: []cli.Flag{
		&requestflag.Flag[string]{
			Name:      "id",
			Required:  true,
			PathParam: "id",
		},
	},
	Action:          handleShareLinksRetrieve,
	HideHelpCommand: true,
}

var shareLinksDeleteData = cli.Command{
	Name:    "delete-data",
	Usage:   "Revoke a share link so its bearer credential can no longer be used. The record\nis retained with a `data_deleted_at` timestamp and is purged by a background\nsweep. This action is irreversible. The share link ID is sent in the body\nbecause it is the bearer credential and must not appear in URLs.",
	Suggest: true,
	Flags: []cli.Flag{
		&requestflag.Flag[string]{
			Name:     "id",
			Usage:    "Share link ID to revoke. Sent in the body — never as a URL path segment — because the ID is itself the bearer credential.",
			Required: true,
			BodyPath: "id",
		},
	},
	Action:          handleShareLinksDeleteData,
	HideHelpCommand: true,
}

var shareLinksListPipelineResults = cli.Command{
	Name:    "list-pipeline-results",
	Usage:   "Paginated results for one pipeline exposed by a share link. The response shape\nmatches the authed pipeline-results endpoints exactly. No authentication is\nrequired — the share-link ID gates access. Pipeline IDs not covered by the link\nreturn 404 indistinguishably from unknown links.",
	Suggest: true,
	Flags: []cli.Flag{
		&requestflag.Flag[string]{
			Name:      "id",
			Required:  true,
			PathParam: "id",
		},
		&requestflag.Flag[string]{
			Name:      "pipeline-id",
			Required:  true,
			PathParam: "pipelineId",
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
		&requestflag.Flag[int64]{
			Name:  "max-items",
			Usage: "The maximum number of items to return (use -1 for unlimited).",
		},
	},
	Action:          handleShareLinksListPipelineResults,
	HideHelpCommand: true,
}

func handleShareLinksCreate(ctx context.Context, cmd *cli.Command) error {
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

	params := boltzapi.ShareLinkNewParams{}

	var res []byte
	options = append(options, option.WithResponseBodyInto(&res))
	_, err = client.ShareLinks.New(ctx, params, options...)
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
		Title:          "share-links create",
		Transform:      transform,
	})
}

func handleShareLinksRetrieve(ctx context.Context, cmd *cli.Command) error {
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
	_, err = client.ShareLinks.Get(ctx, cmd.Value("id").(string), options...)
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
		Title:          "share-links retrieve",
		Transform:      transform,
	})
}

func handleShareLinksDeleteData(ctx context.Context, cmd *cli.Command) error {
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

	params := boltzapi.ShareLinkDeleteDataParams{}

	var res []byte
	options = append(options, option.WithResponseBodyInto(&res))
	_, err = client.ShareLinks.DeleteData(ctx, params, options...)
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
		Title:          "share-links delete-data",
		Transform:      transform,
	})
}

func handleShareLinksListPipelineResults(ctx context.Context, cmd *cli.Command) error {
	client := boltzapi.NewClient(getDefaultRequestOptions(cmd)...)
	unusedArgs := cmd.Args().Slice()
	if !cmd.IsSet("pipeline-id") && len(unusedArgs) > 0 {
		cmd.Set("pipeline-id", unusedArgs[0])
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

	params := boltzapi.ShareLinkListPipelineResultsParams{
		ID: cmd.Value("id").(string),
	}

	format := cmd.Root().String("format")
	explicitFormat := cmd.Root().IsSet("format")
	transform := cmd.Root().String("transform")
	if format == "raw" {
		var res []byte
		options = append(options, option.WithResponseBodyInto(&res))
		_, err = client.ShareLinks.ListPipelineResults(
			ctx,
			cmd.Value("pipeline-id").(string),
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
			Title:          "share-links list-pipeline-results",
			Transform:      transform,
		})
	} else {
		iter := client.ShareLinks.ListPipelineResultsAutoPaging(
			ctx,
			cmd.Value("pipeline-id").(string),
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
			Title:          "share-links list-pipeline-results",
			Transform:      transform,
		})
	}
}
