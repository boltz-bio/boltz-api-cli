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

var smallMoleculeExploreRetrieve = cli.Command{
	Name:    "retrieve",
	Usage:   "Retrieve an exploration by ID. Once library preparation completes, progress\nreports the accepted library size alongside the budget.",
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
	Action:          handleSmallMoleculeExploreRetrieve,
	HideHelpCommand: true,
}

var smallMoleculeExploreListResults = cli.Command{
	Name:    "list-results",
	Usage:   "Retrieve paginated results from an exploration. Results appear as molecules are\nscored, and remain retrievable if the run fails partway.",
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
	Action:          handleSmallMoleculeExploreListResults,
	HideHelpCommand: true,
}

var smallMoleculeExploreResume = cli.Command{
	Name:    "resume",
	Usage:   "Resume a stopped exploration. Selection continues from the surrogate as it\nstood, so molecules already scored still inform what is chosen next.",
	Suggest: true,
	Flags: []cli.Flag{
		&requestflag.Flag[string]{
			Name:      "id",
			Required:  true,
			PathParam: "id",
		},
	},
	Action:          handleSmallMoleculeExploreResume,
	HideHelpCommand: true,
}

var smallMoleculeExploreStart = requestflag.WithInnerFlags(cli.Command{
	Name:    "start",
	Usage:   "Explore a large library against a protein target without screening all of it.\nSubmit the whole library and a budget; molecules are chosen to score as results\narrive, so each choice is informed by everything scored before it.",
	Suggest: true,
	Flags: []cli.Flag{
		&requestflag.Flag[int64]{
			Name:     "budget",
			Usage:    "How many molecules to score. Must not exceed the accepted library size or 4,000,000.",
			Required: true,
			BodyPath: "budget",
		},
		&requestflag.Flag[map[string]any]{
			Name:     "library",
			Usage:    "CSV or TSV molecule library, limited to 300 MiB and 4,000,000 data records. URL sources can use the full file limit. Base64 sources are also subject to the API's 50 MiB JSON request-body limit, so use a URL source for larger files. The file must be UTF-8 and may contain only the selected SMILES and ID columns; column order does not matter. Candidate IDs are limited to 1,024 UTF-8 bytes; missing or blank IDs default to the zero-based data-record index.",
			Required: true,
			BodyPath: "library",
		},
		&requestflag.Flag[map[string]any]{
			Name:     "target",
			Usage:    "Target protein sequences for small molecule design or screening.",
			Required: true,
			BodyPath: "target",
		},
		&requestflag.Flag[string]{
			Name:     "idempotency-key",
			Usage:    "Client-provided key to prevent duplicate submissions on retries",
			BodyPath: "idempotency_key",
		},
		&requestflag.Flag[map[string]any]{
			Name:     "molecule-filters",
			Usage:    "Molecule filtering configuration. Controls both Boltz built-in SMARTS filtering and custom filters.",
			BodyPath: "molecule_filters",
		},
		&requestflag.Flag[string]{
			Name:     "workspace-id",
			Usage:    "Target workspace ID (admin keys only; ignored for workspace keys)",
			BodyPath: "workspace_id",
		},
	},
	Action:          handleSmallMoleculeExploreStart,
	HideHelpCommand: true,
}, map[string][]requestflag.HasOuterFlag{
	"library": {
		&requestflag.InnerFlag[string]{
			Name:       "library.format",
			Usage:      "Delimited-text format of the molecule library.",
			InnerField: "format",
		},
		&requestflag.InnerFlag[map[string]any]{
			Name:       "library.source",
			Usage:      "How to provide a file to the API",
			InnerField: "source",
		},
		&requestflag.InnerFlag[string]{
			Name:       "library.id-column",
			Usage:      "Column containing candidate IDs. When omitted, a distinct `id` column is used if present; otherwise IDs are generated from zero-based data-record indexes.",
			InnerField: "id_column",
		},
		&requestflag.InnerFlag[string]{
			Name:       "library.smiles-column",
			Usage:      "Column containing SMILES strings. Defaults to `smiles` when omitted.",
			InnerField: "smiles_column",
		},
	},
	"target": {
		&requestflag.InnerFlag[[]map[string]any]{
			Name:       "target.entities",
			Usage:      "Protein and glycan entities defining the target structure. At least one protein entity is required.",
			InnerField: "entities",
		},
		&requestflag.InnerFlag[[]map[string]any]{
			Name:       "target.bonds",
			Usage:      "Covalent bond constraints between atoms in the target complex. Ligand atom references support CCD atom names and explicitly atom-mapped SMILES atoms.",
			InnerField: "bonds",
		},
		&requestflag.InnerFlag[[]map[string]any]{
			Name:       "target.constraints",
			Usage:      "Structural constraints (pocket and contact). Ligand atom references support CCD atom names and explicitly atom-mapped SMILES atoms.",
			InnerField: "constraints",
		},
		&requestflag.InnerFlag[map[string]any]{
			Name:       "target.pocket-residues",
			Usage:      `Binding pocket residues, keyed by chain ID. Each key is a chain ID (e.g. "A") and the value is an array of 0-indexed residue indices that define the binding pocket on that chain. When provided, these residues guide pocket extraction and add a derived pocket constraint during affinity predictions. That derived constraint remains separate from any explicit pocket constraints in target.constraints. When omitted, the model auto-detects the pocket.`,
			InnerField: "pocket_residues",
		},
		&requestflag.InnerFlag[[]string]{
			Name:       "target.reference-ligands",
			Usage:      "Reference ligands as SMILES strings that help the model identify the binding pocket. When omitted, a set of drug-like default ligands is used for pocket detection.",
			InnerField: "reference_ligands",
		},
		&requestflag.InnerFlag[string]{
			Name:       "target.type",
			Usage:      "Target is defined directly by protein sequences rather than a structure template.",
			InnerField: "type",
		},
	},
	"molecule-filters": {
		&requestflag.InnerFlag[string]{
			Name:       "molecule-filters.boltz-smarts-catalog-filter-level",
			Usage:      "Controls the stringency of Boltz's built-in SMARTS structural alert filtering, which removes molecules matching known problematic substructures. 'recommended' (default): applies a curated set of alerts balancing safety and hit rate. 'extra': adds additional alerts beyond the recommended set for stricter filtering. 'aggressive': applies the most comprehensive alert set — may reject viable molecules. 'disabled': turns off Boltz SMARTS filtering entirely; only custom_filters will be applied.",
			InnerField: "boltz_smarts_catalog_filter_level",
		},
		&requestflag.InnerFlag[[]map[string]any]{
			Name:       "molecule-filters.custom-filters",
			Usage:      "Custom filters to apply. Molecules must pass all filters (AND logic).",
			InnerField: "custom_filters",
		},
	},
})

var smallMoleculeExploreStop = cli.Command{
	Name:    "stop",
	Usage:   "Stop an in-progress exploration early. Molecules already scored are kept and\nremain retrievable; no further molecules are selected.",
	Suggest: true,
	Flags: []cli.Flag{
		&requestflag.Flag[string]{
			Name:      "id",
			Required:  true,
			PathParam: "id",
		},
	},
	Action:          handleSmallMoleculeExploreStop,
	HideHelpCommand: true,
}

func handleSmallMoleculeExploreRetrieve(ctx context.Context, cmd *cli.Command) error {
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

	params := boltzapi.SmallMoleculeExploreGetParams{}

	var res []byte
	options = append(options, option.WithResponseBodyInto(&res))
	_, err = client.SmallMolecule.Explore.Get(
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
		Title:          "small-molecule:explore retrieve",
		Transform:      transform,
	})
}

func handleSmallMoleculeExploreListResults(ctx context.Context, cmd *cli.Command) error {
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

	params := boltzapi.SmallMoleculeExploreListResultsParams{}

	format := cmd.Root().String("format")
	explicitFormat := cmd.Root().IsSet("format")
	transform := cmd.Root().String("transform")
	if format == "raw" {
		var res []byte
		options = append(options, option.WithResponseBodyInto(&res))
		_, err = client.SmallMolecule.Explore.ListResults(
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
			Title:          "small-molecule:explore list-results",
			Transform:      transform,
		})
	} else {
		iter := client.SmallMolecule.Explore.ListResultsAutoPaging(
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
			Title:          "small-molecule:explore list-results",
			Transform:      transform,
		})
	}
}

func handleSmallMoleculeExploreResume(ctx context.Context, cmd *cli.Command) error {
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
	_, err = client.SmallMolecule.Explore.Resume(ctx, cmd.Value("id").(string), options...)
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
		Title:          "small-molecule:explore resume",
		Transform:      transform,
	})
}

func handleSmallMoleculeExploreStart(ctx context.Context, cmd *cli.Command) error {
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

	params := boltzapi.SmallMoleculeExploreStartParams{}

	var res []byte
	options = append(options, option.WithResponseBodyInto(&res))
	_, err = client.SmallMolecule.Explore.Start(ctx, params, options...)
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
		Title:          "small-molecule:explore start",
		Transform:      transform,
	})
}

func handleSmallMoleculeExploreStop(ctx context.Context, cmd *cli.Command) error {
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
	_, err = client.SmallMolecule.Explore.Stop(ctx, cmd.Value("id").(string), options...)
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
		Title:          "small-molecule:explore stop",
		Transform:      transform,
	})
}
