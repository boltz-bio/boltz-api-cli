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

var smallMoleculeLibraryScreenRetrieve = cli.Command{
	Name:    "retrieve",
	Usage:   "Retrieve a library screen by ID, including progress and status",
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
	Action:          handleSmallMoleculeLibraryScreenRetrieve,
	HideHelpCommand: true,
}

var smallMoleculeLibraryScreenList = cli.Command{
	Name:    "list",
	Usage:   "List small molecule library screens, optionally filtered by workspace",
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
	Action:          handleSmallMoleculeLibraryScreenList,
	HideHelpCommand: true,
}

var smallMoleculeLibraryScreenDeleteData = cli.Command{
	Name:    "delete-data",
	Usage:   "Permanently delete the input, output, and result data associated with this\nlibrary screen. The library screen record itself is retained with a\n`data_deleted_at` timestamp. This action is irreversible.",
	Suggest: true,
	Flags: []cli.Flag{
		&requestflag.Flag[string]{
			Name:      "id",
			Required:  true,
			PathParam: "id",
		},
	},
	Action:          handleSmallMoleculeLibraryScreenDeleteData,
	HideHelpCommand: true,
}

var smallMoleculeLibraryScreenEstimateCost = requestflag.WithInnerFlags(cli.Command{
	Name:    "estimate-cost",
	Usage:   "Estimate the cost of a small molecule library screen without creating any\nresource or consuming GPU.",
	Suggest: true,
	Flags: []cli.Flag{
		&requestflag.Flag[[]map[string]any]{
			Name:     "molecule",
			Usage:    "List of small molecules to screen.",
			Required: true,
			BodyPath: "molecules",
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
	Action:          handleSmallMoleculeLibraryScreenEstimateCost,
	HideHelpCommand: true,
}, map[string][]requestflag.HasOuterFlag{
	"molecule": {
		&requestflag.InnerFlag[string]{
			Name:       "molecule.smiles",
			Usage:      "SMILES string of the molecule",
			InnerField: "smiles",
		},
		&requestflag.InnerFlag[string]{
			Name:       "molecule.id",
			Usage:      "Optional identifier for this molecule",
			InnerField: "id",
		},
	},
	"target": {
		&requestflag.InnerFlag[[]map[string]any]{
			Name:       "target.entities",
			Usage:      "Protein entities defining the target structure. Each entity represents a protein chain.",
			InnerField: "entities",
		},
		&requestflag.InnerFlag[[]map[string]any]{
			Name:       "target.bonds",
			Usage:      "Covalent bond constraints between atoms in the target complex. Atom-level ligand references currently support ligand_ccd only; ligand_smiles is unsupported.",
			InnerField: "bonds",
		},
		&requestflag.InnerFlag[[]map[string]any]{
			Name:       "target.constraints",
			Usage:      "Structural constraints (pocket and contact). Atom-level ligand references currently support ligand_ccd only; ligand_smiles is unsupported.",
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

var smallMoleculeLibraryScreenListResults = cli.Command{
	Name:    "list-results",
	Usage:   "Retrieve paginated results from a library screen",
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
	Action:          handleSmallMoleculeLibraryScreenListResults,
	HideHelpCommand: true,
}

var smallMoleculeLibraryScreenStart = requestflag.WithInnerFlags(cli.Command{
	Name:    "start",
	Usage:   "Screen a set of small molecule candidates against a protein target",
	Suggest: true,
	Flags: []cli.Flag{
		&requestflag.Flag[[]map[string]any]{
			Name:     "molecule",
			Usage:    "List of small molecules to screen.",
			Required: true,
			BodyPath: "molecules",
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
	Action:          handleSmallMoleculeLibraryScreenStart,
	HideHelpCommand: true,
}, map[string][]requestflag.HasOuterFlag{
	"molecule": {
		&requestflag.InnerFlag[string]{
			Name:       "molecule.smiles",
			Usage:      "SMILES string of the molecule",
			InnerField: "smiles",
		},
		&requestflag.InnerFlag[string]{
			Name:       "molecule.id",
			Usage:      "Optional identifier for this molecule",
			InnerField: "id",
		},
	},
	"target": {
		&requestflag.InnerFlag[[]map[string]any]{
			Name:       "target.entities",
			Usage:      "Protein entities defining the target structure. Each entity represents a protein chain.",
			InnerField: "entities",
		},
		&requestflag.InnerFlag[[]map[string]any]{
			Name:       "target.bonds",
			Usage:      "Covalent bond constraints between atoms in the target complex. Atom-level ligand references currently support ligand_ccd only; ligand_smiles is unsupported.",
			InnerField: "bonds",
		},
		&requestflag.InnerFlag[[]map[string]any]{
			Name:       "target.constraints",
			Usage:      "Structural constraints (pocket and contact). Atom-level ligand references currently support ligand_ccd only; ligand_smiles is unsupported.",
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

var smallMoleculeLibraryScreenStop = cli.Command{
	Name:    "stop",
	Usage:   "Stop an in-progress library screen early",
	Suggest: true,
	Flags: []cli.Flag{
		&requestflag.Flag[string]{
			Name:      "id",
			Required:  true,
			PathParam: "id",
		},
	},
	Action:          handleSmallMoleculeLibraryScreenStop,
	HideHelpCommand: true,
}

func handleSmallMoleculeLibraryScreenRetrieve(ctx context.Context, cmd *cli.Command) error {
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

	params := boltzapi.SmallMoleculeLibraryScreenGetParams{}

	var res []byte
	options = append(options, option.WithResponseBodyInto(&res))
	_, err = client.SmallMolecule.LibraryScreen.Get(
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
		Title:          "small-molecule:library-screen retrieve",
		Transform:      transform,
	})
}

func handleSmallMoleculeLibraryScreenList(ctx context.Context, cmd *cli.Command) error {
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

	params := boltzapi.SmallMoleculeLibraryScreenListParams{}

	format := cmd.Root().String("format")
	explicitFormat := cmd.Root().IsSet("format")
	transform := cmd.Root().String("transform")
	if format == "raw" {
		var res []byte
		options = append(options, option.WithResponseBodyInto(&res))
		_, err = client.SmallMolecule.LibraryScreen.List(ctx, params, options...)
		if err != nil {
			return err
		}
		obj := gjson.ParseBytes(res)
		return ShowJSON(obj, ShowJSONOpts{
			ExplicitFormat: explicitFormat,
			Format:         format,
			RawOutput:      cmd.Root().Bool("raw-output"),
			Title:          "small-molecule:library-screen list",
			Transform:      transform,
		})
	} else {
		iter := client.SmallMolecule.LibraryScreen.ListAutoPaging(ctx, params, options...)
		maxItems := int64(-1)
		if cmd.IsSet("max-items") {
			maxItems = cmd.Value("max-items").(int64)
		}
		return ShowJSONIterator(iter, maxItems, ShowJSONOpts{
			ExplicitFormat: explicitFormat,
			Format:         format,
			RawOutput:      cmd.Root().Bool("raw-output"),
			Title:          "small-molecule:library-screen list",
			Transform:      transform,
		})
	}
}

func handleSmallMoleculeLibraryScreenDeleteData(ctx context.Context, cmd *cli.Command) error {
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
	_, err = client.SmallMolecule.LibraryScreen.DeleteData(ctx, cmd.Value("id").(string), options...)
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
		Title:          "small-molecule:library-screen delete-data",
		Transform:      transform,
	})
}

func handleSmallMoleculeLibraryScreenEstimateCost(ctx context.Context, cmd *cli.Command) error {
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

	params := boltzapi.SmallMoleculeLibraryScreenEstimateCostParams{}

	var res []byte
	options = append(options, option.WithResponseBodyInto(&res))
	_, err = client.SmallMolecule.LibraryScreen.EstimateCost(ctx, params, options...)
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
		Title:          "small-molecule:library-screen estimate-cost",
		Transform:      transform,
	})
}

func handleSmallMoleculeLibraryScreenListResults(ctx context.Context, cmd *cli.Command) error {
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

	params := boltzapi.SmallMoleculeLibraryScreenListResultsParams{}

	format := cmd.Root().String("format")
	explicitFormat := cmd.Root().IsSet("format")
	transform := cmd.Root().String("transform")
	if format == "raw" {
		var res []byte
		options = append(options, option.WithResponseBodyInto(&res))
		_, err = client.SmallMolecule.LibraryScreen.ListResults(
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
			Title:          "small-molecule:library-screen list-results",
			Transform:      transform,
		})
	} else {
		iter := client.SmallMolecule.LibraryScreen.ListResultsAutoPaging(
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
			Title:          "small-molecule:library-screen list-results",
			Transform:      transform,
		})
	}
}

func handleSmallMoleculeLibraryScreenStart(ctx context.Context, cmd *cli.Command) error {
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

	params := boltzapi.SmallMoleculeLibraryScreenStartParams{}

	var res []byte
	options = append(options, option.WithResponseBodyInto(&res))
	_, err = client.SmallMolecule.LibraryScreen.Start(ctx, params, options...)
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
		Title:          "small-molecule:library-screen start",
		Transform:      transform,
	})
}

func handleSmallMoleculeLibraryScreenStop(ctx context.Context, cmd *cli.Command) error {
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
	_, err = client.SmallMolecule.LibraryScreen.Stop(ctx, cmd.Value("id").(string), options...)
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
		Title:          "small-molecule:library-screen stop",
		Transform:      transform,
	})
}
