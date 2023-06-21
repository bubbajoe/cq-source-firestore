package client

import (
	"context"
	"fmt"

	"cloud.google.com/go/firestore"
	"github.com/cloudquery/plugin-pb-go/specs"
	"github.com/cloudquery/plugin-sdk/v3/plugins/source"
	"github.com/cloudquery/plugin-sdk/v3/schema"
	"github.com/rs/zerolog"
	"google.golang.org/api/option"
)

type Client struct {
	logger                  zerolog.Logger
	metrics                 *source.Metrics
	Tables                  schema.Tables
	client                  *firestore.Client
	nestedCollections       bool
	nestedCollectionsTables bool
	maxBatchSize            int
	orderBy                 string
	orderDirection          string
}

var _ schema.ClientMeta = (*Client)(nil)

func (*Client) ID() string {
	return "source-firestore"
}

func Configure(ctx context.Context, logger zerolog.Logger, spec specs.Source, _ source.Options) (schema.ClientMeta, error) {
	var firestoreSpec Spec
	err := spec.UnmarshalSpec(&firestoreSpec)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal spec: %w", err)
	}

	if err := firestoreSpec.Validate(); err != nil {
		return nil, fmt.Errorf("failed to validate spec: %w", err)
	}

	firestoreSpec.SetDefaults()

	var client *firestore.Client
	if firestoreSpec.ServiceAccountJSON == "" {
		// Use Application Default Credentials
		client, err = firestore.NewClient(ctx, firestoreSpec.ProjectID)
		if err != nil {
			return nil, fmt.Errorf("failed to create firestore client: %w", err)
		}
	} else {
		creds := option.WithCredentialsJSON([]byte(firestoreSpec.ServiceAccountJSON))
		client, err = firestore.NewClient(ctx, firestoreSpec.ProjectID, creds)
		if err != nil {
			return nil, fmt.Errorf("failed to create firestore client: %w", err)
		}
	}
	zctx := logger.With().Str("module", "firestore-source")
	c := &Client{
		logger:                  zctx.Logger(),
		client:                  client,
		nestedCollections:       firestoreSpec.NestedCollections,
		nestedCollectionsTables: firestoreSpec.NestedCollectionsTables,
		maxBatchSize:            firestoreSpec.MaxBatchSize,
		orderBy:                 firestoreSpec.OrderBy,
		orderDirection:          firestoreSpec.OrderDirection,
	}

	c.Tables, err = c.listTables(ctx, client)
	if err != nil {
		return nil, fmt.Errorf("failed to list tables: %w", err)
	}
	if len(c.Tables) == 0 {
		return nil, fmt.Errorf("no tables found")
	} else {
		logger.Info().Msgf("Found tables: %s", c.Tables.TableNames())
	}
	c.Tables, err = c.Tables.FilterDfs(spec.Tables, spec.SkipTables, spec.SkipDependentTables)
	if err != nil {
		return nil, fmt.Errorf("failed to apply config to tables: %w", err)
	}

	return c, nil
}
