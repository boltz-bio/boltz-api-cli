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
	Usage:   "Create a read-only share link covering one or more predictions and/or pipelines\nthat all live in the same workspace. Public links require only the returned\nbearer ID; email-restricted links also require a signed-in viewer whose email is\nallowed. Treat the returned `id` as a secret.",
	Suggest: true,
	Flags: []cli.Flag{
		&requestflag.Flag[string]{
			Name:     "expires-at",
			Required: true,
			BodyPath: "expires_at",
		},
		&requestflag.Flag[map[string]any]{
			Name:     "access-parameters",
			Usage:    "Access-control parameters for the share link. Discriminated by `access_mode`: `public` requires no other fields; `email` requires a non-empty `allowed_emails` list.",
			BodyPath: "access_parameters",
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
			Usage:    "Workspace to target. Admin API keys and OAuth callers may select an authorized workspace; for workspace-scoped keys the value must match the key assignment.",
			BodyPath: "workspace_id",
		},
	},
	Action:          handleShareLinksCreate,
	HideHelpCommand: true,
}

var shareLinksRetrieve = cli.Command{
	Name:    "retrieve",
	Usage:   "Retrieve metadata for a share link owned by the authenticated organization.\nArchived and expired links remain retrievable.",
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

var shareLinksArchive = cli.Command{
	Name:    "archive",
	Usage:   "Archive a share link so it no longer grants public access. Metadata remains\nretrievable and repeated calls preserve the first archive timestamp.",
	Suggest: true,
	Flags: []cli.Flag{
		&requestflag.Flag[string]{
			Name:      "id",
			Required:  true,
			PathParam: "id",
		},
	},
	Action:          handleShareLinksArchive,
	HideHelpCommand: true,
}

var shareLinksListPipelineResults = cli.Command{
	Name:    "list-pipeline-results",
	Usage:   "Paginated results for one pipeline exposed by a share link. The response shape\nmatches the authed pipeline-results endpoints exactly. Access is gated by the\nshare-link ID and — for email-mode links — a signed compute-API JWT. Pipeline\nIDs not covered by the link return 404 indistinguishably from unknown links.",
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

var shareLinksRead = cli.Command{
	Name:    "read",
	Usage:   "Read the predictions and pipelines exposed by a share link. Public links require\nno authentication — the share link ID itself is the access credential.\nEmail-mode links additionally require a signed compute-API JWT (minted for the\nbrowser session by Lab, or presented directly by CLI/SDK callers). Returns 404\nindistinguishably for unknown, expired, or archived links.",
	Suggest: true,
	Flags: []cli.Flag{
		&requestflag.Flag[string]{
			Name:      "id",
			Required:  true,
			PathParam: "id",
		},
	},
	Action:          handleShareLinksRead,
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

func handleShareLinksArchive(ctx context.Context, cmd *cli.Command) error {
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
	_, err = client.ShareLinks.Archive(ctx, cmd.Value("id").(string), options...)
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
		Title:          "share-links archive",
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

func handleShareLinksRead(ctx context.Context, cmd *cli.Command) error {
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
	_, err = client.ShareLinks.Read(ctx, cmd.Value("id").(string), options...)
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
		Title:          "share-links read",
		Transform:      transform,
	})
}
