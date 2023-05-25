package videotranscoder

import (
	pb "cloud.google.com/go/video/transcoder/apiv1/transcoderpb"
	"github.com/apache/arrow/go/v13/arrow"
	"github.com/cloudquery/plugin-sdk/v3/schema"
	"github.com/cloudquery/plugin-sdk/v3/transformers"
	"github.com/cloudquery/plugins/source/gcp/client"
)

func JobTemplates() *schema.Table {
	return &schema.Table{
		Name:        "gcp_videotranscoder_job_templates",
		Description: `https://cloud.google.com/transcoder/docs/reference/rest/v1/projects.locations.jobTemplates`,
		Resolver:    fetchJobTemplates,
		Multiplex:   client.ProjectMultiplexEnabledServices("transcoder.googleapis.com"),
		Transform:   client.TransformWithStruct(&pb.JobTemplate{}, transformers.WithPrimaryKeys("Name")),
		Columns: []schema.Column{
			{
				Name:       "project_id",
				Type:       arrow.BinaryTypes.String,
				Resolver:   client.ResolveProject,
				PrimaryKey: true,
			},
		},
	}
}
