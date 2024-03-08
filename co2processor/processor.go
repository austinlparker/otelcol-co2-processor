package co2processor

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/giantswarm/cloud-carbon/pkg/footprint"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/ptrace"
)

type spanProcessor struct {
	config           Config
	instanceMappings map[string]string
}

func newSpanProcessor(config Config) (*spanProcessor, error) {
	sp := &spanProcessor{
		config: config,
	}
	return sp, nil
}

func (sp *spanProcessor) processTraces(ctx context.Context, td pdata.Traces) (pdata.Traces, error) {
	// Process each span in the trace data.
	rs := td.ResourceSpans()
	for i := 0; i < rs.Len(); i++ {
		rs := rs.At(i)
		// Extract the AWS instance ID and region from the resources.
		instanceID, region := sp.extractAWSInfo(rs)
		// Process each instrumentation library in the resource spans.
		ils := rs.InstrumentationLibrarySpans()
		for j := 0; j < ils.Len(); j++ {
			ils := ils.At(j)
			// Process each span in the instrumentation library.
			spans := ils.Spans()
			for k := 0; k < spans.Len(); k++ {
				span := spans.At(k)
				// Pass the AWS instance ID and region to the processSpan function.
				sp.processSpan(span, instanceID, region)
			}
		}
	}
	return td, nil
}

func (sp *spanProcessor) processSpan(span pdata.Span, instanceID string, region string) {
	startTime := span.StartTime()
	endTime := span.EndTime()
	duration := endTime.AsTime().Sub(startTime.AsTime())
	footprint := AWS(region, instanceID, duration)
	// Add the footprint as an attribute to the span.
	span.Attributes().InsertString("co2.footprint", fmt.Sprintf("%.2f", footprint))
}

func (sp *spanProcessor) extractAWSInfo(rs pdata.ResourceSpans) (string, string) {
	instanceID := ""
	region := ""
	// Iterate over the attributes of the resource.
	attrs := rs.Resource().Attributes()
	attrs.ForEach(func(k string, v pdata.AttributeValue) {
		// Check if the attribute key matches the instance ID key.
		if k == "aws.ec2.instance.id" {
			instanceID = v.StringVal()
		}
		// Check if the attribute key matches the region key.
		if k == "aws.region" {
			region = v.StringVal()
		}
	})

	return instanceID, region
}
