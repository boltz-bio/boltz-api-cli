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

var smallMoleculeDesignRetrieve = cli.Command{
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
	Action:          handleSmallMoleculeDesignRetrieve,
	HideHelpCommand: true,
}

var smallMoleculeDesignList = cli.Command{
	Name:    "list",
	Usage:   "List small molecule design runs, optionally filtered by workspace",
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
	Action:          handleSmallMoleculeDesignList,
	HideHelpCommand: true,
}

var smallMoleculeDesignDeleteData = cli.Command{
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
	Action:          handleSmallMoleculeDesignDeleteData,
	HideHelpCommand: true,
}

var smallMoleculeDesignEstimateCost = requestflag.WithInnerFlags(cli.Command{
	Name:    "estimate-cost",
	Usage:   "Estimate the billed cost of a small molecule design run without creating any\nresource or consuming GPU. Includes generation charges implied by the scheduler\niteration cap plus structure-scoring charges for each requested molecule.",
	Suggest: true,
	Flags: []cli.Flag{
		&requestflag.Flag[int64]{
			Name:     "num-molecules",
			Usage:    "Number of molecules to generate. Must be between 10 and 1,000,000.",
			Required: true,
			BodyPath: "num_molecules",
		},
		&requestflag.Flag[map[string]any]{
			Name:     "target",
			Usage:    "Target protein sequences for small molecule design or screening.",
			Required: true,
			BodyPath: "target",
		},
		&requestflag.Flag[string]{
			Name:     "chemical-space",
			Usage:    "Chemical space to constrain generated molecules. Use 'enamine_real' for the Enamine REAL chemical space or 'none' to disable chemical-space filtering.",
			Default:  "enamine_real",
			BodyPath: "chemical_space",
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
	Action:          handleSmallMoleculeDesignEstimateCost,
	HideHelpCommand: true,
}, map[string][]requestflag.HasOuterFlag{
	"target": {
		&requestflag.InnerFlag[[]map[string]any]{
			Name:       "target.entities",
			Usage:      "Protein entities defining the target structure. Each entity represents a protein chain.",
			InnerField: "entities",
		},
		&requestflag.InnerFlag[[]map[string]any]{
			Name:       "target.bonds",
			Usage:      "Covalent bond constraints between atoms in the target complex. Atom-level ligand references support CCD atom names and explicitly atom-mapped SMILES atoms; for SMILES ligands, use numeric atom-map notation (for example, [C:1] is referenced as C1).",
			InnerField: "bonds",
		},
		&requestflag.InnerFlag[[]map[string]any]{
			Name:       "target.constraints",
			Usage:      "Structural constraints (pocket and contact). Atom-level ligand references support CCD atom names and explicitly atom-mapped SMILES atoms; for SMILES ligands, use numeric atom-map notation (for example, [C:1] is referenced as C1).",
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

var smallMoleculeDesignListResults = cli.Command{
	Name:    "list-results",
	Usage:   "Retrieve paginated results from a design run",
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
	Action:          handleSmallMoleculeDesignListResults,
	HideHelpCommand: true,
}

var smallMoleculeDesignResume = cli.Command{
	Name:    "resume",
	Usage:   "Resume a stopped small molecule design run from its last checkpoint",
	Suggest: true,
	Flags: []cli.Flag{
		&requestflag.Flag[string]{
			Name:      "id",
			Required:  true,
			PathParam: "id",
		},
	},
	Action:          handleSmallMoleculeDesignResume,
	HideHelpCommand: true,
}

var smallMoleculeDesignStart = requestflag.WithInnerFlags(cli.Command{
	Name:    "start",
	Usage:   "Create a new design run that generates novel small molecule candidates for a\nprotein target",
	Suggest: true,
	Flags: []cli.Flag{
		&requestflag.Flag[int64]{
			Name:     "num-molecules",
			Usage:    "Number of molecules to generate. Must be between 10 and 1,000,000.",
			Required: true,
			BodyPath: "num_molecules",
		},
		&requestflag.Flag[map[string]any]{
			Name:     "target",
			Usage:    "Target protein sequences for small molecule design or screening.",
			Required: true,
			BodyPath: "target",
		},
		&requestflag.Flag[string]{
			Name:     "chemical-space",
			Usage:    "Chemical space to constrain generated molecules. Use 'enamine_real' for the Enamine REAL chemical space or 'none' to disable chemical-space filtering.",
			Default:  "enamine_real",
			BodyPath: "chemical_space",
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
	Action:          handleSmallMoleculeDesignStart,
	HideHelpCommand: true,
}, map[string][]requestflag.HasOuterFlag{
	"target": {
		&requestflag.InnerFlag[[]map[string]any]{
			Name:       "target.entities",
			Usage:      "Protein entities defining the target structure. Each entity represents a protein chain.",
			InnerField: "entities",
		},
		&requestflag.InnerFlag[[]map[string]any]{
			Name:       "target.bonds",
			Usage:      "Covalent bond constraints between atoms in the target complex. Atom-level ligand references support CCD atom names and explicitly atom-mapped SMILES atoms; for SMILES ligands, use numeric atom-map notation (for example, [C:1] is referenced as C1).",
			InnerField: "bonds",
		},
		&requestflag.InnerFlag[[]map[string]any]{
			Name:       "target.constraints",
			Usage:      "Structural constraints (pocket and contact). Atom-level ligand references support CCD atom names and explicitly atom-mapped SMILES atoms; for SMILES ligands, use numeric atom-map notation (for example, [C:1] is referenced as C1).",
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

var smallMoleculeDesignStop = cli.Command{
	Name:    "stop",
	Usage:   "Stop an in-progress design run early",
	Suggest: true,
	Flags: []cli.Flag{
		&requestflag.Flag[string]{
			Name:      "id",
			Required:  true,
			PathParam: "id",
		},
	},
	Action:          handleSmallMoleculeDesignStop,
	HideHelpCommand: true,
}

func handleSmallMoleculeDesignRetrieve(ctx context.Context, cmd *cli.Command) error {
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

	params := boltzapi.SmallMoleculeDesignGetParams{}

	var res []byte
	options = append(options, option.WithResponseBodyInto(&res))
	_, err = client.SmallMolecule.Design.Get(
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
		Title:          "small-molecule:design retrieve",
		Transform:      transform,
	})
}

func handleSmallMoleculeDesignList(ctx context.Context, cmd *cli.Command) error {
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

	params := boltzapi.SmallMoleculeDesignListParams{}

	format := cmd.Root().String("format")
	explicitFormat := cmd.Root().IsSet("format")
	transform := cmd.Root().String("transform")
	if format == "raw" {
		var res []byte
		options = append(options, option.WithResponseBodyInto(&res))
		_, err = client.SmallMolecule.Design.List(ctx, params, options...)
		if err != nil {
			return err
		}
		obj := gjson.ParseBytes(res)
		return ShowJSON(obj, ShowJSONOpts{
			ExplicitFormat: explicitFormat,
			Format:         format,
			RawOutput:      cmd.Root().Bool("raw-output"),
			Title:          "small-molecule:design list",
			Transform:      transform,
		})
	} else {
		iter := client.SmallMolecule.Design.ListAutoPaging(ctx, params, options...)
		maxItems := int64(-1)
		if cmd.IsSet("max-items") {
			maxItems = cmd.Value("max-items").(int64)
		}
		return ShowJSONIterator(iter, maxItems, ShowJSONOpts{
			ExplicitFormat: explicitFormat,
			Format:         format,
			RawOutput:      cmd.Root().Bool("raw-output"),
			Title:          "small-molecule:design list",
			Transform:      transform,
		})
	}
}

func handleSmallMoleculeDesignDeleteData(ctx context.Context, cmd *cli.Command) error {
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
	_, err = client.SmallMolecule.Design.DeleteData(ctx, cmd.Value("id").(string), options...)
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
		Title:          "small-molecule:design delete-data",
		Transform:      transform,
	})
}

func handleSmallMoleculeDesignEstimateCost(ctx context.Context, cmd *cli.Command) error {
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

	params := boltzapi.SmallMoleculeDesignEstimateCostParams{}

	var res []byte
	options = append(options, option.WithResponseBodyInto(&res))
	_, err = client.SmallMolecule.Design.EstimateCost(ctx, params, options...)
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
		Title:          "small-molecule:design estimate-cost",
		Transform:      transform,
	})
}

func handleSmallMoleculeDesignListResults(ctx context.Context, cmd *cli.Command) error {
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

	params := boltzapi.SmallMoleculeDesignListResultsParams{}

	format := cmd.Root().String("format")
	explicitFormat := cmd.Root().IsSet("format")
	transform := cmd.Root().String("transform")
	if format == "raw" {
		var res []byte
		options = append(options, option.WithResponseBodyInto(&res))
		_, err = client.SmallMolecule.Design.ListResults(
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
			Title:          "small-molecule:design list-results",
			Transform:      transform,
		})
	} else {
		iter := client.SmallMolecule.Design.ListResultsAutoPaging(
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
			Title:          "small-molecule:design list-results",
			Transform:      transform,
		})
	}
}

func handleSmallMoleculeDesignResume(ctx context.Context, cmd *cli.Command) error {
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
	_, err = client.SmallMolecule.Design.Resume(ctx, cmd.Value("id").(string), options...)
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
		Title:          "small-molecule:design resume",
		Transform:      transform,
	})
}

func handleSmallMoleculeDesignStart(ctx context.Context, cmd *cli.Command) error {
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

	params := boltzapi.SmallMoleculeDesignStartParams{}

	var res []byte
	options = append(options, option.WithResponseBodyInto(&res))
	_, err = client.SmallMolecule.Design.Start(ctx, params, options...)
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
		Title:          "small-molecule:design start",
		Transform:      transform,
	})
}

func handleSmallMoleculeDesignStop(ctx context.Context, cmd *cli.Command) error {
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
	_, err = client.SmallMolecule.Design.Stop(ctx, cmd.Value("id").(string), options...)
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
		Title:          "small-molecule:design stop",
		Transform:      transform,
	})
}
