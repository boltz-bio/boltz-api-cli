// Custom CLI extension code. Not generated.
package cmd

import (
	"context"
	"fmt"

	boltzapi "github.com/boltz-bio/boltz-api-go"
	"github.com/boltz-bio/boltz-api-go/option"
	"github.com/tidwall/gjson"
	"github.com/urfave/cli/v3"
)

func addProteinDesignCuratedSpecificationsCommand(app *cli.Command) {
	parent := findCommandByPath(app, "protein:design")
	if parent == nil || hasCommand(parent.Commands, "list-curated-specifications") {
		return
	}

	parent.Commands = insertCommandAfter(
		parent.Commands,
		"list",
		newProteinDesignListCuratedSpecificationsCommand(),
	)
}

func newProteinDesignListCuratedSpecificationsCommand() *cli.Command {
	return &cli.Command{
		Name:    "list-curated-specifications",
		Usage:   "List binder-side protein design specifications from Boltz-managed curated nanobody or antibody libraries",
		Suggest: true,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "type",
				Usage:    "Curated binder library to retrieve (one of: nanobody, antibody)",
				Required: true,
				Validator: func(value string) error {
					if value != "nanobody" && value != "antibody" {
						return fmt.Errorf("type must be one of: nanobody, antibody")
					}
					return nil
				},
			},
		},
		Action:          handleProteinDesignListCuratedSpecifications,
		HideHelpCommand: true,
	}
}

func handleProteinDesignListCuratedSpecifications(ctx context.Context, cmd *cli.Command) error {
	if unusedArgs := cmd.Args().Slice(); len(unusedArgs) > 0 {
		return fmt.Errorf("Unexpected extra arguments: %v", unusedArgs)
	}

	client := boltzapi.NewClient(getDefaultRequestOptions(cmd)...)
	params := boltzapi.ProteinDesignListCuratedSpecificationsParams{
		Type: boltzapi.ProteinDesignListCuratedSpecificationsParamsType(cmd.String("type")),
	}

	var res []byte
	_, err := client.Protein.Design.ListCuratedSpecifications(
		ctx,
		params,
		option.WithResponseBodyInto(&res),
	)
	if err != nil {
		return err
	}

	return ShowJSON(gjson.ParseBytes(res), ShowJSONOpts{
		ExplicitFormat: cmd.Root().IsSet("format"),
		Format:         cmd.Root().String("format"),
		RawOutput:      cmd.Root().Bool("raw-output"),
		Title:          "protein:design list-curated-specifications",
		Transform:      cmd.Root().String("transform"),
	})
}
